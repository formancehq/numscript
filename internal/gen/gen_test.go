package gen_test

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/formancehq/numscript/internal/gen"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestPortionsListSumsToOne(t *testing.T) {
	// portionsList isn't exported; exercise it indirectly via many
	// generated programs and check every allotment clause set found sums
	// to 1 (mirrors TestGen.hs's own portionsList property).
	for seed := range 200 {
		rng := rand.New(rand.NewSource(int64(seed)))
		p := gen.GenerateProgram(rng)
		for _, stmt := range p {
			checkSourcePortions(t, stmt.Source)
			checkDestinationPortions(t, stmt.Destination)
		}
	}
}

func checkSourcePortions(t *testing.T, s gen.Source) {
	t.Helper()
	switch s.Kind {
	case gen.SrcCapped:
		checkSourcePortions(t, *s.Inner)
	case gen.SrcInorder:
		for _, inner := range s.Sources {
			checkSourcePortions(t, inner)
		}
	case gen.SrcAllotment:
		total := new(big.Rat)
		for _, c := range s.Clauses {
			total.Add(total, c.Portion)
			checkSourcePortions(t, c.Source)
		}
		require.True(t, total.Cmp(big.NewRat(1, 1)) == 0, "source allotment portions must sum to 1, got %s", total)
	}
}

func checkDestinationPortions(t *testing.T, d gen.Destination) {
	t.Helper()
	switch d.Kind {
	case gen.DestInorder:
		for _, c := range d.InorderClauses {
			checkKeptOrDestPortions(t, c.KeptOrDest)
		}
		if d.Remaining != nil {
			checkKeptOrDestPortions(t, *d.Remaining)
		}
	case gen.DestAllotment:
		total := new(big.Rat)
		for _, c := range d.AllotClauses {
			total.Add(total, c.Portion)
			checkKeptOrDestPortions(t, c.KeptOrDest)
		}
		require.True(t, total.Cmp(big.NewRat(1, 1)) == 0, "destination allotment portions must sum to 1, got %s", total)
	}
}

func checkKeptOrDestPortions(t *testing.T, k gen.KeptOrDest) {
	t.Helper()
	if k.Kind == gen.To {
		checkDestinationPortions(t, *k.Dest)
	}
}

func TestGenerateScriptParses(t *testing.T) {
	// Every generated script must be syntactically valid numscript. This
	// doesn't guarantee it also compiles against the oracle (that's the
	// difftest harness's job), but a parse failure here would mean the
	// generator/cleanup pass produced something structurally broken.
	for seed := range 500 {
		rng := rand.New(rand.NewSource(int64(seed)))
		_, script := gen.GenerateScript(rng)

		res := parser.Parse(script)
		require.Empty(t, res.Errors, "seed %d produced an unparseable script:\n%s", seed, script)
	}
}
