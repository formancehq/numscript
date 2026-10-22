package prop

import (
	"fmt"
	"testing"

	"github.com/formancehq/numscript/internal/prove/smt"
	"github.com/stretchr/testify/require"
)

func TestLower(t *testing.T) {
	resolve := func(a Atom) smt.Term {
		if a.Kind == Moved {
			return smt.Symbol(fmt.Sprintf("moved(@%s,@%s,%s)", a.Account, a.Account2, a.Asset), smt.SortInt)
		}
		return smt.Symbol(fmt.Sprintf("%s(@%s,%s)", a.Kind, a.Account, a.Asset), smt.SortInt)
	}
	success := smt.Symbol("success", smt.SortBool)
	cases := []struct{ src, want string }{
		{"success => after(@src, USD/2) == before(@src, USD/2) - 10", "(=> success (= |after(@src,USD/2)| (- |before(@src,USD/2)| 10)))"},
		{"fails", "(not success)"},
		{"true", "true"},
		{"delta(@a, USD/2) >= 0", "(<= 0 (- |after(@a,USD/2)| |before(@a,USD/2)|))"},
		{"sent(@a, USD/2) * 2 <= 10", "(<= (* 2 |sent(@a,USD/2)|) 10)"},
		{"2 * 3 * sent(@a, USD/2) == 6", "(= (* 6 |sent(@a,USD/2)|) 6)"},
		{"before(@a, USD/2) / 2 % 3 == min(1, sent(@a, USD/2))", "(= (mod (div |before(@a,USD/2)| 2) 3) (ite (< 1 |sent(@a,USD/2)|) 1 |sent(@a,USD/2)|))"},
		{"not success and true", "(not success)"},
		{"-moved(@a, @b, USD/2) < 0 or fails", "(or (< (- |moved(@a,@b,USD/2)|) 0) (not success))"},
		{"max(sent(@a, USD/2), 0) != 1", "(not (= (ite (< |sent(@a,USD/2)| 0) 0 |sent(@a,USD/2)|) 1))"},
		{"received(@a, USD/2) > 1 <=> success", "(= (< 1 |received(@a,USD/2)|) success)"},
		{"-5 / 2 == -3 and -5 % 2 == 1", "true"},
		{"sent(@a, USD/2) - 0 == sent(@a, USD/2) * 1", "true"},
		{"min(sent(@a, USD/2), received(@a, USD/2), 7) == 7", "(= (ite (< (ite (< |sent(@a,USD/2)| |received(@a,USD/2)|) |sent(@a,USD/2)| |received(@a,USD/2)|) 7) (ite (< |sent(@a,USD/2)| |received(@a,USD/2)|) |sent(@a,USD/2)| |received(@a,USD/2)|) 7) 7)"},
	}
	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			e := mustParse(t, c.src)
			require.Empty(t, Typecheck(e, Known{}))
			require.Equal(t, c.want, smt.Print(Lower(e, resolve, success)))
		})
	}
}
