package prop

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

var roundTripFormulas = []string{
	"success => after(@src, USD/2) == before(@src, USD/2) - 10",
	"not sent(@a, USD/2) == 0",
	"not (success and fails)",
	"(success => fails) => true",
	"success => fails => true",
	"(success <=> fails) <=> true",
	"before(@a, USD/2) - (before(@b, USD/2) - 1) < 3",
	"before(@a, USD/2) * 2 + 1 % 2 >= -5",
	"-(before(@a, USD/2) + 1) != --3",
	"min(1, max(2, 3), delta(@a, EUR/2)) <= moved(@a, @b:001, COIN)",
	"success and not fails or true and false",
	"received(@users:001, USD/2) / 2 * 3 == 1000000",
	"success or fails and (true or false)",
	"1 - -1 == 2 and not not success",
}

func TestDumpRoundTrip(t *testing.T) {
	for _, src := range roundTripFormulas {
		t.Run(src, func(t *testing.T) {
			e := mustParse(t, src)
			require.Equal(t, src, Dump(e))
			again := mustParse(t, Dump(e))
			require.True(t, sameExpr(e, again), "reparse of %q differs: %s vs %s", src, tree(e), tree(again))
		})
	}
}

func TestDumpParenthesizes(t *testing.T) {
	a, b, c := &Success{}, &Fails{}, &BoolLit{Value: true}
	x := &AtomExpr{Atom: Atom{Kind: Sent, Account: "a", Asset: "USD/2"}}
	one, two := &IntLit{Value: big.NewInt(1)}, &IntLit{Value: big.NewInt(2)}
	cases := []struct {
		e    Expr
		want string
	}{
		{&Binary{Op: And, L: a, R: &Binary{Op: And, L: b, R: c}}, "success and (fails and true)"},
		{&Binary{Op: And, L: &Binary{Op: And, L: a, R: b}, R: c}, "success and fails and true"},
		{&Binary{Op: Implies, L: &Binary{Op: Implies, L: a, R: b}, R: c}, "(success => fails) => true"},
		{&Binary{Op: Implies, L: a, R: &Binary{Op: Implies, L: b, R: c}}, "success => fails => true"},
		{&Binary{Op: Sub, L: x, R: &Binary{Op: Sub, L: one, R: two}}, "sent(@a, USD/2) - (1 - 2)"},
		{&Binary{Op: Sub, L: &Binary{Op: Sub, L: x, R: one}, R: two}, "sent(@a, USD/2) - 1 - 2"},
		{&Unary{Op: Neg, X: &Binary{Op: Mul, L: x, R: two}}, "-(sent(@a, USD/2) * 2)"},
		{&Binary{Op: Mul, L: &Unary{Op: Neg, X: x}, R: two}, "-sent(@a, USD/2) * 2"},
		{&Unary{Op: Not, X: &Binary{Op: And, L: a, R: b}}, "not (success and fails)"},
		{&Unary{Op: Not, X: &Binary{Op: Eq, L: x, R: one}}, "not sent(@a, USD/2) == 1"},
		{&Binary{Op: And, L: &Unary{Op: Not, X: a}, R: b}, "not success and fails"},
		{&Binary{Op: Lt, L: &Binary{Op: Add, L: x, R: one}, R: &Unary{Op: Neg, X: two}}, "sent(@a, USD/2) + 1 < -2"},
		{&MinMax{IsMax: true, Args: []Expr{x, &Binary{Op: Add, L: one, R: two}}}, "max(sent(@a, USD/2), 1 + 2)"},
		{&Binary{Op: Eq, L: &Delta{Account: "a", Asset: "USD/2"}, R: &IntLit{Value: big.NewInt(-3)}}, "delta(@a, USD/2) == -3"},
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			require.Equal(t, c.want, Dump(c.e))
			require.Equal(t, c.want, Dump(mustParse(t, c.want)))
		})
	}
}

func FuzzParse(f *testing.F) {
	for _, src := range roundTripFormulas {
		f.Add(src)
	}
	f.Add("name: success")
	f.Add("$x == [USD/2 100]")
	f.Add("1 < 2 < 3")
	f.Add("before(@a:, USD")
	f.Add("/* unterminated")
	f.Add("--1 - -1")
	f.Fuzz(func(t *testing.T, src string) {
		p, errs := ParseInline(src)
		if len(errs) > 0 {
			return
		}
		d := Dump(p.Formula)
		again, errs := ParseInline(d)
		if len(errs) > 0 {
			t.Fatalf("Dump of %q gives %q, which does not parse: %v", src, d, errs)
		}
		if Dump(again.Formula) != d {
			t.Fatalf("Dump of %q is %q, reparsed as %q", src, d, Dump(again.Formula))
		}
	})
}
