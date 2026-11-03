package cmd

import (
	"fmt"
	"os"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/parser"

	"github.com/spf13/cobra"
)

type IRArgs struct {
	OutputPath string
}

// compileScriptToIR reads a numscript source file and compiles it to IR,
// printing parse and compile errors on stderr the way `run` does.
func compileScriptToIR(scriptPath string) ([]ir.Instr, error) {
	content, err := readIRSource(scriptPath)
	if err != nil {
		return nil, err
	}
	src := string(content)

	parseResult := parser.Parse(src)
	if len(parseResult.Errors) != 0 {
		fmt.Fprint(os.Stderr, parser.ParseErrorsToString(parseResult.Errors, src))
		return nil, fmt.Errorf("parsing failed")
	}

	instrs, _, err := compiler.CompileToIR(parseResult.Value, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		if ranged, ok := err.(parser.Ranged); ok {
			fmt.Fprintln(os.Stderr, ranged.GetRange().ShowOnSource(src))
		}
		return nil, fmt.Errorf("compilation failed")
	}
	return instrs, nil
}

func dumpIR(scriptPath string, opts IRArgs) error {
	instrs, err := compileScriptToIR(scriptPath)
	if err != nil {
		return err
	}

	text := ir.Dump(instrs)
	if opts.OutputPath == "" || opts.OutputPath == stdioPath {
		_, err := os.Stdout.WriteString(text)
		return err
	}
	return os.WriteFile(opts.OutputPath, []byte(text), 0o644)
}

func getIRCmd() *cobra.Command {
	opts := IRArgs{}

	cmd := cobra.Command{
		Use:   "ir <file.num>",
		Short: "Compile a numscript file and dump its textual IR",
		Long: `Compile a numscript file and dump the textual IR the compiler produces, before
it is assembled into bytecode. The output goes to stdout unless --output says
otherwise, so it can be piped into assemble:
numscript ir folder/my-script.num | numscript assemble - > folder/my-script.numb

Pass - as the path to read the script from stdin.

The IR format tracks an unstable instruction set and is not a public interface.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := dumpIR(args[0], opts)
			if err != nil {
				cmd.SilenceErrors = true
				cmd.SilenceUsage = true
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "", "Path where to write the IR ('-' for stdout)")

	return &cmd
}
