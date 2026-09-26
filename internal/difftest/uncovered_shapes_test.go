package difftest

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/gen"
)

// TestUncoveredShapeAgreements pins agreement on shapes both grammars share
// but internal/gen does not emit, so the sweep says nothing about them:
// percent-form portions, balance()/meta() origin vars in cap and account
// positions, and portion sums that land exactly on 100% through vars. Each
// case ran green when the shape was first probed (2026-09-25); a regression in
// either engine's handling turns it into a mismatch here.
func TestUncoveredShapeAgreements(t *testing.T) {
	testCases := []struct {
		name     string
		script   string
		vars     map[string]string
		balances map[gen.BalanceKey]*big.Int
		metadata map[gen.MetaKey]string
		// bothReject: the agreement is that neither engine commits anything
		// (each rejects at some stage), rather than that both completed.
		bothReject bool
	}{
		{
			// The generator renders portions as n/d only, never p% — and a
			// decimal percent has its own parse path in both engines.
			name: "percent literal with decimals",
			script: `send [COIN 1000] (
  source = @world
  destination = {
    12.5% to @acc1
    remaining to @acc2
  }
)`,
		},
		{
			name: "percent literal on an indivisible amount",
			script: `send [COIN 99] (
  source = @world
  destination = {
    33% to @acc1
    remaining to @acc2
  }
)`,
		},
		{
			// Portion vars summing exactly to 100% next to a remaining clause:
			// the remaining portion is exactly zero at run time. The generator
			// keeps var sums strictly below 100% (portionsList), so only this
			// test reaches the boundary.
			name: "portion vars summing exactly 100% plus remaining",
			vars: map[string]string{"p": "1/3", "q": "2/3"},
			script: `vars {
  portion $p
  portion $q
}

send [COIN 100] (
  source = @world
  destination = {
    $p to @acc1
    $q to @acc2
    remaining to @acc3
  }
)`,
		},
		{
			// meta()-origin portion in allotment position; the generator only
			// uses meta-portion vars as set_tx_meta values.
			name:     "portion var from meta in allotment",
			metadata: map[gen.MetaKey]string{{Account: "acc0", Key: "k"}: "1/3"},
			script: `vars {
  portion $p = meta(@acc0, "k")
}

send [COIN 100] (
  source = @world
  destination = {
    $p to @acc1
    remaining to @acc2
  }
)`,
		},
		{
			// balance()-origin monetary in the three cap positions the
			// generator never routes it through (it only uses such vars as
			// send/save amounts).
			name: "balance var as source max cap",
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "acc0", Asset: "COIN"}: big.NewInt(37),
				{Account: "acc1", Asset: "COIN"}: big.NewInt(500),
			},
			script: `vars {
  monetary $b = balance(@acc0, COIN)
}

send [COIN 100] (
  source = {
    max $b from @acc1
    @world
  }
  destination = @acc2
)`,
		},
		{
			name: "balance var as bounded overdraft cap",
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "acc0", Asset: "COIN"}: big.NewInt(37),
				{Account: "acc1", Asset: "COIN"}: big.NewInt(10),
			},
			script: `vars {
  monetary $b = balance(@acc0, COIN)
}

send [COIN 30] (
  source = @acc1 allowing overdraft up to $b
  destination = @acc2
)`,
		},
		{
			name:     "balance var as destination max",
			balances: map[gen.BalanceKey]*big.Int{{Account: "acc0", Asset: "COIN"}: big.NewInt(37)},
			script: `vars {
  monetary $b = balance(@acc0, COIN)
}

send [COIN 100] (
  source = @world
  destination = {
    max $b to @acc1
    remaining to @acc2
  }
)`,
		},
		{
			// A negative balance cannot reach a cap through balance(): both
			// engines refuse the read itself (the oracle at resolve, the
			// interpreter at run), which is what keeps balance()-fed caps out
			// of DIVERGENCES.md #4/#5 territory.
			name:       "negative balance var is rejected by both engines",
			balances:   map[gen.BalanceKey]*big.Int{{Account: "acc0", Asset: "COIN"}: big.NewInt(-20)},
			bothReject: true,
			script: `vars {
  monetary $b = balance(@acc0, COIN)
}

send [COIN 100] (
  source = @world
  destination = {
    max $b to @acc1
    remaining to @acc2
  }
)`,
		},
		{
			// meta()-origin account vars in account position; the generator
			// only uses them as set_tx_meta values.
			name:     "meta account var as source",
			metadata: map[gen.MetaKey]string{{Account: "acc0", Key: "k"}: "acc1"},
			balances: map[gen.BalanceKey]*big.Int{{Account: "acc1", Asset: "COIN"}: big.NewInt(200)},
			script: `vars {
  account $a = meta(@acc0, "k")
}

send [COIN 100] (
  source = $a
  destination = @acc2
)`,
		},
		{
			name:     "meta account var as destination",
			metadata: map[gen.MetaKey]string{{Account: "acc0", Key: "k"}: "acc1"},
			script: `vars {
  account $a = meta(@acc0, "k")
}

send [COIN 100] (
  source = @world
  destination = $a
)`,
		},
		{
			name:     "meta monetary var as send amount",
			metadata: map[gen.MetaKey]string{{Account: "acc0", Key: "k"}: "COIN 55"},
			script: `vars {
  monetary $m = meta(@acc0, "k")
}

send $m (
  source = @world
  destination = @acc1
)`,
		},
		{
			name:     "meta asset var in a monetary literal",
			metadata: map[gen.MetaKey]string{{Account: "acc0", Key: "k"}: "COIN"},
			script: `vars {
  asset $as = meta(@acc0, "k")
}

send [$as 10] (
  source = @world
  destination = @acc1
)`,
		},
		{
			// `remaining` between two specific portions: the builder only
			// renders it last, but both grammars accept any position, and the
			// part order feeds each engine's front-first leftover top-up.
			name: "remaining clause mid-allotment",
			script: `send [COIN 100] (
  source = @world
  destination = {
    1/3 to @acc1
    remaining to @acc2
    1/3 to @acc3
  }
)`,
		},
		{
			// balance() of a runtime-bound var holding "world": a legal,
			// always-zero read on both engines. The generator reads
			// balance(@world, ...) inline but never routes world through a
			// var into balance().
			name:     "balance of a runtime-bound world var",
			vars:     map[string]string{"a": "world"},
			balances: map[gen.BalanceKey]*big.Int{{Account: "acc1", Asset: "COIN"}: big.NewInt(50)},
			script: `vars {
  account $a
  monetary $b = balance($a, COIN)
}

send $b (
  source = @acc1
  destination = @acc2
)`,
		},
		{
			// save with a meta()-origin monetary; the generator's saves only
			// take literals or balance()-origin vars.
			name:     "save with a meta monetary var, observed by a drain",
			metadata: map[gen.MetaKey]string{{Account: "acc0", Key: "k"}: "COIN 40"},
			balances: map[gen.BalanceKey]*big.Int{{Account: "acc1", Asset: "COIN"}: big.NewInt(100)},
			script: `vars {
  monetary $m = meta(@acc0, "k")
}

save $m from @acc1

send [COIN *] (
  source = @acc1
  destination = @acc2
)`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			newRes := runNew(ctx, tc.script, tc.vars, tc.balances, tc.metadata, nil)
			oracleRes := runOracle(ctx, tc.script, tc.vars, tc.balances, tc.metadata)
			vmRes := runVM(ctx, tc.script, tc.vars, tc.balances, tc.metadata, nil)

			v := Compare(newRes, oracleRes, "new interpreter", "oracle")
			if v.Mismatch {
				t.Fatalf("mismatch: %s\nnew: %+v\noracle: %+v", v.Reason, newRes, oracleRes)
			}
			if v.Tolerated != "" {
				t.Fatalf("nothing was compared (tolerated: %s)\nnew: %+v\noracle: %+v", v.Tolerated, newRes, oracleRes)
			}
			if v := Compare(vmRes, newRes, "vm", "new interpreter"); v.Mismatch {
				t.Fatalf("vm mismatch: %s\nvm: %+v\nnew: %+v", v.Reason, vmRes, newRes)
			}
			if tc.bothReject {
				if !newRes.Failed() || !oracleRes.Failed() || !vmRes.Failed() {
					t.Fatalf("expected all three engines to reject\nnew: %+v\noracle: %+v\nvm: %+v", newRes, oracleRes, vmRes)
				}
				return
			}
			if newRes.Failed() || oracleRes.Failed() || vmRes.Failed() {
				t.Fatalf("expected all three engines to complete\nnew: %+v\noracle: %+v\nvm: %+v", newRes, oracleRes, vmRes)
			}
		})
	}
}

