package analysis_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidType(t *testing.T) {
	t.Parallel()

	input := `vars { invalid $my_var }
send [C 10] (
	source = $my_var
	destination = $my_var
)`

	res := analysis.CheckSource(input)
	require.Lenf(t, res.Diagnostics, 1, "xs: %#v", res.Diagnostics)

	d1 := res.Diagnostics[0]
	assert.Equal(t,
		parser.Range{
			Start: parser.Position{Character: 7},
			End:   parser.Position{Character: 7 + len("invalid")},
		},
		d1.Range,
	)

	assert.Equal(t,
		&analysis.InvalidType{Name: "invalid"},
		d1.Kind,
	)
}

func TestValidType(t *testing.T) {
	t.Parallel()

	input := `vars { account $my_var }
send [C 10] (
	source = $my_var
	destination = $my_var
)`
	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 0)
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

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		parser.Range{
			Start: parser.Position{Line: 3, Character: 10},
			End:   parser.Position{Line: 3, Character: 10 + len("$x")},
		},
		d1.Range,
	)

	assert.Equal(t,
		&analysis.DuplicateVariable{Name: "x"},
		d1.Kind,
	)
}

func TestUnboundVarInSaveAccount(t *testing.T) {
	t.Parallel()

	input := `save $unbound_mon from $unbound_acc`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 2)

	assert.Equal(t,
		[]analysis.Diagnostic{
			{
				Kind:  &analysis.UnboundVariable{Name: "unbound_mon"},
				Range: parser.RangeOfIndexed(input, "$unbound_mon", 0),
			},
			{
				Kind:  &analysis.UnboundVariable{Name: "unbound_acc"},
				Range: parser.RangeOfIndexed(input, "$unbound_acc", 0),
			},
		},
		diagnostics,
	)
}

func TestUnboundVarInInfixOp(t *testing.T) {
	t.Parallel()

	input := `
		send [COIN 10] + $unbound_mon1 (
			source = max [COIN 10] + $unbound_mon2 from @world
			destination = @b
		)
	`

	parseResult := parser.Parse(input)
	require.Empty(t, parseResult.Errors)

	assert.Equal(t,
		[]analysis.Diagnostic{
			{
				Kind:  &analysis.UnboundVariable{Name: "unbound_mon1"},
				Range: parser.RangeOfIndexed(input, "$unbound_mon1", 0),
			},
			{
				Kind:  &analysis.UnboundVariable{Name: "unbound_mon2"},
				Range: parser.RangeOfIndexed(input, "$unbound_mon2", 0),
			},
		},
		analysis.CheckProgram(parseResult.Value).Diagnostics,
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

	assert.Equal(t,
		[]analysis.Diagnostic{
			{
				Kind:  &analysis.TypeMismatch{Expected: "monetary", Got: "string"},
				Range: parser.RangeOfIndexed(input, "$str", 1),
			},
			{
				Kind:  &analysis.TypeMismatch{Expected: "account", Got: "number"},
				Range: parser.RangeOfIndexed(input, "$n", 1),
			},
		},
		diagnostics,
	)
}

func TestUnboundVarInSource(t *testing.T) {
	t.Parallel()

	input := `send [C 1] (
  source = { max [C 1] from $unbound_var }
  destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		parser.Range{
			Start: parser.Position{Line: 1, Character: 28},
			End:   parser.Position{Line: 1, Character: 28 + len("$unbound_var")},
		},
		d1.Range,
	)

	assert.Equal(t,
		&analysis.UnboundVariable{Name: "unbound_var"},
		d1.Kind,
	)
}

func TestUnboundVarInSourceOneof(t *testing.T) {
	t.Parallel()

	input := `send [C 1] (
  source = oneof { $unbound_var }
  destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	assert.Equal(t,
		[]analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$unbound_var", 0),
				Kind:  &analysis.UnboundVariable{Name: "unbound_var"},
			},
		},
		diagnostics,
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

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Lenf(t, diagnostics, 1, "expected len to be 1")

	d1 := diagnostics[0]
	assert.Equal(t,
		parser.RangeOfIndexed(input, "$unbound_var", 0),
		d1.Range,
	)
	assert.Equal(t,
		&analysis.UnboundVariable{Name: "unbound_var"},
		d1.Kind,
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

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 2)
}

func TestUnboundCurrenciesVars(t *testing.T) {
	t.Parallel()

	input := `send $unbound1 (
  	source = {
		max $unbound2 from @a
	}
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 2)
}

// TODO unbound vars in declr

func TestUnusedVarInSource(t *testing.T) {
	t.Parallel()

	input := `vars { monetary $unused_var }`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		parser.Range{
			Start: parser.Position{Character: 16},
			End:   parser.Position{Character: 16 + len("$unused_var")},
		},
		d1.Range,
	)

	assert.Equal(t,
		&analysis.UnusedVar{Name: "unused_var"},
		d1.Kind,
	)
}

func TestWrongTypeForMonetaryLitAsset(t *testing.T) {
	t.Parallel()

	input := `vars { account $a }

send [$a 100] (
  	source = @src
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "asset",
			Got:      "account",
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "$a", 1),
		d1.Range,
	)
}

