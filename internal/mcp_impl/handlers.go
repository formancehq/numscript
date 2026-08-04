package mcp_impl

import (
	"context"
	"fmt"
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
			mcp.Description(`The accounts' balances. A list of entries, each an object with an "account", an "asset", an integer "amount", an optional "color", and an optional "scope".
			The (account, asset, color, scope) tuple of each entry must be unique within the list.
			For example: [ { "account": "alice", "asset": "USD/2", "amount": 100 }, { "account": "alice", "asset": "EUR/2", "amount": -42 }, { "account": "bob", "asset": "BTC", "amount": 1 } ]
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

// maxExactJSONInt is the largest integer float64 can represent without loss
// (2^53 - 1). The MCP transport decodes incoming JSON numbers into a generic
// float64 (via Params.Arguments any) before this handler ever runs, so a
// balance amount past this magnitude may already have been silently rounded
// by the time BindArguments hands it to *big.Int - it will look like a
// perfectly valid, exact integer at that point, indistinguishable from one
// that was never corrupted. We can't recover the original value, so we
// refuse to execute with one instead of risking a silently wrong balance.
const maxExactJSONInt = float64(9_007_199_254_740_991)

// checkBalanceAmountsInSafeRange rejects any "balances" entry whose amount
// falls outside the range a JSON number can represent exactly. Must run
// before the amount is converted to *big.Int, since that conversion can no
// longer tell a corrupted value from a genuine one.
func checkBalanceAmountsInSafeRange(args map[string]any) *mcp.CallToolResult {
	balancesRaw, ok := args["balances"].([]any)
	if !ok {
		return nil
	}

	for _, rowRaw := range balancesRaw {
		row, ok := rowRaw.(map[string]any)
		if !ok {
			continue
		}

		amount, ok := row["amount"].(float64)
		if !ok {
			continue
		}

		if amount < -maxExactJSONInt || amount > maxExactJSONInt {
			return mcp.NewToolResultError(fmt.Sprintf(
				"amount %v for account=%v asset=%v exceeds the range a JSON number can represent exactly (±(2^53-1)); it may have already lost precision before reaching the server",
				amount, row["account"], row["asset"],
			))
		}
	}

	return nil
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

	if result := checkBalanceAmountsInSafeRange(request.GetArguments()); result != nil {
		return result, nil
	}

	var args struct {
		Vars     map[string]string    `json:"vars"`
		Balances interpreter.Balances `json:"balances"`
	}
	err = request.BindArguments(&args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if dup, ok := args.Balances.FirstDuplicate(); ok {
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
			Balances: args.Balances,
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
