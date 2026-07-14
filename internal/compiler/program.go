package compiler

import (
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/vm"
)

// CompiledProgram is a program lowered to VM bytecode, ready to run. Its fields
// are unexported so it stays opaque when re-exported by the public API.
type CompiledProgram struct {
	encoder VarsEncoder
	program vm.Program
}

// CompileProgram lowers a parsed program to a runnable CompiledProgram.
func CompileProgram(program parser.Program) (CompiledProgram, error) {
	encoder, prog, err := Compile(program)
	if err != nil {
		return CompiledProgram{}, err
	}
	return CompiledProgram{encoder: encoder, program: prog}, nil
}

// Run executes the compiled program with the given variables against store.
func (c CompiledProgram) Run(vars map[string]string, store vm.Store) (runtime.ExecutionResult, error) {
	encoded, err := c.encoder.Encode(vars)
	if err != nil {
		return runtime.ExecutionResult{}, err
	}
	res, execErr := vm.Exec(vm.NewVm(c.program), &encoded, store)
	if execErr != nil {
		return runtime.ExecutionResult{}, execErr
	}
	return res, nil
}
