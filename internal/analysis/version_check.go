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

func (res *CheckResult) checkAccountInterpolationVersion(expr parser.AccountInterpLiteral) {
	isInterpolation := slices.ContainsFunc(expr.Parts, func(part parser.AccountNamePart) bool {
		_, isVar := part.(*parser.Variable)
		return isVar
	})

	if !isInterpolation {
		return
	}

	res.requireVersion(expr.Range, flaggedUntilStable(parser.NewVersionInterpreter(0, 0, 15), flags.ExperimentalAccountInterpolationFlag)...)
}

func (res *CheckResult) checkMidScriptFnCallVersion(rng parser.Range) {
	res.requireVersion(rng, flaggedUntilStable(parser.NewVersionInterpreter(0, 0, 15), flags.ExperimentalMidScriptFunctionCall)...)
}

func (res *CheckResult) checkColorVersion(color parser.ValueExpr) {
	if color == nil {
		return
	}

	res.requireVersion(color.GetRange(), flaggedUntilStable(parser.NewVersionInterpreter(0, 0, 17), flags.ExperimentalAssetColors)...)
}

// -- version check utilities

func flaggedUntilStable(since parser.VersionInterpreter, flag flags.FeatureFlag) []VersionClause {
	return []VersionClause{
		{Version: since, FeatureFlag: flag},
		{Version: parser.NewVersionInterpreter(1, 0, 0)},
	}
}

type VersionClause struct {
	Version     parser.Version
	FeatureFlag flags.FeatureFlag
}

// Clauses are alternatives: when none is satisfied, the first one's failure is reported.
func (res *CheckResult) requireVersion(
	rng parser.Range,
	clauses ...VersionClause,
) {
	flags := res.Program.GetFlags()
	actualVersion := res.Program.GetVersion()
	if actualVersion == nil || len(clauses) == 0 {
		return
	}

	for _, clause := range clauses {
		if clauseFailure(actualVersion, flags, clause) == nil {
			return
		}
	}

	res.pushDiagnostic(rng, clauseFailure(actualVersion, flags, clauses[0]))
}

func clauseFailure(
	actualVersion parser.Version,
	flags map[string]struct{},
	clause VersionClause,
) DiagnosticKind {
	switch requiredVersion := clause.Version.(type) {
	case parser.VersionMachine:
		if _, ok := actualVersion.(parser.VersionMachine); !ok {
			return VersionMismatch{GotVersion: actualVersion, RequiredVersion: requiredVersion}
		}
		return nil

	case parser.VersionInterpreter:
		interpreterActualVersion, ok := actualVersion.(parser.VersionInterpreter)
		if !ok || !interpreterActualVersion.GtEq(requiredVersion) {
			return VersionMismatch{GotVersion: actualVersion, RequiredVersion: requiredVersion}
		}

		if clause.FeatureFlag == "" {
			return nil
		}

		if _, flagEnabled := flags[clause.FeatureFlag]; !flagEnabled {
			return ExperimentalFeature{Name: clause.FeatureFlag}
		}
		return nil
	}

	return nil
}
