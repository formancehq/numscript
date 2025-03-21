package analysis

import (
	"github.com/formancehq/numscript/internal/parser"
)

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

func (res *CheckResult) requireVersion(
	rng parser.Range,
	requiredVersion parser.Version,
) {
	actualVersion := res.Program.GetVersion()
	if actualVersion == nil {
		return
	}

	switch requiredVersion := requiredVersion.(type) {
	case parser.VersionMachine:
		_, ok := actualVersion.(parser.VersionMachine)
		if !ok {

			res.pushDiagnostic(rng.GetRange(), VersionMismatch{
				GotVersion:      actualVersion,
				RequiredVersion: requiredVersion,
			})
		}

	case parser.VersionInterpreter:
		interpreterActualVersion, ok := actualVersion.(parser.VersionInterpreter)

		if !ok || !interpreterActualVersion.GtEq(requiredVersion) {
			res.pushDiagnostic(rng, VersionMismatch{
				GotVersion:      actualVersion,
				RequiredVersion: requiredVersion,
			})
		}

	}
}
