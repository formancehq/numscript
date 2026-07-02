package interpreter

import (
	"errors"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

// errStopWalk is a sentinel that a hook can return to abort the walk without it
// being treated as a failure by the caller.
var errStopWalk = errors.New("stop walk")

type astVisitor struct {
	// Any hook may be nil (then it is skipped). Returning a non-nil error
	// aborts the whole walk and that error is propagated to walkProgram's
	// caller. Hooks fire pre-order (parent before children).
	OnVariable    func(*parser.Variable) error
	OnAsset       func(*parser.AssetLiteral) error
	OnNumber      func(*parser.NumberLiteral) error
	OnString      func(*parser.StringLiteral) error
	OnPercentage  func(*parser.PercentageLiteral) error
	OnMonetary    func(*parser.MonetaryLiteral) error
	OnAccount     func(*parser.AccountInterpLiteral) error
	OnBinaryInfix func(*parser.BinaryInfix) error
	OnPrefix      func(*parser.Prefix) error
	OnFnCall      func(*parser.FnCall) error

	// OnMeta fires for meta() calls in place of OnFnCall. originType is the
	// declared var type when the call is directly a var origin (the only place
	// meta() is valid), and nil when meta() appears anywhere else — i.e. nested
	// in a sub-expression or in a statement.
	OnMeta func(originType *string, fnCall *parser.FnCall) error

	OnSource        func(parser.Source) error
	OnDestination   func(parser.Destination) error
	OnStatement     func(parser.Statement) error
	OnSaveStatement func(*parser.SaveStatement) error
	OnVarDecl       func(parser.VarDeclaration) error
}

func (v astVisitor) walkProgram(program parser.Program) error {
	if program.Vars != nil {
		for _, decl := range program.Vars.Declarations {
			if err := v.walkVarDecl(decl); err != nil {
				return err
			}
		}
	}
	for _, stmt := range program.Statements {
		if err := v.walkStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (v astVisitor) walkVarDecl(decl parser.VarDeclaration) error {
	if v.OnVarDecl != nil {
		if err := v.OnVarDecl(decl); err != nil {
			return err
		}
	}
	if decl.Origin == nil {
		return nil
	}
	origin := *decl.Origin

	// A meta() that is directly the var origin is the only valid meta position:
	// route it to OnMeta with the declared type and walk its args, bypassing
	// the generic (nil-type) meta dispatch in walkFnCall.
	if fnCall, ok := origin.(*parser.FnCall); ok && fnCall.Caller.Name == analysis.FnVarOriginMeta {
		if v.OnMeta != nil {
			if err := v.OnMeta(&decl.Type.Name, fnCall); err != nil {
				return err
			}
		}
		return v.walkFnCallArgs(fnCall)
	}

	return v.walkValueExpr(origin)
}

func (v astVisitor) walkStatement(statement parser.Statement) error {
	if v.OnStatement != nil {
		if err := v.OnStatement(statement); err != nil {
			return err
		}
	}
	switch statement := statement.(type) {
	case *parser.FnCall:
		return v.walkFnCall(statement)

	case *parser.SendStatement:
		if err := v.walkSentValue(statement.SentValue); err != nil {
			return err
		}
		if err := v.walkSource(statement.Source); err != nil {
			return err
		}
		return v.walkDestination(statement.Destination)

	case *parser.SaveStatement:
		if v.OnSaveStatement != nil {
			if err := v.OnSaveStatement(statement); err != nil {
				return err
			}
		}
		if err := v.walkSentValue(statement.SentValue); err != nil {
			return err
		}
		return v.walkValueExpr(statement.Account)

	default:
		utils.NonExhaustiveMatchPanic[any](statement)
		return nil
	}
}

func (v astVisitor) walkSentValue(sentValue parser.SentValue) error {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueLiteral:
		return v.walkValueExpr(sentValue.Monetary)
	case *parser.SentValueAll:
		return v.walkValueExpr(sentValue.Asset)
	default:
		utils.NonExhaustiveMatchPanic[any](sentValue)
		return nil
	}
}

func (v astVisitor) walkSource(source parser.Source) error {
	if v.OnSource != nil {
		if err := v.OnSource(source); err != nil {
			return err
		}
	}
	switch source := source.(type) {
	case *parser.SourceAccount:
		if source.Color != nil {
			if err := v.walkValueExpr(source.Color); err != nil {
				return err
			}
		}
		return v.walkValueExpr(source.ValueExpr)

	case *parser.SourceInorder:
		for _, sub := range source.Sources {
			if err := v.walkSource(sub); err != nil {
				return err
			}
		}
		return nil

	case *parser.SourceOneof:
		for _, sub := range source.Sources {
			if err := v.walkSource(sub); err != nil {
				return err
			}
		}
		return nil

	case *parser.SourceAllotment:
		for _, item := range source.Items {
			if err := v.walkAllotmentValue(item.Allotment); err != nil {
				return err
			}
			if err := v.walkSource(item.From); err != nil {
				return err
			}
		}
		return nil

	case *parser.SourceCapped:
		if err := v.walkValueExpr(source.Cap); err != nil {
			return err
		}
		return v.walkSource(source.From)

	case *parser.SourceOverdraft:
		if source.Color != nil {
			if err := v.walkValueExpr(source.Color); err != nil {
				return err
			}
		}
		if err := v.walkValueExpr(source.Address); err != nil {
			return err
		}
		if source.Bounded != nil {
			return v.walkValueExpr(*source.Bounded)
		}
		return nil

	case *parser.SourceWithScaling:
		if err := v.walkValueExpr(source.Address); err != nil {
			return err
		}
		return v.walkValueExpr(source.Through)

	default:
		utils.NonExhaustiveMatchPanic[any](source)
		return nil
	}
}

func (v astVisitor) walkDestination(destination parser.Destination) error {
	if v.OnDestination != nil {
		if err := v.OnDestination(destination); err != nil {
			return err
		}
	}
	switch destination := destination.(type) {
	case *parser.DestinationAccount:
		return v.walkValueExpr(destination.ValueExpr)

	case *parser.DestinationInorder:
		for _, clause := range destination.Clauses {
			if err := v.walkCappedKeptOrDestination(clause); err != nil {
				return err
			}
		}
		return v.walkKeptOrDestination(destination.Remaining)

	case *parser.DestinationOneof:
		for _, clause := range destination.Clauses {
			if err := v.walkCappedKeptOrDestination(clause); err != nil {
				return err
			}
		}
		return v.walkKeptOrDestination(destination.Remaining)

	case *parser.DestinationAllotment:
		for _, item := range destination.Items {
			if err := v.walkAllotmentValue(item.Allotment); err != nil {
				return err
			}
			if err := v.walkKeptOrDestination(item.To); err != nil {
				return err
			}
		}
		return nil

	default:
		utils.NonExhaustiveMatchPanic[any](destination)
		return nil
	}
}

func (v astVisitor) walkCappedKeptOrDestination(clause parser.CappedKeptOrDestination) error {
	if err := v.walkValueExpr(clause.Cap); err != nil {
		return err
	}
	return v.walkKeptOrDestination(clause.To)
}

func (v astVisitor) walkKeptOrDestination(keptOrDestination parser.KeptOrDestination) error {
	switch keptOrDestination := keptOrDestination.(type) {
	case *parser.DestinationKept:
		return nil
	case *parser.DestinationTo:
		return v.walkDestination(keptOrDestination.Destination)
	default:
		utils.NonExhaustiveMatchPanic[any](keptOrDestination)
		return nil
	}
}

func (v astVisitor) walkAllotmentValue(allotmentValue parser.AllotmentValue) error {
	switch allotmentValue := allotmentValue.(type) {
	case *parser.ValueExprAllotment:
		return v.walkValueExpr(allotmentValue.Value)
	case *parser.RemainingAllotment:
		return nil
	default:
		utils.NonExhaustiveMatchPanic[any](allotmentValue)
		return nil
	}
}

func (v astVisitor) walkValueExpr(valueExpr parser.ValueExpr) error {
	switch valueExpr := valueExpr.(type) {
	case *parser.Variable:
		if v.OnVariable != nil {
			return v.OnVariable(valueExpr)
		}
		return nil

	case *parser.AssetLiteral:
		if v.OnAsset != nil {
			return v.OnAsset(valueExpr)
		}
		return nil

	case *parser.NumberLiteral:
		if v.OnNumber != nil {
			return v.OnNumber(valueExpr)
		}
		return nil

	case *parser.StringLiteral:
		if v.OnString != nil {
			return v.OnString(valueExpr)
		}
		return nil

	case *parser.PercentageLiteral:
		if v.OnPercentage != nil {
			return v.OnPercentage(valueExpr)
		}
		return nil

	case *parser.MonetaryLiteral:
		if v.OnMonetary != nil {
			if err := v.OnMonetary(valueExpr); err != nil {
				return err
			}
		}
		if err := v.walkValueExpr(valueExpr.Asset); err != nil {
			return err
		}
		return v.walkValueExpr(valueExpr.Amount)

	case *parser.AccountInterpLiteral:
		if v.OnAccount != nil {
			if err := v.OnAccount(valueExpr); err != nil {
				return err
			}
		}
		// account name parts can only be text or *Variable; the variables are
		// worth descending into so OnVariable fires for interpolations.
		for _, part := range valueExpr.Parts {
			if variable, ok := part.(*parser.Variable); ok && v.OnVariable != nil {
				if err := v.OnVariable(variable); err != nil {
					return err
				}
			}
		}
		return nil

	case *parser.BinaryInfix:
		if v.OnBinaryInfix != nil {
			if err := v.OnBinaryInfix(valueExpr); err != nil {
				return err
			}
		}
		if err := v.walkValueExpr(valueExpr.Left); err != nil {
			return err
		}
		return v.walkValueExpr(valueExpr.Right)

	case *parser.Prefix:
		if v.OnPrefix != nil {
			if err := v.OnPrefix(valueExpr); err != nil {
				return err
			}
		}
		return v.walkValueExpr(valueExpr.Expr)

	case *parser.FnCall:
		return v.walkFnCall(valueExpr)

	default:
		utils.NonExhaustiveMatchPanic[any](valueExpr)
		return nil
	}
}

func (v astVisitor) walkFnCall(fnCall *parser.FnCall) error {
	if fnCall.Caller.Name == analysis.FnVarOriginMeta {
		// Reached through the generic walk, so this meta() is never a top-level
		// var origin (those are intercepted in walkVarDecl): its type is nil.
		if v.OnMeta != nil {
			if err := v.OnMeta(nil, fnCall); err != nil {
				return err
			}
		}
	} else if v.OnFnCall != nil {
		if err := v.OnFnCall(fnCall); err != nil {
			return err
		}
	}
	return v.walkFnCallArgs(fnCall)
}

func (v astVisitor) walkFnCallArgs(fnCall *parser.FnCall) error {
	for _, arg := range fnCall.Args {
		if err := v.walkValueExpr(arg); err != nil {
			return err
		}
	}
	return nil
}
