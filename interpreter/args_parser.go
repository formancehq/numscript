package interpreter

type argsParser struct {
	parsedArgsCount int
	args            []Value
	err             error
}

func NewArgsParser(args []Value) *argsParser {
	return &argsParser{
		args: args,
	}
}

func parseArg[T any](p *argsParser, expect func(Value) (*T, error)) *T {
	index := p.parsedArgsCount
	p.parsedArgsCount++

	if p.err != nil || index >= len(p.args) {
		return nil
	}

	arg := p.args[index]
	parsed, err := expect(arg)
	if err != nil {
		p.err = err
		return nil
	}
	return parsed
}

func (p *argsParser) parse() error {
	if len(p.args) != p.parsedArgsCount {
		p.err = BadArityErr{
			ExpectedArity:  p.parsedArgsCount,
			GivenArguments: len(p.args),
		}
	}

	return p.err
}
