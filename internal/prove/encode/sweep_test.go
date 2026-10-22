package encode

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"math/rand"
	"path/filepath"
	"sort"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/gen"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/stretchr/testify/require"
)

const scriptsFolder = "../../interpreter/testdata/script-tests"

type coverage struct {
	encoded int
	cases   int
	skipped map[string]int
}

func (c *coverage) skip(reason string) {
	if c.skipped == nil {
		c.skipped = map[string]int{}
	}
	c.skipped[reason]++
}

func (c *coverage) log(t *testing.T, what string) {
	t.Helper()
	reasons := make([]string, 0, len(c.skipped))
	for r := range c.skipped {
		reasons = append(reasons, r)
	}
	sort.Strings(reasons)
	t.Logf("%s: %d encoded, %d cases checked", what, c.encoded, c.cases)
	for _, r := range reasons {
		t.Logf("  skipped %d: %s", c.skipped[r], r)
	}
}

func beforeOf(rows interpreter.Balances) (map[Key]*big.Int, bool) {
	out := map[Key]*big.Int{}
	for _, row := range rows {
		if row.Color != "" || row.Scope != "" {
			return nil, false
		}
		amount := row.Amount
		if amount == nil {
			amount = new(big.Int)
		}
		out[Key{row.Account, row.Asset}] = amount
	}
	return out, true
}

// Every spec script the encoder accepts must agree with the VM on every test
// case's balances.
func TestSpecCorpusAgreesWithVM(t *testing.T) {
	rawSpecs, err := specs_format.ReadSpecsFiles([]string{scriptsFolder})
	require.NoError(t, err)

	var cov coverage
	for _, rawSpec := range rawSpecs {
		rel, err := filepath.Rel(scriptsFolder, rawSpec.NumscriptPath)
		require.NoError(t, err)

		var specs specs_format.Specs
		require.NoError(t, json.Unmarshal(rawSpec.SpecsFileContent, &specs))

		parsed := parser.Parse(rawSpec.NumscriptContent)
		require.Empty(t, parsed.Errors, rel)
		flags := map[string]struct{}{}
		for _, f := range specs.FeatureFlags {
			flags[f] = struct{}{}
		}
		instrs, _, err := compiler.CompileToIR(parsed.Value, flags)
		if err != nil {
			cov.skip("compile: " + err.Error())
			continue
		}
		p, err := Encode(instrs, nil, nil)
		var unsupported UnsupportedError
		if errors.As(err, &unsupported) {
			cov.skip(unsupported.Reason)
			continue
		}
		require.NoError(t, err, rel)
		cov.encoded++

		for _, tc := range specs.TestCases {
			before, ok := beforeOf(specs_format.MergeBalances(specs.Balances, tc.Balances))
			if !ok {
				cov.skip("colored or scoped balances")
				continue
			}
			t.Run(rel+"/"+tc.It, func(t *testing.T) {
				model := p.OutcomeFromBalances(before)
				replay, err := Replay(context.Background(), instrs, before, p.Keys, p.MoveKeys)
				require.NoError(t, err)
				require.Empty(t, Compare(model, replay))
			})
			cov.cases++
		}
	}
	cov.log(t, "spec corpus")
	require.Greater(t, cov.encoded, 40)
}

// Generated scripts lean on vars and metadata, which v1 rejects; the ones
// that encode must still agree with the VM.
func TestGeneratedScriptsAgreeWithVM(t *testing.T) {
	var cov coverage
	for seed := int64(0); seed < 200; seed++ {
		g := gen.Generate(rand.New(rand.NewSource(seed)))
		parsed := parser.Parse(g.Script)
		if len(parsed.Errors) != 0 {
			cov.skip("parse error")
			continue
		}
		flags := map[string]struct{}{}
		for _, f := range g.Flags {
			flags[f] = struct{}{}
		}
		instrs, _, err := compiler.CompileToIR(parsed.Value, flags)
		if err != nil {
			cov.skip("compile error")
			continue
		}
		p, err := Encode(instrs, nil, nil)
		var unsupported UnsupportedError
		if errors.As(err, &unsupported) {
			cov.skip(unsupported.Reason)
			continue
		}
		require.NoError(t, err, "seed %d", seed)
		cov.encoded++

		before := map[Key]*big.Int{}
		for k, v := range g.Balances {
			before[Key{k.Account, k.Asset}] = v
		}
		model := p.OutcomeFromBalances(before)
		replay, err := Replay(context.Background(), instrs, before, p.Keys, p.MoveKeys)
		require.NoError(t, err, "seed %d", seed)
		require.Empty(t, Compare(model, replay), "seed %d:\n%s", seed, g.Script)
		cov.cases++
	}
	cov.log(t, "generated scripts")
}
