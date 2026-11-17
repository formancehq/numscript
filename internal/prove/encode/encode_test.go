package encode

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/prove/smt"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func compileIR(t *testing.T, src string) []ir.Instr {
	t.Helper()
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	instrs, _, err := compiler.CompileToIR(parsed.Value, nil)
	require.NoError(t, err)
	return instrs
}

func parseIR(t *testing.T, src string) []ir.Instr {
	t.Helper()
	instrs, errs := ir.Parse(src)
	require.Empty(t, errs)
	return instrs
}

func requireZ3(t *testing.T) {
	t.Helper()
	if err := (smt.Solver{}).Available(); err != nil {
		if os.Getenv("NUMSCRIPT_REQUIRE_Z3") != "" {
			t.Fatal(err)
		}
		t.Skip(err)
	}
}

const simpleSend = `send [USD/2 10] (
  source = @src
  destination = @dest
)`

func TestSimpleSendText(t *testing.T) {
	p, err := Encode(compileIR(t, simpleSend), nil, nil)
	require.NoError(t, err)
	require.Equal(t, []Key{{"dest", "USD/2"}, {"src", "USD/2"}}, p.Keys)
	require.Equal(t, []MoveKey{{"src", "dest", "USD/2"}}, p.MoveKeys)
	require.Equal(t, []string{"dest", "src"}, p.Accounts)
	require.Equal(t, []string{"USD/2"}, p.Assets)
	require.Len(t, p.FailSites, 1)
	require.Equal(t, MissingFunds, p.FailSites[0].Kind)
	snaps.MatchSnapshot(t, p.Text())
}

// checkFixed evaluates the model under concrete balances and requires the
// real VM to agree on everything.
func checkFixed(t *testing.T, instrs []ir.Instr, before map[Key]*big.Int) Outcome {
	t.Helper()
	p, err := Encode(instrs, nil, nil)
	require.NoError(t, err)
	model := p.OutcomeFromBalances(before)
	replay, err := Replay(context.Background(), instrs, before, p.Keys, p.MoveKeys)
	require.NoError(t, err)
	require.Empty(t, Compare(model, replay), "model vs vm")
	return model
}

func bal(pairs ...any) map[Key]*big.Int {
	out := map[Key]*big.Int{}
	for i := 0; i < len(pairs); i += 3 {
		out[Key{pairs[i].(string), pairs[i+1].(string)}] = big.NewInt(int64(pairs[i+2].(int)))
	}
	return out
}

func TestSimpleSendFixedInputs(t *testing.T) {
	instrs := compileIR(t, simpleSend)
	out := checkFixed(t, instrs, bal("src", "USD/2", 25))
	require.True(t, out.Success)
	require.Equal(t, "10", out.Sent[Key{"src", "USD/2"}].String())
	require.Equal(t, "15", out.After[Key{"src", "USD/2"}].String())
	require.Equal(t, "10", out.After[Key{"dest", "USD/2"}].String())

	out = checkFixed(t, instrs, bal("src", "USD/2", 3))
	require.False(t, out.Success)
	require.Equal(t, MissingFunds, out.FailKind)
	require.Equal(t, "3", out.After[Key{"src", "USD/2"}].String())

	checkFixed(t, instrs, bal("src", "USD/2", -3))
	checkFixed(t, instrs, bal("src", "USD/2", 10))
}

func TestSimpleSendWithZ3(t *testing.T) {
	requireZ3(t)
	p, err := Encode(compileIR(t, simpleSend), nil, nil)
	require.NoError(t, err)
	solver := smt.Solver{Timeout: 5 * time.Second}
	src, dest := Key{"src", "USD/2"}, Key{"dest", "USD/2"}

	// success => after(src) = before(src) - 10
	prop := smt.Implies(p.Success, smt.Eq(p.After[src], smt.Sub(p.Before[src], smt.Int64(10))))
	res, err := solver.CheckSat(context.Background(), p.Text()+"(assert "+smt.Print(smt.Not(prop))+")\n", nil)
	require.NoError(t, err)
	require.Equal(t, smt.Unsat, res.Status)

	// after(dest) = before(dest) + 10 has a counterexample: the run fails
	prop = smt.Eq(p.After[dest], smt.Add(p.Before[dest], smt.Int64(10)))
	res, err = solver.CheckSat(context.Background(), p.Text()+"(assert "+smt.Print(smt.Not(prop))+")\n", append(p.InputNames(), p.OutputNames()...))
	require.NoError(t, err)
	require.Equal(t, smt.Sat, res.Status)
	before, model, err := p.OutcomeFromValues(res.Values)
	require.NoError(t, err)
	require.False(t, model.Success)
	replay, err := Replay(context.Background(), compileIR(t, simpleSend), before, p.Keys, p.MoveKeys)
	require.NoError(t, err)
	require.Empty(t, Compare(model, replay))
}
