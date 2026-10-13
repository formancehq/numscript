package prop

import (
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestParsePrecedence(t *testing.T) {
	cases := []struct{ src, want string }{
		{"success or true and false", "(or success (and true false))"},
		{"not sent(@a, USD/2) == 0", "(not (== sent(@a, USD/2) 0))"},
		{"success => true => false", "(=> success (=> true false))"},
		{"(success <=> true) <=> false", "(<=> (<=> success true) false)"},
		{"success => true <=> false", "(<=> (=> success true) false)"},
		{"1 + 2 * 3", "(+ 1 (* 2 3))"},
		{"1 - 2 - 3", "(- (- 1 2) 3)"},
		{"-1 * 2", "(* (- 1) 2)"},
		{"success and not fails or true", "(or (and success (not fails)) true)"},
		{"not not success", "(not (not success))"},
		{"--1 == 1", "(== (- (- 1)) 1)"},
		{"1_000_000 == 1000000", "(== 1000000 1000000)"},
		{"before(@a, USD/2) / 2 % 3 == 0", "(== (% (/ before(@a, USD/2) 2) 3) 0)"},
		{"min(1, 2, 3) == max(1, 2)", "(== (min 1 2 3) (max 1 2))"},
		{"not 1 + 2 < 3 and fails", "(and (not (< (+ 1 2) 3)) fails)"},
		{"1 < 2 and 2 < 3 => success", "(=> (and (< 1 2) (< 2 3)) success)"},
		{"moved(@a, @b:001, USD/2) == delta(@users:x_1, COIN)", "(== moved(@a, @b:001, USD/2) delta(@users:x_1, COIN))"},
		{"after(@a-b, USD) != received(@A1, USD/2)", "(!= after(@a-b, USD) received(@A1, USD/2))"},
		{"1 >= 2 or 3 <= 4 or 5 > 6", "(or (or (>= 1 2) (<= 3 4)) (> 5 6))"},
		{"before(@a, USD/2) / 2 == 1", "(== (/ before(@a, USD/2) 2) 1)"},
		{"/* c */ success // trailing\n and fails", "(and success fails)"},
	}
	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			require.Equal(t, c.want, tree(mustParse(t, c.src)))
		})
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct{ src, want string }{
		{"$x == 0", "1:1: script variables are not supported in properties yet"},
		{`"a" == 0`, "1:1: strings are not supported in properties"},
		{"[USD/2 100] == 0", "1:1: monetary literals are not allowed in properties; amounts are plain integers in the asset's minor unit: write 100"},
		{"forall a in accounts", "1:1: quantifiers are not supported yet"},
		{"success and exists a", "1:13: quantifiers are not supported yet"},
		{"sum(1, 2) == 0", "1:1: sums are not supported yet"},
		{`meta(@a, "k") == 0`, "1:1: metadata is not supported yet"},
		{"before(@a, USD/2) = 0", "1:19: unexpected character '='"},
		{"before(@a, USD/2) == 0 and", "1:27: unexpected end of input"},
		{"(1 + 2", "1:7: expected ')' but found end of input"},
		{"before(@a USD/2)", "1:11: expected ',' but found 'USD/2'"},
		{"before(a, USD/2)", "1:8: expected an account but found 'a'"},
		{"before(@a, usd)", "1:12: expected an asset but found 'usd'"},
		{"1 2", "1:3: expected end of input but found '2'"},
		{"/* x", "1:1: unterminated block comment"},
		{`"abc`, "1:1: unterminated string"},
		{"@", "1:1: expected an account name after '@'"},
		{"before(@a:, USD/2)", "1:8: expected an account segment after ':'"},
		{"min(1)", "1:1: min takes at least two arguments"},
		{"max(1 < 2, 3)", "1:7: expected ',' or ')' but found '<'"},
		{"foo", "1:1: unknown name 'foo'"},
		{"and", "1:1: expected an expression but found 'and'"},
		{"prop a: success", "1:1: expected an expression but found 'prop'"},
		{"success\n  and $x", "2:7: script variables are not supported in properties yet"},
		{"success <=> fails <=> true", "1:19: '<=>' does not chain: write (a <=> b) <=> c"},
		{"1 < 2 < 3", "1:7: comparisons do not chain: write (a < b) and (b < c)"},
		{"1 == 2 != 3", "1:8: comparisons do not chain: write (a < b) and (b < c)"},
		{"success ?", "1:9: unexpected character '?'"},
		{"x:", "1:3: unexpected end of input"},
		{"1 +", "1:4: unexpected end of input"},
		{"", "1:1: unexpected end of input"},
		{"(success))", "1:10: expected end of input but found ')'"},
	}
	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			_, errs := ParseInline(c.src)
			require.Equal(t, []string{c.want}, errStrings(errs))
		})
	}
}

