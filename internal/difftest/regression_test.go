package difftest

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/gen"
)

// TestKnownBugRepros locks in the real divergences found and fixed while
// building this harness, independent of whether
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			newRes := runNew(ctx, tc.script, nil, tc.balances, nil)
			oracleRes := runOracle(ctx, tc.script, nil, tc.balances, nil)

			v := Compare(tc.script, newRes, oracleRes, "new interpreter", "oracle")
			if v.Mismatch {
				t.Fatalf("mismatch: %s\nnew: %+v\noracle: %+v", v.Reason, newRes, oracleRes)
			}
		})
	}
}

// TestSourceSideNegativeMaxClauseTolerated locks in a deliberate, accepted
// gap: a negative amount in a source-side `max ... from` clause is a hard
// reject on the oracle, but the interpreter and vm both silently treat it as
// contributing zero instead (the same "clamp instead of error" bug class
// already fixed on the destination side — see "negative max-clause amount"
// in TestKnownBugRepros above — but deliberately left unfixed here, since
// closing it would mean changing the interpreter's "ground truth" behavior,
// not just the vm catching up to it). With a fallback source in the same
// list, the interpreter draws the shortfall from there and succeed, while
// the oracle never gets that far. SideResult.MissingFunds and Compare's
// classification-based tolerance (see its doc comment) exist specifically
// so this doesn't produce fuzzer noise: as long as the underlying
// funds-adequacy outcome agrees across engines, this divergence in what's
// *accepted* is tolerated.
func TestSourceSideNegativeMaxClauseTolerated(t *testing.T) {
	script := `send [EUR/2 100] (
  source = {
    max [EUR/2 0] - [EUR/2 35] from @acc2 allowing unbounded overdraft
    @acc3 allowing unbounded overdraft
  }
  destination = @acc1
)`

	ctx := context.Background()
	newRes := runNew(ctx, script, nil, nil, nil)
	oracleRes := runOracle(ctx, script, nil, nil, nil)

	for _, pair := range []struct {
		name string
		v    Verdict
	}{
		{"new vs oracle", Compare(script, newRes, oracleRes, "new interpreter", "oracle")},
	} {
		if pair.v.Mismatch {
			t.Errorf("%s: unexpected mismatch: %s\nnew: %+v\noracle: %+v",
				pair.name, pair.v.Reason, newRes, oracleRes)
		}
	}

	// Pin down *why* this is expected to be tolerated, so the test fails
	// loudly (for the right reason) if the underlying gap ever gets fixed,
	// instead of silently passing on a script that no longer exercises it.
	if oracleRes.RunErr == "" {
		t.Fatalf("expected the oracle to reject this script; it didn't: %+v", oracleRes)
	}
	if newRes.Failed() {
		t.Fatalf("expected the interpreter to succeed; got new=%+v", newRes)
	}
}

// TestMissingFundsClassificationMismatchStillCaught is
// TestSourceSideNegativeMaxClauseTolerated's companion: the same underlying
// gap, but without a fallback source, so the zeroed-out clause leaves
// nothing to cover the send and the interpreter fails — specifically
// with a missing-funds error, whereas the oracle fails for an unrelated
// reason (the negative amount itself). This asymmetry is exactly what
// Compare's relaxation does NOT tolerate (see SideResult.MissingFunds and
// Compare's doc comments): it must still be flagged as a mismatch.
func TestMissingFundsClassificationMismatchStillCaught(t *testing.T) {
	script := `send [EUR/2 100] (
  source = max [EUR/2 0] - [EUR/2 35] from @acc2 allowing unbounded overdraft
  destination = @acc1
)`
	ctx := context.Background()
	newRes := runNew(ctx, script, nil, nil, nil)
	oracleRes := runOracle(ctx, script, nil, nil, nil)

	if !oracleRes.Failed() || oracleRes.MissingFunds {
		t.Fatalf("expected the oracle to fail for a non-missing-funds reason; got %+v", oracleRes)
	}
	if !newRes.MissingFunds {
		t.Fatalf("expected the interpreter to fail specifically due to missing funds; got new=%+v", newRes)
	}

	if v := Compare(script, newRes, oracleRes, "new interpreter", "oracle"); !v.Mismatch {
		t.Fatalf("expected new-vs-oracle to be flagged as a mismatch, got none")
	}
}

// TestKnownOpenDivergences pins the numscript/ledger disagreements that are
// real and still undecided — see internal/oracle/DIVERGENCES.md §3. None of
// them is an oracle defect: the oracle is faithful to ledger on all three, and
// each was previously invisible because the oracle had been bent toward
// numscript to hide it.
//
// These assert that a mismatch IS still reported. If one starts passing,
// something changed the semantics — update DIVERGENCES.md and move the case
// into TestKnownBugRepros rather than deleting it.
func TestKnownOpenDivergences(t *testing.T) {
	testCases := []struct {
		name     string
		script   string
		balances map[gen.BalanceKey]*big.Int
		why      string
	}{
		{
			// DIVERGENCES.md §3.2. numscript's runSaveStatement floors the
			// saved amount at the balance; ledger subtracts unfloored and
			// goes negative. Invisible until a later bounded-overdraft draw
			// computes its available room from the two different balances.
			name: "save beyond the account's balance",
			why:  "numscript floors save at zero, ledger goes negative",
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
		{
			// DIVERGENCES.md §3.1. `kept` decides which account keeps the
			// money, so it changes balances: numscript takes it off the
			// front of the pool (@acc1 keeps 400), ledger off the bottom
			// (@acc0 keeps it, @acc1 is drained). The second statement then
			// moves 400 on numscript and nothing on ledger.
			name: "kept attribution, observed by a later statement",
			why:  "kept comes off the front (numscript) vs the bottom (ledger)",
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "acc0", Asset: "COIN"}: big.NewInt(1000),
				{Account: "acc1", Asset: "COIN"}: big.NewInt(1000),
			},
			script: `send [COIN *] (
  source = {
    @acc1
    @acc0
  }
  destination = {
    max [COIN 400] kept
    remaining to @dst
  }
)

send [COIN *] (
  source = @acc1
  destination = @sink
)`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			newRes := runNew(ctx, tc.script, nil, tc.balances, nil)
			oracleRes := runOracle(ctx, tc.script, nil, tc.balances, nil)

			v := Compare(tc.script, newRes, oracleRes, "new interpreter", "oracle")
			if !v.Mismatch {
				t.Fatalf("expected a divergence (%s), got none\nnew: %+v\noracle: %+v",
					tc.why, newRes, oracleRes)
			}
			t.Logf("still diverging, as expected (%s): %s", tc.why, v.Reason)
		})
	}
}
