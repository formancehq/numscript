package difftest

import (
	"bytes"
	"context"
	"errors"
	"math/big"

	"github.com/formancehq/numscript"
	"github.com/formancehq/numscript/internal/gen"
	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/vm"
)

// vmStore is a vm.Store over the same (account, asset) -> amount /
// (account, key) -> value maps runNew/runOracle already build their stores
// from. internal/gen presets only uncolored, unscoped balances and metadata,
// so a query for any other color or scope answers zero/absent — exactly what
// runNew's StaticStore does, since its rows all carry the empty color and
// scope. Answering the uncolored balance regardless of color (as this store
// once did) hands the vm phantom colored funds the interpreter doesn't see.
type vmStore struct {
	balances map[gen.BalanceKey]*big.Int
	metadata map[gen.MetaKey]string
}

func (s vmStore) GetBalance(_ context.Context, account, scope, asset, color string) (*big.Int, error) {
	if scope == "" && color == "" {
		if amount, ok := s.balances[gen.BalanceKey{Account: account, Asset: asset}]; ok {
			return new(big.Int).Set(amount), nil
		}
	}
	return new(big.Int), nil
}

func (s vmStore) GetMetadata(_ context.Context, account, scope, key string) (string, bool, error) {
	if scope != "" {
		return "", false, nil
	}
	value, ok := s.metadata[gen.MetaKey{Account: account, Key: key}]
	return value, ok, nil
}

func runVM(ctx context.Context, script string, vars map[string]string, balances map[gen.BalanceKey]*big.Int, metadata map[gen.MetaKey]string, featureFlags []string) SideResult {
	flagSet := make(map[string]struct{}, len(featureFlags))
	for _, f := range featureFlags {
		flagSet[f] = struct{}{}
	}
	varsEncoder, program, err := numscript.CompileWithFeatureFlags(script, flagSet)
	if err != nil {
		return SideResult{
			CompileErr:       err.Error(),
			RegisterOverflow: errors.Is(err, ir.ErrRegisterBankOverflow),
			ProgramTooLarge:  errors.Is(err, ir.ErrProgramTooLarge),
		}
	}

	// Encode/DecodeProgram must round-trip on everything the compiler emits;
	// nothing else in the differential loop exercises the codec, and generated
	// scripts reach value shapes (multi-word big.Ints, long pools) the codec
	// tests don't. Byte-stability (re-encoding the decoded program reproduces
	// the bytes) is the cheap full-structure check; behavioral equivalence is
	// pinned separately by the corpus.
	encoded := program.Encode()
	decoded, decErr := numscript.DecodeCompiledProgram(encoded)
	if decErr != nil {
		return SideResult{InternalErr: "DecodeProgram failed on Encode output: " + decErr.Error()}
	}
	if !bytes.Equal(decoded.Encode(), encoded) {
		return SideResult{InternalErr: "Encode/Decode/Encode is not byte-stable"}
	}

	// Binding the vars is the compile-to-run boundary, like the machine's
	// SetVarsFromJSON: report it as the resolve stage so Compare flags the vm
	// rejecting bindings the other engines accepted.
	encodedVars, err := varsEncoder.Encode(vars)
	if err != nil {
		return SideResult{ResolveErr: err.Error()}
	}

	// The verifier is opt-in and the compiler is trusted not to need it, so this
	// is not defending the run — it is checking that claim on every generated
	// script. A failure means the compiler emitted bytecode the VM's own static
	// rules reject, which is a bug in this repo rather than a disagreement with
	// the oracle, hence InternalErr and not CompileErr.
	//
	// Worth having here specifically because internal/gen reaches shapes the
	// hand-written corpus doesn't, and it does so on inputs nobody chose.
	if err := vm.VerifyWithVars(program, &encodedVars); err != nil {
		return SideResult{InternalErr: "compiled program failed verification: " + err.Error()}
	}

	store := vmStore{balances: balances, metadata: metadata}

	// The public entry point, so this leg exercises exactly what an integrator
	// calls — including its metadata contract.
	execResult, execErr := numscript.ExecVm(ctx, numscript.NewVm(program), &encodedVars, store)
	if execErr != nil {
		var missingFunds vm.MissingFundsError
		var negativeAmount vm.NegativeAmountError
		return SideResult{
			RunErr:         execErr.Error(),
			MissingFunds:   errors.As(execErr, &missingFunds),
			NegativeAmount: errors.As(execErr, &negativeAmount),
		}
	}

	postings := make([]Posting, 0, len(execResult.Postings))
	for _, p := range execResult.Postings {
		postings = append(postings, Posting{
			Source:      p.Source,
			Destination: p.Destination,
			Asset:       p.Asset,
			Color:       p.Color,
			Amount:      p.Amount,
		})
	}

	txMeta := make(map[string]string, len(execResult.Metadata))
	for k, v := range execResult.Metadata {
		txMeta[k] = v
	}
	accountMeta := make(map[string]string, len(execResult.AccountsMetadata))
	for _, row := range execResult.AccountsMetadata {
		accountMeta[metaKey(row.Account, row.Key)] = row.Value
	}

	return SideResult{Postings: postings, TxMeta: txMeta, AccountMeta: accountMeta}
}
