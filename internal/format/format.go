package format

import (
	"fmt"
	"math/big"
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
		if lit.Denominator.Cmp(big.NewInt(100)) == 0 {
			return lit.Numerator.String() + "%"
		}
		return fmt.Sprintf("%s/%s", lit.Numerator.String(), lit.Denominator.String())

	case *parser.NumberLiteral:
		return fmt.Sprint(lit.Number)

	case *parser.StringLiteral:
		return fmt.Sprintf(`"%s"`, lit.String)

	default:
		return utils.NonExhaustiveMatchPanic[string](lit)
	}

}

func fmtSentValue(sentValue parser.SentValue) string {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueAll:
		return fmt.Sprintf("[%s *]", fmtLit(sentValue.Asset))

	case *parser.SentValueLiteral:
		return fmtLit(sentValue.Monetary)

	default:
		return utils.NonExhaustiveMatchPanic[string](sentValue)
	}
}

func fmtAllotmentValue(allot parser.AllotmentValue) string {
	switch allot := allot.(type) {
	case *parser.RatioLiteral:
		return fmtLit(allot)

	case *parser.VariableLiteral:
		return fmtLit(allot)

	case *parser.RemainingAllotment:
		return "remaining"

	default:
		return utils.NonExhaustiveMatchPanic[string](allot)
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
		return block(
			strings.Join(lines, "\n"),
		)

	case *parser.SourceCapped:
		return fmt.Sprintf("max %s from %s", fmtLit(src.Cap), fmtSrc(src.From))

	case *parser.SourceAllotment:
		var lines []string
		for _, item := range src.Items {
			s := fmt.Sprintf("%s from %s", fmtAllotmentValue(item.Allotment), fmtSrc(item.From))
			lines = append(lines, s)
		}
		return block(strings.Join(lines, "\n"))

	case *parser.SourceOverdraft:
		if src.Bounded == nil {
			return fmt.Sprintf("%s allowing unbounded overdraft", fmtLit(src.Address))
		}

		return fmt.Sprintf("%s allowing overdraft up to %s", fmtLit(src.Address), fmtLit(*src.Bounded))

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

func fmtProgram(program parser.Program) string {
	var statementsDocs []string
	for _, statement := range program.Statements {
		statementsDocs = append(statementsDocs, fmtStatement(statement))
	}
	return strings.Join(statementsDocs, "\n\n")
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
		var args []string
		for _, arg := range statement.Args {
			args = append(args, fmtLit(arg))
		}
		return fmt.Sprintf("%s(%s)", statement.Caller.Name, strings.Join(args, ", "))

	default:
		return utils.NonExhaustiveMatchPanic[string](statement)
	}
}
