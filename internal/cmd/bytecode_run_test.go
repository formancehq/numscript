package cmd

import (
	"context"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

const varsIR = `
  $asset = load_var<str>(0)
  set_current_asset($asset)
  $amount = load_var<int>(0)
  $src = load_var<str>(1)
  $overdraft = load_var<int>(1)
  $pulled = pull_account(account: $src, cap: $amount, overdraft: $overdraft)
  check_enough_funds($pulled, $amount)
  $dest = load_var<str>(2)
  send_to_account(account: $dest)
`

func TestDefaultBytecodePath(t *testing.T) {
	require.Equal(t, "folder/x.numb", defaultBytecodePath("folder/x.ir"))
	require.Equal(t, "folder/x.numb", defaultBytecodePath("folder/x"))
	require.Equal(t, "folder/x.num.numb", defaultBytecodePath("folder/x.num"))
	// stdin has no path to derive an output name from
	require.Equal(t, "-", defaultBytecodePath("-"))
}

// Reading the IR from stdin must assemble to the same program as reading it
// from a file, and default to writing the bytecode to stdout.
func TestAssembleFromStdin(t *testing.T) {
	dir := t.TempDir()

	stdin, err := os.Create(filepath.Join(dir, "stdin"))
	require.NoError(t, err)
	_, err = stdin.WriteString(varsIR)
	require.NoError(t, err)
	require.NoError(t, stdin.Close())
	stdin, err = os.Open(filepath.Join(dir, "stdin"))
	require.NoError(t, err)

	captured, err := os.Create(filepath.Join(dir, "stdout"))
	require.NoError(t, err)

	realStdin, realStdout := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = stdin, captured
	err = assemble("-", AssembleArgs{})
	os.Stdin, os.Stdout = realStdin, realStdout
	require.NoError(t, stdin.Close())
	require.NoError(t, captured.Close())
	require.NoError(t, err)

	written, err := os.ReadFile(filepath.Join(dir, "stdout"))
	require.NoError(t, err)
	fromStdin, err := vm.DecodeProgram(written)
	require.NoError(t, err)

	instrs, irErrs := ir.Parse(varsIR)
	require.Empty(t, irErrs)
	expected, err := ir.Assemble(instrs)
	require.NoError(t, err)
	require.Equal(t, expected.Instructions, fromStdin.Instructions)

	// nothing was written next to a "-" path
	_, err = os.Stat("-.numb")
	require.True(t, os.IsNotExist(err))
}

// The file assemble writes must decode back to exactly what the assembler
// produced in memory.
func TestAssembleWritesADecodableProgram(t *testing.T) {
	dir := t.TempDir()
	irPath := filepath.Join(dir, "prog.ir")
	require.NoError(t, os.WriteFile(irPath, []byte(varsIR), 0o644))

	require.NoError(t, assemble(irPath, AssembleArgs{}))

	written, err := os.ReadFile(filepath.Join(dir, "prog.numb"))
	require.NoError(t, err)
	decoded, err := vm.DecodeProgram(written)
	require.NoError(t, err)

	instrs, irErrs := ir.Parse(varsIR)
	require.Empty(t, irErrs)
	require.NoError(t, ir.Typecheck(instrs))
	expected, err := ir.Assemble(instrs)
	require.NoError(t, err)

	require.Equal(t, expected.Instructions, decoded.Instructions)
	require.Equal(t, expected.MaxRegString, decoded.MaxRegString)
	require.Equal(t, expected.MaxRegInt, decoded.MaxRegInt)
	require.Equal(t, expected.MaxRegPortion, decoded.MaxRegPortion)
	require.Equal(t, expected.MaxRegBool, decoded.MaxRegBool)
	// compared by content, not with require.Equal on the whole Program: this
	// program has no constants, and an empty pool assembles to a nil slice but
	// decodes to an empty one (parseStringsPool's make([]T, 0))
	require.Empty(t, decoded.StringsPool)
	require.Empty(t, decoded.IntsPool)
}

func TestAssembleToStdoutLeavesNoFile(t *testing.T) {
	dir := t.TempDir()
	irPath := filepath.Join(dir, "prog.ir")
	require.NoError(t, os.WriteFile(irPath, []byte(varsIR), 0o644))

	// the bytecode is binary: capture it instead of letting it into the test log
	captured, err := os.Create(filepath.Join(dir, "stdout"))
	require.NoError(t, err)
	realStdout := os.Stdout
	os.Stdout = captured
	err = assemble(irPath, AssembleArgs{OutputPath: "-"})
	os.Stdout = realStdout
	require.NoError(t, captured.Close())
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dir, "prog.numb"))
	require.True(t, os.IsNotExist(err))

	written, err := os.ReadFile(filepath.Join(dir, "stdout"))
	require.NoError(t, err)
	_, err = vm.DecodeProgram(written)
	require.NoError(t, err)
}

