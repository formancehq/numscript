package prop

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypecheckErrors(t *testing.T) {
	cases := []struct {
		src  string
		want []string
	}{
		{"success", nil},
		{"sent(@a, USD/2) * 2 + 1 % 2 >= -5", nil},
		{"2 * 3 * sent(@a, USD/2) == 0", nil},
		{"sent(@a, USD/2) * (10 - 3) == 0", nil},
		{"sent(@a, USD/2) % (1 + 1) == 0", nil},
		{"sent(@a, USD/2) / min(2, 3) == 0", nil},
		{"success + 1", []string{
			"1:1: expected an integer expression, found a boolean",
			"1:1: expected a boolean expression, found an integer",
		}},
		{"1 and success", []string{"1:1: expected a boolean expression, found an integer"}},
		{"not 1", []string{"1:5: expected a boolean expression, found an integer"}},
		{"min(1, success) == 1", []string{"1:8: expected an integer expression, found a boolean"}},
		{"-success == 1", []string{"1:2: expected an integer expression, found a boolean"}},
		{"success == true", []string{
			"1:1: expected an integer expression, found a boolean",
			"1:12: expected an integer expression, found a boolean",
		}},
		{"1", []string{"1:1: expected a boolean expression, found an integer"}},
		{"before(@a, USD/2) * sent(@a, USD/2) == 0", []string{"1:1: multiplication must have a literal operand (linear arithmetic only)"}},
		{"sent(@a, USD/2) / 0 == 0", []string{"1:19: division must be by a positive literal"}},
		{"sent(@a, USD/2) / -2 == 0", []string{"1:19: division must be by a positive literal"}},
		{"sent(@a, USD/2) / sent(@b, USD/2) == 0", []string{"1:19: division must be by a positive literal"}},
		{"sent(@a, USD/2) % 0 == 0", []string{"1:19: modulo must be by a positive literal"}},
		{"(1 / 0) * sent(@a, USD/2) == 0", []string{
			"1:6: division must be by a positive literal",
			"1:2: multiplication must have a literal operand (linear arithmetic only)",
		}},
		{"success and 1 or 2", []string{
			"1:13: expected a boolean expression, found an integer",
			"1:18: expected a boolean expression, found an integer",
		}},
	}
	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			errs := Typecheck(mustParse(t, c.src), Known{})
			require.Equal(t, c.want, errStrings(errs))
		})
	}
}

func TestTypecheckKnown(t *testing.T) {
	known := Known{
		Accounts: map[string]struct{}{"src": {}, "dest": {}, "world": {}},
		Assets:   map[string]struct{}{"USD/2": {}},
	}
	cases := []struct {
		src  string
		want []string
	}{
		{"before(@src, USD/2) == 0", nil},
		{"moved(@src, @dest, USD/2) == delta(@world, USD/2)", nil},
		{"before(@dset, USD/2) == 0", []string{"1:1: account @dset is not referenced by the script (known: @dest, @src, @world)"}},
		{"before(@src, EUR/2) == 0", []string{"1:1: asset EUR/2 is not referenced by the script (known: USD/2)"}},
		{"moved(@src, @nope, USD/2) == 0", []string{"1:1: account @nope is not referenced by the script (known: @dest, @src, @world)"}},
		{"0 == delta(@x, EUR/2)", []string{
			"1:6: account @x is not referenced by the script (known: @dest, @src, @world)",
			"1:6: asset EUR/2 is not referenced by the script (known: USD/2)",
		}},
	}
	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			errs := Typecheck(mustParse(t, c.src), known)
			require.Equal(t, c.want, errStrings(errs))
		})
	}

	e := mustParse(t, "before(@nobody, XYZ) == 0")
	require.Empty(t, Typecheck(e, Known{}))
	require.Equal(t, []string{
		"1:1: account @nobody is not referenced by the script (known: none)",
		"1:1: asset XYZ is not referenced by the script (known: none)",
	}, errStrings(Typecheck(e, Known{Accounts: map[string]struct{}{}, Assets: map[string]struct{}{}})))
}

func TestAtoms(t *testing.T) {
	e := mustParse(t, "delta(@b, USD/2) + sent(@a, USD/2) == moved(@a, @b, USD/2) + sent(@a, USD/2) - min(received(@a, EUR/2), 1)")
	require.Equal(t, []Atom{
		{Kind: After, Account: "b", Asset: "USD/2"},
		{Kind: Before, Account: "b", Asset: "USD/2"},
		{Kind: Moved, Account: "a", Account2: "b", Asset: "USD/2"},
		{Kind: Received, Account: "a", Asset: "EUR/2"},
		{Kind: Sent, Account: "a", Asset: "USD/2"},
	}, Atoms(e))
	require.Empty(t, Atoms(mustParse(t, "success")))
}

func TestAtomString(t *testing.T) {
	require.Equal(t, "before(@src, USD/2)", Atom{Kind: Before, Account: "src", Asset: "USD/2"}.String())
	require.Equal(t, "moved(@a, @b, USD/2)", Atom{Kind: Moved, Account: "a", Account2: "b", Asset: "USD/2"}.String())
	require.Equal(t, "received", Received.String())
}
