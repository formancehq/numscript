package format

import (
	"fmt"
	"strings"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

func Format(program parser.Program) string {
	return fmtProgram(program)
}

func block(inner string) string {
	return fmt.Sprintf(
		"{\n%s}",
		nest(inner),
	)
}

func nest(srcs ...string) string {
	src := strings.Join(srcs, "")

	lines := strings.Split(src, "\n")
	ret := ""
	for _, line := range lines {
		ret += "  " + line + "\n"
	}
	return ret
}

func fmtExpr(lit parser.ValueExpr) string {
	switch lit := lit.(type) {
	case *parser.Variable:
		return "$" + lit.Name

	case *parser.AccountInterpLiteral:
		return "@" + lit.String()

	case *parser.MonetaryLiteral:
		return fmt.Sprintf("[%s %s]", fmtExpr(lit.Asset), fmtExpr(lit.Amount))

	case *parser.AssetLiteral:
		return lit.Asset

	case *parser.PercentageLiteral:
		// TODO handle floating digits
		return fmt.Sprintf("%s%s", lit.Amount.String(), "%")

	case *parser.NumberLiteral:
		return fmt.Sprint(lit.Number)

	case *parser.StringLiteral:
		return fmt.Sprintf(`"%s"`, lit.String)

	case *parser.BinaryInfix:
		fmtLeft := fmtExpr(lit.Left)
		fmtRight := fmtExpr(lit.Right)

		leftPrec := operatorPrec[lit.Operator]
		var rightPrec uint8 = 255
		if right, ok := lit.Right.(*parser.BinaryInfix); ok {
			rightPrec = operatorPrec[right.Operator]
		}

		if leftPrec > rightPrec {
			fmtRight = fmt.Sprintf("(%s)", fmtRight)
		}

		// Do not use whitespace when formatting ratios, e.g. 1/2
		if lit.Operator == parser.InfixOperatorDiv {
			return fmt.Sprintf("%s%s%s", fmtLeft, lit.Operator, fmtRight)
		}

		return fmt.Sprintf("%s %s %s", fmtLeft, lit.Operator, fmtRight)

	default:
		return utils.NonExhaustiveMatchPanic[string](lit)
	}

}

// Reference: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Operator_precedence#table
var operatorPrec = map[parser.InfixOperator]uint8{
	"+": 11,
	"-": 11,

	"/": 12,
	"*": 12,
}

func fmtSentValue(sentValue parser.SentValue) string {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueAll:
		return fmt.Sprintf("[%s *]", fmtExpr(sentValue.Asset))

	case *parser.SentValueLiteral:
		return fmtExpr(sentValue.Monetary)

	default:
		return utils.NonExhaustiveMatchPanic[string](sentValue)
	}
}

func fmtAllotmentValue(allot parser.AllotmentValue) string {
	switch allot := allot.(type) {
	case *parser.ValueExprAllotment:
		return fmtExpr(allot.Value)

	case *parser.RemainingAllotment:
		return "remaining"

	default:
		return utils.NonExhaustiveMatchPanic[string](allot)
	}
}

func fmtSrc(src parser.Source) string {
	switch src := src.(type) {
	case *parser.SourceAccount:
		return fmtExpr(src.ValueExpr)

	case *parser.SourceInorder:
		var lines []string
		for _, subSrc := range src.Sources {
			lines = append(lines, fmtSrc(subSrc))
		}
		return block(
			strings.Join(lines, "\n"),
		)

	case *parser.SourceCapped:
		return fmt.Sprintf("max %s from %s", fmtExpr(src.Cap), fmtSrc(src.From))

	case *parser.SourceAllotment:
		var lines []string
		for _, item := range src.Items {
			s := fmt.Sprintf("%s from %s", fmtAllotmentValue(item.Allotment), fmtSrc(item.From))
			lines = append(lines, s)
		}
		return block(strings.Join(lines, "\n"))

	case *parser.SourceOverdraft:
		if src.Bounded == nil {
			return fmt.Sprintf("%s allowing unbounded overdraft", fmtExpr(src.Address))
		}

		return fmt.Sprintf("%s allowing overdraft up to %s", fmtExpr(src.Address), fmtExpr(*src.Bounded))

	default:
		return utils.NonExhaustiveMatchPanic[string](src)
	}
}

