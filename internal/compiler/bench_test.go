package compiler

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/vm"
)

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
	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}
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
		rs.Pull(pulled, "src", ten, zero, "")
		_ = pulled.Cmp(ten) // CheckEnoughFunds
		rs.SendUncapped(&dest, nil)
		_ = rs.GetPostings()
	}
}

// BenchmarkCompiledVM measures the compiled bytecode on the register VM, reusing
// a single Vm instance across iterations (its register banks are not realloc'd).
func BenchmarkCompiledVM(b *testing.B) {
	parsed := parser.Parse(benchSrc)
	if len(parsed.Errors) != 0 {
		b.Fatalf("parse errors: %v", parsed.Errors)
	}
	compiled, cErr := compileProgramToVirtual(parsed.Value)
	if cErr != nil {
		b.Fatalf("compile: %v", cErr)
	}
	program, aErr := Assemble(compiled.instructions)
	if aErr != nil {
		b.Fatalf("assemble: %v", aErr)
	}
	store := e2eStore{balances: map[runtime.PairKey]*big.Int{
		{Account: "src", Asset: "USD/2", Color: ""}: big.NewInt(100),
	}}

	machine := vm.NewVm(program) // reused across iterations

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vm.Exec(machine, nil, store)
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

func cappedStore() e2eStore {
	return e2eStore{balances: map[runtime.PairKey]*big.Int{
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
	store := cappedStore()
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
		rs.Pull(pulled, "a", remaining, zero, "")
		total.Add(total, pulled)
		remaining.Sub(remaining, pulled)

		if remaining.Sign() != 0 { // jmp_if_zero(remaining)
			// max [USD/2 5] from @b  ->  cap = min(5, remaining)
			if five.Cmp(remaining) < 0 {
				capB.Set(five)
			} else {
				capB.Set(remaining)
			}
			rs.Pull(pulled, "b", capB, zero, "")
			total.Add(total, pulled)
			remaining.Sub(remaining, pulled)

			if remaining.Sign() != 0 {
				rs.Pull(pulled, "c", remaining, zero, "") // @c (cap = remaining)
				total.Add(total, pulled)
			}
		}

		_ = total.Cmp(ten) // check_enough_funds
		rs.SendUncapped(&dest, nil)
		_ = rs.GetPostings()
	}
}

func BenchmarkCompiledVMCapped(b *testing.B) {
	parsed := parser.Parse(benchSrcCapped)
	if len(parsed.Errors) != 0 {
		b.Fatalf("parse errors: %v", parsed.Errors)
	}
	compiled, cErr := compileProgramToVirtual(parsed.Value)
	if cErr != nil {
		b.Fatalf("compile: %v", cErr)
	}
	program, aErr := Assemble(compiled.instructions)
	if aErr != nil {
		b.Fatalf("assemble: %v", aErr)
	}
	store := cappedStore()

	machine := vm.NewVm(program) // reused across iterations

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vm.Exec(machine, nil, store)
		if err != nil {
			b.Fatalf("exec: %v", err)
		}
	}
}
