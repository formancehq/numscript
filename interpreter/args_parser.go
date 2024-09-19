package interpreter

import "github.com/formancehq/numscript/parser"

type argsParser struct {
	parsedArgsCount int
	args            []Value
	err             InterpreterError
}

func NewArgsParser(args []Value) *argsParser {
	return &argsParser{
		args: args,
	}
}

func parseArg[T any](p *argsParser, r parser.Range, expect func(Value, parser.Range) (*T, InterpreterError)) *T {
	index := p.parsedArgsCount
	p.parsedArgsCount++

	if p.err != nil || index >= len(p.args) {
		return nil
	}

	arg := p.args[index]
	parsed, err := expect(arg, r)
	if err != nil {
		p.err = err
		return nil
	}
	return parsed
}

func (p *argsParser) parse() InterpreterError {
	if len(p.args) != p.parsedArgsCount {
		p.err = BadArityErr{
			ExpectedArity:  p.parsedArgsCount,
			GivenArguments: len(p.args),
		}
	}

	return p.err
}