func fmtKeptorDest(keptOrDest parser.KeptOrDestination) string {
	switch keptOrDest := keptOrDest.(type) {
	case *parser.DestinationKept:
		return "kept"

	case *parser.DestinationTo:
		return fmt.Sprintf("to %s", fmtDest(keptOrDest.Destination))

	default:
		return utils.NonExhaustiveMatchPanic[string](keptOrDest)
	}
}

func fmtDest(dest parser.Destination) string {
	switch dest := dest.(type) {
	case *parser.DestinationAccount:
		return fmtExpr(dest.ValueExpr)

	case *parser.DestinationInorder:
		var lines []string
		for _, subDest := range dest.Clauses {
			s := fmt.Sprintf("max %s %s", fmtExpr(subDest.Cap), fmtKeptorDest(subDest.To))
			lines = append(lines, s)
		}

		s := fmt.Sprintf("remaining %s", fmtKeptorDest(dest.Remaining))
		lines = append(lines, s)
		return block(
			strings.Join(lines, "\n"),
		)

	case *parser.DestinationAllotment:
		var lines []string
		for _, item := range dest.Items {
			s := fmt.Sprintf("%s %s", fmtAllotmentValue(item.Allotment), fmtKeptorDest(item.To))
			lines = append(lines, s)
		}
		return block(strings.Join(lines, "\n"))

	default:
		return utils.NonExhaustiveMatchPanic[string](dest)
	}
}

func fmtVars(vars []parser.VarDeclaration) string {
	if len(vars) == 0 {
		return ""
	}

	var lines []string
	for _, varDecl := range vars {
		origin := ""
		if varDecl.Origin != nil {
			origin = " = " + fmtFnCall(*varDecl.Origin)
		}

		s := fmt.Sprintf("%s $%s%s", varDecl.Type.Name, varDecl.Name.Name, origin)
		lines = append(lines, s)
	}

	return fmt.Sprintf("vars %s\n", block(strings.Join(lines, "\n"))) + "\n"
}

func fmtStatements(comments []parser.Comment, statements []parser.Statement) string {
	if len(statements) == 0 {
		return ""
	}

	var statementsDocs []string
	for _, statement := range statements {
		if len(comments) != 0 {
			lastComment := comments[len(comments)-1]

			if !lastComment.End.GtEq(statement.GetRange().Start) {
				cmt := strings.TrimRight(lastComment.Content, "\n")
				statementsDocs = append(statementsDocs, cmt)
				comments = comments[0 : len(comments)-1]
			}

		}

		statementsDocs = append(statementsDocs, fmtStatement(statement))
	}
	return strings.Join(statementsDocs, "\n\n") + "\n"
}

func fmtProgram(program parser.Program) string {
	return fmt.Sprint(fmtVars(program.Vars), fmtStatements(program.Comments, program.Statements))
}

func fmtStatement(statement parser.Statement) string {
	switch statement := statement.(type) {
	case *parser.SendStatement:
		return fmt.Sprint(
			"send ",
			fmtSentValue(statement.SentValue),
			" (\n",
			nest(
				"source = ",
				fmtSrc(statement.Source),
				"\n",
				"destination = ",
				fmtDest(statement.Destination),
			),
			")",
		)

	case *parser.FnCall:
		return fmtFnCall(*statement)

	default:
		return utils.NonExhaustiveMatchPanic[string](statement)
	}
}

func fmtFnCall(fnCall parser.FnCall) string {
	var args []string
	for _, arg := range fnCall.Args {
		args = append(args, fmtExpr(arg))
	}
	return fmt.Sprintf("%s(%s)", fnCall.Caller.Name, strings.Join(args, ", "))
}
