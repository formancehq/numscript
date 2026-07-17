package compiler_test

import (
	"context"
	"math/big"
	"strconv"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/vm"
)

// benchStore is a minimal vm.Store for the benchmarks.
type benchStore struct {
	balances map[runtime.PairKey]*big.Int
}

func (s benchStore) GetBalance(ctx context.Context, account, asset, color string) (*big.Int, error) {
	if v, ok := s.balances[runtime.PairKey{Account: account, Asset: asset, Color: color}]; ok {
		return v, nil
	}
	return new(big.Int), nil
}

func (benchStore) GetMetadata(ctx context.Context, k, v string) (string, bool, error) {
	return "", false, nil
}

type runtimeStoreAdapter struct {
	store vm.Store
}

func (s runtimeStoreAdapter) GetBalance(
	account string,
	asset string,
	color string,
) (*big.Int, error) {
	return s.store.GetBalance(context.Background(), account, asset, color)
}

// Both benchmarks run the SAME program with the same starting balance; only the
// per-iteration RUN is measured (parse/compile/assemble happen once, up front).
const benchSrc = `send [USD/2 10] (
	source = @src
	destination = @dest
)`

// BenchmarkTreeWalker measures the tree-walking interpreter on a pre-parsed AST.
func BenchmarkTreeWalker(b *testing.B) {
	parsed := parser.Parse(benchSrc)
	if len(parsed.Errors) != 0 {
		b.Fatalf("parse errors: %v", parsed.Errors)
	}
	store := interpreter.StaticStore{
		Balances: interpreter.Balances{
			{Account: "src", Asset: "USD/2", Amount: big.NewInt(100)},
		},
	}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := interpreter.RunProgram(ctx, parsed.Value, nil, store, nil)
		if err != nil {
			b.Fatalf("run: %v", err)
		}
	}
}

// BenchmarkRuntimeBaseline is the floor: it drives runtime.RunState directly,
// performing exactly the funds operations the program lowers to — with no AST
// walk and no bytecode dispatch. It reuses one RunState (like the VM reuses its
// runstate) and hoists the constants (the compiler would pool them). The gap
// between this and BenchmarkCompiledVM is the VM's dispatch/register overhead;
// the gap to BenchmarkTreeWalker is the interpreter's front-end overhead.
func BenchmarkRuntimeBaseline(b *testing.B) {
	store := runtimeStoreAdapter{
		store: benchStore{balances: map[runtime.PairKey]*big.Int{
			{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
		}},
	}

	rs := runtime.New(store)

	ten := big.NewInt(10)  // the sent amount / pull cap
	zero := big.NewInt(0)  // bounded overdraft of 0
	pulled := new(big.Int) // reused output register
	dest := "dest"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs.Reset(store)
		rs.SetCurrentAsset("USD/2")
		rs.Pull(pulled, "src", "", ten, zero, "")
		_ = pulled.Cmp(ten) // CheckEnoughFunds
		rs.SendUncapped(&dest, "", nil)
		_ = rs.PostingsRef()
	}
}

// BenchmarkRuntimeBaselineUnique is the WORST case for the balanceEntry
// generation cache: every iteration touches a brand-new source account, so the
// cache never hits — each run allocates a fresh entry (like the pre-cache impl)
// AND the map grows unbounded (the prototype has no eviction). Contrast with
// BenchmarkRuntimeBaseline (same hot accounts every run = best case) to bracket
// the cache's real-workload behavior.
func BenchmarkRuntimeBaselineUnique(b *testing.B) {
	store := runtimeStoreAdapter{store: benchStore{balances: map[runtime.PairKey]*big.Int{}}}
	rs := runtime.New(store)

	ten := big.NewInt(10)
	zero := big.NewInt(0)
	pulled := new(big.Int)
	dest := "dest"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src := "src" + strconv.Itoa(i) // unique every iteration -> always a cache miss
		rs.Reset(store)
		rs.SetCurrentAsset("USD/2")
		rs.Pull(pulled, src, "", ten, zero, "")
		_ = pulled.Cmp(ten)
		rs.SendUncapped(&dest, "", nil)
		_ = rs.PostingsRef()
	}
}

// BenchmarkCompiledVM measures the compiled bytecode on the register VM, reusing
// a single Vm instance across iterations (its register banks are not realloc'd).
func BenchmarkCompiledVM(b *testing.B) {
	parsed := parser.Parse(benchSrc)
	if len(parsed.Errors) != 0 {
		b.Fatalf("parse errors: %v", parsed.Errors)
	}
	_, program, err := compiler.Compile(parsed.Value)
	if err != nil {
		b.Fatalf("compile: %v", err)
	}
	store := benchStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	machine := vm.NewVm(program) // reused across iterations

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vm.Exec(context.Background(), machine, nil, store)
		if err != nil {
			b.Fatalf("exec: %v", err)
		}
	}
}

// --- Capped inorder script: `{ @a ; max [USD/2 5] from @b ; @c }` -----------
// Same methodology as above, on a more representative script (inorder traversal,
// a `max` cap with a min_int, running total, and an early-exit jump). Balances:
// a=3, b=100 (capped to 5), c=100 → pulls 3 / 5 / 2.
const benchSrcCapped = `send [USD/2 10] (
	source = {
		@a
		max [USD/2 5] from @b
		@c
	}
	destination = @dest
)`

