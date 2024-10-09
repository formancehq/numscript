package format

import (
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/pretty"
	"github.com/formancehq/numscript/internal/utils"
)

func Format(program parser.Program) string {
	doc := programToDoc(program)
	return pretty.PrintDefault(doc)
}

func literalToDoc(lit parser.Literal) pretty.Document {
	switch lit := lit.(type) {
	case *parser.VariableLiteral:
		return pretty.Text("$" + lit.Name)

	default:
		return utils.NonExhaustiveMatchPanic[pretty.Document](lit)
	}

}

func sentValueToDoc(sentValue parser.SentValue) pretty.Document {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueAll:
		panic("TODO send*")

	case *parser.SentValueLiteral:
		return literalToDoc(sentValue.Monetary)

	default:
		return utils.NonExhaustiveMatchPanic[pretty.Document](sentValue)
	}
}

func sourceToDoc(parser.Source) pretty.Document {
	return pretty.Text("@src")
}

func destinationToDoc(parser.Source) pretty.Document {
	return pretty.Empty()
}

func programToDoc(program parser.Program) pretty.Document {
	var statementsDocs []pretty.Document
	for _, statement := range program.Statements {
		statementsDocs = append(statementsDocs, statementToDoc(statement))
	}
	// TODO newlines after statements
	return pretty.Concat(statementsDocs...)
}

func statementToDoc(statement parser.Statement) pretty.Document {
	switch statement := statement.(type) {
	case *parser.SendStatement:
		return pretty.Concat(
			pretty.Text("send "),
			sentValueToDoc(statement.SentValue),
			pretty.Text(" ("),
			pretty.Lines(0),
			pretty.Nest(
				pretty.Concat(
					pretty.SpaceBreak(),
					pretty.Text("source = "),
					sourceToDoc(statement.Source),
				),
			),
		)

	case *parser.FnCall:
		panic("fnCall")

	default:
		return utils.NonExhaustiveMatchPanic[pretty.Document](statement)
	}
}
