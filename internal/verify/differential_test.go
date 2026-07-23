package verify

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	nsparser "github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

// capped exercises a mixed inorder source with a per-source cap.
const capped = `
	send [USD/2 10] (
		source = {
			@a
			max [USD/2 5] from @b
			@c
		}
		destination = @dest
	)
`

// Differential test: the symbolic encoding must agree with the real VM.
//
// For a script and a concrete assignment of starting balances, we
//  1. run the actual vm.Exec and read its postings (or failure), and
//  2. evaluate the symbolic encoding with the same balances pinned,
//
// then assert the two agree on `fail` and on every account's sent/received
// totals. This validates faithfulness against ground truth, not just against
// hand-written queries.

// --- an in-memory VM store --------------------------------------------------

type memStore map[string]*big.Int // key: account + "\x00" + asset

func balKey(account, asset string) string { return account + "\x00" + asset }

func (m memStore) GetBalance(_ context.Context, account, asset, _ string) (*big.Int, error) {
	if v, ok := m[balKey(account, asset)]; ok {
		return new(big.Int).Set(v), nil
	}
	return big.NewInt(0), nil
}

func (m memStore) GetMetadata(_ context.Context, _, _ string) (string, bool, error) {
	return "", false, nil
}

type diffCase struct {
	name     string
	src      string
	balances map[string]int64 // "account\x00asset" -> starting balance
}

func TestDifferentialAgainstVM(t *testing.T) {
	requireZ3(t)

	usd := func(account string, n int64) (string, int64) { return balKey(account, "USD/2"), n }
	bals := func(pairs ...struct {
		k string
		v int64
	}) map[string]int64 {
		m := map[string]int64{}
		for _, p := range pairs {
			m[p.k] = p.v
		}
		return m
	}
	kv := func(account string, n int64) struct {
		k string
		v int64
	} {
		k, v := usd(account, n)
		return struct {
			k string
			v int64
		}{k, v}
	}

	cases := []diffCase{
		{"simple/enough", simple, bals(kv("src", 10))},
		{"simple/exact", simple, bals(kv("src", 10))},
		{"simple/surplus", simple, bals(kv("src", 100))},
		{"simple/short", simple, bals(kv("src", 5))},
		{"simple/empty", simple, bals()},
		{"inorder/split", inorder, bals(kv("a", 3), kv("b", 3), kv("c", 10))},
		{"inorder/first-covers", inorder, bals(kv("a", 50), kv("b", 0), kv("c", 0))},
		{"inorder/short", inorder, bals(kv("a", 1), kv("b", 1), kv("c", 1))},
		{"world/split", world, bals()},
		{"capped/enough", capped, bals(kv("a", 2), kv("b", 100), kv("c", 100))},
		{"capped/b-capped-at-5", capped, bals(kv("a", 0), kv("b", 100), kv("c", 100))},
		{"capped/short", capped, bals(kv("a", 1), kv("b", 2), kv("c", 0))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vmFail, vmSent, vmRecv := runVM(t, tc.src, tc.balances)
			symFail, symSent, symRecv := runSymbolic(t, tc.src, tc.balances)

			require.Equal(t, vmFail, symFail, "fail flag disagrees (vm=%v sym=%v)", vmFail, symFail)
			require.Equal(t, vmSent, symSent, "sent totals disagree")
			require.Equal(t, vmRecv, symRecv, "received totals disagree")
		})
	}
}

// runVM executes the compiled program on the real VM and returns whether it
// failed, plus per-(account,asset) sent/received totals from the postings.
func runVM(t *testing.T, src string, balances map[string]int64) (failed bool, sent, recv map[string]int64) {
	t.Helper()
	pr := nsparser.Parse(src)
	require.Empty(t, pr.Errors)
	_, prog, err := compiler.Compile(pr.Value)
	require.NoError(t, err)

	store := memStore{}
	for k, v := range balances {
		store[k] = big.NewInt(v)
	}

	machine := vm.NewVm(prog)
	res, execErr := vm.Exec(context.Background(), machine, &vm.Vars{}, store)

	sent, recv = map[string]int64{}, map[string]int64{}
	if execErr != nil {
		return true, sent, recv
	}
	for _, p := range res.Postings {
		sent[balKey(p.Source, p.Asset)] += p.Amount.Int64()
		recv[balKey(p.Destination, p.Asset)] += p.Amount.Int64()
	}
	return false, sent, recv
}

