package interpreter

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestParseEmptyArgsValidArity(t *testing.T) {
	t.Parallel()

	p := NewArgsParser([]Value{})
	err := p.parse()
	require.Nil(t, err)
}

func TestParseInvalidArity(t *testing.T) {
	t.Parallel()

	p := NewArgsParser([]Value{})
	parseArg(p, parser.Range{}, expectAccount)
	parseArg(p, parser.Range{}, expectAsset)

	err := p.parse()

	require.Equal(t, err, BadArityErr{
		ExpectedArity:  2,
		GivenArguments: 0,
	})
}

func TestParseValid(t *testing.T) {
	t.Parallel()

	p := NewArgsParser([]Value{
		NewMonetaryInt(42),
		Account{Repr: AccountAddress("user:001")},
	})
	a1 := parseArg(p, parser.Range{}, expectNumber)
	a2 := parseArg(p, parser.Range{}, expectAccount)
	err := p.parse()

	require.Nil(t, err)

	require.NotNil(t, a1, "a1 should not be nil")
	require.NotNil(t, a2, "a2 should not be nil")

	require.Equal(t, *big.NewInt(42), *a1)
	require.Equal(t, Account{AccountAddress("user:001")}, *a2)
}

func TestParseBadType(t *testing.T) {
	t.Parallel()

	p := NewArgsParser([]Value{
		NewMonetaryInt(42),
		Account{Repr: AccountAddress("user:001")},
	})
	parseArg(p, parser.Range{}, expectMonetary)
	parseArg(p, parser.Range{}, expectAccount)
	err := p.parse()

	require.Equal(t, err, TypeError{
		Expected: "monetary",
		Value:    NewMonetaryInt(42),
	})
}