func TestWrongTypeForMonetaryLitNumber(t *testing.T) {
	t.Parallel()

	input := `vars { account $n }

send [EUR/2 $n] (
  	source = @src
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "number",
			Got:      "account",
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "$n", 1),
		d1.Range,
	)
}

func TestWrongTypeForCap(t *testing.T) {
	t.Parallel()

	input := `vars { account $account }

send [COIN 100] (
  	source = max $account from @a
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "monetary",
			Got:      "account",
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "$account", 1),
		d1.Range,
	)
}

func TestWrongTypeForSrcAccount(t *testing.T) {
	t.Parallel()

	input := `vars { portion $x }

send [COIN 100] (
  	source = $x
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "account",
			Got:      "portion",
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "$x", 1),
		d1.Range,
	)
}

func TestWrongTypeForDestAccount(t *testing.T) {
	t.Parallel()

	input := `vars { portion $x }

send [COIN 100] (
  	source = @src
  	destination = $x
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "account",
			Got:      "portion",
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "$x", 1),
		d1.Range,
	)
}

func TestWrongTypeForUnboundedAccount(t *testing.T) {
	t.Parallel()

	input := `vars { portion $x }

send [COIN 100] (
  	source = $x allowing unbounded overdraft
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "account",
			Got:      "portion",
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "$x", 1),
		d1.Range,
	)
}

func TestWrongTypeForBoundedOverdraftCap(t *testing.T) {
	t.Parallel()

	input := `vars { portion $x }

send [COIN 100] (
  	source = @x allowing overdraft up to $x
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "monetary",
			Got:      "portion",
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "$x", 1),
		d1.Range,
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

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "portion",
			Got:      "string",
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "$p", 1),
		d1.Range,
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

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "portion",
			Got:      "string",
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "$p", 1),
		d1.Range,
	)
}

func TestBadRemainingInSource(t *testing.T) {
	t.Parallel()

	input := `send [COIN 100] (
  	source = {
		1/2 from @a
		remaining from @b
		1/2 from @c
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.RemainingIsNotLast{},
		d1.Kind,
	)

	assert.Equal(t,
		parser.Range{
			Start: parser.Position{Line: 1, Character: 12},
			End:   parser.Position{Line: 5, Character: 5},
		},
		d1.Range,
	)

}

func TestBadRemainingInDest(t *testing.T) {
	t.Parallel()

	input := `send [COIN 100] (
  	source = @a
  	destination = {
			1/2 from @a
			remaining from @b
			1/2 from @c
    }
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.RemainingIsNotLast{},
		d1.Kind,
	)

	assert.Equal(t,
		parser.Range{
			Start: parser.Position{Line: 2, Character: 17},
			End:   parser.Position{Line: 6, Character: 5},
		},
		d1.Range,
	)

}