func BenchmarkTreeWalkerCapped(b *testing.B) {
	parsed := parser.Parse(benchSrcCapped)
	if len(parsed.Errors) != 0 {
		b.Fatalf("parse errors: %v", parsed.Errors)
	}
	store := interpreter.StaticStore{
		Balances: interpreter.Balances{
			{Account: "a", Asset: "USD/2", Amount: big.NewInt(3)},
			{Account: "b", Asset: "USD/2", Amount: big.NewInt(100)},
			{Account: "c", Asset: "USD/2", Amount: big.NewInt(100)},
		},
	}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := interpreter.RunProgram(ctx, parsed.Value, nil, store, nil)
		if err != nil {
			b.Fatalf("run: %v", err)
		}
	}
}

func cappedStore() benchStore {
	return benchStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "a", Asset: "USD/2", Color: ""}: big.NewInt(3),
		{Account: "b", Asset: "USD/2", Color: ""}: big.NewInt(100),
		{Account: "c", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}
}

// BenchmarkRuntimeBaselineCapped is the floor: it drives runtime.RunState
// directly, performing the funds ops the capped-inorder script lowers to (with
// the cap/running-total/early-exit arithmetic done inline on reused big.Ints) —
// no AST walk, no bytecode dispatch. RunState reused across iterations.
func BenchmarkRuntimeBaselineCapped(b *testing.B) {
	store := runtimeStoreAdapter{store: cappedStore()}
	rs := runtime.New(store)

	zero := big.NewInt(0)
	ten := big.NewInt(10)
	five := big.NewInt(5)
	remaining := new(big.Int)
	capB := new(big.Int)
	pulled := new(big.Int)
	total := new(big.Int)
	dest := "dest"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs.Reset(store)
		rs.SetCurrentAsset("USD/2")
		total.SetInt64(0)
		remaining.Set(ten) // inorder cap = copy(amount)

		// @a (cap = remaining)
		rs.Pull(pulled, "a", "", remaining, zero, "")
		total.Add(total, pulled)
		remaining.Sub(remaining, pulled)

		if remaining.Sign() != 0 { // jmp_if_zero(remaining)
			// max [USD/2 5] from @b  ->  cap = min(5, remaining)
			if five.Cmp(remaining) < 0 {
				capB.Set(five)
			} else {
				capB.Set(remaining)
			}
			rs.Pull(pulled, "b", "", capB, zero, "")
			total.Add(total, pulled)
			remaining.Sub(remaining, pulled)

			if remaining.Sign() != 0 {
				rs.Pull(pulled, "c", "", remaining, zero, "") // @c (cap = remaining)
				total.Add(total, pulled)
			}
		}

		_ = total.Cmp(ten) // check_enough_funds
		rs.SendUncapped(&dest, "", nil)
		_ = rs.PostingsRef()
	}
}

func BenchmarkCompiledVMCapped(b *testing.B) {
	parsed := parser.Parse(benchSrcCapped)
	if len(parsed.Errors) != 0 {
		b.Fatalf("parse errors: %v", parsed.Errors)
	}
	_, program, err := compiler.Compile(parsed.Value)
	if err != nil {
		b.Fatalf("compile: %v", err)
	}
	store := cappedStore()

	machine := vm.NewVm(program) // reused across iterations

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vm.Exec(context.Background(), machine, nil, store)
		if err != nil {
			b.Fatalf("exec: %v", err)
		}
	}
}

// --- peephole-optimized VM (compile -> optimize -> assemble -> reused VM) ----

func benchCompiledVMOpt(b *testing.B, src string, store benchStore) {
	parsed := parser.Parse(src)
	if len(parsed.Errors) != 0 {
		b.Fatalf("parse errors: %v", parsed.Errors)
	}
	_, program, cErr := compiler.CompileWithOptimizations(parsed.Value)
	if cErr != nil {
		b.Fatalf("compile: %v", cErr)
	}
	machine := vm.NewVm(program)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := vm.Exec(context.Background(), machine, nil, store); err != nil {
			b.Fatalf("exec: %v", err)
		}
	}
}

func BenchmarkCompiledVMOpt(b *testing.B) {
	benchCompiledVMOpt(b, benchSrc, benchStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}})
}

func BenchmarkCompiledVMOptCapped(b *testing.B) {
	benchCompiledVMOpt(b, benchSrcCapped, cappedStore())
}

// Fan-out allotment: 1 source -> {1/2 @a; 1/2 @b}. Exercises MakeAllotment's
// big.Rat arithmetic and the (not-bypassed) queue drain across two capped sends.
const benchSrcAllotment = `send [USD/2 100] (
	source = @src
	destination = {
		1/2 to @a
		1/2 to @b
	}
)`

func BenchmarkCompiledVMOptAllotment(b *testing.B) {
	benchCompiledVMOpt(b, benchSrcAllotment, benchStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(1000),
	}})
}

// --- cold VM (fresh Vm per iteration): exposes the funds-queue allocation the
// funds-bypass saves, which a reused VM hides via its big.Int free pool. -------

func benchCompiledVMCold(b *testing.B, src string, store benchStore, optimize bool) {
	parsed := parser.Parse(src)
	if len(parsed.Errors) != 0 {
		b.Fatalf("parse errors: %v", parsed.Errors)
	}
	compile := compiler.Compile
	if optimize {
		compile = compiler.CompileWithOptimizations
	}
	_, program, err := compile(parsed.Value)
	if err != nil {
		b.Fatalf("compile: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		machine := vm.NewVm(program) // fresh runstate each iteration (no pooling)
		if _, err := vm.Exec(context.Background(), machine, nil, store); err != nil {
			b.Fatalf("exec: %v", err)
		}
	}
}

func BenchmarkCompiledVMCold(b *testing.B) {
	benchCompiledVMCold(b, benchSrc, benchStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}, false)
}

func BenchmarkCompiledVMOptCold(b *testing.B) {
	benchCompiledVMCold(b, benchSrc, benchStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}, true)
}
