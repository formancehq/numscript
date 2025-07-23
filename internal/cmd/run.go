package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"

	"github.com/spf13/cobra"
)

const (
	OutputFormatPretty = "pretty"
	OutputFormatJson   = "json"
)

type InputsFile struct {
	FeatureFlags []string                     `json:"featureFlags"`
	Variables    map[string]string            `json:"variables"`
	Meta         interpreter.AccountsMetadata `json:"metadata"`
	Balances     interpreter.Balances         `json:"balances"`
}

type RunArgs struct {
	InputsPath   string
	OutFormatOpt string
}

func run(scriptPath string, opts RunArgs) error {
	numscriptContent, err := os.ReadFile(scriptPath)
	if err != nil {
		return err
	}

	parseResult := parser.Parse(string(numscriptContent))
	if len(parseResult.Errors) != 0 {
		fmt.Fprint(os.Stderr, parser.ParseErrorsToString(parseResult.Errors, string(numscriptContent)))
		return fmt.Errorf("parsing failed")
	}

	inputsPath := opts.InputsPath
	if inputsPath == "" {
		inputsPath = scriptPath + ".inputs.json"
	}

	inputsContent, err := os.ReadFile(inputsPath)
	if err != nil {
		return err
	}

	var inputs InputsFile
	err = json.Unmarshal(inputsContent, &inputs)
	if err != nil {
		return fmt.Errorf("failed to parse inputs file '%s' as JSON: %w", inputsPath, err)
	}

	featureFlags := map[string]struct{}{}
	for _, flag := range inputs.FeatureFlags {
		featureFlags[flag] = struct{}{}
	}

	result, iErr := interpreter.RunProgram(context.Background(), parseResult.Value, inputs.Variables, interpreter.StaticStore{
		Balances: inputs.Balances,
		Meta:     inputs.Meta,
	}, featureFlags)

	if iErr != nil {
		rng := iErr.GetRange()
		fmt.Fprint(os.Stderr, iErr.Error())
		if rng.Start != rng.End {
			fmt.Fprint(os.Stderr, "\n")
			fmt.Fprint(os.Stderr, iErr.GetRange().ShowOnSource(parseResult.Source))
		}
		return fmt.Errorf("execution failed")
	}

	switch opts.OutFormatOpt {
	case OutputFormatJson:
		return showJson(result)
	case OutputFormatPretty:
		return showPretty(result)
	default:
		return fmt.Errorf("invalid output format: %s", opts.OutFormatOpt)
	}
}

func showJson(result *interpreter.ExecutionResult) error {
	out, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("error marshaling result to JSON: %w", err)
	}

	_, err = os.Stdout.Write(out)
	return err
}

func showPretty(result *interpreter.ExecutionResult) error {
	fmt.Println("Postings:")
	fmt.Println(interpreter.PrettyPrintPostings(result.Postings))

	if len(result.Metadata) != 0 {
		fmt.Println("Meta:")
		fmt.Println(interpreter.PrettyPrintMeta(result.Metadata))
	}

	return nil
}

func getRunCmd() *cobra.Command {
	opts := RunArgs{}

	cmd := cobra.Command{
		Use:   "run",
		Short: "Evaluate a numscript file",
		Long: `Evaluate a numscript file, taking as inputs a json file containing balances, variables and metadata.

The inputs file has to have the same name as the numscript file plus a ".inputs.json" suffix, for example:
run folder/my-script.num
will expect a 'folder/my-script.num.inputs.json' file where to read inputs from.

You can use explicitly specify where the inputs file should be using the optional --inputs argument.
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]

			err := run(path, opts)
			if err != nil {
				fmt.Fprint(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVar(&opts.InputsPath, "inputs", "", "Path of a json file containing the inputs")
	cmd.Flags().StringVarP(&opts.OutFormatOpt, "output-format", "o", OutputFormatPretty, "Set the output format. Available options: pretty, json.")

	return &cmd
}
