package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/formancehq/numscript/internal/ir"

	"github.com/spf13/cobra"
)

type AssembleArgs struct {
	OutputPath string
}

// defaultBytecodePath derives the output path from the IR path: "x.ir" becomes
// "x.numb", anything else just gains the suffix.
func defaultBytecodePath(irPath string) string {
	return strings.TrimSuffix(irPath, ".ir") + ".numb"
}

func assemble(irPath string, opts AssembleArgs) error {
	content, err := os.ReadFile(irPath)
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
	if outputPath == "-" {
		_, err := os.Stdout.Write(bytecode)
		return err
	}

	return os.WriteFile(outputPath, bytecode, 0o644)
}

func getAssembleCmd() *cobra.Command {
	opts := AssembleArgs{}

	cmd := cobra.Command{
		Use:   "assemble",
		Short: "Assemble a textual IR file into bytecode",
		Long: `Assemble a textual IR file into the binary bytecode the vm executes.

The output goes to the input path with a ".numb" extension, for example:
assemble folder/my-script.ir
will write 'folder/my-script.numb'.

Use --output to write elsewhere, or --output - to write the bytecode to stdout.

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
