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

func (o *inputOpts) fromRaw(opts runArgs) {
	if opts.RawOpt == "" {
		return
	}

	err := json.Unmarshal([]byte(opts.RawOpt), o)
	if err != nil {
		panic(err)
	}
}

func (o *inputOpts) fromStdin(opts runArgs) {
	if !opts.StdinFlag {
		return
	}

	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(bytes, o)
	if err != nil {
		panic(err)
	}
}

func (o *inputOpts) fromOptions(path string, opts runArgs) {
	if path != "" {
		numscriptContent, err := os.ReadFile(path)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		o.Script = string(numscriptContent)
	}

	if opts.BalancesOpt != "" {
		content, err := os.ReadFile(opts.BalancesOpt)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &o.Balances)
	}

	if opts.MetaOpt != "" {
		content, err := os.ReadFile(opts.MetaOpt)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &o.Meta)
	}

	if opts.VariablesOpt != "" {
		content, err := os.ReadFile(opts.VariablesOpt)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &o.Variables)
	}
}

func run(path string, opts runArgs) {
	opt := inputOpts{
		Variables: make(map[string]string),
		Meta:      make(interpreter.AccountsMetadata),
		Balances:  make(interpreter.Balances),
	}

	opt.fromRaw(opts)
	opt.fromOptions(path, opts)
	opt.fromStdin(opts)

	parseResult := parser.Parse(opt.Script)
	if len(parseResult.Errors) != 0 {
		os.Stderr.Write([]byte(parser.ParseErrorsToString(parseResult.Errors, opt.Script)))
		os.Exit(1)
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
		os.Stderr.Write([]byte(err.Error()))
		if rng.Start != rng.End {
			os.Stderr.Write([]byte("\n"))
			os.Stderr.Write([]byte(err.GetRange().ShowOnSource(parseResult.Source)))
		}
		os.Exit(1)
		return
	}

	switch opts.OutFormatOpt {
	case OutputFormatJson:
		showJson(result)
	case OutputFormatPretty:
		showPretty(result)
	default:
		// TODO handle err
		panic("Invalid option: " + opts.OutFormatOpt)
	}

}

func showJson(result *interpreter.ExecutionResult) {
	out, err := json.Marshal(result)
	if err != nil {
		// TODO handle err
		panic(err)
	}

	os.Stdout.Write(out)
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
		Run: func(cmd *cobra.Command, args []string) {
			var path string
			if len(args) > 0 {
				path = args[0]
			}
			run(path, opts)
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