func TestBadAllotmentSumInSourceLessThanOne(t *testing.T) {
	t.Parallel()

	input := `send [COIN 100] (
   source = {
		1/3 from @s1
		1/3 from @s2
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Lenf(t, diagnostics, 1, "wrong diagnostics len")

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.BadAllotmentSum{
			Sum: *big.NewRat(2, 3),
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.Range{
			Start: parser.Position{Line: 1, Character: 12},
			End:   parser.Position{Line: 4, Character: 5},
		},
		d1.Range,
	)
}

func TestBadAllotmentSumInSourceMoreThanOne(t *testing.T) {
	t.Parallel()

	input := `send [COIN 100] (
   source = {
		2/3 from @s1
		2/3 from @s2
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Lenf(t, diagnostics, 1, "wrong diagnostics len")

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.BadAllotmentSum{
			Sum: *big.NewRat(4, 3),
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.Range{
			Start: parser.Position{Line: 1, Character: 12},
			End:   parser.Position{Line: 4, Character: 5},
		},
		d1.Range,
	)

}

func TestBadAllotmentSumInDestinationLessThanOne(t *testing.T) {
	t.Parallel()

	input := `send [COIN 100] (
   source = @src
  	destination = {
		1/3 to @d1
		1/3 to @d2
    }
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Lenf(t, diagnostics, 1, "wrong diagnostics len")

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.BadAllotmentSum{
			Sum: *big.NewRat(2, 3),
		},
		d1.Kind,
	)
}

func TestNoAllotmentLt1ErrIfVariable(t *testing.T) {
	t.Parallel()

	input := `vars {
	portion $portion1
	portion $portion2
}

send [COIN 100] (
   source = {
		1/3 from @s1
		1/3 from @s2
		$portion1 from @s3
		$portion2 from @s4
    }
  	destination = @d
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 0)
}

func TestAllotmentGt1ErrIfVariable(t *testing.T) {
	t.Parallel()

	input := `vars {
	portion $portion1
	portion $portion2
}

send [COIN 100] (
   source = @src
  	destination = {
			2/3 to @d1
			2/3 to @d2
			$portion1 to @d3
			$portion2 to @d4
    }
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	assert.IsType(t, diagnostics[0].Kind, &analysis.BadAllotmentSum{})
}

func TestAllotmentErrOnlyOneVar(t *testing.T) {
	t.Parallel()

	input := `vars { portion $portion }

send [COIN 100] (
   source = {
		2/3 from @s1
		$portion from @s2
   }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	assert.Equal(t, diagnostics[0], analysis.Diagnostic{
		Kind: &analysis.FixedPortionVariable{
			Value: *big.NewRat(1, 3),
		},
		Range: parser.RangeOfIndexed(input, "$portion", 1),
	})
}

func TestAllotmentErrWhenVarIsZero(t *testing.T) {
	t.Parallel()

	input := `vars {
	portion $portion1
	portion $portion2
}

send [COIN 100] (
   source = {
		2/3 from @s1
		1/3 from @s2
		$portion1 from @s3
		$portion2 from @s4
   }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 2)

	assert.Equal(t, diagnostics[0], analysis.Diagnostic{
		Kind: &analysis.FixedPortionVariable{
			Value: *big.NewRat(0, 1),
		},
		Range: parser.RangeOfIndexed(input, "$portion1", 1),
	})

	assert.Equal(t, diagnostics[1], analysis.Diagnostic{
		Kind: &analysis.FixedPortionVariable{
			Value: *big.NewRat(0, 1),
		},
		Range: parser.RangeOfIndexed(input, "$portion2", 1),
	})
}

func TestNoBadAllotmentWhenRemaining(t *testing.T) {
	t.Parallel()

	input := `send [COIN 100] (
   source = {
		1/3 from @s1
		1/3 from @s2
		remaining from @s3
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Lenf(t, diagnostics, 0, "wrong diagnostics len")
}

func TestBadAllotmentWhenRemainingButGt1(t *testing.T) {
	t.Parallel()

	input := `send [COIN 100] (
   source = {
		2/3 from @s1
		2/3 from @s2
		remaining from @s3
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Lenf(t, diagnostics, 1, "wrong diagnostics len")

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.BadAllotmentSum{
			Sum: *big.NewRat(4, 3),
		},
		d1.Kind,
	)
}

func TestRedundantRemainingWhenSumIsOne(t *testing.T) {
	t.Parallel()

	input := `send [COIN 100] (
   source = {
		2/3 from @s1
		1/3 from @s2
		remaining from @s3
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Lenf(t, diagnostics, 1, "wrong diagnostics len")

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.RedundantRemaining{},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "remaining", 0),
		d1.Range,
	)
}

func TestNoSingleAllotmentVariable(t *testing.T) {
	t.Parallel()

	input := `vars { portion $allot }

send [COIN 100] (
   source = {
		$allot from @s1
		remaining from @s2
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Lenf(t, diagnostics, 0, "wrong diagnostics len")
}

func TestCheckNoUnboundFunctionCall(t *testing.T) {
	t.Parallel()

	input := `invalid_fn_call()`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.UnknownFunction{Name: "invalid_fn_call"},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "invalid_fn_call", 0),
		d1.Range,
	)
}

func TestAllowedFnCall(t *testing.T) {
	t.Parallel()

	input := `set_tx_meta("for_cone", "true")`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 0)
}

func TestCheckFnCallTypesWrongType(t *testing.T) {
	t.Parallel()

	input := `set_tx_meta(@addr, 42)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "string",
			Got:      "account",
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, "@addr", 0),
		d1.Range,
	)
}

func TestTooFewFnArgs(t *testing.T) {
	t.Parallel()

	input := `set_tx_meta("arg")`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.BadArity{
			Expected: 2,
			Actual:   1,
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, `set_tx_meta("arg")`, 0),
		d1.Range,
	)
}

func TestTooManyFnArgs(t *testing.T) {
	t.Parallel()

	input := `set_tx_meta("arg", "ok", 10, 20)`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.BadArity{
			Expected: 2,
			Actual:   4,
		},
		d1.Kind,
	)

	assert.Equal(t,
		parser.RangeOfIndexed(input, `10, 20`, 0),
		d1.Range,
	)
}

func TestCheckTrailingCommaFnCall(t *testing.T) {
	t.Parallel()

	input := `set_tx_meta("ciao", 42, 10, )`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)
}

