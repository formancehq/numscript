package difftest

import (
	"context"
	"math/big"

	"github.com/formancehq/numscript"
	"github.com/formancehq/numscript/internal/gen"
	"github.com/formancehq/numscript/internal/vm"
)

// vmStore is a numscript.VMStore over the same (account, asset) -> amount /
// (account, key) -> value maps runNew/runOracle already build their stores
// from. internal/gen never produces colored assets or scoped constructs, so
// the color and scope parameters are ignored (the vm never queries anything
// but the empty color and scope for generator output).
type vmStore struct {
	balances map[gen.BalanceKey]*big.Int
	metadata map[gen.MetaKey]string
}

func (s vmStore) GetBalance(_ context.Context, account, _, asset, _ string) (*big.Int, error) {
	if amount, ok := s.balances[gen.BalanceKey{Account: account, Asset: asset}]; ok {
		return new(big.Int).Set(amount), nil
	}
	return new(big.Int), nil
}

func (s vmStore) GetMetadata(_ context.Context, account, _, key string) (string, bool, error) {
	value, ok := s.metadata[gen.MetaKey{Account: account, Key: key}]
	return value, ok, nil
}

func runVM(ctx context.Context, script string, vars map[string]string, balances map[gen.BalanceKey]*big.Int, metadata map[gen.MetaKey]string) SideResult {
	varsEncoder, program, err := numscript.Compile(script)
	if err != nil {
		return SideResult{CompileErr: err.Error()}
	}

	encodedVars, err := varsEncoder.Encode(vars)
	if err != nil {
		return SideResult{CompileErr: err.Error()}
	}

	// The verifier is opt-in and the compiler is trusted not to need it, so this
	// is not defending the run — it is checking that claim on every generated
	// script. A failure means the compiler emitted bytecode the VM's own static
	// rules reject, which is a bug in this repo rather than a disagreement with
	// the oracle, hence InternalErr and not CompileErr.
	//
	// Worth having here specifically because internal/gen reaches shapes the
	// hand-written corpus doesn't, and it does so on inputs nobody chose.
	if err := numscript.VerifyCompiledProgramWithVars(program, &encodedVars); err != nil {
		return SideResult{InternalErr: "compiled program failed verification: " + err.Error()}
	}

	store := vmStore{balances: balances, metadata: metadata}

	execResult, execErr := numscript.ExecVm(ctx, numscript.NewVm(program), &encodedVars, store)
	if execErr != nil {
		_, missingFunds := execErr.(vm.MissingFundsError)
		return SideResult{RunErr: execErr.Error(), MissingFunds: missingFunds}
	}

	postings := make([]Posting, 0, len(execResult.Postings))
	for _, p := range execResult.Postings {
		postings = append(postings, Posting{
			Source:      p.Source,
			Destination: p.Destination,
			Asset:       p.Asset,
			Amount:      p.Amount,
		})
	}

	return SideResult{Postings: postings}
}
