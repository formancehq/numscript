package difftest

import (
	"context"
	"math/big"
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/gen"
)

// TestNumscriptOnlyShapeAgreements pins NewVsVM agreement on shapes the oracle
// cannot parse, so only the two in-repo engines can check each other: oneof,
// colored sources, account interpolation, mid-script calls, division-expression
// portions and mixed-asset caps. The sweep reaches most of these since the
// generator learned them, but each case here is a deterministic witness of a
// specific behavior — several were live NewVsVM bugs when first probed
// (2026-09-25): the eager/lazy clause-evaluation family, the negative
// allotment share family, and the oneof covered-check under a negative cap.
//
// The specs fixture corpus (internal/interpreter/testdata/script-tests) pins
// the same families wherever the expectation is postings or a typed error;
// cases whose expected outcome is "both engines reject, same classification"
// live here, since the specs format cannot express generic errors.
func TestNumscriptOnlyShapeAgreements(t *testing.T) {
	testCases := []struct {
		name     string
		script   string
		flags    []string
		vars     map[string]string
		balances map[gen.BalanceKey]*big.Int
		metadata map[gen.MetaKey]string
		// bothReject: the agreement is that neither engine commits anything.
		bothReject bool
	}{
		{
			// The interpreter evaluates every source-inorder clause even after
			// the cap is exhausted; the vm once jumped out early and committed
			// where the interpreter errors on the wrong-asset cap.
			name:       "exhausted source-inorder clause still evaluates its cap",
			bothReject: true,
			balances:   map[gen.BalanceKey]*big.Int{{Account: "a", Asset: "COIN"}: big.NewInt(100)},
			script: `send [COIN 10] (
  source = {
    max [COIN 10] from @a
    max [EUR 1] from @b
  }
  destination = @d
)`,
		},
		{
			// A oneof under a negative allotment share: the interpreter clamps
			// the requested amount at zero on entry, so the first branch
			// trivially covers it and later branches are never evaluated; the
			// vm once compared pulls against the raw negative cap and walked
			// into the second branch, whose over-100% allotment then errored
			// with the wrong classification (found by the sweep, seed 571).
			name:       "oneof under a negative allotment share stops at the first branch",
			flags:      []string{flags.ExperimentalOneofFeatureFlag},
			bothReject: true,
			vars:       map[string]string{"n": "-1"},
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "a", Asset: "COIN"}: big.NewInt(500),
				{Account: "b", Asset: "COIN"}: big.NewInt(500),
			},
			script: `vars {
  number $n
}

send [COIN 90] (
  source = {
    $n/3 from oneof {
      @a
      {
        3/2 from @b
        remaining from @b
      }
    }
    remaining from @b
  }
  destination = @d
)`,
		},
		{
			name:  "oneof source: first branch short, second covers",
			flags: []string{flags.ExperimentalOneofFeatureFlag},
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "a", Asset: "COIN"}: big.NewInt(5),
				{Account: "b", Asset: "COIN"}: big.NewInt(50),
			},
			script: `send [COIN 10] (
  source = oneof { @a @b }
  destination = @d
)`,
		},
		{
			name:       "oneof source: all branches short fails as missing funds",
			flags:      []string{flags.ExperimentalOneofFeatureFlag},
			bothReject: true,
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "a", Asset: "COIN"}: big.NewInt(5),
				{Account: "b", Asset: "COIN"}: big.NewInt(7),
			},
			script: `send [COIN 10] (
  source = oneof { @a @b }
  destination = @d
)`,
		},
		{
			// The first branch's partial pulls must be rolled back before the
			// second branch runs, or source attribution differs.
			name:  "oneof source: nested inorder branch rolls back",
			flags: []string{flags.ExperimentalOneofFeatureFlag},
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "a1", Asset: "COIN"}: big.NewInt(4),
				{Account: "a2", Asset: "COIN"}: big.NewInt(4),
				{Account: "b1", Asset: "COIN"}: big.NewInt(30),
			},
			script: `send [COIN 10] (
  source = oneof {
    { @a1 @a2 }
    { @b1 }
  }
  destination = @d
)`,
		},
		{
			name:  "oneof source under send-all takes the first branch only",
			flags: []string{flags.ExperimentalOneofFeatureFlag},
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "a", Asset: "COIN"}: big.NewInt(5),
				{Account: "b", Asset: "COIN"}: big.NewInt(50),
			},
			script: `send [COIN *] (
  source = oneof { @a @b }
  destination = @d
)`,
		},
		{
			name:  "oneof destination: clause choice by cap",
			flags: []string{flags.ExperimentalOneofFeatureFlag},
			script: `send [COIN 10] (
  source = @world
  destination = oneof {
    max [COIN 5] to @a
    max [COIN 10] to @b
    remaining to @c
  }
)`,
		},
		{
			name:  "oneof destination: no clause covers, remaining takes all",
			flags: []string{flags.ExperimentalOneofFeatureFlag},
			script: `send [COIN 10] (
  source = @world
  destination = oneof {
    max [COIN 5] to @a
    max [COIN 9] to @b
    remaining to @c
  }
)`,
		},
		{
			name:  "colored funds pulled from world stay colored downstream",
			flags: []string{flags.ExperimentalAssetColors},
			script: `send [COIN 30] (
  source = @world \ "RED"
  destination = @a
)

send [COIN 10] (
  source = @a \ "RED"
  destination = @b
)`,
		},
		{
			// A colored pull must not see the uncolored store balance: the
			// account has 25 RED and 7 uncolored, and only the 25 move.
			name:  "colored pull reads only the colored balance",
			flags: []string{flags.ExperimentalAssetColors},
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "b", Asset: "COIN"}: big.NewInt(7),
			},
			script: `send [COIN 30] (
  source = @world \ "RED"
  destination = @b
)

send [COIN *] (
  source = @b \ "RED"
  destination = @c
)`,
		},
		{
			name:       "invalid color via var rejected by both",
			flags:      []string{flags.ExperimentalAssetColors},
			bothReject: true,
			vars:       map[string]string{"c": "not-valid-color"},
			balances:   map[gen.BalanceKey]*big.Int{{Account: "a", Asset: "COIN"}: big.NewInt(25)},
			script: `vars {
  string $c
}

send [COIN 20] (
  source = @a \ $c
  destination = @b
)`,
		},
		{
			name:  "interpolated account from a number var",
			flags: []string{flags.ExperimentalAccountInterpolationFlag},
			vars:  map[string]string{"id": "42"},
			script: `vars {
  number $id
}

send [COIN 10] (
  source = @world
  destination = @users:$id
)`,
		},
		{
			name:       "invalid interpolated account rejected by both",
			flags:      []string{flags.ExperimentalAccountInterpolationFlag},
			bothReject: true,
			vars:       map[string]string{"id": "no spaces!"},
			script: `vars {
  string $id
}

send [COIN 10] (
  source = @world
  destination = @users:$id
)`,
		},
		{
			name:  "mid-script balance reflects earlier statements",
			flags: []string{flags.ExperimentalMidScriptFunctionCall},
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "dest", Asset: "COIN"}: big.NewInt(7),
			},
			script: `send [COIN 50] (
  source = @world
  destination = @dest
)

set_tx_meta("bal", balance(@dest, COIN))`,
		},
		{
			name:  "mid-script balance as a send amount",
			flags: []string{flags.ExperimentalMidScriptFunctionCall},
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "a", Asset: "COIN"}: big.NewInt(30),
			},
			script: `send [COIN 20] (
  source = @world
  destination = @a
)

send balance(@a, COIN) (
  source = @a
  destination = @b
)`,
		},
		{
			name:       "negative mid-script balance rejected by both",
			flags:      []string{flags.ExperimentalMidScriptFunctionCall},
			bothReject: true,
			balances:   map[gen.BalanceKey]*big.Int{{Account: "a", Asset: "COIN"}: big.NewInt(-5)},
			script:     `set_tx_meta("bal", balance(@a, COIN))`,
		},
		{
			name: "amounts beyond int64 round-trip the whole pipeline",
			script: `send [COIN 123456789012345678901234567890] (
  source = @world
  destination = @a
)`,
		},
		{
			// long strings through the vars payload, the string pools and the
			// program/vars codecs (runVM byte-stability check included)
			name: "long account name and long meta value round-trip encoding",
			vars: map[string]string{
				"acc": "acc-" + strings.Repeat("x", 400) + ":seg-" + strings.Repeat("y", 300),
			},
			balances: map[gen.BalanceKey]*big.Int{
				{Account: "acc-" + strings.Repeat("x", 400) + ":seg-" + strings.Repeat("y", 300), Asset: "COIN"}: big.NewInt(40),
			},
			script: `vars {
  account $acc
}

send [COIN 25] (
  source = $acc
  destination = @d
)

set_tx_meta("note", "` + strings.Repeat("z", 900) + `")`,
		},
		{
			name:       "negative monetary var send amount rejected by both",
			bothReject: true,
			vars:       map[string]string{"m": "COIN -5"},
			balances:   map[gen.BalanceKey]*big.Int{{Account: "a", Asset: "COIN"}: big.NewInt(100)},
			script: `vars {
  monetary $m
}

send $m (
  source = @a
  destination = @b
)`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			newRes := runNew(ctx, tc.script, tc.vars, tc.balances, tc.metadata, tc.flags)
			vmRes := runVM(ctx, tc.script, tc.vars, tc.balances, tc.metadata, tc.flags)

			if v := Compare(vmRes, newRes, "vm", "new interpreter"); v.Mismatch {
				t.Fatalf("vm mismatch: %s\nvm: %+v\nnew: %+v", v.Reason, vmRes, newRes)
			}
			if tc.bothReject {
				if !newRes.Failed() || !vmRes.Failed() {
					t.Fatalf("expected both engines to reject\nnew: %+v\nvm: %+v", newRes, vmRes)
				}
				return
			}
			if newRes.Failed() || vmRes.Failed() {
				t.Fatalf("expected both engines to complete\nnew: %+v\nvm: %+v", newRes, vmRes)
			}
		})
	}
}
