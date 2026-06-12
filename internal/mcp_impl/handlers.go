package mcp_impl

import (
	"context"
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
			mcp.Description(`The accounts' balances. A list of entries, each an object with an "account", an "asset", an integer "amount", and an optional "color".
			The (account, asset, color) triple of each entry must be unique within the list.
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

	var args struct {
		Vars     map[string]string    `json:"vars"`
		Balances interpreter.Balances `json:"balances"`
	}
	err = request.BindArguments(&args)
	if err != nil {
		return nil, err
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
