package difftest

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/gen"
)

// TestKnownBugRepros locks in the real divergences found and fixed while
// building this harness (see DIFFTEST_HANDOFF.md), independent of whether
// the fuzzer's corpus/coverage happens to rediscover them. Each case failed
// before its corresponding fix and must pass now.
func TestKnownBugRepros(t *testing.T) {
	testCases := []struct {
		name     string
		script   string
		balances map[gen.BalanceKey]*big.Int
	}{
		{
			// vm.StaticStore.GetBalances (oracle) used to reinitialize its
			// per-account map on every asset iteration, silently dropping
			// every asset but the last queried for an account used with 2+
			// assets. Fixed in internal/oracle/machine/vm/store.go.
			name: "multi-asset same source account",
			script: `send [COIN 10] (
  source = @world
  destination = @acc0
)

send [EUR/2 10] (
  source = @world
  destination = @acc0
)

send [COIN 5] (
  source = @acc0
  destination = @acc1
)

send [EUR/2 5] (
  source = @acc0
  destination = @acc1
)`,
		},
		{
			// Machine.UnresolvedResourceBalances (oracle) used to be a
			// single-int map, so two balance()-origin vars on the same
			// account silently collided and left one resource unresolved
			// (nil Amount), later crashing the VM. Fixed in
			// internal/oracle/machine/vm/machine.go.
			name: "duplicate balance()-origin var, same account and asset",
			script: `vars {
  monetary $a = balance(@acc0, COIN)
  monetary $b = balance(@acc0, COIN)
}

send $a (
  source = @acc1 allowing unbounded overdraft
  destination = @acc2
)

send $b (
  source = @acc1 allowing unbounded overdraft
  destination = @acc3
)`,
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "acc0", Asset: "COIN"}: big.NewInt(100),
			},
		},
		{
			// Same root cause as above, one layer up
			// (assignBalanceAsResource), for two balance()-origin vars on
			// the same account but different assets, one of which has a
			// negative preset balance (must error identically on both
			// sides, not crash).
			name: "duplicate balance()-origin var, same account different asset",
			script: `vars {
  monetary $a = balance(@acc0, COIN)
  monetary $b = balance(@acc0, EUR/2)
}

send $a (
  source = @acc1 allowing unbounded overdraft
  destination = @acc2
)`,
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "acc0", Asset: "COIN"}:  big.NewInt(-50),
				{Account: "acc0", Asset: "EUR/2"}: big.NewInt(100),
			},
		},
		{
			// The new interpreter used to silently clamp a negative
			// max-clause destination amount to zero instead of erroring,
			// while the oracle correctly raises a runtime error. Fixed in
			// internal/interpreter/interpreter.go's sendTo
			// (*parser.DestinationInorder case).
			name: "negative max-clause amount",
			script: `send [EUR/2 100] (
  source = @acc2 allowing unbounded overdraft
  destination = {
    max [EUR/2 0] - [EUR/2 35] to @acc1
    remaining to @acc2
  }
)`,
		},
		{
			// `save` for more than an account's actual balance used to leave
			// the oracle's cached balance negative (a plain, unfloored
			// subtraction), while the new interpreter floors at zero. That
			// difference is invisible until a later bounded-overdraft draw
			// on the same account computes a different available "room"
			// from the two different balances. Fixed in
			// internal/oracle/machine/vm/machine.go's OP_SAVE handler to
			// floor at zero too, matching internal/interpreter's
			// runSaveStatement.
			name: "save more than the account's balance",
			script: `send [COIN 100] (
  source = @world
  destination = @acc0
)

save [COIN 900] from @acc0

send [COIN 250] (
  source = @acc0 allowing overdraft up to [COIN 1000]
  destination = @acc1
)`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			newRes := runNew(ctx, tc.script, nil, tc.balances, nil)
			oracleRes := runOracle(ctx, tc.script, nil, tc.balances, nil)

			v := Compare(newRes, oracleRes, "new interpreter", "oracle")
			if v.Mismatch {
				t.Fatalf("mismatch: %s\nnew: %+v\noracle: %+v", v.Reason, newRes, oracleRes)
			}
		})
	}
}
