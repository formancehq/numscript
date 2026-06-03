package numscript

import (
	"context"

	"github.com/formancehq/numscript/accounts"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
)

// This struct represents a parsed numscript source code
type ParseResult struct {
	parseResult parser.ParseResult
}

// Returns a map from a variable's name to its type.
//
// doesn't include variables whose value is already defined within the script
func (p ParseResult) GetNeededVariables() map[string]string {
	m := make(map[string]string)

	if p.parseResult.Value.Vars == nil {
		return m
	}

	for _, varDecl := range p.parseResult.Value.Vars.Declarations {
		if varDecl.Name == nil || varDecl.Origin != nil {
			continue
		}

		m[varDecl.Name.Name] = varDecl.Type.Name
	}

	return m
}

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
	// For each account, list of the needed (asset, color) pairs
	BalanceQuery = interpreter.BalanceQuery
	// AssetColor identifies a (asset, color) pair to query.
	AssetColor     = interpreter.AssetColor
	MetadataQuery  = interpreter.MetadataQuery
	AccountBalance = interpreter.AccountBalance
	ColorBalance   = interpreter.ColorBalance
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
	MissingFundsErr  = interpreter.MissingFundsErr
)

// Uncolored wraps an amount as a ColorBalance under the empty color key
// (the "no color" bucket). Useful when building a Balances literal in tests.
var Uncolored = interpreter.Uncolored

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

func (p ParseResult) GetInvolvedAccounts(vars VariablesMap) ([]accounts.InvolvedAccount, []accounts.InvolvedMeta, InterpreterError) {
	return interpreter.GetInvolvedAccounts(vars, p.parseResult.Value)
}
