package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"numscript/ansi"
	"numscript/interpreter"
	"numscript/parser"
	"os"

	"github.com/spf13/cobra"
)

var runVariablesOpt string
var runBalancesOpt string
var runMetaOpt string
var runRawOpt string
var runStdinFlag bool

type rawOptions struct {
	Script    string                          `json:"script"`
	Variables map[string]string               `json:"variables"`
	Meta      map[string]interpreter.Metadata `json:"meta"`
	Balances  interpreter.StaticStore         `json:"balances"`
}

func (o *rawOptions) fromRaw() {
	if runRawOpt == "" {
		return
	}

	err := json.Unmarshal([]byte(runRawOpt), o)
	if err != nil {
		panic(err)
	}
}

func (o *rawOptions) fromStdin() {
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

func (o *rawOptions) fromOptions(path string) {
	if path != "" {
		numscriptContent, err := os.ReadFile(path)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		o.Script = string(numscriptContent)
	}

	store := make(interpreter.StaticStore)
	if runBalancesOpt != "" {
		content, err := os.ReadFile(runBalancesOpt)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &store)
	}

	meta := make(map[string]interpreter.Metadata)
	if runMetaOpt != "" {
		content, err := os.ReadFile(runMetaOpt)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &meta)
	}

	vars := make(map[string]string)
	if runVariablesOpt != "" {
		content, err := os.ReadFile(runVariablesOpt)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &vars)
	}
}

func run(path string) {
	opt := rawOptions{
		Variables: make(map[string]string),
		Meta:      make(map[string]interpreter.Metadata),
		Balances:  make(interpreter.StaticStore),
	}

	opt.fromRaw()
	opt.fromOptions(path)
	opt.fromStdin()

	parseResult := parser.Parse(opt.Script)
	if len(parseResult.Errors) != 0 {
		// TODO better output
		fmt.Printf("Got errors while parsing\n")
		os.Exit(1)
	}

	result, err := interpreter.RunProgram(parseResult.Value, opt.Variables, opt.Balances, opt.Meta)
	showResult(result, err)
}

func showResult(result *interpreter.ExecutionResult, err error) {
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	fmt.Println(ansi.ColorCyan("Postings:"))
	postingsJson, err := json.MarshalIndent(result.Postings, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(postingsJson))

	fmt.Println()

	fmt.Println(ansi.ColorCyan("Meta:"))
	txMetaJson, err := json.MarshalIndent(result.TxMeta, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(txMetaJson))
}

func getRunCmd() *cobra.Command {
	cmd := cobra.Command{
		// Keep the command as hidden as long as it's unstable
		Hidden: true,

		// Other ideas: simulate, eval, exec
		Use:   "run",
		Short: "Evaluate a numscript file",
		Long:  "Evaluate a numscript file. This command is unstable and still being developed",
		Run: func(cmd *cobra.Command, args []string) {
			var path string
			if len(args) > 0 {
				path = args[0]
			}
			run(path)
		},
	}

	cmd.Flags().StringVarP(&runVariablesOpt, "variables", "v", "", "Path of a json file containing the variables")
	cmd.Flags().StringVarP(&runBalancesOpt, "balances", "b", "", "Path of a json file containing the balances")
	cmd.Flags().StringVarP(&runMetaOpt, "meta", "m", "", "Path of a json file containing the accounts metadata")

	cmd.Flags().StringVarP(&runRawOpt, "raw", "r", "", "Raw json input containing script, variables, balances, metadata")

	cmd.Flags().BoolVar(&runStdinFlag, "stdin", false, "example")

	return &cmd
}
