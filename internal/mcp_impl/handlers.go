package mcp_impl

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// maxSafeAmountFloat is 2^53. Integers whose absolute value is strictly below
// this bound are exactly representable as float64, so a JSON number that
// decodes to such a float64 is guaranteed to carry the caller's exact value.
const maxSafeAmountFloat = float64(1 << 53)

// parseAmount converts a JSON-decoded balance amount into a big.Int without
// any precision loss.
//
// Amounts can be sent either as JSON numbers or as strings. JSON numbers are
// decoded to float64 by the MCP transport, so they are only accepted when
// they are exact integers within the float64 safe-integer range (|n| < 2^53);
// anything else must be sent as a string to preserve precision.
func parseAmount(amountRaw any) (*big.Int, error) {
	switch amount := amountRaw.(type) {
	case string:
		n, ok := new(big.Int).SetString(strings.TrimSpace(amount), 10)
		if !ok {
			return nil, fmt.Errorf("expected an integer string as amount, got: %q", amount)
		}
		return n, nil

	case json.Number:
		n, ok := new(big.Int).SetString(amount.String(), 10)
		if !ok {
			return nil, fmt.Errorf("expected an integer amount, got: %v", amount)
		}
		return n, nil

	case float64:
		if amount != math.Trunc(amount) {
			return nil, fmt.Errorf("expected an integer amount, got a non-integer number: %v", amount)
		}
		if math.Abs(amount) >= maxSafeAmountFloat {
			return nil, fmt.Errorf("amount %v is outside the range JSON numbers can represent exactly; pass the amount as a string instead, e.g. \"9007199254740993\"", strconv.FormatFloat(amount, 'f', -1, 64))
		}
		return big.NewInt(int64(amount)), nil

	default:
		return nil, fmt.Errorf("expected a number or an integer string as amount, got: %v", amountRaw)
	}
}

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
			n, err := parseAmount(amountRaw)
			if err != nil {
				return interpreter.Balances{}, mcp.NewToolResultError(fmt.Sprintf("Invalid amount for account %q, asset %q: %v", account, asset, err))
			}
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
			Amounts whose absolute value is 2^53 or greater must be passed as decimal strings to avoid precision loss, e.g. { "alice": { "USD/2": "9007199254740993" } }
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
	s.AddTool(tool, handleEvalTool)
}

func handleEvalTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		return mcp.NewToolResultError(strings.Join(out, ", ")), nil
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
