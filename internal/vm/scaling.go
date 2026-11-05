package vm

import (
	"context"
	"errors"
	"math/big"

	"github.com/formancehq/numscript/internal/funds"
)

// scaleAccount is the interpreter's `with scaling through` conversion step.
//
// The interpreter sees the account's family through a batched `base/*` balance
// query (skipped for @world) prewarmed into the run state, plus whatever the run
// already touched; fetching the family here and prewarming it reproduces that
// view, since Prewarm never overwrites a loaded entry.
func scaleAccount[S Store](
	ctx context.Context,
	store S,
	runstate *funds.RunState,
	asset, account, scope, swap, swapScope string,
	cap *big.Int,
) ExecutionError {
	baseAsset, assetScale := funds.GetBaseAndScale(asset)

	if account != "world" {
		fetched, err := store.GetScaledBalances(ctx, account, scope, baseAsset)
		if err != nil {
			return StoreError{Wrapped: err}
		}
		seed := make(map[funds.PairKey]*big.Int, len(fetched))
		for fetchedAsset, amount := range fetched {
			seed[funds.PairKey{Account: account, Scope: scope, Asset: fetchedAsset}] = amount
		}
		runstate.Prewarm(seed)
	}

	balances, err := runstate.AccountBalances(account, scope, baseAsset)
	if err != nil {
		return StoreError{Wrapped: err}
	}
	if len(balances) == 0 {
		return InvalidScalingSource{Account: account}
	}

	if cap != nil && cap.Sign() < 0 {
		cap = new(big.Int)
	}
	sol, swapped := funds.FindScalingSolution(cap, assetScale, funds.GetAssets(balances, baseAsset))

	for _, pair := range sol {
		if err := forcePosting(runstate, account, scope, swap, swapScope, funds.BuildScaledAsset(baseAsset, pair.Scale), pair.Amount); err != nil {
			return err
		}
	}
	return forcePosting(runstate, swap, swapScope, account, scope, asset, swapped)
}

func forcePosting(runstate *funds.RunState, src, srcScope, dst, dstScope, asset string, amount *big.Int) ExecutionError {
	err := runstate.ForcePosting(src, srcScope, dst, dstScope, asset, "", amount)
	if errors.Is(err, funds.ErrNegativePosting) {
		return InternalError{Err: InvalidPostingError{Posting: funds.Posting{
			Source:           src,
			SourceScope:      srcScope,
			Destination:      dst,
			DestinationScope: dstScope,
			Asset:            asset,
			Amount:           new(big.Int).Set(amount),
		}}}
	}
	if err != nil {
		return StoreError{Wrapped: err}
	}
	return nil
}
