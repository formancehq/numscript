package numscript

import (
	"math/big"

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

type Posting struct {
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Amount      *big.Int `json:"amount"`
	Asset       string   `json:"asset"`
}

type ExecutionResult struct {
	Postings     []Posting                    `json:"postings"`
	TxMeta       map[string]string            `json:"txMeta"`
	AccountsMeta map[string]map[string]string `json:"accountsMeta"`
}

// For each account, list of the needed assets
type BalanceQuery = interpreter.BalanceQuery

type MetadataQuery = interpreter.MetadataQuery

type AccountBalance = interpreter.AccountBalance
type Balances = interpreter.Balances

type AccountMetadata = interpreter.AccountMetadata
type Metadata = interpreter.Metadata

type Store = interpreter.Store

func (p ParseResult) Run(store Store) (ExecutionResult, error) {
	interpreter.RunProgram(p.parseResult.Value, store)

	return ExecutionResult{}, nil
}
