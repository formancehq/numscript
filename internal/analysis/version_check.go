package analysis

import "github.com/formancehq/numscript/internal/parser"

func (res *CheckResult) checkInfixVersion(expr parser.BinaryInfix) {
	_, isLeftANumberLit := expr.Left.(*parser.NumberLiteral)
	_, isRightANumberLit := expr.Right.(*parser.NumberLiteral)
	if isLeftANumberLit && isRightANumberLit {
		return
	}

	res.requireVersion(expr.Range,
		parser.NewVersionInterpreter(0, 0, 15),
	)
}
