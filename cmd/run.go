package cmd

import (
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
	content, err := os.ReadFile(path)
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		return
	}

	parseResult := parser.Parse(string(content))
	if len(parseResult.Errors) != 0 {
		// TODO better output
		fmt.Printf("Got errors while parsing\n")
		return
	}

	program := parseResult.Value

	store := interpreter.StaticStore{}
	// TODO vars, store, meta
	result, err := interpreter.RunProgram(program, nil, store, nil)
	if err != nil {
		panic(err)
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
