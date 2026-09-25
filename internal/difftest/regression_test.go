package difftest

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/gen"
)

// TestKnownBugRepros locks in the real divergences found and fixed while
// building this harness, independent of whether the fuzzer rediscovers them.
// Each case failed before its fix and must pass now.
func TestKnownBugRepros(t *testing.T) {
	testCases := []struct {
		name     string
		script   string
		vars     map[string]string
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
			// DIVERGENCES.md #3. numscript's runSaveStatement floors the saved
			// amount at the balance; ledger subtracts unfloored and goes negative,
			// so a later bounded-overdraft draw sees less room. The engines agree
			// because the oracle was changed to floor as well (2026-09-18), not
			// because ledger does: on ledger acc0 sits at -800 after the save, the
			// overdraft of 1000 leaves 200 of room and the send of 250 fails. Both
			// engines here move 250.
			name: "save beyond the account's balance",
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
			// Same root cause as above, from the other side of zero: a save of
			// nothing on a negative balance. numscript floors the result at zero,
			// which raises the balance from -50 to 0, so the whole overdraft is
			// available. Ledger leaves -50 and moves 50; both engines here move 100.
			name: "save on a negative balance",
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "acc0", Asset: "COIN"}: big.NewInt(-50),
			},
			script: `save [COIN 0] from @acc0

send [COIN *] (
  source = @acc0 allowing overdraft up to [COIN 100]
  destination = @acc1
)`,
		},
		{
			// DIVERGENCES.md #6, destination side. Portions bound past 100% with a
			// `remaining` clause. The interpreter used to compute a negative
			// remaining portion and commit world->acc1 60, world->acc2 30 where the
			// oracle fails OP_MAKE_ALLOTMENT ("sum of portions exceeded 100%");
			// fixed by rejecting the negative remaining in makeAllotment
			// (InvalidAllotmentSum), so both engines now fail for a non-funds
			// reason. The literal form never reached Compare either way: the
			// oracle rejects it at compile time, which is tolerated. The generator
			// keeps every emitted sum below 100% (portionsList), so only this test
			// reaches the shape.
			name: "allotment portions above 100% with remaining, destination side",
			vars: map[string]string{"p": "2/3", "q": "2/3"},
			script: `vars {
  portion $p
  portion $q
}

send [COIN 90] (
  source = @world
  destination = {
    $p to @acc1
    $q to @acc2
    remaining to @acc3
  }
)`,
		},
		{
			// DIVERGENCES.md #6, source side. The same parts [60, 60, -30] used to
			// reach tryTakingExact(-30) and report missing funds ("Needed
			// [COIN -30]") although every account holds 500 -- a classification
			// mismatch against the oracle's non-funds rejection. Same fix.
			name: "allotment portions above 100% with remaining, source side",
			vars: map[string]string{"p": "2/3", "q": "2/3"},
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "acc1", Asset: "COIN"}: big.NewInt(500),
				{Account: "acc2", Asset: "COIN"}: big.NewInt(500),
				{Account: "acc3", Asset: "COIN"}: big.NewInt(500),
			},
			script: `vars {
  portion $p
  portion $q
}

send [COIN 90] (
  source = {
    $p from @acc1
    $q from @acc2
    remaining from @acc3
  }
  destination = @acc4
)`,
		},
		{
			// Ledger does not consume a kept funding, the interpreter does. The
			// oracle was changed to consume it too, deliberately diverging from
			// ledger -- DIVERGENCES.md #2. Here @acc1 keeps 400; on ledger @acc0
			// keeps it and @acc1 is drained, so the second statement moves 400 here
			// and 0 there.
			name: "kept attribution, observed by a later statement",
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
			newRes := runNew(ctx, tc.script, tc.vars, tc.balances, nil)
			oracleRes := runOracle(ctx, tc.script, tc.vars, tc.balances, nil)
			vmRes := runVM(ctx, tc.script, tc.vars, tc.balances, nil)

			if v := Compare(newRes, oracleRes, "new interpreter", "oracle"); v.Mismatch {
				t.Fatalf("mismatch: %s\nnew: %+v\noracle: %+v", v.Reason, newRes, oracleRes)
			}
			if v := Compare(vmRes, oracleRes, "vm", "oracle"); v.Mismatch {
				t.Fatalf("mismatch: %s\nvm: %+v\noracle: %+v", v.Reason, vmRes, oracleRes)
			}
			if v := Compare(vmRes, newRes, "vm", "new interpreter"); v.Mismatch {
				t.Fatalf("mismatch: %s\nvm: %+v\nnew: %+v", v.Reason, vmRes, newRes)
			}
		})
	}
}

