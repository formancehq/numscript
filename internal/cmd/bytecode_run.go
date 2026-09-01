package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strings"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/vm"

	"github.com/spf13/cobra"
)

// VarsPoolFile is the raw form of vm.Vars: the compiler's VarsEncoder maps
// declared variable names onto pool slots, but it lives in the source, not in
// the bytecode, so a bytecode-only run has to name the slots by index. Ints are
// strings so that values beyond float64/int64 survive the JSON round-trip.
type VarsPoolFile struct {
	Strings []string `json:"strings"`
	Ints    []string `json:"ints"`
}

// BytecodeInputsFile is the `run` inputs file plus the vars pool. It is a
// separate type so that the vm's index-addressed vars stay out of the
// interpreter's inputs shape; the shared fields keep the same json names, so one
// .inputs.json works for both commands.
type BytecodeInputsFile struct {
	Meta     interpreter.AccountsMetadata `json:"metadata"`
	Balances interpreter.Balances         `json:"balances"`
	VarsPool *VarsPoolFile                `json:"varsPool"`
}

type BytecodeRunArgs struct {
	InputsPath   string
	VarsPath     string
	OutFormatOpt string
}

// vmMetaKey identifies one metadata slot: account, scope and key.
type vmMetaKey struct {
	account string
	scope   string
	key     string
}

// vmStore is a vm.Store over the rows of an inputs file.
type vmStore struct {
	balances map[runtime.PairKey]*big.Int
	meta     map[vmMetaKey]string
}

func (s vmStore) GetBalance(_ context.Context, account, scope, asset, color string) (*big.Int, error) {
	// the caller owns what it gets: the run state mutates balances in place
	if v, ok := s.balances[runtime.PairKey{Account: account, Scope: scope, Asset: asset, Color: color}]; ok {
		return new(big.Int).Set(v), nil
	}
	return new(big.Int), nil
}

func (s vmStore) GetMetadata(_ context.Context, account, scope, key string) (string, bool, error) {
	v, ok := s.meta[vmMetaKey{account: account, scope: scope, key: key}]
	return v, ok, nil
}

func newVmStore(inputsPath string, inputs BytecodeInputsFile) (vmStore, error) {
	store := vmStore{
		balances: make(map[runtime.PairKey]*big.Int, len(inputs.Balances)),
		meta:     make(map[vmMetaKey]string, len(inputs.Meta)),
	}

	for _, row := range inputs.Balances {
		amount := row.Amount
		if amount == nil {
			amount = new(big.Int)
		}
		store.balances[runtime.PairKey{Account: row.Account, Scope: row.Scope, Asset: row.Asset, Color: row.Color}] = amount
	}

	for _, row := range inputs.Meta {
		store.meta[vmMetaKey{account: row.Account, scope: row.Scope, key: row.Key}] = row.Value
	}

	return store, nil
}

// loadVars resolves the vars pool from either the inputs file or an encoded
// .nvar blob. Both absent is legal: vm.Exec accepts a nil *Vars.
func loadVars(inputsPath string, inputs BytecodeInputsFile, varsPath string) (*vm.Vars, error) {
	if inputs.VarsPool != nil && varsPath != "" {
		return nil, fmt.Errorf("cannot use --vars together with the 'varsPool' key of '%s'", inputsPath)
	}

	if varsPath != "" {
		content, err := os.ReadFile(varsPath)
		if err != nil {
			return nil, err
		}
		vars, err := vm.DecodeVars(content)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vars file '%s': %w", varsPath, err)
		}
		return &vars, nil
	}

	if inputs.VarsPool == nil {
		return nil, nil
	}

	ints := make([]big.Int, len(inputs.VarsPool.Ints))
	for i, raw := range inputs.VarsPool.Ints {
		if _, ok := ints[i].SetString(raw, 10); !ok {
			return nil, fmt.Errorf("invalid inputs file '%s': varsPool.ints[%d] is not an integer: %q", inputsPath, i, raw)
		}
	}

	return &vm.Vars{
		StringsPool: inputs.VarsPool.Strings,
		IntsPool:    ints,
	}, nil
}

