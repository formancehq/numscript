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
