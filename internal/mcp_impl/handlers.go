package mcp_impl

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// maxExactJSONInt is the largest integer that float64 can represent
// without loss (2^53 - 1). JSON-decoded amounts past this magnitude have
// already lost precision before reaching the handler, so we reject them.
const maxExactJSONInt = float64(9_007_199_254_740_991)

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
				return interpreter.Balances{}, mcp.NewToolResultError(fmt.Sprintf("Expected number for amount on %s/%s, got: <%#v>", account, asset, amountRaw))
			}

			// JSON numbers arrive here as float64. Reject anything that
			// cannot be losslessly represented as an integer in float64
			// precision: fractional values and magnitudes past 2^53 - 1.
			// Silent truncation / rounding on a balance is not acceptable
			// for a financial DSL — the caller should switch to a safer
			// encoding when they need values outside the safe range.
			if amount < -maxExactJSONInt || amount > maxExactJSONInt || amount != float64(int64(amount)) {
				return interpreter.Balances{}, mcp.NewToolResultError(fmt.Sprintf("amount for %s/%s must be an exact integer in [-(2^53-1), 2^53-1], got: %v", account, asset, amount))
			}

			iBalances[account][asset] = big.NewInt(int64(amount))
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
			return map[string]string{}, mcp.NewToolResultError(fmt.Sprintf("Expected %s var to be a string, got: %T instead", key, rawValue))
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
			ctx,
			parsed.Value,
			vars,
			interpreter.StaticStore{
				Balances: balances,
			},
			map[string]struct{}{},
		)
		if iErr != nil {
			return mcp.NewToolResultError(iErr.Error()), nil
		}
		return mcp.NewToolResultJSON(*out)
	})
}

func addCheckTool(s *server.MCPServer) {
	tool := mcp.NewTool("check",
		mcp.WithDescription("Check a program for parsing error or static analysis errors"),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithOpenWorldHintAnnotation(false),
		mcp.WithString("script",
			mcp.Required(),
			mcp.Description("The numscript source"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		script, err := request.RequireString("script")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		checkResult := analysis.CheckSource(script)

		var errors []any
		for _, d := range checkResult.Diagnostics {
			errors = append(errors, map[string]any{
				"kind":     d.Kind.Message(),
				"severity": analysis.SeverityToString(d.Kind.Severity()),
				"span":     d.Range,
			})
		}

		return mcp.NewToolResultJSON(map[string]any{
			"errors": errors,
		})
	})
}

func RunServer() error {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Numscript",
		"0.0.1",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
		server.WithInstructions(`
		You're a Numscript expert AI assistant. Numscript is a DSL that allows modeling financial transactions in an easy and declarative way. Numscript scripts always terminate.
		`),
	)
	addEvalTool(s)
	addCheckTool(s)

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		return err
	}

	return nil
}
