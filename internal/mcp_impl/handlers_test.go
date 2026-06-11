package mcp_impl

import (
	"context"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

func TestHandleEvalToolRejectsParseErrors(t *testing.T) {
	result, err := handleEvalTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"script":   "send [COIN 100] (",
				"balances": map[string]any{},
				"vars":     map[string]any{},
			},
		},
	})

	require.NoError(t, err)
	require.True(t, result.IsError)
	require.NotEmpty(t, result.Content)
	text, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, text.Text, "mismatched input")
}

// errorText extracts the text of an error tool result.
func errorText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.NotNil(t, result)
	require.True(t, result.IsError)
	require.NotEmpty(t, result.Content)
	text, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok)
	return text.Text
}

func TestParseAmountExactSmallInteger(t *testing.T) {
	n, err := parseAmount(float64(100))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(100), n)
}

func TestParseAmountNegativeInteger(t *testing.T) {
	n, err := parseAmount(float64(-42))
	require.NoError(t, err)
	require.Equal(t, big.NewInt(-42), n)
}

func TestParseAmountLargeIntegerAsString(t *testing.T) {
	// 2^53 + 1 cannot be represented exactly as a float64,
	// but must round-trip exactly when sent as a string
	n, err := parseAmount("9007199254740993")
	require.NoError(t, err)

	expected, ok := new(big.Int).SetString("9007199254740993", 10)
	require.True(t, ok)
	require.Equal(t, expected, n)
}

func TestParseAmountNegativeLargeIntegerAsString(t *testing.T) {
	n, err := parseAmount("-9007199254740993")
	require.NoError(t, err)

	expected, ok := new(big.Int).SetString("-9007199254740993", 10)
	require.True(t, ok)
	require.Equal(t, expected, n)
}

func TestParseAmountJsonNumber(t *testing.T) {
	n, err := parseAmount(json.Number("9007199254740993"))
	require.NoError(t, err)

	expected, ok := new(big.Int).SetString("9007199254740993", 10)
	require.True(t, ok)
	require.Equal(t, expected, n)
}

func TestParseAmountRejectsUnsafeFloat(t *testing.T) {
	// 2^53 + 1 sent as a JSON number arrives as float64(9007199254740992):
	// the precision loss is undetectable, so the amount must be rejected
	// instead of being silently rounded
	_, err := parseAmount(float64(9007199254740993))
	require.Error(t, err)
	require.Contains(t, err.Error(), "pass the amount as a string")
}

func TestParseAmountRejectsFractional(t *testing.T) {
	_, err := parseAmount(float64(100.5))
	require.Error(t, err)
	require.Contains(t, err.Error(), "non-integer")
}

func TestParseAmountRejectsNonNumericString(t *testing.T) {
	_, err := parseAmount("100.5")
	require.Error(t, err)

	_, err = parseAmount("not a number")
	require.Error(t, err)
}

func TestParseAmountRejectsInvalidType(t *testing.T) {
	_, err := parseAmount(true)
	require.Error(t, err)

	_, err = parseAmount(nil)
	require.Error(t, err)
}

func TestParseBalancesJsonExactAmounts(t *testing.T) {
	// simulate the MCP transport: arguments arrive as the result of
	// json.Unmarshal into map[string]any (numbers become float64)
	var balancesRaw any
	require.NoError(t, json.Unmarshal(
		[]byte(`{ "alice": { "USD/2": 100, "EUR/2": -42 }, "bob": { "BTC": "9007199254740993" } }`),
		&balancesRaw,
	))

	balances, mcpErr := parseBalancesJson(balancesRaw)
	require.Nil(t, mcpErr)

	require.Equal(t, big.NewInt(100), balances["alice"]["USD/2"])
	require.Equal(t, big.NewInt(-42), balances["alice"]["EUR/2"])

	expected, ok := new(big.Int).SetString("9007199254740993", 10)
	require.True(t, ok)
	require.Equal(t, expected, balances["bob"]["BTC"])
}

func TestParseBalancesJsonRejectsUnsafeFloat(t *testing.T) {
	var balancesRaw any
	require.NoError(t, json.Unmarshal(
		[]byte(`{ "alice": { "USD/2": 9007199254740993 } }`),
		&balancesRaw,
	))

	_, mcpErr := parseBalancesJson(balancesRaw)
	text := errorText(t, mcpErr)
	require.Contains(t, text, `account "alice"`)
	require.Contains(t, text, `asset "USD/2"`)
	require.Contains(t, text, "pass the amount as a string")
}

func TestParseBalancesJsonRejectsFractional(t *testing.T) {
	var balancesRaw any
	require.NoError(t, json.Unmarshal(
		[]byte(`{ "alice": { "USD/2": 100.5 } }`),
		&balancesRaw,
	))

	_, mcpErr := parseBalancesJson(balancesRaw)
	require.Contains(t, errorText(t, mcpErr), "non-integer")
}

func TestHandleEvalToolUsesExactBalances(t *testing.T) {
	// end-to-end: a balance above 2^53 passed as a string is used exactly
	result, err := handleEvalTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"script": `send [COIN 9007199254740993] (
	source = @alice
	destination = @bob
)`,
				"balances": map[string]any{
					"alice": map[string]any{
						"COIN": "9007199254740993",
					},
				},
				"vars": map[string]any{},
			},
		},
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.IsError)
	require.NotEmpty(t, result.Content)
	text, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, text.Text, "9007199254740993")
}
