package compiler

// White-box tests asserting the concrete CompilerError produced for invalid
// programs. They call compileProgramToIR directly, since the public Compile
// stringifies the error and would lose the type.

import (
	"testing"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/typecheck"
	"github.com/stretchr/testify/require"
)

func TestE2E_RejectsUnboundVariable(t *testing.T) {
	parsed := parser.Parse(`send [C 10] (source = $undeclared destination = @d)`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToIR(parsed.Value, nil)
	require.IsType(t, TypeError{}, cErr)
	require.IsType(t, typecheck.UnboundVariable{}, cErr.(TypeError).Kind)
}

func TestE2E_RejectsTypeMismatch(t *testing.T) {
	parsed := parser.Parse(`vars { string $s } send [C 10] (source = $s destination = @d)`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToIR(parsed.Value, nil)
	require.IsType(t, TypeError{}, cErr)
	require.IsType(t, typecheck.TypeMismatch{}, cErr.(TypeError).Kind)
}

func TestE2E_RejectsMetaOutsideVarOrigin(t *testing.T) {
	// meta() is only supported as a direct variable origin; nested in an
	// expression it must be a compile error, not a panic.
	parsed := parser.Parse(`
		#![feature("experimental-mid-script-function-call")]
		vars {
			account $a
			number $n = meta($a, "k") + 1
		}
		send [C $n] (source = @world destination = @d)
	`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToIR(parsed.Value, nil)
	require.IsType(t, InvalidMetaPosition{}, cErr)
}

func TestE2E_RejectsNonCastableInterpVar(t *testing.T) {
	// a monetary var has no string form: interpolating it must be a compile
	// error (matching the interpreter's runtime CannotCastToString), not a panic.
	parsed := parser.Parse(`
		#![feature("experimental-account-interpolation")]
		vars { monetary $m }
		set_tx_meta("k", @acc:$m)
	`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToIR(parsed.Value, nil)
	require.IsType(t, CannotCastToString{}, cErr)
	require.Equal(t, typecheck.TypeMonetary, cErr.(CannotCastToString).Type)
}

func TestE2E_AllotmentDuplicateRemaining(t *testing.T) {
	parsed := parser.Parse(`
		send [USD/2 100] (
			source = @world
			destination = {
				remaining to @a
				remaining to @b
			}
		)
	`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToIR(parsed.Value, nil)
	require.IsType(t, DuplicateRemaining{}, cErr)
}

// --- feature flags

// each case is a construct gated behind a feature flag: compiling it without the
// flag must fail, and compiling it with the flag must get past the gate.
func TestFeatureFlagGating(t *testing.T) {
	testCases := []struct {
		name string
		flag flags.FeatureFlag
		src  string
	}{
		{
			name: "oneof in source",
			flag: flags.ExperimentalOneofFeatureFlag,
			src: `send [C 10] (
				source = oneof { @a @b }
				destination = @d
			)`,
		},
		{
			name: "oneof in destination",
			flag: flags.ExperimentalOneofFeatureFlag,
			src: `send [C 10] (
				source = @world
				destination = oneof {
					max [C 3] to @a
					remaining to @b
				}
			)`,
		},
		{
			name: "account interpolation",
			flag: flags.ExperimentalAccountInterpolationFlag,
			src: `vars { string $s }
			send [C 10] (source = @world destination = @dest:$s)`,
		},
		{
			name: "mid-script function call",
			flag: flags.ExperimentalMidScriptFunctionCall,
			src:  `send balance(@a, C) (source = @world destination = @d)`,
		},
		{
			name: "overdraft function",
			flag: flags.ExperimentalOverdraftFunctionFeatureFlag,
			src: `vars { monetary $m = overdraft(@a, C) }
			send $m (source = @world destination = @d)`,
		},
		{
			name: "get_asset function",
			flag: flags.ExperimentalGetAssetFunctionFeatureFlag,
			src: `vars { monetary $m asset $a = get_asset($m) }
			send [$a 10] (source = @world destination = @d)`,
		},
		{
			name: "get_amount function",
			flag: flags.ExperimentalGetAmountFunctionFeatureFlag,
			src: `vars { monetary $m number $n = get_amount($m) }
			send [C $n] (source = @world destination = @d)`,
		},
		{
			name: "asset colors",
			flag: flags.ExperimentalAssetColors,
			src:  `send [C 10] (source = @a \ "RED" destination = @d)`,
		},
		{
			name: "asset scaling",
			flag: flags.AssetScaling,
			src: `send [C 10] (
				source = @src with scaling through @swap
				destination = @d
			)`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parsed := parser.Parse(tc.src)
			require.Empty(t, parsed.Errors)

			_, cErr := compileProgramToIR(parsed.Value, nil)
			require.IsType(t, ExperimentalFeature{}, cErr)
			require.Equal(t, tc.flag, cErr.(ExperimentalFeature).FlagName)

			// with the flag on, whatever comes back must not be about the flag
			// (scaling still hits FeatureNotImplemented)
			_, cErr = compileProgramToIR(parsed.Value, map[string]struct{}{tc.flag: {}})
			_, stillGated := cErr.(ExperimentalFeature)
			require.False(t, stillGated, "still gated with the flag on: %v", cErr)
		})
	}
}

// a function call that *is* the variable's origin is not a mid-script call
func TestFnCallAsVarOriginIsNotMidScript(t *testing.T) {
	parsed := parser.Parse(`
		vars { monetary $m = balance(@a, C) }
		send $m (source = @world destination = @d)
	`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToIR(parsed.Value, nil)
	require.Nil(t, cErr)
}

// ... but one nested inside the origin expression is
func TestNestedFnCallInVarOriginIsMidScript(t *testing.T) {
	parsed := parser.Parse(`
		vars { monetary $m = balance(@a, C) + balance(@b, C) }
		send $m (source = @world destination = @d)
	`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToIR(parsed.Value, nil)
	require.IsType(t, ExperimentalFeature{}, cErr)
	require.Equal(t, flags.ExperimentalMidScriptFunctionCall, cErr.(ExperimentalFeature).FlagName)
}

// #![feature(..)] in the source enables a flag the host didn't pass
func TestInSourceFeatureDeclaration(t *testing.T) {
	parsed := parser.Parse(`
		#![feature("experimental-oneof")]
		send [C 10] (
			source = oneof { @a @b }
			destination = @d
		)
	`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToIR(parsed.Value, nil)
	require.Nil(t, cErr)
}

func TestInSourceFeatureDeclarationRejectsUnknownFlag(t *testing.T) {
	parsed := parser.Parse(`
		#![feature("not-a-flag")]
		send [C 10] (source = @world destination = @d)
	`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToIR(parsed.Value, nil)
	require.IsType(t, InvalidFeature{}, cErr)
	require.Equal(t, "not-a-flag", cErr.(InvalidFeature).Feature)
}

// Every CompilerError must carry a human-readable message: CompilerError is
// parser.Ranged + compileError(), so a type missing Error() still satisfies the
// interface and Compile's fmt.Errorf("%v") would print the raw struct instead.
func TestCompilerErrorMessages(t *testing.T) {
	testCases := []struct {
		name string
		err  CompilerError
		msg  string
	}{
		{"UnboundVar", UnboundVar{Var: "x"}, "the variable '$x' was not declared"},
		{"TypeError", TypeError{Kind: typecheck.UnboundVariable{Name: "x"}}, "The variable '$x' was not declared"},
		{"InvalidUncappedSource", InvalidUncappedSource{}, "cannot take all balance of an unbounded source"},
		{"DuplicateRemaining", DuplicateRemaining{}, "a 'remaining' clause should be the last in an allotment expression"},
		{"InvalidMetaPosition", InvalidMetaPosition{}, "meta() is only allowed as a variable origin"},
		{"CannotCastToString", CannotCastToString{Type: typecheck.TypeMonetary}, "cannot cast a value of type monetary to string"},
		{"FeatureNotImplemented", FeatureNotImplemented{Feature: "scaling"}, "internal error: feature not implemented: scaling"},
		{"ExperimentalFeature", ExperimentalFeature{FlagName: flags.ExperimentalAssetColors}, "You need the 'experimental-asset-colors' feature flag to enable it"},
		{"InvalidFeature", InvalidFeature{Feature: "nope"}, "Invalid feature: nope"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err, ok := tc.err.(error)
			require.True(t, ok, "%T does not implement error", tc.err)
			require.Contains(t, err.Error(), tc.msg)
		})
	}
}
