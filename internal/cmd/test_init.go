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

func MakeSpecsFile(source string) (specs_format.Specs, error) {
	parseResult := parser.Parse(source)
	if len(parseResult.Errors) != 0 {
		fmt.Fprint(os.Stderr, parser.ParseErrorsToString(parseResult.Errors, source))
		return specs_format.Specs{}, fmt.Errorf("parsing failed")
	}

	checkResult := analysis.CheckProgram(parseResult.Value)

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

	return makeSpecsFile(
		parseResult.Value,
		vars,
		map[string]struct{}{},
		big.NewInt(100),
	)
}

func makeSpecsFile(
	program parser.Program,
	vars map[string]string,
	featureFlags map[string]struct{},
	defaultBalance *big.Int,
) (specs_format.Specs, error) {

	store := TestInitStore{
		// TODO use max of numeric vars as default
		DefaultBalance: defaultBalance,
		Balances:       make(interpreter.Balances),
		Meta:           make(interpreter.AccountsMetadata),
	}

	res, iErr := interpreter.RunProgram(
		context.Background(),
		program,
		vars,
		store,
		featureFlags,
	)

	if iErr != nil {
		missingFundsErr, missingFunds := iErr.(interpreter.MissingFundsErr)
		if missingFunds {
			// TODO we could have better heuristics with a balance for each account/asset pair
			return makeSpecsFile(
				program,
				vars,
				featureFlags,
				&missingFundsErr.Needed,
			)
		}

		expFeatErr, missingFeatureFlag := iErr.(interpreter.ExperimentalFeature)
		if missingFeatureFlag {
			featureFlags[expFeatErr.FlagName] = struct{}{}
			return makeSpecsFile(
				program,
				vars,
				featureFlags,
				&missingFundsErr.Needed,
			)
		}

		return specs_format.Specs{}, iErr
	}

	var featureFlags_ []string
	for k := range featureFlags {
		featureFlags_ = append(featureFlags_, k)
	}

	specs := specs_format.Specs{
		Schema:       "https://raw.githubusercontent.com/formancehq/numscript/main/specs.schema.json",
		Balances:     store.Balances,
		Vars:         vars,
		FeatureFlags: featureFlags_,
		TestCases: []specs_format.TestCase{
			{
				It:             "example spec",
				ExpectPostings: res.Postings,
			},
		},
	}

	return specs, nil
}

func runTestInitCmd(opts testInitArgs) error {
	// TODO check there isn't a specsfile already

	numscriptContent, err := os.ReadFile(opts.path)
	if err != nil {
		return err
	}

	specs, err := MakeSpecsFile(string(numscriptContent))

	if err != nil {
		return err
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

			outputAccountBalance := utils.MapGetOrPutDefault(outputBalance, queriedAccount, func() interpreter.AccountBalance {
				return interpreter.AccountBalance{}
			})

			outputAccountBalance[curr] = new(big.Int).Set(amt)
		}
	}

	return outputBalance, nil
}

func (s TestInitStore) GetAccountsMetadata(c context.Context, q interpreter.MetadataQuery) (interpreter.AccountsMetadata, error) {
	outputMeta := interpreter.AccountsMetadata{}
	for queriedAccount, queriedCurrencies := range q {
		for _, curr := range queriedCurrencies {
			outputAccountMeta := utils.MapGetOrPutDefault(outputMeta, queriedAccount, func() interpreter.AccountMetadata {
				return interpreter.AccountMetadata{}
			})
			outputAccountMeta[curr] = ""
		}
	}

	return outputMeta, nil
}