func TestCheckTypesOriginFn(t *testing.T) {
	t.Parallel()

	input := `
	vars {
		monetary $mon = meta(42, "str")
	}

	send $mon (
		source = @s
		destination = @d
	)
	`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "account",
			Got:      "number",
		},
		d1.Kind,
	)
}

func TestCheckReturnTypeOriginFn(t *testing.T) {
	t.Parallel()

	input := `
	vars {
		account $mon = balance(@account, EUR/2)
	}

	send [EUR/2 100] (
		source = $mon
		destination = @d
	)
	`

	program := parser.Parse(input).Value

	diagnostics := analysis.CheckProgram(program).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "monetary",
			Got:      "account",
		},
		d1.Kind,
	)
}

func TestWorldOverdraft(t *testing.T) {
	t.Parallel()

	input := `
	send [EUR/2 100] (
		source = {
			@a
			@world allowing unbounded overdraft
		}
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.InvalidWorldOverdraft{},
		d1.Kind,
	)

	assert.Equal(t, d1.Range, parser.RangeOfIndexed(input, "@world", 0))
}

func TestForbidAllotmentInSendAll(t *testing.T) {
	t.Parallel()

	input := `
	send [EUR/2 *] (
		source = {
			1/2 from @s1
			remaining from @s2
		}
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.NoAllotmentInSendAll{},
		d1.Kind,
	)
}

func TestAllowAllotmentInCappedSendAll(t *testing.T) {
	t.Parallel()

	input := `
	send [EUR/2 *] (
		source = {
			max [EUR/2 10] from {
				1/2 from @s1
				remaining from @s2
			}
		}
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Empty(t, diagnostics)
}

func TestDisallowAllotmentInCappedSendAllOutsideMax(t *testing.T) {
	t.Parallel()

	input := `
	send [EUR/2 *] (
		source = {
			max [EUR/2 10] from @a
			{
				1/2 from @s1
				remaining from @s2
			}
		}
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.NoAllotmentInSendAll{},
		d1.Kind,
	)
}