// TestDestinationSideNegativeMaxTolerated pins the destination-side twin of
// DIVERGENCES.md #4: the interpreter clamps a negative `max` destination clause
// to zero (sendTo, *parser.DestinationInorder) and routes the whole amount
// through `remaining`; the oracle rejects the script in OP_TAKE_MAX. Compare
// does not flag it: neither side reports missing funds, and the rejection is
// ledger's OP_TAKE_MAX guard, so the named "negative max clause" tolerance
// fires instead of a mismatch. This test is the only check on the shape, so it
// asserts the exact asymmetry rather than just "no mismatch".
func TestDestinationSideNegativeMaxTolerated(t *testing.T) {
	script := `send [EUR/2 100] (
  source = @acc2 allowing unbounded overdraft
  destination = {
    max [EUR/2 0] - [EUR/2 35] to @acc1
    remaining to @acc2
  }
)`
	ctx := context.Background()
	newRes := runNew(ctx, script, nil, nil, nil)
	oracleRes := runOracle(ctx, script, nil, nil, nil)

	v := Compare(newRes, oracleRes, "new interpreter", "oracle")
	if v.Mismatch {
		t.Fatalf("unexpected mismatch: %s\nnew: %+v\noracle: %+v", v.Reason, newRes, oracleRes)
	}
	if v.Tolerated != "negative max clause" {
		t.Fatalf("expected the negative max clause tolerance to fire, got %+v\nnew: %+v\noracle: %+v", v, newRes, oracleRes)
	}
	if newRes.Failed() {
		t.Fatalf("expected the interpreter to clamp and succeed; got %+v", newRes)
	}
	if oracleRes.RunErr == "" || oracleRes.MissingFunds {
		t.Fatalf("expected the oracle to reject the negative max for a non-missing-funds reason; got %+v", oracleRes)
	}

	// The vm clamps like the interpreter, so the same tolerance fires on its leg.
	vmRes := runVM(ctx, script, nil, nil, nil)
	if vmRes.Failed() {
		t.Fatalf("expected the vm to clamp and succeed; got %+v", vmRes)
	}
	if v := Compare(vmRes, oracleRes, "vm", "oracle"); v.Mismatch || v.Tolerated != "negative max clause" {
		t.Fatalf("expected the negative max clause tolerance on the vm leg, got %+v\nvm: %+v", v, vmRes)
	}
}

// TestSourceSideNegativeMaxClauseTolerated locks in an accepted gap: a negative
// amount in a source-side `max ... from` clause is a hard reject on the oracle
// but contributes zero on the interpreter (DIVERGENCES.md #4). Left unfixed
// because closing it means changing the interpreter's ground truth, not just
// catching up.
//
// With a fallback source in the list the interpreter covers the shortfall and
// succeeds where the oracle never gets that far. SideResult.MissingFunds and
// Compare's classification tolerance exist so this is not fuzzer noise.
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

	if v := Compare(newRes, oracleRes, "new interpreter", "oracle"); v.Mismatch {
		t.Errorf("unexpected mismatch: %s\nnew: %+v\noracle: %+v", v.Reason, newRes, oracleRes)
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

	// The vm must side with the interpreter on the gap.
	vmRes := runVM(ctx, script, nil, nil, nil)
	if v := Compare(vmRes, newRes, "vm", "new interpreter"); v.Mismatch {
		t.Fatalf("the vm does not side with the interpreter: %s\nvm: %+v\nnew: %+v", v.Reason, vmRes, newRes)
	}
}

// TestMissingFundsClassificationMismatchStillCaught is the companion to
// TestSourceSideNegativeMaxClauseTolerated: same gap, no fallback source, so
// the interpreter fails with a missing-funds error while the oracle fails over
// the negative amount itself. Compare must still flag that asymmetry.
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

	if v := Compare(newRes, oracleRes, "new interpreter", "oracle"); !v.Mismatch {
		t.Fatalf("expected new-vs-oracle to be flagged as a mismatch, got none")
	}

	// Same classification on the vm, so its oracle leg is flagged too.
	vmRes := runVM(ctx, script, nil, nil, nil)
	if !vmRes.MissingFunds {
		t.Fatalf("expected the vm to fail specifically due to missing funds; got vm=%+v", vmRes)
	}
	if v := Compare(vmRes, oracleRes, "vm", "oracle"); !v.Mismatch {
		t.Fatalf("expected vm-vs-oracle to be flagged as a mismatch, got none")
	}
}

// TestKnownOpenDivergences pins the numscript/ledger disagreements that are
// real and still undecided — internal/oracle/DIVERGENCES.md. Neither is an
// oracle defect: the oracle is faithful to ledger on both.
//
// These assert that a mismatch IS still reported. If one starts passing,
// something changed the semantics — update DIVERGENCES.md and move the case
// into TestKnownBugRepros rather than deleting it.
func TestKnownOpenDivergences(t *testing.T) {
	testCases := []struct {
		name     string
		script   string
		vars     map[string]string
		balances map[gen.BalanceKey]*big.Int
		why      string
	}{
		{
			// DIVERGENCES.md #5. A bounded overdraft written as a negative
			// expression. numscript clamps the cap to zero (tryTakingUpTo and
			// takeAll, *parser.SourceOverdraft) and moves the 50 that is there;
			// ledger adds -10 to the balance in withdrawAll and moves 40. The
			// generator never emits a negative cap, so only this test reaches it.
			name: "negative bounded overdraft cap",
			why:  "numscript clamps the cap to zero, ledger applies it as-is",
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "acc0", Asset: "COIN"}: big.NewInt(50),
			},
			script: `send [COIN *] (
  source = @acc0 allowing overdraft up to [COIN 0] - [COIN 10]
  destination = @acc1
)`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			newRes := runNew(ctx, tc.script, tc.vars, tc.balances, nil)
			oracleRes := runOracle(ctx, tc.script, tc.vars, tc.balances, nil)
			vmRes := runVM(ctx, tc.script, tc.vars, tc.balances, nil)

			v := Compare(newRes, oracleRes, "new interpreter", "oracle")
			if !v.Mismatch {
				t.Fatalf("expected a divergence (%s), got none\nnew: %+v\noracle: %+v",
					tc.why, newRes, oracleRes)
			}
			t.Logf("still diverging, as expected (%s): %s", tc.why, v.Reason)

			// The divergence is numscript-vs-ledger; within numscript the two
			// engines must still agree on it.
			if v := Compare(vmRes, newRes, "vm", "new interpreter"); v.Mismatch {
				t.Fatalf("the vm does not side with the interpreter: %s\nvm: %+v\nnew: %+v", v.Reason, vmRes, newRes)
			}
		})
	}
}
