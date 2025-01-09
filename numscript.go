package numscript

import (
	"context"

	"github.com/PagoPlus/numscript-wasm/internal/interpreter"
	"github.com/PagoPlus/numscript-wasm/internal/parser"
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

var ParseErrorsToString = parser.ParseErrorsToString

func (p ParseResult) GetParsingErrors() []ParserError {
	return p.parseResult.Errors
}

type (
	VariablesMap    = interpreter.VariablesMap
	Posting         = interpreter.Posting
	ExecutionResult = interpreter.ExecutionResult
	// For each account, list of the needed assets
	BalanceQuery   = interpreter.BalanceQuery
	MetadataQuery  = interpreter.MetadataQuery
	AccountBalance = interpreter.AccountBalance
	Balances       = interpreter.Balances

	AccountMetadata = interpreter.AccountMetadata

	// The newly defined account metadata after the execution
	AccountsMetadata = interpreter.AccountsMetadata

	// The transaction metadata, set by set_tx_meta()
	Metadata = interpreter.Metadata

	Store = interpreter.Store

	StaticStore = interpreter.StaticStore

	Value = interpreter.Value

	InterpreterError = interpreter.InterpreterError
)

func (p ParseResult) Run(ctx context.Context, vars VariablesMap, store Store) (ExecutionResult, InterpreterError) {
	return p.RunWithFeatureFlags(ctx, vars, store, nil)
}

func (p ParseResult) RunWithFeatureFlags(
	ctx context.Context,
	vars VariablesMap,
	store Store,
	featureFlags map[string]struct{},
) (ExecutionResult, InterpreterError) {
	if featureFlags == nil {
		featureFlags = make(map[string]struct{})
	}

	res, err := interpreter.RunProgram(ctx, p.parseResult.Value, vars, store, featureFlags)
	if err != nil {
		return ExecutionResult{}, err
	}
	return *res, nil
}

func (p ParseResult) GetSource() string {
	return p.parseResult.Source
}