// TestAllotmentFullSumPlusRemainingRejectedByOracleOnly pins an asymmetry the
// sweep can only ever count, not flag: literal portions summing to exactly
// 100% followed by a `remaining` clause. The oracle rejects the block at
// compile time ("known portions are already equal to 100%"), the interpreter
// runs it (the remaining clause receives zero and is trimmed). Compare
// tolerates the one-sided compile rejection by design, so the generator must
// not emit the shape — portionsList keeps sums strictly below 100% whenever a
// remaining clause is present.
func TestAllotmentFullSumPlusRemainingRejectedByOracleOnly(t *testing.T) {
	script := `send [COIN 100] (
  source = @world
  destination = {
    1/2 to @acc1
    1/2 to @acc2
    remaining to @acc3
  }
)`
	ctx := context.Background()
	newRes := runNew(ctx, script, nil, nil, nil, nil)
	oracleRes := runOracle(ctx, script, nil, nil, nil)

	v := Compare(newRes, oracleRes, "new interpreter", "oracle")
	if v.Mismatch {
		t.Fatalf("unexpected mismatch: %s\nnew: %+v\noracle: %+v", v.Reason, newRes, oracleRes)
	}
	if v.Tolerated != "b-side compile rejection" {
		t.Fatalf("expected the compile-rejection tolerance to fire, got %+v\nnew: %+v\noracle: %+v", v, newRes, oracleRes)
	}
	if newRes.Failed() {
		t.Fatalf("expected the interpreter to run the script; got %+v", newRes)
	}

	// The vm runs it like the interpreter: the remaining clause receives zero.
	vmRes := runVM(context.Background(), script, nil, nil, nil, nil)
	if vmRes.Failed() {
		t.Fatalf("expected the vm to run the script; got %+v", vmRes)
	}
	if v := Compare(vmRes, newRes, "vm", "new interpreter"); v.Mismatch {
		t.Fatalf("vm mismatch: %s\nvm: %+v\nnew: %+v", v.Reason, vmRes, newRes)
	}
}
