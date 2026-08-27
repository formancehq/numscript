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
// bindings map it needs at runtime and the pre-set starting balances/
// metadata it assumes (keyed by (account, asset) / (account, key); see
// internal/difftest's runNew/runOracle, which feed these into each engine's
// StaticStore before executing).
func GenerateScript(rng *rand.Rand) (vars map[string]string, balances map[BalanceKey]*big.Int, metadata map[MetaKey]string, script string) {
	s := GenerateScriptAST(rng)
	stmts, accountVarFills := ToBuilderScript(s)

	var varsEnv builder.VarsEnv
	vars, varsEnv, script = builder.BuildProgram(stmts...)

	// Plain account-typed vars (see AccountVarFill) have no compiler-computed
	// origin, so — unlike VarFromBalance/VarFromMeta vars — their value must
	// be supplied here, the same way a real caller would.
	for _, f := range accountVarFills {
		name, value := varsEnv.FillAccount(f.Var, f.Value)
		if name != "" { // "" means this var was declared but never referenced
			vars[name] = value
		}
	}

	return vars, s.Balances, s.Metadata, script
}