func TestParseInlineName(t *testing.T) {
	p, errs := ParseInline("safe: success")
	require.Empty(t, errs)
	require.Equal(t, "safe", p.Name)
	require.Equal(t, "success", Dump(p.Formula))
	require.Equal(t, parser.Range{Start: parser.Position{}, End: parser.Position{Character: 13}}, p.Range)

	p, errs = ParseInline("success => fails")
	require.Empty(t, errs)
	require.Equal(t, "", p.Name)
	require.Equal(t, parser.Range{Start: parser.Position{}, End: parser.Position{Character: 16}}, p.Range)
	require.Equal(t, p.Range, p.Formula.GetRange())
}

func TestParseFile(t *testing.T) {
	src := `// safety
prop conserve: // trailing
  success => after(@src, USD/2) == before(@src, USD/2) - 10
/* block
   comment */
prop fee_bound: sent(@src, USD/2) <= 100 and received(@fees, USD/2) >= 0`
	props, errs := ParseFile(src)
	require.Empty(t, errs)
	require.Len(t, props, 2)
	require.Equal(t, "conserve", props[0].Name)
	require.Equal(t, "success => after(@src, USD/2) == before(@src, USD/2) - 10", Dump(props[0].Formula))
	require.Equal(t, parser.Position{Line: 1}, props[0].Range.Start)
	require.Equal(t, parser.Position{Line: 2, Character: 59}, props[0].Range.End)
	require.Equal(t, "fee_bound", props[1].Name)
	require.Equal(t, "sent(@src, USD/2) <= 100 and received(@fees, USD/2) >= 0", Dump(props[1].Formula))
	require.Equal(t, parser.Position{Line: 5}, props[1].Range.Start)
}

func TestParseFileErrors(t *testing.T) {
	cases := []struct {
		src  string
		want []string
	}{
		{"", []string{"1:1: no properties found"}},
		{"// just a comment\n", []string{"1:1: no properties found"}},
		{"prop a: success\nprop a: fails", []string{`2:6: property "a" is defined twice`}},
		{"prop a: success\nprop b: fails\nprop a: true\nprop b: false", []string{`3:6: property "a" is defined twice`, `4:6: property "b" is defined twice`}},
		{"success", []string{"1:1: expected 'prop' but found 'success'"}},
		{"prop a: success fails", []string{"1:17: expected 'prop' but found 'fails'"}},
		{"prop : success", []string{"1:6: expected a property name but found ':'"}},
		{"prop a success", []string{"1:8: expected ':' but found 'success'"}},
		{"prop a:", []string{"1:8: unexpected end of input"}},
	}
	for _, c := range cases {
		t.Run(c.src, func(t *testing.T) {
			props, errs := ParseFile(c.src)
			require.Nil(t, props)
			require.Equal(t, c.want, errStrings(errs))
		})
	}
}

func TestErrorFormat(t *testing.T) {
	e := Error{Range: parser.Range{Start: parser.Position{Line: 2, Character: 4}}, Msg: "boom"}
	require.Equal(t, "3:5: boom", e.Error())
}
