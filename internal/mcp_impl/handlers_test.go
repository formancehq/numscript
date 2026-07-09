package mcp_impl

import (
	"context"
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
