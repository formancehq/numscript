package mcp_impl

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

const evalScript = `send [USD/2 100] (
	source = @alice
	destination = @bob
)`

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

func TestHandleEvalToolRejectsDuplicateBalancesWithScope(t *testing.T) {
	result, err := handleEvalTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"script": `
		send [USD/2 1] (
			source = @world
			destination = @a
		)
		`,
				"balances": []any{
					map[string]any{"account": "alice", "asset": "USD/2", "amount": 1, "scope": "x"},
					map[string]any{"account": "alice", "asset": "USD/2", "amount": 2, "scope": "x"},
				},
				"vars": map[string]any{},
			},
		},
	})

	require.NoError(t, err)
	require.True(t, result.IsError)
	text, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok)
	require.Contains(t, text.Text, "must not contain duplicate entries")
	require.Contains(t, text.Text, `scope="x"`)
}

func TestHandleEvalToolAllowsSameBalanceKeyDifferentScope(t *testing.T) {
	result, err := handleEvalTool(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"script": `
				send [USD/2 1] (
					source = @world
					destination = @a
				)
				`,
				"balances": []any{
					map[string]any{"account": "alice", "asset": "USD/2", "amount": 1},
					map[string]any{"account": "alice", "asset": "USD/2", "amount": 2, "scope": "x"},
				},
				"vars": map[string]any{},
			},
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)
}

func TestHandleEvalToolRejectsAmountBeyondSafeIntegerRange(t *testing.T) {
	// simulates the real MCP transport: the incoming JSON-RPC message is
	// decoded generically before this handler ever runs, so an amount past
	// float64's exact-integer range (2^53 - 1) is already rounded by the
	// time we see it - 9007199254740993 becomes 9007199254740992.
	wire := []byte(`{
		"method": "tools/call",
		"params": {
			"name": "evaluate",
			"arguments": {
				"script": ` + jsonString(evalScript) + `,
				"vars": {},
				"balances": [{"account":"alice","asset":"USD/2","amount": 9007199254740993}]
			}
		}
	}`)

	var request mcp.CallToolRequest
	require.NoError(t, json.Unmarshal(wire, &request))

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsError, "expected an error result for an amount beyond the safe integer range, got: %#v", result)
}

func TestHandleEvalToolRejectsUnsafelyLargeNegativeAmount(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "evaluate",
			Arguments: map[string]any{
				"script": evalScript,
				"vars":   map[string]any{},
				"balances": []any{
					map[string]any{"account": "alice", "asset": "USD/2", "amount": float64(-1e18)},
				},
			},
		},
	}

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsError, "expected an error result for a negative amount beyond the safe integer range, got: %#v", result)
}

func TestHandleEvalToolAcceptsAmountsWithinSafeRange(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "evaluate",
			Arguments: map[string]any{
				"script": evalScript,
				"vars":   map[string]any{},
				"balances": []any{
					map[string]any{"account": "alice", "asset": "USD/2", "amount": float64(100)},
				},
			},
		},
	}

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError, "expected a successful result, got: %#v", result)
}

func TestHandleEvalToolRejectsFractionalAmount(t *testing.T) {
	// unrelated to the safe-integer-range check: a fractional amount is
	// still rejected downstream by *big.Int's own JSON unmarshaling.
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "evaluate",
			Arguments: map[string]any{
				"script": evalScript,
				"vars":   map[string]any{},
				"balances": []any{
					map[string]any{"account": "alice", "asset": "USD/2", "amount": float64(100.9)},
				},
			},
		},
	}

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsError, "expected an error result for a fractional amount, got: %#v", result)
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
