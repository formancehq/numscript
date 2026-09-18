package gen

import (
	"hash/fnv"
	"math/big"
	"math/rand"

	"github.com/formancehq/numscript/builder"
)

// RandFromBytes derives an RNG from arbitrary bytes, so a Go fuzz target can
// drive generation reproducibly from its []byte corpus.
func RandFromBytes(b []byte) *rand.Rand {
	h := fnv.New64a()
	_, _ = h.Write(b) // hash.Hash.Write never errors
	return rand.New(rand.NewSource(int64(h.Sum64())))
}

// GenerateScript formats a random Script as a runnable numscript program, with
// the vars bindings it needs at run time and the starting balances and metadata
// it assumes. The caller loads those into each engine store before executing.
func GenerateScript(rng *rand.Rand) (vars map[string]string, balances map[BalanceKey]*big.Int, metadata map[MetaKey]string, script string) {
	s := generateScriptAST(rng)
	stmts, accountVarFills := toBuilderScript(s)

	var varsEnv builder.VarsEnv
	vars, varsEnv, script = builder.BuildProgram(stmts...)

	// Account-typed vars have no compiler-computed origin, so their value is
	// supplied here the way a real caller would.
	for _, f := range accountVarFills {
		name, value := varsEnv.FillAccount(f.Var, f.Value)
		if name != "" { // "" means this var was declared but never referenced
			vars[name] = value
		}
	}

	return vars, s.Balances, s.Metadata, script
}
