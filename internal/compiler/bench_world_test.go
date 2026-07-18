package compiler_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
)

// send [USD/2 42] (source = @world ; destination = @dest)
const benchSrcWorld = `send [USD/2 42] (
	source = @world
	destination = @dest
)`

func worldStore() benchStore {
	// @world is unbounded; no balance entry needed.
	return benchStore{balances: map[runtime.PairKey]*big.Int{}}
}

// TestDumpWorld prints the naive and optimized bytecode for the world script so
// we can see exactly what the funds-bypass peephole leaves behind.
func TestDumpWorld(t *testing.T) {
	parsed := parser.Parse(benchSrcWorld)
	if len(parsed.Errors) != 0 {
		t.Fatalf("parse: %v", parsed.Errors)
	}
	_, naive, _ := compiler.Compile(parsed.Value)
	_, opt, _ := compiler.CompileWithOptimizations(parsed.Value)

	fmt.Println("=== NAIVE ===")
	for i, ins := range naive.Instructions {
		fmt.Printf("  %2d  op=%-3d A=%d B=%d C=%d\n", i, ins.Opcode, ins.A, ins.B, ins.C)
	}
	fmt.Println("=== OPT ===")
	for i, ins := range opt.Instructions {
		fmt.Printf("  %2d  op=%-3d A=%d B=%d C=%d\n", i, ins.Opcode, ins.A, ins.B, ins.C)
	}
}

func BenchmarkWorldNaive(b *testing.B) {
	benchCompiledVMNaive(b, benchSrcWorld, worldStore())
}

func BenchmarkWorldOpt(b *testing.B) {
	benchCompiledVMOpt(b, benchSrcWorld, worldStore())
}

// BenchmarkWorldBaselineTakePost is the floor of the CURRENT bypass path: a Take
// (unbounded world: set cap + record the world-balance debit delta) followed by
// a PostDirect. This is what Op_Take + Op_Post lower to, minus VM dispatch.
func BenchmarkWorldBaselineTakePost(b *testing.B) {
	store := runtimeStoreAdapter{store: worldStore()}
	rs := runtime.New(store)
	fortyTwo := big.NewInt(42)
	out := new(big.Int)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs.Reset(store)
		rs.SetCurrentAsset("USD/2")
		_ = rs.Take(out, "world", "", fortyTwo, nil, "") // unbounded take
		_ = rs.PostDirect("world", "", "dest", "", "", out)
		_ = rs.PostingsRef()
	}
}

// BenchmarkWorldBaselineDirectPost is the PROPOSED floor: a single fused
// "emit posting from @world" — no Take, no out register, no world-balance delta
// tracking. Just append the {world -> dest, amount} posting. This is what a
// hypothetical Op_PostFromWorld would do.
func BenchmarkWorldBaselineDirectPost(b *testing.B) {
	store := runtimeStoreAdapter{store: worldStore()}
	rs := runtime.New(store)
	fortyTwo := big.NewInt(42)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs.Reset(store)
		rs.SetCurrentAsset("USD/2")
		_ = rs.PostDirect("world", "", "dest", "", "", fortyTwo)
		_ = rs.PostingsRef()
	}
}
