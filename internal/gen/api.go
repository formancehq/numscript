package gen

import (
	"hash/fnv"
	"math/big"
	"math/rand"

	"github.com/formancehq/numscript/builder"
)

// RandFromBytes deterministically derives an RNG from arbitrary bytes, so a
// Go fuzz target (which mutates a []byte corpus) can drive this package's
// generation reproducibly: the same bytes always yield the same program.
func RandFromBytes(b []byte) *rand.Rand {
	h := fnv.New64a()
	_, _ = h.Write(b) // hash.Hash.Write never errors
	return rand.New(rand.NewSource(int64(h.Sum64())))
}

// GenerateScript generates a full random Script (see GenerateScriptAST) and
// formats it as a single runnable numscript program, alongside the vars
// bindings map it needs at runtime and the pre-set starting balances it
// assumes (keyed by (account, asset); see internal/difftest's runNew/
// runOracle, which feed this into each engine's StaticStore before
// executing).
func GenerateScript(rng *rand.Rand) (vars map[string]string, balances map[BalanceKey]*big.Int, script string) {
	s := GenerateScriptAST(rng)
	stmts := ToBuilderScript(s)
	vars, _, script = builder.BuildProgram(stmts...)
	return vars, s.Balances, script
}
