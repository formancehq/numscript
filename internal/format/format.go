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

func nest(srcs ...string) string {
	src := strings.Join(srcs, "")

	lines := strings.Split(src, "\n")
	ret := ""
	for _, line := range lines {
		ret += "  " + line + "\n"
	}
	return ret
}

func fmtLit(lit parser.Literal) string {
	switch lit := lit.(type) {
	case *parser.VariableLiteral:
		return "$" + lit.Name

	case *parser.AccountLiteral:
		return "@" + lit.Name

	case *parser.MonetaryLiteral:
		return fmt.Sprintf("[%s %s]", fmtLit(lit.Asset), fmtLit(lit.Amount))

	case *parser.AssetLiteral:
		return lit.Asset

	case *parser.RatioLiteral:
		panic("TODO ratio lit")

	case *parser.NumberLiteral:
		return fmt.Sprint(lit.Number)

	default:
		return utils.NonExhaustiveMatchPanic[string](lit)
	}

}

func fmtSentValue(sentValue parser.SentValue) string {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueAll:
		panic("TODO send*")

	case *parser.SentValueLiteral:
		return fmtLit(sentValue.Monetary)

	default:
		return utils.NonExhaustiveMatchPanic[string](sentValue)
	}
}

func fmtSrc(src parser.Source) string {
	switch src := src.(type) {
	case *parser.SourceAccount:
		return fmtLit(src.Literal)

	case *parser.SourceInorder:
		var lines []string
		for _, subSrc := range src.Sources {
			lines = append(lines, fmtSrc(subSrc))
		}
		return fmt.Sprintf(
			"{\n%s}",
			nest(strings.Join(lines, "\n")),
		)

	case *parser.SourceCapped:
		return fmt.Sprintf("max %s from %s", fmtLit(src.Cap), fmtSrc(src.From))

	case *parser.SourceAllotment:
		panic("TODO src allot")

	case *parser.SourceOverdraft:
		panic("TODO src overdraft")

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
		return fmtLit(dest.Literal)

	case *parser.DestinationInorder:
		var lines []string
		for _, subDest := range dest.Clauses {
			s := fmt.Sprintf("max %s %s", fmtLit(subDest.Cap), fmtKeptorDest(subDest.To))
			lines = append(lines, s)
		}

		s := fmt.Sprintf("remaining %s", fmtKeptorDest(dest.Remaining))
		lines = append(lines, s)

		return fmt.Sprintf(
			"{\n%s}",
			nest(strings.Join(lines, "\n")),
		)

	case *parser.DestinationAllotment:
		panic("TODO Dest allot")

	default:
		return utils.NonExhaustiveMatchPanic[string](dest)
	}
}

func fmtProgram(program parser.Program) string {
	var statementsDocs []string
	for _, statement := range program.Statements {
		statementsDocs = append(statementsDocs, fmtStatement(statement))
	}
	return strings.Join(statementsDocs, "\n")
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
		panic("fnCall")

	default:
		return utils.NonExhaustiveMatchPanic[string](statement)
	}
}
