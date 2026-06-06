package mcp_impl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// parseBalancesJson routes the caller-supplied balances payload through the
// canonical interpreter.Balances unmarshaller, so the MCP `evaluate` tool
// accepts exactly the same shapes documented elsewhere — bare number for the
// uncolored shorthand, single value-object, or array of value-objects.
//
// The MCP framework hands us an already-decoded `any`; re-marshalling round
// trips through JSON, but big-integer amounts stay safe because the
// re-encoded payload feeds into Balances.UnmarshalJSON which decodes each
// amount directly into a *big.Int.
func parseBalancesJson(balancesRaw any) (interpreter.Balances, *mcp.CallToolResult) {
	if balancesRaw == nil {
		return interpreter.Balances{}, nil
	}
	encoded, err := json.Marshal(balancesRaw)
	if err != nil {
		return interpreter.Balances{}, mcp.NewToolResultError(fmt.Sprintf("Could not re-encode balances: %v", err))
	}
	var balances interpreter.Balances
	if err := json.Unmarshal(encoded, &balances); err != nil {
		return interpreter.Balances{}, mcp.NewToolResultError(fmt.Sprintf("Invalid balances payload: %v", err))
	}
	return balances, nil
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
			mcp.Description(`The accounts' balances. A nested map from the account name, to the asset, to the held amount.

			Each per-asset entry accepts three forms:
			  - a bare integer for the uncolored bucket (shorthand)
			  - a single value-object: { "amount": N } or { "color": "RED", "amount": N }
			  - an array of value-objects when several colors coexist on the same asset

			Examples:
			  { "alice": { "USD/2": 100, "EUR/2": -42 }, "bob": { "BTC": 1 } }
			  { "alice": { "USD/2": [{ "amount": 100 }, { "color": "RED", "amount": 50 }] } }

			Color is a first-class dimension on the emitted postings (see Posting.color);
			the empty/missing color is its own bucket, distinct from any non-empty one.
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
