package numscript

import (
	"context"

	"github.com/formancehq/numscript/accounts"
	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/vm"
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
	// For each account, list of the needed assets
	BalanceQuery     = interpreter.BalanceQuery
	BalanceQueryItem = interpreter.BalanceQueryItem
	MetadataQuery    = interpreter.MetadataQuery
	AccountBalance   = interpreter.AccountBalance
	Balances         = interpreter.Balances
	BalanceRow       = interpreter.BalanceRow

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

func (p ParseResult) Run(ctx context.Context, vars VariablesMap, store Store) (ExecutionResult, InterpreterError) {
	return p.RunWithFeatureFlags(ctx, vars, store, nil)
}

func (p ParseResult) RunWithFeatureFlags(
	ctx context.Context,
	vars VariablesMap,
	store Store,
	featureFlags map[string]struct{},
) (ExecutionResult, InterpreterError) {
	if len(p.parseResult.Errors) != 0 {
		return ExecutionResult{}, p.parseResult.Errors[0]
	}

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

type (
	VarsEncoder     = compiler.VarsEncoder
	CompiledProgram = vm.Program
	VMStore         = vm.Store
	Vm              = vm.Vm
	Vars            = vm.Vars
)

var NewVm = vm.NewVm

var DecodeVars = vm.DecodeVars

func (p ParseResult) Compile() (VarsEncoder, CompiledProgram, error) {
	if len(p.parseResult.Errors) != 0 {
		return VarsEncoder{}, CompiledProgram{}, p.parseResult.Errors[0]
	}
	return compiler.Compile(p.parseResult.Value)
}

func Compile(source string) (VarsEncoder, CompiledProgram, error) {
	return Parse(source).Compile()
}

var DecodeCompiledProgram = vm.DecodeProgram

func ExecVm[S VMStore](machine *Vm, vars *Vars, store S) (ExecutionResult, error) {
	res, execErr := vm.Exec(machine, vars, store)
	if execErr != nil {
		return ExecutionResult{}, execErr
	}
	return res, nil
}
