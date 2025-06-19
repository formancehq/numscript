package specs_format

import (
	"context"
	"reflect"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
)

// --- Specs:
type Specs struct {
	It               string                       `json:"it"`
	Balances         interpreter.Balances         `json:"balances,omitempty"`
	Vars             interpreter.VariablesMap     `json:"vars,omitempty"`
	Meta             interpreter.AccountsMetadata `json:"accountsMeta,omitempty"`
	TestCases        []Specs                      `json:"testCases,omitempty"`
	ExpectedPostings []interpreter.Posting        `json:"expectedPostings,omitempty"`
	// TODO expected tx meta
	// TODO expected accountsMeta
}

type SpecOutput struct {
	It               string
	Success          bool
	ExpectedPostings []interpreter.Posting
	ActualPostings   []interpreter.Posting

	// TODO expected tx meta, accountsMeta
}

func Run(program parser.Program, specs Specs) (SpecOutput, error) {

	result, err := interpreter.RunProgram(
		context.Background(),
		program,
		specs.Vars,
		interpreter.StaticStore{
			Balances: specs.Balances,
			Meta:     specs.Meta,
		}, nil)

	if err != nil {
		return SpecOutput{}, err
	}

	success := reflect.DeepEqual(result.Postings, specs.ExpectedPostings)

	return SpecOutput{
		It:               specs.It,
		Success:          success,
		ExpectedPostings: specs.ExpectedPostings,
		ActualPostings:   result.Postings,
	}, nil
}
