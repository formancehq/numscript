package analysis_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkSource(input string) []analysis.Diagnostic {
	res := analysis.CheckSource(input)
	for i := range res.Diagnostics {
		res.Diagnostics[i].Id = 0
	}
	return res.Diagnostics
}

func TestInvalidType(t *testing.T) {
	t.Parallel()

	input := `vars { invalid $my_var }
send [C 10] (
	source = $my_var
	destination = $my_var
)`

	require.Equal(t, []analysis.Diagnostic{
		{
			Range: parser.RangeOfIndexed(input, "invalid", 0),
			Kind:  analysis.InvalidType{Name: "invalid"},
		},
	}, checkSource(input))
}

func TestValidType(t *testing.T) {
	t.Parallel()

	input := `vars { account $my_var }
send [C 10] (
	source = $my_var
	destination = $my_var
)`

	require.Empty(t, checkSource(input))
}

func TestDuplicateVariable(t *testing.T) {
	t.Parallel()

	input := `vars {
  account $x
  account $y
  portion $x
}
  send [C 10] (
	source = { $x $y }
	destination = @dest
)`

	require.Equal(t, []analysis.Diagnostic{
		{
			Range: parser.RangeOfIndexed(input, "$x", 1),
			Kind:  analysis.DuplicateVariable{Name: "x"},
		},
	}, checkSource(input))
}

func TestUnboundVarInSaveAccount(t *testing.T) {
	t.Parallel()

	input := `save $unbound_mon from $unbound_acc`

	require.Equal(t, []analysis.Diagnostic{
		{
			Range: parser.RangeOfIndexed(input, "$unbound_mon", 0),
			Kind:  analysis.UnboundVariable{Name: "unbound_mon", Type: "monetary"},
		},
		{
			Range: parser.RangeOfIndexed(input, "$unbound_acc", 0),
			Kind:  analysis.UnboundVariable{Name: "unbound_acc", Type: "account"},
		},
	}, checkSource(input))

}

func TestUnboundVarInInfixOp(t *testing.T) {
	t.Parallel()

	input := `
		send [COIN 10] + $unbound_mon1 (
			source = max [COIN 10] + $unbound_mon2 from @world
			destination = @b
		)
	`

	assert.Equal(t,
		[]analysis.Diagnostic{
			{
				Kind:  analysis.UnboundVariable{Name: "unbound_mon1", Type: analysis.TypeMonetary},
				Range: parser.RangeOfIndexed(input, "$unbound_mon1", 0),
			},
			{
				Kind:  analysis.UnboundVariable{Name: "unbound_mon2", Type: analysis.TypeMonetary},
				Range: parser.RangeOfIndexed(input, "$unbound_mon2", 0),
			},
		},
		checkSource(input),
	)
}

func TestMismatchedTypeInSave(t *testing.T) {
	t.Parallel()

	input := `vars {
	string $str
	number $n
}
	
save $str from $n
`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 2)

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Kind:  analysis.TypeMismatch{Expected: "monetary", Got: "string"},
				Range: parser.RangeOfIndexed(input, "$str", 1),
			},
			{
				Kind:  analysis.TypeMismatch{Expected: "account", Got: "number"},
				Range: parser.RangeOfIndexed(input, "$n", 1),
			},
		},
		checkSource(input),
	)
}