// runSymbolic evaluates the symbolic encoding with the given balances pinned,
// returning the same shape as runVM by reading z3's model.
func runSymbolic(t *testing.T, src string, balances map[string]int64) (failed bool, sent, recv map[string]int64) {
	t.Helper()
	enc, err := compiler.SymbolicEncodeSource(src)
	require.NoError(t, err)

	// Build the ordered list of symbols to evaluate: fail, then each account's
	// sent/received totals. `probes` records what VM key each eval corresponds
	// to so we can rebuild the maps.
	var evals []string
	type probe struct {
		kind string // "sent" | "recv"
		key  string // balKey
	}
	var probes []probe
	evals = append(evals, enc.Symbols.Fail)

	for _, aa := range enc.Symbols.Order {
		key := balKey(aa.Account, aa.Asset)
		if s, ok := enc.Symbols.Sent[aa]; ok {
			evals = append(evals, s)
			probes = append(probes, probe{"sent", key})
		}
		if s, ok := enc.Symbols.Received[aa]; ok {
			evals = append(evals, s)
			probes = append(probes, probe{"recv", key})
		}
	}

	var b strings.Builder
	b.WriteString("(set-option :produce-models true)\n")
	b.WriteString(enc.SMTLIB)
	// pin each declared starting balance to its concrete value
	pinned := make([]string, 0, len(enc.Symbols.Start))
	for aa, symName := range enc.Symbols.Start {
		pinned = append(pinned, fmt.Sprintf("(assert (= %s %d))", symName, balances[balKey(aa.Account, aa.Asset)]))
	}
	sort.Strings(pinned) // deterministic
	for _, p := range pinned {
		b.WriteString(p)
		b.WriteByte('\n')
	}
	b.WriteString("(check-sat)\n")
	for _, e := range evals {
		fmt.Fprintf(&b, "(eval %s)\n", e)
	}

	sat, model, raw, err := runZ3(context.Background(), "", b.String())
	require.NoError(t, err, "z3:\n%s", raw)
	require.Equal(t, "sat", sat, "fully-pinned encoding must be sat; z3:\n%s", raw)

	values := strings.Split(strings.TrimSpace(model), "\n")
	require.Len(t, values, len(evals), "expected one eval result per symbol; z3:\n%s", raw)

	failed = strings.TrimSpace(values[0]) == "true"
	sent, recv = map[string]int64{}, map[string]int64{}
	for i, pr := range probes {
		v := parseSMTInt(t, values[i+1])
		if v == 0 {
			continue // omit zeros so the maps match the VM's (postings-only) maps
		}
		switch pr.kind {
		case "sent":
			sent[pr.key] += v
		case "recv":
			recv[pr.key] += v
		}
	}
	return failed, sent, recv
}

// parseSMTInt parses an SMT integer literal, which may be "(- 5)" for negatives.
func parseSMTInt(t *testing.T, s string) int64 {
	t.Helper()
	s = strings.TrimSpace(s)
	neg := false
	if strings.HasPrefix(s, "(-") {
		neg = true
		s = strings.TrimSpace(strings.TrimPrefix(s, "(-"))
		s = strings.TrimSpace(strings.TrimSuffix(s, ")"))
	}
	n := new(big.Int)
	_, ok := n.SetString(s, 10)
	require.True(t, ok, "cannot parse SMT int %q", s)
	if neg {
		n.Neg(n)
	}
	return n.Int64()
}
