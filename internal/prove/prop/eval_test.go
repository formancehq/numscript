package prop

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEval(t *testing.T) {
	big20, ok := new(big.Int).SetString("100000000000000000000", 10)
	require.True(t, ok)
	values := map[string]*big.Int{
		"before(@a, USD/2)": big.NewInt(10),
		"after(@a, USD/2)":  big.NewInt(25),
		"sent(@a, USD/2)":   big.NewInt(5),
		"big(@a, USD/2)":    big20,
		"before(@b, USD/2)": big20,
	}
	lookup := func(a Atom) *big.Int {
		v, ok := values[a.String()]
		if !ok {
			panic("no value for " + a.String())
		}
		return v
	}
	cases := []struct {
		src     string
		success bool
		want    bool
	}{
		{"-5 / 2 == -3", true, true},
		{"-5 % 2 == 1", true, true},
		{"5 / 2 == 2", true, true},
		{"5 % 2 == 1", true, true},
		{"-7 / 3 == -3", true, true},
		{"-7 % 3 == 2", true, true},
		{"-6 / 3 == -2 and -6 % 3 == 0", true, true},
		{"min(3, 1, 2) == 1", true, true},
		{"max(3, 1, 2) == 3", true, true},
		{"min(-1, 0) == -1 and max(-1, 0) == 0", true, true},
		{"false => false", true, true},
		{"false => true", true, true},
		{"true => true", true, true},
		{"true => false", true, false},
		{"fails", true, false},
		{"fails", false, true},
		{"fails <=> not success", true, true},
		{"fails <=> not success", false, true},
		{"success <=> fails", true, false},
		{"true and false or true", true, true},
		{"not 1 == 2", true, true},
		{"1 != 2 and 1 < 2 and 1 <= 1 and 2 > 1 and 2 >= 2", true, true},
		{"delta(@a, USD/2) == 15", true, true},
		{"not sent(@a, USD/2) == 0", true, true},
		{"before(@b, USD/2) * 2 == 200000000000000000000", true, true},
		{"before(@b, USD/2) > 9223372036854775807", true, true},
		{"before(@b, USD/2) / 3 == 33333333333333333333", true, true},
		{"before(@b, USD/2) % 7 == 100000000000000000000 % 7", true, true},
		{"-before(@b, USD/2) < -9223372036854775808", true, true},
		{"after(@a, USD/2) - before(@a, USD/2) == sent(@a, USD/2) * 3", true, true},
		{"1_000 == 1000", true, true},
	}
	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			e := mustParse(t, c.src)
			require.Empty(t, Typecheck(e, Known{}))
			require.Equal(t, c.want, Eval(e, lookup, c.success))
		})
	}
}

func TestEvalDoesNotMutateLookupValues(t *testing.T) {
	v := big.NewInt(10)
	lookup := func(Atom) *big.Int { return v }
	e := mustParse(t, "-sent(@a, USD/2) + sent(@a, USD/2) * 2 - sent(@a, USD/2) / 3 % 2 == 9")
	require.True(t, Eval(e, lookup, true))
	require.Equal(t, int64(10), v.Int64())
}
