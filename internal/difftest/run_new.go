package difftest

import (
	"context"
	"errors"
	"maps"
	"math/big"

	"github.com/formancehq/numscript"
	"github.com/formancehq/numscript/internal/gen"
)

// Posting is the normalized shape both engines' postings are reduced to. The
// interpreter also tracks scopes and colors; internal/gen never emits the
// syntax for either, so they are not compared.
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
	// ResolveErr is oracle-only: set if compilation succeeded but the machine
	// rejected the script before executing it, resolving vars/resources/
	// balances against the store (SetVarsFromJSON, ResolveResources,
	// ResolveBalances). The new interpreter has no separate resolve stage --
	// the equivalent work happens inside Run and surfaces as RunErr -- so this
	// is always empty on that side.
	//
	// Kept apart from CompileErr so Compare can tell "the generator's cleanup
	// pass didn't reach the b-side's grammar" from "the b-side refused to even
	// start resolving the script", and apart from RunErr because no side
	// executed anything: comparing MissingFunds here would be meaningless, the
	// oracle never got as far as classifying a funds failure.
	ResolveErr string
	// RunErr is set if compilation and resolution succeeded but execution
	// failed.
	RunErr string
	// MissingFunds is only meaningful when RunErr is set: true iff the failure
	// was an insufficient-funds error rather than some other runtime rejection.
	// Compare treats this classification, not the RunErr text, as what must agree
	// across engines.
	MissingFunds bool
	// NegativeAmount is only meaningful when RunErr is set on the new
	// interpreter's side: true iff the failure was NegativeAmountErr. Compare
	// uses this typed classification, not RunErr text, to spot the negative
	// send amount divergence (oracle/DIVERGENCES.md #1).
	NegativeAmount bool
	// NegativeMaxReject is only meaningful when RunErr is set on the oracle's
	// side: true iff the failure was ledger's OP_TAKE_MAX guard rejecting a
	// negative `max` clause outright, source- or destination-side. Compare
	// uses this, not RunErr text, to spot the one known one-sided-failure
	// divergence (oracle/DIVERGENCES.md #4) without tolerating every other
	// one.
	NegativeMaxReject bool
	// RegisterOverflow is only meaningful when CompileErr is set on the vm's
	// side: true iff the compiler refused the script because it needs more
	// simultaneously-live registers than the bytecode encoding's one-byte
	// operands can address (ir.ErrRegisterBankOverflow). A known capacity
	// bound, not a semantic rejection: Compare tolerates it by name so any
	// other vm-side rejection of a script another engine ran stays a
	// mismatch.
	RegisterOverflow bool
	// InternalErr is set when an engine broke its own contract, as opposed to
	// rejecting the script. Deliberately not CompileErr: Compare tolerates one
	// side rejecting what the other accepted, so a self-inconsistency reported as
	// a compile error would be swallowed as expected. A non-empty InternalErr on
	// either side is an unconditional mismatch.
	//
	// No engine produces it yet; the compiler+VM will, on a bytecode-verifier
	// failure.
	InternalErr string
	// Postings is nil unless both CompileErr and RunErr are empty.
	Postings []Posting

	// TxMeta and AccountMeta are the metadata each engine wrote, normalized to
	// strings — the new interpreter already reports strings, the legacy machine
	// reports typed values. Both are nil unless execution succeeded.
	//
	// AccountMeta is keyed "<account>\x00<key>": a flat map compares as a set
	// without needing a nested equality helper, and NUL cannot occur in either
	// an account address or a meta key.
	TxMeta      map[string]string
	AccountMeta map[string]string
}

func (r SideResult) Failed() bool {
	return r.CompileErr != "" || r.ResolveErr != "" || r.RunErr != ""
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

	// Defensive copy: vars is shared with runOracle's call in RunOne.
	execResult, err := parseResult.Run(ctx, maps.Clone(vars), store)
	if err != nil {
		var missingFundsErr numscript.MissingFundsErr
		var negativeAmountErr numscript.NegativeAmountErr
		missingFunds := errors.As(err, &missingFundsErr)
		negativeAmount := errors.As(err, &negativeAmountErr)
		return SideResult{RunErr: err.Error(), MissingFunds: missingFunds, NegativeAmount: negativeAmount}
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

	txMeta := make(map[string]string, len(execResult.Metadata))
	for k, v := range execResult.Metadata {
		txMeta[k] = v
	}
	accountMeta := make(map[string]string, len(execResult.AccountsMetadata))
	for _, row := range execResult.AccountsMetadata {
		accountMeta[metaKey(row.Account, row.Key)] = row.Value
	}

	return SideResult{Postings: postings, TxMeta: txMeta, AccountMeta: accountMeta}
}

// metaKey joins an account and a meta key into one flat map key. NUL cannot
// appear in either, so the join is unambiguous.
func metaKey(account, key string) string {
	return account + "\x00" + key
}
