package mcp_impl

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func parseBalancesJson(balancesRaw any) (interpreter.Balances, *mcp.CallToolResult) {
	balances, ok := balancesRaw.(map[string]any)
	if !ok {
		return interpreter.Balances{}, mcp.NewToolResultError(fmt.Sprintf("Expected an object as balances, got: <%#v>", balancesRaw))
	}

	iBalances := interpreter.Balances{}
	for account, assetsRaw := range balances {
		if iBalances[account] == nil {
			iBalances[account] = interpreter.AccountBalance{}
		}

		assets, ok := assetsRaw.(map[string]any)
		if !ok {
			return interpreter.Balances{}, mcp.NewToolResultError(fmt.Sprintf("Expected nested object for account %v", account))
		}

		for asset, amountRaw := range assets {
			amount, ok := amountRaw.(float64)
			if !ok {
				return interpreter.Balances{}, mcp.NewToolResultError(fmt.Sprintf("Expected float for amount: %v", amountRaw))
			}

			n, _ := big.NewFloat(amount).Int(new(big.Int))
			iBalances[account][asset] = n
		}
	}
	return iBalances, nil
}

func parseVarsJson(varsRaw any) (map[string]string, *mcp.CallToolResult) {
	vars, ok := varsRaw.(map[string]any)
	if !ok {
		return map[string]string{}, mcp.NewToolResultError(fmt.Sprintf("Expected an object as vars, got: <%#v>", varsRaw))
	}

	iVars := map[string]string{}
	for key, rawValue := range vars {

		value, ok := rawValue.(string)
		if !ok {
			return map[string]string{}, mcp.NewToolResultError(fmt.Sprintf("Expected stringified var, got: %v", key))
		}

		iVars[key] = value
	}

	return iVars, nil
}

func addEvalTool(s *server.MCPServer) {
	tool := mcp.NewTool("evaluate",
		mcp.WithDescription("Evaluate a numscript program"),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("script",
			mcp.Required(),
			mcp.Description("The numscript source"),
		),
		mcp.WithObject("balances",
			mcp.Required(),
			mcp.Description(`The accounts' balances. A nested map from the account name, to the asset, to its integer amount.
			For example: { "alice": { "USD/2": 100, "EUR/2": -42 }, "bob": { "BTC": 1 } }
			`),
		),
		mcp.WithObject("vars",
			mcp.Required(),
			mcp.Description(`The stringified variables to be passed to the script's "vars" block.
			For example: { "acc": "alice", "mon": "EUR 100" } can be passed to the following script:
			vars {
				monetary $mon
				account $acc
			}

			send $mon (
				source = $acc
				destination = @world
			)
			`),
		),
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		script, err := request.RequireString("script")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		parsed := parser.Parse(script)
		if len(parsed.Errors) != 0 {
			// TODO return all errors
			out := make([]string, len(parsed.Errors))
			for index, err := range parsed.Errors {
				out[index] = err.Msg
			}
			mcp.NewToolResultError(strings.Join(out, ", "))
		}

		balances, mcpErr := parseBalancesJson(request.GetArguments()["balances"])
		if mcpErr != nil {
			return mcpErr, nil
		}

		vars, mcpErr := parseVarsJson(request.GetArguments()["vars"])
		if mcpErr != nil {
			return mcpErr, nil
		}

		out, iErr := interpreter.RunProgram(
			context.Background(),
			parsed.Value,
			vars,
			interpreter.StaticStore{
				Balances: balances,
			},
			map[string]struct{}{},
		)
		if iErr != nil {
			mcp.NewToolResultError(iErr.Error())
		}
		return mcp.NewToolResultJSON(*out)
	})
}

func RunServer() error {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Numscript",
		"0.0.1",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)
	addEvalTool(s)

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		return err
	}

	return nil
}
