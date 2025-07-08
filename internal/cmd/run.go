package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"

	"github.com/spf13/cobra"
)

const (
	OutputFormatPretty = "pretty"
	OutputFormatJson   = "json"
)

type runArgs struct {
	VariablesOpt string
	BalancesOpt  string
	MetaOpt      string
	RawOpt       string
	StdinFlag    bool
	OutFormatOpt string
	Flags        []string
}

type inputOpts struct {
	Script    string                       `json:"script"`
	Variables map[string]string            `json:"variables"`
	Meta      interpreter.AccountsMetadata `json:"metadata"`
	Balances  interpreter.Balances         `json:"balances"`
}

func (o *inputOpts) fromRaw(opts runArgs) error {
	if opts.RawOpt == "" {
		return nil
	}

	err := json.Unmarshal([]byte(opts.RawOpt), o)
	if err != nil {
		return fmt.Errorf("invalid raw input JSON: %w", err)
	}
	return nil
}

func (o *inputOpts) fromStdin(opts runArgs) error {
	if !opts.StdinFlag {
		return nil
	}

	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("error reading from stdin: %w", err)
	}

	err = json.Unmarshal(bytes, o)
	if err != nil {
		return fmt.Errorf("invalid stdin JSON: %w", err)
	}
	return nil
}

func (o *inputOpts) fromOptions(path string, opts runArgs) error {
	if path != "" {
		numscriptContent, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading script file: %w", err)
		}
		o.Script = string(numscriptContent)
	}

	if opts.BalancesOpt != "" {
		content, err := os.ReadFile(opts.BalancesOpt)
		if err != nil {
			return fmt.Errorf("error reading balances file: %w", err)
		}
		if err := json.Unmarshal(content, &o.Balances); err != nil {
			return fmt.Errorf("invalid balances JSON: %w", err)
		}
	}

	if opts.MetaOpt != "" {
		content, err := os.ReadFile(opts.MetaOpt)
		if err != nil {
			return fmt.Errorf("error reading metadata file: %w", err)
		}
		if err := json.Unmarshal(content, &o.Meta); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
	}

	if opts.VariablesOpt != "" {
		content, err := os.ReadFile(opts.VariablesOpt)
		if err != nil {
			return fmt.Errorf("error reading variables file: %w", err)
		}
		if err := json.Unmarshal(content, &o.Variables); err != nil {
			return fmt.Errorf("invalid variables JSON: %w", err)
		}
	}
	return nil
}

func run(path string, opts runArgs) error {
	opt := inputOpts{
		Variables: make(map[string]string),
		Meta:      make(interpreter.AccountsMetadata),
		Balances:  make(interpreter.Balances),
	}

	if err := opt.fromRaw(opts); err != nil {
		return err
	}
	if err := opt.fromOptions(path, opts); err != nil {
		return err
	}
	if err := opt.fromStdin(opts); err != nil {
		return err
	}

	parseResult := parser.Parse(opt.Script)
	if len(parseResult.Errors) != 0 {
		fmt.Fprint(os.Stderr, parser.ParseErrorsToString(parseResult.Errors, opt.Script))
		return fmt.Errorf("parsing failed")
	}

	featureFlags := map[string]struct{}{}
	for _, flag := range opts.Flags {
		featureFlags[flag] = struct{}{}
	}

	result, err := interpreter.RunProgram(context.Background(), parseResult.Value, opt.Variables, interpreter.StaticStore{
		Balances: opt.Balances,
		Meta:     opt.Meta,
	}, featureFlags)

	if err != nil {
		rng := err.GetRange()
		fmt.Fprint(os.Stderr, err.Error())
		if rng.Start != rng.End {
			fmt.Fprint(os.Stderr, "\n")
			fmt.Fprint(os.Stderr, err.GetRange().ShowOnSource(parseResult.Source))
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

func showPretty(result *interpreter.ExecutionResult) {
	fmt.Println("Postings:")
	fmt.Println(interpreter.PrettyPrintPostings(result.Postings))

	if len(result.Metadata) != 0 {
		fmt.Println("Meta:")
		fmt.Println(interpreter.PrettyPrintMeta(result.Metadata))
	}
}

func getRunCmd() *cobra.Command {
	opts := runArgs{}

	cmd := cobra.Command{
		Use:   "run",
		Short: "Evaluate a numscript file",
		Long:  "Evaluate a numscript file, using the balances, the current metadata and the variables values as input.",
		RunE: func(cmd *cobra.Command, args []string) error {
			var path string
			if len(args) > 0 {
				path = args[0]
			}
			return run(path, opts)
		},
	}

	// Input args
	cmd.Flags().StringVarP(&opts.VariablesOpt, "variables", "v", "", "Path of a json file containing the variables")
	cmd.Flags().StringVarP(&opts.BalancesOpt, "balances", "b", "", "Path of a json file containing the balances")
	cmd.Flags().StringVarP(&opts.MetaOpt, "meta", "m", "", "Path of a json file containing the accounts metadata")
	cmd.Flags().StringVarP(&opts.RawOpt, "raw", "r", "", "Raw json input containing script, variables, balances, metadata")
	cmd.Flags().BoolVar(&opts.StdinFlag, "stdin", false, "Take input from stdin (same format as the --raw option)")

	// Feature flag
	cmd.Flags().StringSliceVar(&opts.Flags, "flags", nil, fmt.Sprintf("the feature flags to pass to the interpreter. Currently available flags: %s",
		strings.Join(flags.AllFlags, ", "),
	))

	// Output options
	cmd.Flags().StringVar(&opts.OutFormatOpt, "output-format", OutputFormatPretty, "Set the output format. Available options: pretty, json.")

	return &cmd
}