func TestUnboundVarInSource(t *testing.T) {
	t.Parallel()

	input := `send [C 1] (
  source = { max [C 1] from $unbound_var }
  destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$unbound_var", 0),
				Kind:  analysis.UnboundVariable{Name: "unbound_var", Type: analysis.TypeAccount},
			},
		},
		checkSource(input),
	)
}

func TestUnboundVarInSourceOneof(t *testing.T) {
	t.Parallel()

	input := `send [C 1] (
  source = oneof { $unbound_var }
  destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$unbound_var", 0),
				Kind:  analysis.UnboundVariable{Name: "unbound_var", Type: analysis.TypeAccount},
			},
		},
		checkSource(input),
	)
}
func TestUnboundVarInDest(t *testing.T) {
	t.Parallel()

	input := `send [C 1] (
  source = @src
  destination = {
	1/2 to @a
	1/2 to $unbound_var
  }
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$unbound_var", 0),
				Kind:  analysis.UnboundVariable{Name: "unbound_var", Type: analysis.TypeAccount},
			},
		},
		checkSource(input),
	)
}

func TestUnboundMany(t *testing.T) {
	t.Parallel()

	input := `send [C 1] (
  	source = {
  		1/3 from $unbound1
		2/3 from $unbound2
	}
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$unbound1", 0),
				Kind:  analysis.UnboundVariable{Name: "unbound1", Type: analysis.TypeAccount},
			},
			{
				Range: parser.RangeOfIndexed(input, "$unbound2", 0),
				Kind:  analysis.UnboundVariable{Name: "unbound2", Type: analysis.TypeAccount},
			},
		},
		checkSource(input),
	)
}

func TestUnboundCurrenciesVars(t *testing.T) {
	t.Parallel()

	input := `send $unbound1 (
  	source = {
		max $unbound2 from @a
	}
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$unbound1", 0),
				Kind:  analysis.UnboundVariable{Name: "unbound1", Type: analysis.TypeMonetary},
			},
			{
				Range: parser.RangeOfIndexed(input, "$unbound2", 0),
				Kind:  analysis.UnboundVariable{Name: "unbound2", Type: analysis.TypeMonetary},
			},
		},
		checkSource(input),
	)
}

func TestUnusedVarInSource(t *testing.T) {
	t.Parallel()

	input := `vars { monetary $unused_var }`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$unused_var", 0),
				Kind:  analysis.UnusedVar{Name: "unused_var"},
			},
		},
		checkSource(input),
	)
}

func TestWrongTypeForMonetaryLitAsset(t *testing.T) {
	t.Parallel()

	input := `vars { account $a }

send [$a 100] (
  	source = @src
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$a", 1),
				Kind: analysis.TypeMismatch{
					Expected: "asset",
					Got:      "account",
				},
			},
		},
		checkSource(input),
	)
}

func TestWrongTypeForMonetaryLitNumber(t *testing.T) {
	t.Parallel()

	input := `vars { account $n }

send [EUR/2 $n] (
  	source = @src
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$n", 1),
				Kind: analysis.TypeMismatch{
					Expected: "number",
					Got:      "account",
				},
			},
		},
		checkSource(input),
	)
}

func TestWrongTypeForCap(t *testing.T) {
	t.Parallel()

	input := `vars { account $account }

send [COIN 100] (
  	source = max $account from @a
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$account", 1),
				Kind: analysis.TypeMismatch{
					Expected: "monetary",
					Got:      "account",
				},
			},
		},
		checkSource(input),
	)
}

func TestWrongTypeForSrcAccount(t *testing.T) {
	t.Parallel()

	input := `vars { portion $x }

send [COIN 100] (
  	source = $x
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$x", 1),
				Kind: analysis.TypeMismatch{
					Expected: "account",
					Got:      "portion",
				},
			},
		},
		checkSource(input),
	)
}

func TestWrongTypeForDestAccount(t *testing.T) {
	t.Parallel()

	input := `vars { portion $x }

send [COIN 100] (
  	source = @src
  	destination = $x
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$x", 1),
				Kind: analysis.TypeMismatch{
					Expected: "account",
					Got:      "portion",
				},
			},
		},
		checkSource(input),
	)
}

func TestWrongTypeForUnboundedAccount(t *testing.T) {
	t.Parallel()

	input := `vars { portion $x }

send [COIN 100] (
  	source = $x allowing unbounded overdraft
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$x", 1),
				Kind: analysis.TypeMismatch{
					Expected: "account",
					Got:      "portion",
				},
			},
		},
		checkSource(input),
	)
}

func TestWrongTypeForBoundedOverdraftCap(t *testing.T) {
	t.Parallel()

	input := `vars { portion $x }

send [COIN 100] (
  	source = @x allowing overdraft up to $x
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$x", 1),
				Kind: analysis.TypeMismatch{
					Expected: "monetary",
					Got:      "portion",
				},
			},
		},
		checkSource(input),
	)
}