func TestLoadVarsFromPool(t *testing.T) {
	vars, err := loadVars("in.json", BytecodeInputsFile{
		VarsPool: &VarsPoolFile{
			Strings: []string{"USD/2"},
			// wider than an int64, so a naive json number would have lost it
			Ints: []string{"123456789012345678901234567890"},
		},
	}, "")
	require.NoError(t, err)

	expected, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	require.Equal(t, []string{"USD/2"}, vars.StringsPool)
	require.Equal(t, []big.Int{*expected}, vars.IntsPool)
}

func TestLoadVarsAbsentIsNil(t *testing.T) {
	vars, err := loadVars("in.json", BytecodeInputsFile{}, "")
	require.NoError(t, err)
	require.Nil(t, vars)
}

func TestLoadVarsRejectsBothSources(t *testing.T) {
	_, err := loadVars("in.json", BytecodeInputsFile{VarsPool: &VarsPoolFile{}}, "vars.nvar")
	require.ErrorContains(t, err, "cannot use --vars together with")
}

// The --vars path is the leader/node wire format: an encoded vm.Vars blob has to
// decode and drive the program to the same result as the inline pool.
func TestLoadVarsFromEncodedFile(t *testing.T) {
	dir := t.TempDir()
	varsPath := filepath.Join(dir, "prog.nvar")
	encoded := vm.Vars{
		StringsPool: []string{"USD/2", "src", "dest"},
		IntsPool:    []big.Int{*big.NewInt(10), *big.NewInt(0)},
	}.Encode()
	require.NoError(t, os.WriteFile(varsPath, encoded, 0o644))

	vars, err := loadVars("in.json", BytecodeInputsFile{}, varsPath)
	require.NoError(t, err)

	instrs, irErrs := ir.Parse(varsIR)
	require.Empty(t, irErrs)
	program, err := ir.Assemble(instrs)
	require.NoError(t, err)

	store, err := newVmStore("in.json", BytecodeInputsFile{
		Balances: interpreter.Balances{
			{Account: "src", Asset: "USD/2", Amount: big.NewInt(100)},
		},
	})
	require.NoError(t, err)

	res, execErr := vm.Exec(context.Background(), vm.NewVm(program), vars, store)
	require.Nil(t, execErr)
	require.Len(t, res.Postings, 1)
	require.Equal(t, runtime.Posting{
		Source: "src", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(10),
	}, res.Postings[0])
}

func TestLoadVarsRejectsAnUndecodableFile(t *testing.T) {
	dir := t.TempDir()
	varsPath := filepath.Join(dir, "prog.nvar")
	require.NoError(t, os.WriteFile(varsPath, []byte("not a vars blob"), 0o644))

	_, err := loadVars("in.json", BytecodeInputsFile{}, varsPath)
	require.ErrorContains(t, err, "failed to decode vars file")
}

// The store hands out balances the run state is free to mutate, so it must not
// alias the inputs.
func TestVmStoreReturnsACopyOfTheBalance(t *testing.T) {
	amount := big.NewInt(100)
	store, err := newVmStore("in.json", BytecodeInputsFile{
		Balances: interpreter.Balances{
			{Account: "src", Asset: "USD/2", Amount: amount},
		},
	})
	require.NoError(t, err)

	got, err := store.GetBalance(context.Background(), "src", "", "USD/2", "")
	require.NoError(t, err)
	require.Zero(t, got.Cmp(big.NewInt(100)))

	got.SetInt64(0)
	require.Zero(t, amount.Cmp(big.NewInt(100)))
}

func TestVmStoreUnknownAccountIsZeroNotAnError(t *testing.T) {
	store, err := newVmStore("in.json", BytecodeInputsFile{})
	require.NoError(t, err)

	got, err := store.GetBalance(context.Background(), "nobody", "", "USD/2", "")
	require.NoError(t, err)
	require.Zero(t, got.Sign())

	_, ok, err := store.GetMetadata(context.Background(), "nobody", "", "k")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestVmStoreSupportsScopedRows(t *testing.T) {
	store, err := newVmStore("in.json", BytecodeInputsFile{
		Balances: interpreter.Balances{
			{Account: "src", Asset: "USD/2", Amount: big.NewInt(1), Scope: "reserve"},
			{Account: "src", Asset: "USD/2", Amount: big.NewInt(100)},
		},
		Meta: interpreter.AccountsMetadata{
			{Account: "src", Key: "k", Value: "scoped", Scope: "reserve"},
			{Account: "src", Key: "k", Value: "unscoped"},
		},
	})
	require.NoError(t, err)

	scopedBal, err := store.GetBalance(context.Background(), "src", "reserve", "USD/2", "")
	require.NoError(t, err)
	require.Zero(t, scopedBal.Cmp(big.NewInt(1)))

	unscopedBal, err := store.GetBalance(context.Background(), "src", "", "USD/2", "")
	require.NoError(t, err)
	require.Zero(t, unscopedBal.Cmp(big.NewInt(100)))

	scopedMeta, ok, err := store.GetMetadata(context.Background(), "src", "reserve", "k")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "scoped", scopedMeta)

	unscopedMeta, ok, err := store.GetMetadata(context.Background(), "src", "", "k")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "unscoped", unscopedMeta)
}
