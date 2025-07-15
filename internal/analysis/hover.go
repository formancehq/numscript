package analysis

import (
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

type Hover interface{ hover() }

type VariableHover struct {
	parser.Range
	Node *parser.Variable
}

func (*VariableHover) hover() {}

type BuiltinFnHoverContext = uint

const (
	OriginContext BuiltinFnHoverContext = iota
	StatementContext
)

type BuiltinFnHover struct {
	Range   parser.Range
	Node    *parser.FnCall
	Context BuiltinFnHoverContext
}

func (*BuiltinFnHover) hover() {}

func HoverOn(program parser.Program, position parser.Position) Hover {
	if program.Vars != nil {
		for _, varDecl := range program.Vars.Declarations {
			hover := hoverOnVar(varDecl, position)
			if hover != nil {
				return hover
			}
		}
	}

	for _, statement := range program.Statements {
		// TODO binary search into statements
		if statement == nil || !statement.GetRange().Contains(position) {
			continue
		}

		switch statement := statement.(type) {
		case *parser.SendStatement:
			hover := hoverOnSendStatement(*statement, position)
			if hover != nil {
				return hover
			}

		case *parser.SaveStatement:
			hover := hoverOnSaveStatement(*statement, position)
			if hover != nil {
				return hover
			}

		case *parser.FnCall:
			hover := hoverOnFnCall(*statement, position)
			if hover != nil {
				return hover
			}
		}

	}
	return nil
}

func hoverOnVar(varDecl parser.VarDeclaration, position parser.Position) Hover {
	if !varDecl.Contains(position) {
		return nil
	}

	if varDecl.Origin != nil {
		hover := hoverOnExpression(*varDecl.Origin, position)
		if hover != nil {
			return hover
		}
	}

	return nil
}

func hoverOnSentValue(sentValue parser.SentValue, position parser.Position) Hover {
	switch sentValue := sentValue.(type) {
	case nil:
		return nil

	case *parser.SentValueAll:
		return hoverOnExpression(sentValue.Asset, position)

	case *parser.SentValueLiteral:
		return hoverOnExpression(sentValue.Monetary, position)

	default:
		return utils.NonExhaustiveMatchPanic[Hover](sentValue)
	}
}

func hoverOnSaveStatement(saveStatement parser.SaveStatement, position parser.Position) Hover {
	if !saveStatement.Contains(position) {
		return nil
	}

	hover := hoverOnSentValue(saveStatement.SentValue, position)
	if hover != nil {
		return hover
	}

	hover = hoverOnExpression(saveStatement.Amount, position)
	if hover != nil {
		return hover
	}

	return nil
}

func hoverOnSendStatement(sendStatement parser.SendStatement, position parser.Position) Hover {
	if !sendStatement.Contains(position) {
		return nil
	}

	hover := hoverOnSentValue(sendStatement.SentValue, position)
	if hover != nil {
		return hover
	}

	hover = hoverOnSource(sendStatement.Source, position)
	if hover != nil {
		return hover
	}

	hover = hoverOnDestination(sendStatement.Destination, position)
	if hover != nil {
		return hover
	}

	return nil
}

func hoverOnExpression(lit parser.ValueExpr, position parser.Position) Hover {
	if lit == nil || !lit.GetRange().Contains(position) {
		return nil
	}

	switch lit := lit.(type) {
	case *parser.Variable:
		return &VariableHover{
			Range: lit.Range,
			Node:  lit,
		}
	case *parser.AccountInterpLiteral:
		for _, part := range lit.Parts {
			if v, ok := part.(*parser.Variable); ok {

				hover := hoverOnExpression(v, position)
				if hover != nil {
					return hover
				}
			}
		}
	case *parser.MonetaryLiteral:
		hover := hoverOnExpression(lit.Amount, position)
		if hover != nil {
			return hover
		}

		hover = hoverOnExpression(lit.Asset, position)
		if hover != nil {
			return hover
		}
	case *parser.BinaryInfix:
		hover := hoverOnExpression(lit.Left, position)
		if hover != nil {
			return hover
		}

		hover = hoverOnExpression(lit.Right, position)
		if hover != nil {
			return hover
		}

	case *parser.FnCall:
		return hoverOnFnCall(*lit, position)
	}

	return nil
}

func hoverOnSource(source parser.Source, position parser.Position) Hover {
	if source == nil || !source.GetRange().Contains(position) {
		return nil
	}

	switch source := source.(type) {
	case *parser.SourceCapped:
		hover := hoverOnExpression(source.Cap, position)
		if hover != nil {
			return hover
		}
		hover = hoverOnSource(source.From, position)
		if hover != nil {
			return hover
		}
		return nil

	case *parser.SourceOverdraft:
		hover := hoverOnExpression(source.Address, position)
		if hover != nil {
			return hover
		}

		if source.Bounded != nil {
			hover := hoverOnExpression(*source.Bounded, position)
			if hover != nil {
				return hover
			}
		}

		return nil

	case *parser.SourceInorder:
		for _, source := range source.Sources {
			// TODO binary search
			if source == nil || !source.GetRange().Contains(position) {
				continue
			}

			hover := hoverOnSource(source, position)
			if hover != nil {
				return hover
			}
		}

	case *parser.SourceAllotment:
		for _, item := range source.Items {
			// TODO binary search
			if !item.Contains(position) {
				continue
			}

			switch allot := item.Allotment.(type) {
			case *parser.RemainingAllotment: // Do nothing here (no nested expr)
			case *parser.ValueExprAllotment:
				hover := hoverOnExpression(allot.Value, position)
				if hover != nil {
					return hover
				}
			}

			hover := hoverOnSource(item.From, position)
			if hover != nil {
				return hover
			}
		}

	case *parser.SourceAccount:
		return hoverOnExpression(source.ValueExpr, position)
	}

	return nil

}

func hoverOnKeptOrDestination(inorderClause parser.KeptOrDestination, position parser.Position) Hover {
	switch inorderClause := inorderClause.(type) {
	case nil, *parser.DestinationKept:
		return nil

	case *parser.DestinationTo:
		return hoverOnDestination(inorderClause.Destination, position)

	default:
		return utils.NonExhaustiveMatchPanic[Hover](inorderClause)
	}

}

func hoverOnDestination(destination parser.Destination, position parser.Position) Hover {
	if destination == nil || !destination.GetRange().Contains(position) {
		return nil
	}

	switch source := destination.(type) {
	case *parser.DestinationInorder:
		for _, inorderClause := range source.Clauses {
			// TODO binary search
			if !inorderClause.Contains(position) {
				continue
			}

			hover := hoverOnExpression(inorderClause.Cap, position)
			if hover != nil {
				return hover
			}

			hover = hoverOnKeptOrDestination(inorderClause.To, position)
			if hover != nil {
				return hover
			}
		}

		hover := hoverOnKeptOrDestination(source.Remaining, position)
		if hover != nil {
			return hover
		}

	case *parser.DestinationAllotment:
		for _, item := range source.Items {
			// TODO binary search
			if !item.Contains(position) {
				continue
			}

			switch allot := item.Allotment.(type) {
			case *parser.ValueExprAllotment:
				hover := hoverOnExpression(allot.Value, position)
				if hover != nil {
					return hover
				}
			}

			hover := hoverOnKeptOrDestination(item.To, position)
			if hover != nil {
				return hover
			}
		}

	case *parser.DestinationAccount:
		return hoverOnExpression(source.ValueExpr, position)
	}

	return nil

}

func hoverOnFnCall(callStatement parser.FnCall, position parser.Position) Hover {
	if !callStatement.Contains(position) {
		return nil
	}

	if callStatement.Caller.Contains(position) {
		return &BuiltinFnHover{
			Range:   callStatement.Caller.Range,
			Node:    &callStatement,
			Context: StatementContext,
		}
	}

	for _, arg := range callStatement.Args {
		hover := hoverOnExpression(arg, position)
		if hover != nil {
			return hover
		}
	}

	return nil

}