func TestWrongTypeForSrcAllotmentPortion(t *testing.T) {
	t.Parallel()

	input := `vars { string $p }

send [COIN 100] (
  	source = {
		$p from @a
		remaining from @b
	}
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$p", 1),
				Kind: analysis.TypeMismatch{
					Expected: "portion",
					Got:      "string",
				},
			},
		},
		checkSource(input),
	)
}

func TestInvalidUnboundedAccount(t *testing.T) {
	t.Parallel()

	input := `
send [COIN *] (
  source = {
    @s allowing unbounded overdraft
  }
  destination = @dest
)
`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "@s", 0),
				Kind:  analysis.InvalidUnboundedAccount{},
			},
		},
		checkSource(input),
	)
}

func TestNoInvalidUnboundedAccountWhenBounded(t *testing.T) {
	t.Parallel()

	input := `
send [COIN *] (
  source = {
    @s allowing overdraft up to [COIN 2]
  }
  destination = @dest
)
`

	require.Equal(t,
		[]analysis.Diagnostic(nil),
		checkSource(input),
	)
}

func TestWrongTypeForDestAllotmentPortion(t *testing.T) {
	t.Parallel()

	input := `vars { string $p }

send [COIN 100] (
  	source = @s
  	destination = {
		$p to @a
		remaining to @dest
	}
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$p", 1),
				Kind: analysis.TypeMismatch{
					Expected: "portion",
					Got:      "string",
				},
			},
		},
		checkSource(input),
	)
}

func TestCheckPlus(t *testing.T) {
	t.Parallel()

	t.Run("error in number+portion", func(t *testing.T) {
		input := `set_tx_meta("k", 1 + 1/2)`

		require.Equal(t,
			[]analysis.Diagnostic{
				{
					Range: parser.RangeOfIndexed(input, "1/2", 0),
					Kind: analysis.TypeMismatch{
						Expected: "number",
						Got:      "portion",
					},
				},
			},
			checkSource(input),
		)
	})

	t.Run("allow number+number", func(t *testing.T) {
		input := `set_tx_meta("k", 1 + 2)`

		require.Empty(t, checkSource(input))
	})

	t.Run("allow monetary+monetary", func(t *testing.T) {
		input := `set_tx_meta("k", [EUR/2 10] + [EUR/2 20])`

		require.Empty(t, checkSource(input))
	})

	t.Run("error when left side is invalid", func(t *testing.T) {
		input := `set_tx_meta("k", @acc + @acc)`

		require.Equal(t,
			[]analysis.Diagnostic{
				{
					Range: parser.RangeOfIndexed(input, "@acc", 0),
					Kind: analysis.TypeMismatch{
						Expected: "number|monetary",
						Got:      "account",
					},
				},
			},
			checkSource(input),
		)
	})

	t.Run("no type error when left side is any", func(t *testing.T) {
		input := `set_tx_meta("k", $unbound_var + @acc)`

		require.Equal(t,
			[]analysis.Diagnostic{
				{
					Range: parser.RangeOfIndexed(input, "$unbound_var", 0),
					Kind: analysis.UnboundVariable{
						Name: "unbound_var",
						Type: analysis.TypeNumber,
					},
				},
			},
			checkSource(input),
		)
	})
}

