package numscript

import (
	"context"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
)

// This struct represents a parsed numscript source code
type ParseResult struct {
	parseResult parser.ParseResult
}

// ---- TODO useful for the playground
// func (*ParseResult) GetNeededVariables() map[string]ValueType {}
// func (*ParseResult) GetDiagnostics() []Diagnostic {}

type ParserError = parser.ParserError

func Parse(code string) ParseResult {
	return ParseResult{parseResult: parser.Parse(code)}
}

func (p ParseResult) GetParsingErrors() []ParserError {
	return p.parseResult.Errors
}

type VariablesMap = interpreter.VariablesMap

type Posting = interpreter.Posting

type ExecutionResult = interpreter.ExecutionResult

// For each account, list of the needed assets
type BalanceQuery = interpreter.BalanceQuery

type MetadataQuery = interpreter.MetadataQuery

type AccountBalance = interpreter.AccountBalance
type Balances = interpreter.Balances

type AccountMetadata = interpreter.AccountMetadata
type Metadata = interpreter.Metadata

type Store = interpreter.Store

func (p ParseResult) Run(ctx context.Context, vars VariablesMap, store Store) (ExecutionResult, error) {
	res, err := interpreter.RunProgram(ctx, p.parseResult.Value, vars, store)
	if err != nil {
		return ExecutionResult{}, err
	}
	return *res, nil
}
