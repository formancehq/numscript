package analysis_test

import (
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

// The cursor position is marked with `|`
func completeAtCursor(t *testing.T, src string) []analysis.CompletionItem {
	t.Helper()
	pos := parser.PositionOf(src, "|")
	require.NotNil(t, pos, "missing cursor marker")
	return analysis.Complete(strings.Replace(src, "|", "", 1), *pos)
}

func labels(items []analysis.CompletionItem) []string {
	var out []string
	for _, item := range items {
		out = append(out, item.Label)
	}
	return out
}

const varsBlock = `vars {
  monetary $mon
  account $acc
  portion $por
  asset $ast
}
`

func TestCompleteEmptyProgram(t *testing.T) {
	got := labels(completeAtCursor(t, "|"))
	require.Subset(t, got, []string{"vars", "send", "save", "set_tx_meta", "set_account_meta"})
}

func TestCompleteNoVarsAfterVarsBlock(t *testing.T) {
	got := labels(completeAtCursor(t, varsBlock+"|"))
	require.Subset(t, got, []string{"send", "save", "set_tx_meta"})
	require.NotContains(t, got, "vars")
}

func TestCompletePartialKeyword(t *testing.T) {
	items := completeAtCursor(t, "send [COIN 10] (\n  sou|")
	require.Equal(t, []string{"source"}, labels(items))
	// the partial word is replaced
	require.Equal(t, parser.Range{
		Start: parser.Position{Line: 1, Character: 2},
		End:   parser.Position{Line: 1, Character: 5},
	}, items[0].Range)
}

func TestCompleteKeywordsAfterSourceAccount(t *testing.T) {
	got := labels(completeAtCursor(t, "send [COIN 10] (\n  source = @a |\n  destination = @b\n)"))
	require.ElementsMatch(t, []string{"allowing", "with", "destination"}, got)
}

func TestCompleteOverdraftKeywords(t *testing.T) {
	got := labels(completeAtCursor(t, "send [COIN 10] ( source = @a allowing | destination = @b )"))
	require.ElementsMatch(t, []string{"unbounded", "overdraft"}, got)
}

func TestCompleteVariablesByExpectedType(t *testing.T) {
	cases := map[string]string{
		"send $|":                 "$mon",
		"send $mon ( source = $|": "$acc",
		"send $mon ( source = @a destination = $|": "$acc",
		"send [$|": "$ast",
		"send $mon ( source = @a destination = { $| to @b } )": "$por",
		"save $mon from $|":               "$acc",
		`set_tx_meta("k", balance($|`:     "$acc",
		`set_tx_meta("k", balance(@a, $|`: "$ast",
	}
	for src, expected := range cases {
		t.Run(src, func(t *testing.T) {
			got := labels(completeAtCursor(t, varsBlock+src))
			require.Equal(t, []string{expected}, got)
		})
	}
}

func TestCompleteAnyVariableInUntypedPosition(t *testing.T) {
	got := labels(completeAtCursor(t, varsBlock+`set_tx_meta("k", $|`))
	require.ElementsMatch(t, []string{"$mon", "$acc", "$por", "$ast"}, got)
}

func TestCompleteFunctionsByReturnType(t *testing.T) {
	got := labels(completeAtCursor(t, "vars { monetary $m = | }"))
	require.Subset(t, got, []string{"balance", "overdraft", "meta"})
	require.NotContains(t, got, "get_amount")
	require.NotContains(t, got, "set_tx_meta")
}

func TestCompleteFunctionSnippet(t *testing.T) {
	items := completeAtCursor(t, "vars { monetary $m = bal| }")
	var balance *analysis.CompletionItem
	for _, item := range items {
		if item.Label == "balance" {
			balance = &item
		}
	}
	require.NotNil(t, balance)
	require.Equal(t, "balance(${1:account}, ${2:asset})", balance.SnippetText)
	require.Equal(t, "balance(account, asset) -> monetary", balance.Detail)
}

func TestCompleteVarTypes(t *testing.T) {
	got := labels(completeAtCursor(t, "vars {\n  mon|"))
	require.ElementsMatch(t, analysis.AllowedTypes, got)
}

func TestCompleteAccounts(t *testing.T) {
	items := completeAtCursor(t, "send [C 1] ( source = @users:alice destination = @us| )")
	require.ElementsMatch(t, []string{"@world", "@users:alice"}, labels(items))
	// the whole account literal is replaced
	require.Equal(t, parser.Range{
		Start: parser.Position{Character: 49},
		End:   parser.Position{Character: 52},
	}, items[0].Range)
}

func TestCompleteAccountAfterAt(t *testing.T) {
	got := labels(completeAtCursor(t, "send [C 1] ( source = @a destination = @| )"))
	require.ElementsMatch(t, []string{"@world", "@a"}, got)
}

func TestCompleteAccountInterpolation(t *testing.T) {
	got := labels(completeAtCursor(t, varsBlock+"send [C 1] ( source = @users:| )"))
	require.Contains(t, got, "@users:$acc")
}

func TestCompleteAssets(t *testing.T) {
	got := labels(completeAtCursor(t, "send [USD/2 1] ( source = @a destination = @b )\nsend [|"))
	require.Contains(t, got, "USD/2")
}

func TestCompleteSourcesInBraces(t *testing.T) {
	got := labels(completeAtCursor(t, "send [C 1] ( source = { @a\n | } destination = @b )"))
	require.Subset(t, got, []string{"max", "oneof", "@world", "@a"})
}

func TestNoCompletionInStringsAndComments(t *testing.T) {
	require.Empty(t, completeAtCursor(t, `set_tx_meta("a|")`))
	require.Empty(t, completeAtCursor(t, "// a comment|\n"))
}

func signatureAtCursor(t *testing.T, src string) *analysis.SignatureHelp {
	t.Helper()
	pos := parser.PositionOf(src, "|")
	require.NotNil(t, pos, "missing cursor marker")
	return analysis.SignatureHelpAt(strings.Replace(src, "|", "", 1), *pos)
}

func TestSignatureHelp(t *testing.T) {
	help := signatureAtCursor(t, `set_tx_meta("k", balance(@a, |))`)
	require.Equal(t, &analysis.SignatureHelp{
		Label:       "balance(account, asset) -> monetary",
		Params:      []string{"account", "asset"},
		ActiveParam: 1,
		Docs:        "fetch account balance",
	}, help)
}

func TestSignatureHelpInnermostCall(t *testing.T) {
	help := signatureAtCursor(t, `set_tx_meta("k", balance(@a, meta(@b, |)))`)
	require.Equal(t, "meta(account, string) -> any", help.Label)
	require.Equal(t, 1, help.ActiveParam)
}

func TestSignatureHelpIncompleteCall(t *testing.T) {
	help := signatureAtCursor(t, "vars { monetary $m = balance(@a, | }")
	require.Equal(t, 1, help.ActiveParam)

	help = signatureAtCursor(t, "set_tx_meta(|")
	require.Equal(t, "set_tx_meta(string, any)", help.Label)
	require.Equal(t, 0, help.ActiveParam)
}

func TestNoSignatureHelpOutsideCalls(t *testing.T) {
	require.Nil(t, signatureAtCursor(t, `set_tx_meta("k", 1)|`))
	require.Nil(t, signatureAtCursor(t, `unknown_fn(|`))
}