func TestCheckMinus(t *testing.T) {
	t.Parallel()

	t.Run("error in number-portion", func(t *testing.T) {
		input := `set_tx_meta("k", 1 - 1/2)`

		require.Equal(t,
			[]analysis.Diagnostic{
				{
					Range: parser.RangeOfIndexed(input, "1/2", 0),
					Kind: analysis.TypeMismatch{
						Expected: "number",
						Got:      "portion",
					},
				},
			},
			checkSource(input),
		)
	})

	t.Run("allow number-number", func(t *testing.T) {
		input := `set_tx_meta("k", 1 - 2)`

		require.Empty(t, checkSource(input))
	})

	t.Run("allow monetary-monetary", func(t *testing.T) {
		input := `set_tx_meta("k", [EUR/2 10] - [EUR/2 20])`

		require.Empty(t, checkSource(input))
	})

	t.Run("error when left side is invalid", func(t *testing.T) {
		input := `set_tx_meta("k", @acc - @acc)`

		require.Equal(t,
			[]analysis.Diagnostic{
				{
					Range: parser.RangeOfIndexed(input, "@acc", 0),
					Kind: analysis.TypeMismatch{
						Expected: "number|monetary",
						Got:      "account",
					},
				},
			},
			checkSource(input),
		)
	})

	t.Run("no type error when left side is any", func(t *testing.T) {
		input := `set_tx_meta("k", $unbound_var - @acc)`

		require.Equal(t,
			[]analysis.Diagnostic{
				{
					Range: parser.RangeOfIndexed(input, "$unbound_var", 0),
					Kind: analysis.UnboundVariable{
						Name: "unbound_var",
						Type: analysis.TypeNumber,
					},
				},
			},
			checkSource(input),
		)
	})
}

func TestNoUnusedOnStringInterp(t *testing.T) {
	t.Parallel()

	input := `vars { number $id }
send [EUR/2 *] (
  	source = @user:$id:pending
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Empty(t, diagnostics)

}

func TestWrongTypeInsideAccountInterp(t *testing.T) {
	t.Skip("TODO formalize a better type system to model this easy")

	t.Parallel()

	input := `vars { monetary $m }
send [EUR/2 *] (
  	source = @user:$m
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics

	require.Len(t, diagnostics, 1, "diagnostics=%#v\n", diagnostics)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "number|account|string",
			Got:      "monetary",
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "$m", 1),
		d1.Range,
	)
}

func TestDoNoCheckVersionWhenNotSpecified(t *testing.T) {
	t.Parallel()

	input := `
vars {
	number $n
}

send [EUR/2 100] (
  	source = {
			1/$n from @a
			remaining from @b
		}
  	destination = @dest
)
`

	require.Equal(t,
		[]analysis.Diagnostic(nil),
		checkSource(input),
	)

}

func TestRequireVersionForInfixDiv(t *testing.T) {
	t.Parallel()

	input := `
// @version machine

vars {
	number $n
}

send [EUR/2 100] (
  	source = {
			1/$n from @a
			remaining from @b
		}
  	destination = @dest
)
`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "1/$n", 0),
				Kind: analysis.VersionMismatch{
					GotVersion:      parser.VersionMachine{},
					RequiredVersion: parser.NewVersionInterpreter(0, 0, 15),
				},
			},
		},
		checkSource(input),
	)
}

func TestRequireVersionForInfixDivWhenVersionLt(t *testing.T) {
	t.Parallel()

	input := `
// required version is 0.0.15
// @version interpreter 0.0.1

vars {
	number $n
}

send [EUR/2 100] (
  	source = {
			1/$n from @a
			remaining from @b
		}
  	destination = @dest
)
`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "1/$n", 0),
				Kind: analysis.VersionMismatch{
					GotVersion:      parser.NewVersionInterpreter(0, 0, 1),
					RequiredVersion: parser.NewVersionInterpreter(0, 0, 15),
				},
			},
		},
		checkSource(input),
	)
}

func TestRequireFlagForOneofWhenMissing(t *testing.T) {
	t.Parallel()

	input := `
// @version interpreter 0.0.15

send [EUR/2 100] (
  	source = oneof {
			@a
			@b
		}
  	destination = @dest
)
`

	ds := checkSource(input)
	require.Len(t, ds, 1)

	require.Equal(t,
		analysis.ExperimentalFeature{
			Name: flags.ExperimentalOneofFeatureFlag,
		},
		ds[0].Kind,
	)
}

func TestRequireFlagForOneofWhenGiven(t *testing.T) {
	t.Parallel()

	input := `
// @version interpreter 0.0.15
// @feature_flag experimental-oneof

send [EUR/2 100] (
  	source = oneof {
			@a
			@b
		}
  	destination = @dest
)
`

	require.Empty(t, checkSource(input))
}

