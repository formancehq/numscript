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
					map[string]any{"account": "alice", "asset": "USD/2", "amount": "1", "scope": "x"},
					map[string]any{"account": "alice", "asset": "USD/2", "amount": "2", "scope": "x"},
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
					map[string]any{"account": "alice", "asset": "USD/2", "amount": "1"},
					map[string]any{"account": "alice", "asset": "USD/2", "amount": "2", "scope": "x"},
				},
				"vars": map[string]any{},
			},
		},
	})

	require.NoError(t, err)
	require.False(t, result.IsError)
}

func TestHandleEvalToolRejectsMissingAmount(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "evaluate",
			Arguments: map[string]any{
				"script": evalScript,
				"vars":   map[string]any{},
				"balances": []any{
					map[string]any{"account": "alice", "asset": "USD/2"},
				},
			},
		},
	}

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsError, "expected an error result for a missing amount, got: %#v", result)
}

func TestHandleEvalToolRejectsNullAmount(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "evaluate",
			Arguments: map[string]any{
				"script": evalScript,
				"vars":   map[string]any{},
				"balances": []any{
					map[string]any{"account": "alice", "asset": "USD/2", "amount": nil},
				},
			},
		},
	}

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsError, "expected an error result for a null amount, got: %#v", result)
}

func TestHandleEvalToolAcceptsAmountsWithinFloat64SafeRange(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "evaluate",
			Arguments: map[string]any{
				"script": evalScript,
				"vars":   map[string]any{},
				"balances": []any{
					map[string]any{"account": "alice", "asset": "USD/2", "amount": "100"},
				},
			},
		},
	}

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError, "expected a successful result, got: %#v", result)
}

func TestHandleEvalToolAcceptsAmountsBeyondFloat64SafeRange(t *testing.T) {
	// amounts are bound as strings and parsed with big.Int.SetString, so
	// there's no float64 anywhere in the path: an amount far beyond
	// float64's exact-integer range (2^53 - 1) - even beyond int64 - must
	// still come through with full precision, exactly like specs_format's
	// balances (which go straight from file bytes to *big.Int, with no
	// generic-any decoding step either).
	const hugeAmount = "99999999999999999999999999999999"
	wire := []byte(`{
		"method": "tools/call",
		"params": {
			"name": "evaluate",
			"arguments": {
				"script": ` + jsonString(evalScript) + `,
				"vars": {},
				"balances": [{"account":"alice","asset":"USD/2","amount": "` + hugeAmount + `"}]
			}
		}
	}`)

	var request mcp.CallToolRequest
	require.NoError(t, json.Unmarshal(wire, &request))

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError, "expected a successful result, got: %#v", result)
}

func TestHandleEvalToolAcceptsLargeNegativeAmount(t *testing.T) {
	// source = @world so alice's (irrelevant) balance can't cause a "not
	// enough funds" error - this isolates what we're testing here, namely
	// that the negative amount is parsed correctly rather than rejected as
	// malformed.
	const script = `send [USD/2 100] (
		source = @world
		destination = @alice
	)`

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "evaluate",
			Arguments: map[string]any{
				"script": script,
				"vars":   map[string]any{},
				"balances": []any{
					map[string]any{"account": "alice", "asset": "USD/2", "amount": "-1000000000000000000"},
				},
			},
		},
	}

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.False(t, result.IsError, "expected a successful result, got: %#v", result)
}

func TestHandleEvalToolRejectsFractionalAmount(t *testing.T) {
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "evaluate",
			Arguments: map[string]any{
				"script": evalScript,
				"vars":   map[string]any{},
				"balances": []any{
					map[string]any{"account": "alice", "asset": "USD/2", "amount": "100.9"},
				},
			},
		},
	}

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsError, "expected an error result for a fractional amount, got: %#v", result)
}

func TestHandleEvalToolRejectsFractionalAmountNearOldSafeIntegerBoundary(t *testing.T) {
	// reproduces the originally reported bypass: over-the-wire,
	// 9007199254740991.1 as a JSON *number* decodes to the float64 value
	// 9007199254740991, indistinguishable from a genuine in-range integer
	// once rounded. Amounts are now bound as raw strings (never decoded
	// through float64 at all) and parsed with big.Int.SetString, so the
	// fractional part is never lost, and this is rejected outright instead
	// of silently rounding to 9007199254740991.
	wire := []byte(`{
		"method": "tools/call",
		"params": {
			"name": "evaluate",
			"arguments": {
				"script": ` + jsonString(evalScript) + `,
				"vars": {},
				"balances": [{"account":"alice","asset":"USD/2","amount": "9007199254740991.1"}]
			}
		}
	}`)

	var request mcp.CallToolRequest
	require.NoError(t, json.Unmarshal(wire, &request))

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsError, "expected an error result for a fractional amount near the safe integer boundary, got: %#v", result)
}

func TestHandleEvalToolRejectsFractionalDriftNearFormerTightenedBoundary(t *testing.T) {
	// reproduces the second reviewer finding against a since-abandoned
	// approach: capping the *magnitude* of a float64-decoded amount
	// (instead of rejecting JSON numbers outright) can never fully close
	// this, because "1125899906842623.01" rounds to exactly the cap value
	// itself (1125899906842623) - a fractional, out-of-spec amount that's
	// indistinguishable from a genuine in-range integer once decoded, no
	// matter where the cap is set. Binding amounts as strings sidesteps the
	// float64 rounding entirely, so this is rejected outright.
	wire := []byte(`{
		"method": "tools/call",
		"params": {
			"name": "evaluate",
			"arguments": {
				"script": ` + jsonString(evalScript) + `,
				"vars": {},
				"balances": [{"account":"alice","asset":"USD/2","amount": "1125899906842623.01"}]
			}
		}
	}`)

	var request mcp.CallToolRequest
	require.NoError(t, json.Unmarshal(wire, &request))

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsError, "expected an error result for a fractional amount near the former tightened boundary, got: %#v", result)
}

func TestHandleEvalToolRejectsAmountPassedAsJSONNumber(t *testing.T) {
	// amounts must be passed as JSON strings, not JSON numbers: a bare
	// number goes through the MCP transport's generic float64 decoding,
	// which is where precision loss happens in the first place, so this
	// format is rejected outright rather than silently accepted.
	wire := []byte(`{
		"method": "tools/call",
		"params": {
			"name": "evaluate",
			"arguments": {
				"script": ` + jsonString(evalScript) + `,
				"vars": {},
				"balances": [{"account":"alice","asset":"USD/2","amount": 100}]
			}
		}
	}`)

	var request mcp.CallToolRequest
	require.NoError(t, json.Unmarshal(wire, &request))

	result, err := handleEvalTool(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsError, "expected an error result for an amount passed as a JSON number, got: %#v", result)
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
