package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/formancehq/numscript/internal/ir"

	"github.com/spf13/cobra"
)

type AssembleArgs struct {
	OutputPath string
}

// stdioPath is the conventional stand-in for stdin/stdout, accepted both as the
// input path and as --output.
const stdioPath = "-"

// defaultBytecodePath derives the output path from the IR path: "x.ir" becomes
// "x.numb", anything else just gains the suffix. Reading from stdin has no path
// to derive from, so it writes to stdout.
func defaultBytecodePath(irPath string) string {
	if irPath == stdioPath {
		return stdioPath
	}
	return strings.TrimSuffix(irPath, ".ir") + ".numb"
}

func readIRSource(irPath string) ([]byte, error) {
	if irPath == stdioPath {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(irPath)
}

func assemble(irPath string, opts AssembleArgs) error {
	content, err := readIRSource(irPath)
	if err != nil {
		return err
	}
	src := string(content)

	instrs, irErrs := ir.Parse(src)
	if len(irErrs) != 0 {
		for _, irErr := range irErrs {
			fmt.Fprintln(os.Stderr, irErr.Error())
			fmt.Fprint(os.Stderr, irErr.Range.ShowOnSource(src))
		}
		return fmt.Errorf("assembling failed")
	}

	if err := ir.Typecheck(instrs); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return fmt.Errorf("assembling failed")
	}

	program, err := ir.Assemble(instrs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return fmt.Errorf("assembling failed")
	}

	bytecode := program.Encode()

	outputPath := opts.OutputPath
	if outputPath == "" {
		outputPath = defaultBytecodePath(irPath)
	}
	if outputPath == stdioPath {
		_, err := os.Stdout.Write(bytecode)
		return err
	}

	return os.WriteFile(outputPath, bytecode, 0o644)
}

func getAssembleCmd() *cobra.Command {
	opts := AssembleArgs{}

	cmd := cobra.Command{
		Use:   "assemble <file.ir>",
		Short: "Assemble a textual IR file into bytecode",
		Long: `Assemble a textual IR file into the binary bytecode the vm executes.

The output goes to the input path with a ".numb" extension, for example:
assemble folder/my-script.ir
will write 'folder/my-script.numb'.

Use --output to write elsewhere, or --output - to write the bytecode to stdout.

Pass - as the path to read the IR from stdin, in which case the bytecode goes to
stdout unless --output says otherwise:
cat folder/my-script.ir | numscript assemble - > folder/my-script.numb

The IR format tracks an unstable instruction set and is not a public interface.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := assemble(args[0], opts)
			if err != nil {
				cmd.SilenceErrors = true
				cmd.SilenceUsage = true
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.OutputPath, "output", "o", "", "Path where to write the bytecode ('-' for stdout)")

	return &cmd
}
