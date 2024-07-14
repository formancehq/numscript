package analysis_test

import (
	"math/big"
	"numscript/analysis"
	"numscript/parser"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidType(t *testing.T) {
	input := `vars { invalid $my_var }
send [C 10] (
	source = $my_var
	destination = $my_var
)`
	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Lenf(t, diagnostics, 1, "xs: %#v", diagnostics)

	d1 := diagnostics[0]
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
	input := `vars { account $my_var }
send [C 10] (
	source = $my_var
	destination = $my_var
)`
	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 0)
}

func TestDuplicateVariable(t *testing.T) {
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

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

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

func TestUnboundVarInSource(t *testing.T) {
	input := `send [C 1] (
  source = { max [C 1] from $unbound_var }
  destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

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

func TestUnboundVarInDest(t *testing.T) {
	input := `send [C 1] (
  source = @src
  destination = {
	1/2 to @a
	1/2 to $unbound_var
  }
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Lenf(t, diagnostics, 1, "expected len to be 1")

	d1 := diagnostics[0]
	assert.Equal(t,
		RangeOfIndexed(input, "$unbound_var", 0),
		d1.Range,
	)
	assert.Equal(t,
		&analysis.UnboundVariable{Name: "unbound_var"},
		d1.Kind,
	)
}

func TestUnboundMany(t *testing.T) {
	input := `send [C 1] (
  	source = {
  		1/3 from $unbound1
		2/3 from $unbound2
	}
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 2)
}

func TestUnboundCurrenciesVars(t *testing.T) {
	input := `send $unbound1 (
  	source = {
		max $unbound2 from @a
	}
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 2)
}

// TODO unbound vars in declr

func TestUnusedVarInSource(t *testing.T) {
	input := `vars { monetary $unused_var }`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

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

func TestWrongTypeForCap(t *testing.T) {
	input := `vars { account $account }

send [COIN 100] (
  	source = max $account from @a
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "monetary",
			Got:      "account",
		},
		d1.Kind,
	)

	assert.Equal(t,
		RangeOfIndexed(input, "$account", 1),
		d1.Range,
	)
}

func TestWrongTypeForSrcAccount(t *testing.T) {
	input := `vars { portion $x }

send [COIN 100] (
  	source = $x
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "account",
			Got:      "portion",
		},
		d1.Kind,
	)

	assert.Equal(t,
		RangeOfIndexed(input, "$x", 1),
		d1.Range,
	)
}

func TestWrongTypeForDestAccount(t *testing.T) {
	input := `vars { portion $x }

send [COIN 100] (
  	source = @src
  	destination = $x
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "account",
			Got:      "portion",
		},
		d1.Kind,
	)

	assert.Equal(t,
		RangeOfIndexed(input, "$x", 1),
		d1.Range,
	)
}

func TestWrongTypeForUnboundedAccount(t *testing.T) {
	input := `vars { portion $x }

send [COIN 100] (
  	source = $x allowing unbounded overdraft
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "account",
			Got:      "portion",
		},
		d1.Kind,
	)

	assert.Equal(t,
		RangeOfIndexed(input, "$x", 1),
		d1.Range,
	)
}

func TestWrongTypeForBoundedOverdraftCap(t *testing.T) {
	input := `vars { portion $x }

send [COIN 100] (
  	source = @x allowing overdraft up to $x
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "monetary",
			Got:      "portion",
		},
		d1.Kind,
	)

	assert.Equal(t,
		RangeOfIndexed(input, "$x", 1),
		d1.Range,
	)
}

func TestWrongTypeForSrcAllotmentPortion(t *testing.T) {
	input := `vars { string $p }

send [COIN 100] (
  	source = {
		$p from @a
		remaining from @b
	}
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "portion",
			Got:      "string",
		},
		d1.Kind,
	)

	assert.Equal(t,
		RangeOfIndexed(input, "$p", 1),
		d1.Range,
	)
}

func TestWrongTypeForDestAllotmentPortion(t *testing.T) {
	input := `vars { string $p }

send [COIN 100] (
  	source = @s
  	destination = {
		$p to @a
		remaining to @dest
	}
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.TypeMismatch{
			Expected: "portion",
			Got:      "string",
		},
		d1.Kind,
	)

	assert.Equal(t,
		RangeOfIndexed(input, "$p", 1),
		d1.Range,
	)
}

func TestBadRemainingInSource(t *testing.T) {
	input := `send [COIN 100] (
  	source = {
		1/2 from @a
		remaining from @b
		1/2 from @c
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

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
	input := `send [COIN 100] (
  	source = @a
  	destination = {
		1/2 from @a
		remaining from @b
		1/2 from @c
    }
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Len(t, diagnostics, 1)

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
	input := `send [COIN 100] (
   source = {
		1/3 from @s1
		1/3 from @s2
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Lenf(t, diagnostics, 1, "wrong diagnostics len")

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
	input := `send [COIN 100] (
   source = {
		2/3 from @s1
		2/3 from @s2
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Lenf(t, diagnostics, 1, "wrong diagnostics len")

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
	input := `send [COIN 100] (
   source = @src
  	destination = {
		1/3 to @d1
		1/3 to @d2
    }
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Lenf(t, diagnostics, 1, "wrong diagnostics len")

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.BadAllotmentSum{
			Sum: *big.NewRat(2, 3),
		},
		d1.Kind,
	)
}

func TestNoBadAllotmentWhenRemaining(t *testing.T) {
	input := `send [COIN 100] (
   source = {
		1/3 from @s1
		1/3 from @s2
		remaining from @s3
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Lenf(t, diagnostics, 0, "wrong diagnostics len")
}

func TestBadAllotmentWhenRemainingButGt1(t *testing.T) {
	input := `send [COIN 100] (
   source = {
		2/3 from @s1
		2/3 from @s2
		remaining from @s3
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Lenf(t, diagnostics, 1, "wrong diagnostics len")

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.BadAllotmentSum{
			Sum: *big.NewRat(4, 3),
		},
		d1.Kind,
	)
}

func TestRedundantRemainingWhenSumIsOne(t *testing.T) {
	input := `send [COIN 100] (
   source = {
		2/3 from @s1
		1/3 from @s2
		remaining from @s3
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Lenf(t, diagnostics, 1, "wrong diagnostics len")

	d1 := diagnostics[0]
	assert.Equal(t,
		&analysis.RedundantRemaining{},
		d1.Kind,
	)

	assert.Equal(t,
		RangeOfIndexed(input, "remaining", 0),
		d1.Range,
	)
}

func TestNoSingleAllotmentVariable(t *testing.T) {
	input := `vars { portion $allot }

send [COIN 100] (
   source = {
		$allot from @s1
		remaining from @s2
    }
  	destination = @dest
)`

	program := parser.Parse(input).Value

	diagnostics := analysis.Check(program).Diagnostics
	assert.Lenf(t, diagnostics, 0, "wrong diagnostics len")
}
