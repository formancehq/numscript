package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"numscript/ansi"
	"numscript/interpreter"
	"numscript/parser"
	"os"

	"github.com/spf13/cobra"
)

var runVariablesPath string
var runBalancesPath string
var runMetaPath string
var runStdinFlag bool

type rawOptions struct {
	Script    string                          `json:"script"`
	Variables map[string]string               `json:"variables"`
	Meta      map[string]interpreter.Metadata `json:"meta"`
	Balances  map[string]map[string]*big.Int  `json:"balances"`
}

func newRawOptions() rawOptions {
	return rawOptions{
		Variables: make(map[string]string),
		Meta:      make(map[string]interpreter.Metadata),
		Balances:  make(map[string]map[string]*big.Int),
	}
}

func parseScript(script string) parser.Program {
	parseResult := parser.Parse(script)
	if len(parseResult.Errors) != 0 {
		// TODO better output
		fmt.Printf("Got errors while parsing\n")
		os.Exit(1)
	}
	return parseResult.Value
}

func runStdin() {
	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	stdin := newRawOptions()
	err = json.Unmarshal(bytes, &stdin)
	if err != nil {
		panic(err)
	}

	program := parseScript(stdin.Script)
	result, err := interpreter.RunProgram(program, stdin.Variables, stdin.Balances, stdin.Meta)
	showResult(result, err)
}

func runFs(path string) {
	numscriptContent, err := os.ReadFile(path)
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		return
	}
	program := parseScript(string(numscriptContent))
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

	result, err := interpreter.RunProgram(program, vars, store, meta)
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
		// Args:  cobra.ExactArgs(1),
		Args: func(cmd *cobra.Command, args []string) error {
			if runStdinFlag {
				return nil
			} else {
				return cobra.ExactArgs(1)(cmd, args)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if runStdinFlag {
				runStdin()
			} else {
				// path := args[0]
				runFs("")
			}
		},
	}

	cmd.Flags().StringVarP(&runVariablesPath, "variables", "v", "", "Path of a json file containing the variables")
	cmd.Flags().StringVarP(&runBalancesPath, "balances", "b", "", "Path of a json file containing the balances")
	cmd.Flags().StringVarP(&runMetaPath, "meta", "m", "", "Path of a json file containing the accounts metadata")

	cmd.Flags().BoolVar(&runStdinFlag, "stdin", false, "example")

	return &cmd
}
