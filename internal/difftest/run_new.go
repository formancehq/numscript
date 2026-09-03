package difftest

import (
	"context"
	"maps"
	"math/big"

	"github.com/formancehq/numscript"
	"github.com/formancehq/numscript/internal/gen"
)

// Posting is the normalized shape both engines' postings are reduced to for
// comparison. The new interpreter additionally tracks scopes/colors, which
// the generator (internal/gen) never produces (no colored-asset or
// account-interpolation syntax), so they're intentionally not compared.
type Posting struct {
	Source      string
	Destination string
	Asset       string
	Amount      *big.Int
}

// SideResult is one engine's outcome for a single script run, normalized
// enough to compare across engines.
type SideResult struct {
	// CompileErr is set if the script failed to parse/compile.
	CompileErr string
	// RunErr is set if compilation succeeded but execution failed.
	RunErr string
	// MissingFunds is only meaningful when RunErr is set: true iff the
	// failure was specifically an insufficient/missing-funds error, as
	// opposed to some other runtime rejection (e.g. a negative amount).
	// Compare() treats this classification, not the exact RunErr text, as
	// the thing that must agree across engines — see its doc comment for
	// why (source-side negative max-clause amounts are a known, deliberately
	// unfixed gap that produces different non-funds RunErr text without
	// being a funds-adequacy bug).
	MissingFunds bool
	// Postings is nil unless both CompileErr and RunErr are empty.
	Postings []Posting
}

func (r SideResult) Failed() bool {
	return r.CompileErr != "" || r.RunErr != ""
}

func runNew(ctx context.Context, script string, vars map[string]string, balances map[gen.BalanceKey]*big.Int, metadata map[gen.MetaKey]string) SideResult {
	parseResult := numscript.Parse(script)
	if errs := parseResult.GetParsingErrors(); len(errs) != 0 {
		return SideResult{CompileErr: errs[0].Error()}
	}

	store := numscript.StaticStore{}
	for k, amount := range balances {
		store.Balances = append(store.Balances, numscript.BalanceRow{
			Account: k.Account,
			Asset:   k.Asset,
			Amount:  new(big.Int).Set(amount),
		})
	}
	for k, value := range metadata {
		store.Meta = append(store.Meta, numscript.AccountMetadataRow{
			Account: k.Account,
			Key:     k.Key,
			Value:   value,
		})
	}

	// Defensive copy: vars is shared with runOracle's call in RunOne, and
	// nothing here should depend on whether this function mutates its
	// input (it doesn't today, but that's not a documented guarantee).
	execResult, err := parseResult.Run(ctx, maps.Clone(vars), store)
	if err != nil {
		_, missingFunds := err.(numscript.MissingFundsErr)
		return SideResult{RunErr: err.Error(), MissingFunds: missingFunds}
	}

	postings := make([]Posting, 0, len(execResult.Postings))
	for _, p := range execResult.Postings {
		postings = append(postings, Posting{
			Source:      p.Source,
			Destination: p.Destination,
			Asset:       p.Asset,
			Amount:      p.Amount,
		})
	}

	return SideResult{Postings: postings}
}
