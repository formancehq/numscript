package difftest

import (
	"context"
	"math/big"

	"github.com/formancehq/numscript"
	"github.com/formancehq/numscript/internal/gen"
)

// vmStore is a numscript.VMStore over the same (account, asset) -> amount /
// (account, key) -> value maps runNew/runOracle already build their stores
// from. internal/gen never produces colored assets or scoped constructs, so
// the color parameter is ignored (the vm never queries anything but the
// empty color for generator output).
type vmStore struct {
	balances map[gen.BalanceKey]*big.Int
	metadata map[gen.MetaKey]string
}

func (s vmStore) GetBalance(_ context.Context, account, asset, _ string) (*big.Int, error) {
	if amount, ok := s.balances[gen.BalanceKey{Account: account, Asset: asset}]; ok {
		return new(big.Int).Set(amount), nil
	}
	return new(big.Int), nil
}

func (s vmStore) GetMetadata(_ context.Context, account, key string) (string, bool, error) {
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

	store := vmStore{balances: balances, metadata: metadata}

	execResult, execErr := numscript.ExecVm(ctx, numscript.NewVm(program), &encodedVars, store)
	if execErr != nil {
		return SideResult{RunErr: execErr.Error()}
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
