package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/formancehq/numscript/internal/utils"
	"github.com/spf13/cobra"
)

type testInitArgs struct {
	path string
}

func getTestInitCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "test-init path",
		Short: "Create a specs file for the given numscript",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := runTestInitCmd(testInitArgs{
				path: args[0],
			})

			if err != nil {
				cmd.SilenceErrors = true
				cmd.SilenceUsage = true
				return err
			}

			return nil

		},
	}

	return cmd
}

func mkDefaultVar(decl parser.VarDeclaration, varsTypes map[parser.VarDeclaration]analysis.Type) string {
	defaultAmt := 100
	defaultCurr := "USD/2"

	asset := varsTypes[decl].Resolve()
	switch asset := asset.(type) {
	case *analysis.TAsset:
		defaultCurr = string(*asset)
	}

	switch decl.Type.Name {

	case analysis.TypeMonetary:
		return fmt.Sprintf("%s %d", defaultCurr, defaultAmt)

	case analysis.TypeNumber:
		return fmt.Sprintf("%d", defaultAmt)

	case analysis.TypeAccount:
		return decl.Name.Name

	case analysis.TypeString:
		return decl.Name.Name

	case analysis.TypePortion:
		return "1/2"

	case analysis.TypeAsset:
		return string(defaultCurr)
	}

	return ""
}

func runTestInitCmd(opts testInitArgs) error {
	// TODO check there isn't a specsfile already

	numscriptContent, err := os.ReadFile(opts.path)
	if err != nil {
		return err
	}

	parseResult := parser.Parse(string(numscriptContent))
	if len(parseResult.Errors) != 0 {
		fmt.Fprint(os.Stderr, parser.ParseErrorsToString(parseResult.Errors, string(numscriptContent)))
		return fmt.Errorf("parsing failed")
	}

	checkResult := analysis.CheckProgram(parseResult.Value)

	// parseResult.Value.Vars

	// TODO we should have an ad-hoc api for this
	// TODO parse vars
	vars := map[string]string{}

	if parseResult.Value.Vars != nil {
		for _, decl := range parseResult.Value.Vars.Declarations {
			if decl.Origin != nil {
				continue
			}

			value := mkDefaultVar(decl, checkResult.VarTypes)
			vars[decl.Name.Name] = value
		}
	}

	featureFlags := map[string]struct{}{}

	store := TestInitStore{
		// TODO use max of numeric vars as default
		DefaultBalance: big.NewInt(1000),
		Balances:       make(interpreter.Balances),
		Meta:           make(interpreter.AccountsMetadata),
	}

	res, iErr := interpreter.RunProgram(
		context.Background(),
		parseResult.Value,
		vars,
		store,
		featureFlags,
	)

	if iErr != nil {
		panic(iErr)
	}

	// TODO check iErr is catchable err

	specs := specs_format.Specs{
		Schema:   "https://raw.githubusercontent.com/formancehq/numscript/main/specs.schema.json",
		Balances: store.Balances,
		Vars:     vars,
		TestCases: []specs_format.TestCase{
			{
				It:             "example spec",
				ExpectPostings: res.Postings,
			},
		},
	}

	marshaled, _ := json.MarshalIndent(specs, "", "  ")

	os.WriteFile(opts.path+".specs.json", marshaled, 0644)

	fmt.Printf("✅ Created specs file: %s.specs.json\n", opts.path)

	return nil
}

type TestInitStore struct {
	DefaultBalance *big.Int
	Balances       interpreter.Balances
	Meta           interpreter.AccountsMetadata
}

func (s TestInitStore) GetBalances(_ context.Context, q interpreter.BalanceQuery) (interpreter.Balances, error) {
	outputBalance := interpreter.Balances{}
	for queriedAccount, queriedCurrencies := range q {

		for _, curr := range queriedCurrencies {
			amt := utils.NestedMapGetOrPutDefault(s.Balances, queriedAccount, curr, func() *big.Int {
				return new(big.Int).Set(s.DefaultBalance)
			})

			outpuAccountBalance := utils.MapGetOrPutDefault(outputBalance, queriedAccount, func() interpreter.AccountBalance {
				return interpreter.AccountBalance{}
			})

			outpuAccountBalance[curr] = new(big.Int).Set(amt)
		}
	}

	return outputBalance, nil
}

func (s TestInitStore) GetAccountsMetadata(context.Context, interpreter.MetadataQuery) (interpreter.AccountsMetadata, error) {
	panic("TODO")
}