func TestRequireFlagForInterpolation(t *testing.T) {
	t.Parallel()

	input := `
// @version interpreter 0.0.15
vars {
	number $id 
}

send [EUR/2 100] (
  	source = @user:$id
  	destination = @dest
)
`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "@user:$id", 0),
				Kind: analysis.ExperimentalFeature{
					Name: flags.ExperimentalAccountInterpolationFlag,
				},
			},
		},
		checkSource(input),
	)
}

func TestRequireFlagForFunctionOverdraft(t *testing.T) {
	t.Parallel()

	input := `
// @version interpreter 0.0.15
vars {
	monetary $m = overdraft(@acc, USD/2)
}

send $m ( source = @world destination = @dest)
`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "overdraft(@acc, USD/2)", 0),
				Kind: analysis.ExperimentalFeature{
					Name: flags.ExperimentalOverdraftFunctionFeatureFlag,
				},
			},
		},
		checkSource(input),
	)
}

func TestCheckColorVar(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 100] (
  	source = {
			@a \ $col1
			@b \ $col2 allowing unbounded overdraft
		}
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$col1", 0),
				Kind:  analysis.UnboundVariable{Name: "col1", Type: "string"},
			},
			{
				Range: parser.RangeOfIndexed(input, "$col2", 0),
				Kind:  analysis.UnboundVariable{Name: "col2", Type: "string"},
			},
		},
		checkSource(input),
	)
}

func TestInorderRedundantWhenColored(t *testing.T) {
	t.Parallel()

	input := `send [COIN 100] (
  	source = {
			@a
			@a \ "x"
		}
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic(nil),
		checkSource(input),
	)
}

func TestInorderRedundantWhenEmptyColored(t *testing.T) {
	t.Parallel()

	input := `send [COIN 100] (
  	source = {
			@a
			@a \ "" // <- empty color behaves as no color
		}
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "@a", 1),
				Kind:  analysis.EmptiedAccount{Name: "a"},
			},
		},
		checkSource(input),
	)
}

func TestCheckAssetMismatch(t *testing.T) {

	t.Parallel()

	input := `
	
	send [USD 100] (
  	source = max [EUR 10] from @a
  	destination = @dest
)`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "[EUR 10]", 0),
				Kind:  &analysis.AssetMismatch{Expected: "USD", Got: "EUR"},
			},
		},
		checkSource(input),
	)
}

func TestCheckAssetMismatchInVar(t *testing.T) {

	t.Parallel()

	input := `
	
vars {
  monetary $mon
}

send [EUR 0] (
  source = max $mon from @a
  destination = @b
)

send [USD 0] (
  source = max $mon from @a
  destination = @b
)

`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$mon", 2),
				// TODO shoulnd't the error be the other way around?
				Kind: &analysis.AssetMismatch{Expected: "USD", Got: "EUR"},
			},
		},
		checkSource(input),
	)
}

func TestCheckBalanceAssetConstraint(t *testing.T) {
	t.Parallel()

	input := `
vars {
	monetary $mon = balance(@acc, USD/2)
}

send [USD 42] (
  source = max $mon from @a
  destination = @b
)
`

	require.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$mon", 1),
				Kind:  &analysis.AssetMismatch{Expected: "USD", Got: "USD/2"},
			},
		},
		checkSource(input),
	)
}

func TestInferVars(t *testing.T) {
	t.Parallel()

	input := `
vars {
	monetary $mon1
	monetary $mon2
}

send $mon1 (
  source = @a allowing overdraft up to $mon2
  destination = @b
)
`

	res := analysis.CheckSource(input)

	t1 := res.VarTypes[res.DeclaredVars["mon1"]]

	t2 := res.VarTypes[res.DeclaredVars["mon2"]]

	require.Same(t, t1.Resolve(), t2.Resolve())
}

func TestInferGetAsset(t *testing.T) {
	t.Parallel()

	input := `
vars {
	asset $ass = get_asset([USD/2 100])
}
`

	res := analysis.CheckSource(input)

	v := res.DeclaredVars["ass"]
	t1 := res.VarTypes[v]

	expected := analysis.TAsset("USD/2")
	require.Equal(t, &expected, t1.Resolve())
}
