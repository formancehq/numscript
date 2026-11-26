package wasmbinding_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/wasmbinding"
	"github.com/stretchr/testify/require"
)

// ---- Check --------------------------------------------------------------------

type checkEnvelope struct {
	Diagnostics []struct {
		Range struct {
			Start struct {
				Line      int `json:"line"`
				Character int `json:"character"`
			} `json:"start"`
			End struct {
				Line      int `json:"line"`
				Character int `json:"character"`
			} `json:"end"`
		} `json:"range"`
		Severity byte   `json:"severity"`
		Message  string `json:"message"`
		Id       int32  `json:"id"`
	} `json:"diagnostics"`
}

func TestCheckValidScriptHasNoDiagnostics(t *testing.T) {
	out := wasmbinding.Check(`send [COIN 10] (
	source = @world
	destination = @dest
)`)

	var env checkEnvelope
	require.NoError(t, json.Unmarshal(out, &env))
	require.Empty(t, env.Diagnostics)
}

func TestCheckUnboundVariableReportsErrorWithRange(t *testing.T) {
	out := wasmbinding.Check(`send [COIN 10] (
	source = $unbound
	destination = @dest
)`)

	var env checkEnvelope
	require.NoError(t, json.Unmarshal(out, &env))
	require.NotEmpty(t, env.Diagnostics)

	d := env.Diagnostics[0]
	require.EqualValues(t, 1, d.Severity, "unbound variable should be an Error")
	require.Contains(t, strings.ToLower(d.Message), "unbound")
	// The range must be non-degenerate and sit on the second line (0-based).
	require.Equal(t, 1, d.Range.Start.Line)
	require.Greater(t, d.Range.End.Character, d.Range.Start.Character)
}

// ---- Run ----------------------------------------------------------------------

type runEnvelope struct {
	Ok    bool `json:"ok"`
	Value *struct {
		Postings []struct {
			Source      string      `json:"source"`
			Destination string      `json:"destination"`
			Amount      json.Number `json:"amount"`
			Asset       string      `json:"asset"`
		} `json:"postings"`
		TxMeta       map[string]string `json:"txMeta"`
		AccountsMeta []struct {
			Account string `json:"account"`
			Key     string `json:"key"`
			Value   string `json:"value"`
		} `json:"accountsMeta"`
	} `json:"value"`
	Error string `json:"error"`
}

func TestRunSimpleSendFromWorld(t *testing.T) {
	out := wasmbinding.Run([]byte(`{
		"script": "send [COIN 10] (\n source = @world\n destination = @dest\n)",
		"variables": {},
		"balances": {},
		"metadata": {},
		"featureFlags": []
	}`))

	var env runEnvelope
	require.NoError(t, json.Unmarshal(out, &env))
	require.True(t, env.Ok, "unexpected error: %s", env.Error)
	require.NotNil(t, env.Value)
	require.Len(t, env.Value.Postings, 1)

	p := env.Value.Postings[0]
	require.Equal(t, "world", p.Source)
	require.Equal(t, "dest", p.Destination)
	require.Equal(t, "COIN", p.Asset)
	require.Equal(t, "10", p.Amount.String())

	// Non-nil JSON containers so the UI never sees `null`.
	require.NotNil(t, env.Value.TxMeta)
	require.NotNil(t, env.Value.AccountsMeta)
}

func TestRunPreservesLargeIntegerPrecision(t *testing.T) {
	// 10^21 overflows float64's exact-integer range (2^53); json.Number + big.Int
	// must carry it through losslessly.
	const huge = "1000000000000000000000"
	out := wasmbinding.Run([]byte(`{
		"script": "send [COIN ` + huge + `] (\n source = @alice\n destination = @dest\n)",
		"variables": {},
		"balances": { "alice": { "COIN": ` + huge + ` } },
		"metadata": {},
		"featureFlags": []
	}`))

	var env runEnvelope
	require.NoError(t, json.Unmarshal(out, &env))
	require.True(t, env.Ok, "unexpected error: %s", env.Error)
	require.Len(t, env.Value.Postings, 1)
	require.Equal(t, huge, env.Value.Postings[0].Amount.String())
}

func TestRunAcceptsScientificNotationBalance(t *testing.T) {
	// A JS number for a large integer stringifies as "1e+21"; the binding must
	// accept it (exactly, since it is integral) rather than hard-erroring.
	out := wasmbinding.Run([]byte(`{
		"script": "send [COIN 1000000000000000000000] (\n source = @alice\n destination = @dest\n)",
		"balances": { "alice": { "COIN": 1e21 } }
	}`))

	var env runEnvelope
	require.NoError(t, json.Unmarshal(out, &env))
	require.True(t, env.Ok, "unexpected error: %s", env.Error)
	require.Len(t, env.Value.Postings, 1)
	require.Equal(t, "1000000000000000000000", env.Value.Postings[0].Amount.String())
}

func TestRunMissingFundsReturnsAnnotatedError(t *testing.T) {
	out := wasmbinding.Run([]byte(`{
		"script": "send [COIN 10] (\n source = @alice\n destination = @dest\n)",
		"variables": {},
		"balances": {},
		"metadata": {},
		"featureFlags": []
	}`))

	var env runEnvelope
	require.NoError(t, json.Unmarshal(out, &env))
	require.False(t, env.Ok)
	require.NotEmpty(t, env.Error)
}

func TestRunInvalidJSONReturnsErrorEnvelope(t *testing.T) {
	out := wasmbinding.Run([]byte(`not json`))

	var env runEnvelope
	require.NoError(t, json.Unmarshal(out, &env))
	require.False(t, env.Ok)
	require.Contains(t, env.Error, "invalid input")
}

func TestRunInvalidBalanceAmountReturnsErrorEnvelope(t *testing.T) {
	out := wasmbinding.Run([]byte(`{
		"script": "send [COIN 10] (\n source = @alice\n destination = @dest\n)",
		"balances": { "alice": { "COIN": 10.5 } }
	}`))

	var env runEnvelope
	require.NoError(t, json.Unmarshal(out, &env))
	require.False(t, env.Ok)
	require.Contains(t, env.Error, "invalid balance amount")
}

// ---- NeededVariables ----------------------------------------------------------

func TestNeededVariables(t *testing.T) {
	out := wasmbinding.NeededVariables(`vars {
	monetary $amt
	account $dest
}
send $amt (
	source = @world
	destination = $dest
)`)

	var env struct {
		Ok        bool              `json:"ok"`
		Variables map[string]string `json:"variables"`
		Error     string            `json:"error"`
	}
	require.NoError(t, json.Unmarshal(out, &env))
	require.True(t, env.Ok)
	require.Equal(t, map[string]string{"amt": "monetary", "dest": "account"}, env.Variables)
}