func TestNoForbidAllotmentInSendAll(t *testing.T) {
	t.Parallel()

	input := `
	send [EUR/2 *] (
		source = @a
		destination = @dest
	)


	send [EUR/2 100] (
		source = {
			1/2 from @s1
			remaining from @s2
		}
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Empty(t, diagnostics)
}

func TestForbidUnboundedSrcInSendAll(t *testing.T) {
	t.Parallel()

	input := `
	send [GEM *] (
		source = {
			@ok
			@illegal allowing unbounded overdraft // <- err
		}
		destination = @b
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Len(t, diagnostics, 1)

	require.Equal(t,
		diagnostics[0].Kind,
		&analysis.InvalidUnboundedAccount{},
	)

	require.Equal(t,
		diagnostics[0].Range,
		parser.RangeOfIndexed(input, "@illegal", 0),
	)
}

func TestAllowUnboundedSrcInSendAllWhenCapped(t *testing.T) {
	t.Parallel()

	input := `
	send [GEM *] (
		source = max [GEM 100] from {
			@ok
			@illegal allowing unbounded overdraft
		}
		destination = @b
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Empty(t, diagnostics)
}

func TestForbidWorldSrcInSendAll(t *testing.T) {
	t.Parallel()

	input := `
	send [EUR/2 *] (
		source = @world
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Len(t, diagnostics, 1)
}

func TestForbidEmptiedAccount(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 100] (
		source = {
			@a
			@b
			@a // <- err
		}
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Len(t, diagnostics, 1)

	require.Equal(t,
		diagnostics[0].Kind,
		&analysis.EmptiedAccount{Name: "a"},
	)

	require.Equal(t,
		diagnostics[0].Range,
		parser.RangeOfIndexed(input, "@a", 1),
	)
}

func TestResetEmptiedAccount(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 100] (
		source = @a
		destination = @dest
	)

	send [COIN 100] (
		source = @a
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Empty(t, diagnostics)
}

func TestEmptiedAccountInMax(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 100] (
		source = {
			@emptied
			max [COIN 10] from {
				@a
				@emptied // <- err
				@b
			}
			@c
		}
		destination = @b
	)

	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Len(t, diagnostics, 1)

	require.Equal(t,
		diagnostics[0].Kind,
		&analysis.EmptiedAccount{Name: "emptied"},
	)

	require.Equal(t,
		diagnostics[0].Range,
		parser.RangeOfIndexed(input, "@emptied", 1),
	)
}

func TestEmptiedAccountDoNotLeakMaxed(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 100] (
		source = {
			max [COIN 10] from @emptied
			@emptied
		}
		destination = @b
	)

	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Empty(t, diagnostics)
}

func TestDoNotEmptyAccountInMax(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 100] (
		source = {
			@a
			max [COIN 10] from {
				@a1
				@emptied
				@b1
				@emptied  // <- err
			}
		}
		destination = @b
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Len(t, diagnostics, 1)

	require.Equal(t,
		diagnostics[0].Kind,
		&analysis.EmptiedAccount{Name: "emptied"},
	)

	require.Equal(t,
		diagnostics[0].Range,
		parser.RangeOfIndexed(input, "@emptied", 1),
	)
}

func TestDoNotEmitEmptiedAccountOnAllotment(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 100] (
		source = {
			1/2 from @emptied
			1/2 from @emptied
		}
		destination = @b
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Empty(t, diagnostics)
}

func TestDoNotAllowExprAfterWorld(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 100] (
		source = {
			@world
			@another
		}
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Len(t, diagnostics, 1)

	require.Equal(t,
		diagnostics[0].Kind,
		&analysis.UnboundedAccountIsNotLast{},
	)

	require.Equal(t,
		diagnostics[0].Range,
		parser.RangeOfIndexed(input, "@another", 0),
	)
}

func TestAllowWorldInNextExpr(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 1] (
		source = @world
		destination = @dest
	)

	send [COIN 1] (
		source = @world
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Empty(t, diagnostics)

}

func TestAllowWorldInMaxedExpr(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 10] (
		source = {
			max [COIN 1] from @world
			@x
		}
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Empty(t, diagnostics)

}

func TestDoNotAllowExprAfterWorldInsideMaxed(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 10] (
		source = max [COIN 1] from {
			@world
			@x
		}
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Len(t, diagnostics, 1)

	require.Equal(t,
		diagnostics[0].Kind,
		&analysis.UnboundedAccountIsNotLast{},
	)

	require.Equal(t,
		diagnostics[0].Range,
		parser.RangeOfIndexed(input, "@x", 0),
	)
}