func bytecodeRun(bytecodePath string, opts BytecodeRunArgs) error {
	bytecode, err := os.ReadFile(bytecodePath)
	if err != nil {
		return err
	}

	program, err := vm.DecodeProgram(bytecode)
	if err != nil {
		return fmt.Errorf("failed to decode bytecode file '%s': %w", bytecodePath, err)
	}

	inputsPath := opts.InputsPath
	if inputsPath == "" {
		inputsPath = bytecodePath + ".inputs.json"
	}

	inputsContent, err := os.ReadFile(inputsPath)
	if err != nil {
		return err
	}

	var inputs BytecodeInputsFile
	err = json.Unmarshal(inputsContent, &inputs)
	if err != nil {
		return fmt.Errorf("failed to parse inputs file '%s' as JSON: %w", inputsPath, err)
	}

	if err := validateInputRows(inputsPath, inputs.Balances, inputs.Meta); err != nil {
		return err
	}

	store, err := newVmStore(inputsPath, inputs)
	if err != nil {
		return err
	}

	vars, err := loadVars(inputsPath, inputs, opts.VarsPath)
	if err != nil {
		return err
	}

	result, execErr := vm.Exec(context.Background(), vm.NewVm(program), vars, store)
	if execErr != nil {
		fmt.Fprintln(os.Stderr, execErr.Error())
		return fmt.Errorf("execution failed")
	}

	switch opts.OutFormatOpt {
	case OutputFormatJson:
		return showBytecodeJson(result)
	case OutputFormatPretty:
		return showBytecodePretty(result)
	default:
		return fmt.Errorf("invalid output format: %s", opts.OutFormatOpt)
	}
}

func showBytecodeJson(result runtime.ExecutionResult) error {
	out, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("error marshaling result to JSON: %w", err)
	}

	_, err = os.Stdout.Write(out)
	return err
}

func showBytecodePretty(result runtime.ExecutionResult) error {
	fmt.Println("Postings:")
	fmt.Println(interpreter.PrettyPrintPostings(result.Postings))

	// interpreter.PrettyPrintMeta takes map[string]Value; the vm's metadata is
	// already stringified
	if len(result.Metadata) != 0 {
		fmt.Println("Meta:")
		fmt.Print(prettyPrintStringMeta(result.Metadata))
	}

	if len(result.AccountsMetadata) != 0 {
		fmt.Println("Accounts meta:")
		fmt.Print(result.AccountsMetadata.PrettyPrint())
	}

	return nil
}

func prettyPrintStringMeta(meta map[string]string) string {
	keys := make([]string, 0, len(meta))
	for key := range meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&sb, "  %s: %s\n", key, meta[key])
	}
	return sb.String()
}

func getBytecodeRunCmd() *cobra.Command {
	opts := BytecodeRunArgs{}

	cmd := cobra.Command{
		Use:   "bytecode-run",
		Short: "Execute a bytecode file",
		Long: `Execute a bytecode file, taking as inputs a json file containing balances, metadata and the vars pool.

The inputs file has to have the same name as the bytecode file plus a ".inputs.json" suffix, for example:
bytecode-run folder/my-script.numb
will expect a 'folder/my-script.numb.inputs.json' file where to read inputs from.

You can explicitly specify where the inputs file should be using the optional --inputs argument.

Unlike 'run', variables are not passed by name: the bytecode addresses them by
their index in the vars pools, so they are given either as a "varsPool" key of the
inputs file ({"strings": [...], "ints": [...]}) or as an encoded vars blob via --vars.

The bytecode format tracks an unstable instruction set and is not a public interface.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := bytecodeRun(args[0], opts)
			if err != nil {
				cmd.SilenceErrors = true
				cmd.SilenceUsage = true
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.InputsPath, "inputs", "", "Path of a json file containing the inputs")
	cmd.Flags().StringVar(&opts.VarsPath, "vars", "", "Path of a file containing an encoded vars payload")
	cmd.Flags().StringVarP(&opts.OutFormatOpt, "output-format", "o", OutputFormatPretty, "Set the output format. Available options: pretty, json.")

	return &cmd
}
