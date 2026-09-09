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
		mcp.WithArray("balances",
			mcp.Required(),
			mcp.Description(`The accounts' balances. A list of entries, each an object with an "account", an "asset", an "amount", an optional "color", and an optional "scope".
			"amount" must be passed as a string containing a base-10 integer (e.g. "100", not 100): JSON numbers only round-trip exactly up to 2^53-1, so an amount given as a JSON number can be silently corrupted in transit.
			The (account, asset, color, scope) tuple of each entry must be unique within the list.
			For example: [ { "account": "alice", "asset": "USD/2", "amount": "100" }, { "account": "alice", "asset": "EUR/2", "amount": "-42" }, { "account": "bob", "asset": "BTC", "amount": "1" } ]
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

	// balances are bound with a string Amount, not interpreter.Balances'
	// *big.Int, so this never goes through the MCP transport's generic
	// float64 argument decoding (Params.Arguments any): a JSON string
	// round-trips through that decode byte for byte, unlike a JSON number,
	// which can already be silently rounded - in magnitude or, at the right
	// magnitude, in its fractional part - by the time a handler sees it.
	var args struct {
		Vars     map[string]string `json:"vars"`
		Balances []struct {
			Account string `json:"account"`
			Asset   string `json:"asset"`
			Amount  string `json:"amount"`
			Color   string `json:"color"`
			Scope   string `json:"scope"`
		} `json:"balances"`
	}
	err = request.BindArguments(&args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	balances := make(interpreter.Balances, len(args.Balances))
	for i, row := range args.Balances {
		amount, ok := new(big.Int).SetString(row.Amount, 10)
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf(
				"amount %q for account=%q asset=%q is not a valid base-10 integer string",
				row.Amount, row.Account, row.Asset,
			)), nil
		}
		balances[i] = interpreter.BalanceRow{
			Account: row.Account,
			Asset:   row.Asset,
			Amount:  amount,
			Color:   row.Color,
			Scope:   row.Scope,
		}
	}

	if dup, ok := balances.FirstDuplicate(); ok {
		key := fmt.Sprintf("account=%q asset=%q", dup.Account, dup.Asset)
		if dup.Color != "" {
			key += fmt.Sprintf(" color=%q", dup.Color)
		}
		if dup.Scope != "" {
			key += fmt.Sprintf(" scope=%q", dup.Scope)
		}
		return mcp.NewToolResultError("balances must not contain duplicate entries: duplicate entry for " + key), nil
	}

	out, iErr := interpreter.RunProgram(
		ctx,
		parsed.Value,
		args.Vars,
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
