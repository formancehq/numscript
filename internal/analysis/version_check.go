package analysis

import (
	"slices"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"
)

// -- functionalities constraint
func (res *CheckResult) checkInfixVersion(expr parser.BinaryInfix) {
	_, isLeftANumberLit := expr.Left.(*parser.NumberLiteral)
	_, isRightANumberLit := expr.Right.(*parser.NumberLiteral)
	if isLeftANumberLit && isRightANumberLit {
		return
	}

	res.requireVersion(expr.Range,
		VersionClause{
			Version: parser.NewVersionInterpreter(0, 0, 15),
		},
	)
}

func (res *CheckResult) checkOneofVersion(rng parser.Range) {
	res.requireVersion(rng,
		VersionClause{
			Version:     parser.NewVersionInterpreter(0, 0, 15),
			FeatureFlag: flags.ExperimentalOneofFeatureFlag,
		},
	)
}

func (res *CheckResult) checkOvedraftFunctionVersion(fnCall parser.FnCall) {
	if fnCall.Caller.Name != FnVarOriginOverdraft {
		return
	}

	res.requireVersion(fnCall.Range,
		VersionClause{
			Version:     parser.NewVersionInterpreter(0, 0, 15),
			FeatureFlag: flags.ExperimentalOverdraftFunctionFeatureFlag,
		},
	)
}

func (res *CheckResult) checkAccountInterpolationVersion(expr parser.AccountInterpLiteral) {
	isInterpolation := slices.ContainsFunc(expr.Parts, func(part parser.AccountNamePart) bool {
		_, isVar := part.(*parser.Variable)
		return isVar
	})

	if !isInterpolation {
		return
	}

	res.requireVersion(expr.Range,
		VersionClause{
			Version:     parser.NewVersionInterpreter(0, 0, 15),
			FeatureFlag: flags.ExperimentalAccountInterpolationFlag,
		},
	)
}

// -- version check utilities

type VersionClause struct {
	Version     parser.Version
	FeatureFlag flags.FeatureFlag
}

func (res *CheckResult) requireVersion(
	rng parser.Range,
	clauses ...VersionClause,
) {
	flags := res.Program.GetFlags()
	actualVersion := res.Program.GetVersion()
	if actualVersion == nil {
		return
	}

	for _, clause := range clauses {
		switch requiredVersion := clause.Version.(type) {
		case parser.VersionMachine:
			_, ok := actualVersion.(parser.VersionMachine)
			if !ok {

				res.pushDiagnostic(rng.GetRange(), VersionMismatch{
					GotVersion:      actualVersion,
					RequiredVersion: requiredVersion,
				})
				return
			}

		case parser.VersionInterpreter:

			interpreterActualVersion, ok := actualVersion.(parser.VersionInterpreter)

			if !ok || !interpreterActualVersion.GtEq(requiredVersion) {
				res.pushDiagnostic(rng, VersionMismatch{
					GotVersion:      actualVersion,
					RequiredVersion: requiredVersion,
				})
				return
			}

			if clause.FeatureFlag == "" {
				return
			}

			_, flagEnabled := flags[clause.FeatureFlag]
			if !flagEnabled {
				res.pushDiagnostic(rng, ExperimentalFeature{
					Name: clause.FeatureFlag,
				})
			}
		}
	}
}
