package cmd

import (
	"encoding/json"
	"fmt"
	"numscript/interpreter"
	"numscript/parser"
	"os"

	"github.com/spf13/cobra"
)

var runVariablesPath string
var runBalancesPath string
var runMetaPath string

func run(path string) {
	numscriptContent, err := os.ReadFile(path)
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		return
	}

	parseResult := parser.Parse(string(numscriptContent))
	if len(parseResult.Errors) != 0 {
		// TODO better output
		fmt.Printf("Got errors while parsing\n")
		return
	}

	store := make(interpreter.StaticStore)
	if runBalancesPath != "" {
		content, err := os.ReadFile(runBalancesPath)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &store)
	}

	meta := make(map[string]interpreter.Metadata)
	if runMetaPath != "" {
		content, err := os.ReadFile(runMetaPath)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &meta)
	}

	vars := make(map[string]string)
	if runVariablesPath != "" {
		content, err := os.ReadFile(runVariablesPath)
		if err != nil {
			os.Stderr.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(content, &vars)
	}

	result, err := interpreter.RunProgram(parseResult.Value, vars, store, meta)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}

	fmt.Println("Postings:")
	for _, posting := range result.Postings {
		fmt.Printf("{ from = @%s, amount = [%s %s], to = %s}\n", posting.Source, posting.Asset, posting.Amount.String(), posting.Destination)
	}
}

func getRunCmd() *cobra.Command {
	cmd := cobra.Command{
		// Keep the command as hidden as long as it's unstable
		Hidden: true,

		// Other ideas: simulate, eval, exec
		Use:   "run",
		Short: "Evaluate a numscript file",
		Long:  "Evaluate a numscript file. This command is unstable and still being developed",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := args[0]
			run(path)
		},
	}

	cmd.Flags().StringVarP(&runVariablesPath, "variables", "v", "", "Path of a json file containing the variables")
	cmd.Flags().StringVarP(&runBalancesPath, "balances", "b", "", "Path of a json file containing the balances")
	cmd.Flags().StringVarP(&runMetaPath, "meta", "m", "", "Path of a json file containing the accounts metadata")
	return &cmd
}
