package gen

import (
	"hash/fnv"
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

// GenerateScript generates a funding-seed program (world -> acc0..accN)
// followed by a random program (send-only, matching Gen.hs's scope),
// formatted as a single runnable script plus the vars bindings map it needs
// at runtime. Seeding and generation share one builder pool/vars block, so
// the whole thing is one script both engines can run as-is.
func GenerateScript(rng *rand.Rand) (vars map[string]string, script string) {
	stmts := append(ToBuilder(GenerateSeeds(rng)), ToBuilder(GenerateProgram(rng))...)
	vars, _, script = builder.BuildProgram(stmts...)
	return vars, script
}
