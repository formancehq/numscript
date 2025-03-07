package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/formancehq/numscript/internal/ansi"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"

	"github.com/spf13/cobra"
)

const (
	OutputFormatPretty = "pretty"
	OutputFormatJson   = "json"
)

var runVariablesOpt string
var runBalancesOpt string
var runMetaOpt string
var runRawOpt string
var runStdinFlag bool
var runOutFormatOpt string

var overdraftFeatureFlag bool
var oneOfFeatureFlag bool
var accountInterpolationFlag bool

type inputOpts struct {
	Script    string                       `json:"script"`
	Variables map[string]string            `json:"variables"`
	Meta      interpreter.AccountsMetadata `json:"metadata"`
	Balances  interpreter.Balances         `json:"balances"`
}

func (o *inputOpts) fromRaw() {
	if runRawOpt == "" {
		return
	}

	err := json.Unmarshal([]byte(runRawOpt), o)
	if err != nil {
		panic(err)
	}
}

func (o *inputOpts) fromStdin() {
	if !runStdinFlag {
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

func (o *inputOpts) fromOptions(path string) {
	if path != "" {
		numscriptContent, err := os.ReadFile(path)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		o.Script = string(numscriptContent)
	}

	if runBalancesOpt != "" {
		content, err := os.ReadFile(runBalancesOpt)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &o.Balances)
	}

	if runMetaOpt != "" {
		content, err := os.ReadFile(runMetaOpt)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &o.Meta)
	}

	if runVariablesOpt != "" {
		content, err := os.ReadFile(runVariablesOpt)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &o.Variables)
	}
}

func run(path string) {
	opt := inputOpts{
		Variables: make(map[string]string),
		Meta:      make(interpreter.AccountsMetadata),
		Balances:  make(interpreter.Balances),
	}

	opt.fromRaw()
	opt.fromOptions(path)
	opt.fromStdin()

	parseResult := parser.Parse(opt.Script)
	if len(parseResult.Errors) != 0 {
		os.Stderr.Write([]byte(parser.ParseErrorsToString(parseResult.Errors, opt.Script)))
		os.Exit(1)
	}

	featureFlags := map[string]struct{}{}
	if overdraftFeatureFlag {
		featureFlags[interpreter.ExperimentalOverdraftFunctionFeatureFlag] = struct{}{}
	}
	if oneOfFeatureFlag {
		featureFlags[interpreter.ExperimentalOneofFeatureFlag] = struct{}{}
	}
	if accountInterpolationFlag {
		featureFlags[interpreter.ExperimentalAccountInterpolationFlag] = struct{}{}
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

	switch runOutFormatOpt {
	case OutputFormatJson:
		showJson(result)
	case OutputFormatPretty:
		showPretty(result)
	default:
		// TODO handle err
		panic("Invalid option: " + runBalancesOpt)
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
	fmt.Println(ansi.ColorCyan("Postings:"))
	postingsJson, err := json.MarshalIndent(result.Postings, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(postingsJson))

	fmt.Println()

	fmt.Println(ansi.ColorCyan("Meta:"))
	txMetaJson, err := json.MarshalIndent(result.Metadata, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(txMetaJson))
}

func getRunCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "run",
		Short: "Evaluate a numscript file",
		Long:  "Evaluate a numscript file, using the balances, the current metadata and the variables values as input.",
		Run: func(cmd *cobra.Command, args []string) {
			var path string
			if len(args) > 0 {
				path = args[0]
			}
			run(path)
		},
	}

	// Input args
	cmd.Flags().StringVarP(&runVariablesOpt, "variables", "v", "", "Path of a json file containing the variables")
	cmd.Flags().StringVarP(&runBalancesOpt, "balances", "b", "", "Path of a json file containing the balances")
	cmd.Flags().StringVarP(&runMetaOpt, "meta", "m", "", "Path of a json file containing the accounts metadata")
	cmd.Flags().StringVarP(&runRawOpt, "raw", "r", "", "Raw json input containing script, variables, balances, metadata")
	cmd.Flags().BoolVar(&runStdinFlag, "stdin", false, "Take input from stdin (same format as the --raw option)")

	// Feature flag
	cmd.Flags().BoolVar(&overdraftFeatureFlag, interpreter.ExperimentalOverdraftFunctionFeatureFlag, false, "feature flag to enable the overdraft() function")
	cmd.Flags().BoolVar(&oneOfFeatureFlag, interpreter.ExperimentalOneofFeatureFlag, false, "feature flag to enable the oneof combinator")
	cmd.Flags().BoolVar(&accountInterpolationFlag, interpreter.ExperimentalAccountInterpolationFlag, false, "enables an account interpolation syntax, e.g. @users:$id:pending")

	// Output options
	cmd.Flags().StringVar(&runOutFormatOpt, "output-format", OutputFormatPretty, "Set the output format. Available options: pretty, json.")

	return &cmd
}