func TestDoNotAllowExprAfterUnbounded(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 100] (
		source = {
			@unbounded allowing unbounded overdraft
			@another
		}
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Len(t, diagnostics, 1)

	require.Equal(t,
		diagnostics[0].Kind,
		&analysis.UnboundedAccountIsNotLast{},
	)

	require.Equal(t,
		diagnostics[0].Range,
		parser.RangeOfIndexed(input, "@another", 0),
	)
}

func TestAllowExprAfterBoundedOverdraft(t *testing.T) {
	t.Parallel()

	input := `
	send [COIN 100] (
		source = {
			@unbounded allowing overdraft up to [COIN 10]
			@another
		}
		destination = @dest
	)
	`

	diagnostics := analysis.CheckSource(input).Diagnostics
	require.Empty(t, diagnostics)
}

func TestCheckPlus(t *testing.T) {
	t.Parallel()

	t.Run("error in number+portion", func(t *testing.T) {
		input := `set_tx_meta("k", 1 + 1/2)`

		diagnostics := analysis.CheckSource(input).Diagnostics
		require.Equal(t, []analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "1/2", 0),
				Kind: &analysis.TypeMismatch{
					Expected: "number",
					Got:      "portion",
				},
			},
		}, diagnostics)
	})

	t.Run("allow number+number", func(t *testing.T) {
		input := `set_tx_meta("k", 1 + 2)`

		diagnostics := analysis.CheckSource(input).Diagnostics
		require.Empty(t, diagnostics)
	})

	t.Run("allow monetary+monetary", func(t *testing.T) {
		input := `set_tx_meta("k", [EUR/2 10] + [EUR/2 20])`

		diagnostics := analysis.CheckSource(input).Diagnostics
		require.Empty(t, diagnostics)
	})

	t.Run("error when left side is invalid", func(t *testing.T) {
		input := `set_tx_meta("k", @acc + @acc)`

		diagnostics := analysis.CheckSource(input).Diagnostics
		require.Equal(t, []analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "@acc", 0),
				Kind: &analysis.TypeMismatch{
					Expected: "number|monetary",
					Got:      "account",
				},
			},
		}, diagnostics)
	})

	t.Run("no type error when left side is any", func(t *testing.T) {
		input := `set_tx_meta("k", $unbound_var + @acc)`

		diagnostics := analysis.CheckSource(input).Diagnostics
		require.Equal(t, []analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$unbound_var", 0),
				Kind: &analysis.UnboundVariable{
					Name: "unbound_var",
				},
			},
		}, diagnostics)
	})
}

func TestCheckMinus(t *testing.T) {
	t.Parallel()

	t.Run("error in number-portion", func(t *testing.T) {
		input := `set_tx_meta("k", 1 - 1/2)`

		diagnostics := analysis.CheckSource(input).Diagnostics
		require.Equal(t, []analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "1/2", 0),
				Kind: &analysis.TypeMismatch{
					Expected: "number",
					Got:      "portion",
				},
			},
		}, diagnostics)
	})

	t.Run("allow number-number", func(t *testing.T) {
		input := `set_tx_meta("k", 1 - 2)`

		diagnostics := analysis.CheckSource(input).Diagnostics
		require.Empty(t, diagnostics)
	})

	t.Run("allow monetary-monetary", func(t *testing.T) {
		input := `set_tx_meta("k", [EUR/2 10] - [EUR/2 20])`

		diagnostics := analysis.CheckSource(input).Diagnostics
		require.Empty(t, diagnostics)
	})

	t.Run("error when left side is invalid", func(t *testing.T) {
		input := `set_tx_meta("k", @acc - @acc)`

		diagnostics := analysis.CheckSource(input).Diagnostics
		require.Equal(t, []analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "@acc", 0),
				Kind: &analysis.TypeMismatch{
					Expected: "number|monetary",
					Got:      "account",
				},
			},
		}, diagnostics)
	})

	t.Run("no type error when left side is any", func(t *testing.T) {
		input := `set_tx_meta("k", $unbound_var - @acc)`

		diagnostics := analysis.CheckSource(input).Diagnostics
		require.Equal(t, []analysis.Diagnostic{
			{
				Range: parser.RangeOfIndexed(input, "$unbound_var", 0),
				Kind: &analysis.UnboundVariable{
					Name: "unbound_var",
				},
			},
		}, diagnostics)
	})
}
