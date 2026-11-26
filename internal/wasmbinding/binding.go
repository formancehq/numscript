// Package wasmbinding is the host-testable core of the Numscript WebAssembly
// binding. It is deliberately free of syscall/js so its JSON/shape logic can be
// unit-tested with the normal Go toolchain; the js/wasm shim in ../../cmd/wasm
// is a thin wrapper that calls these functions and marshals the result across
// the JS boundary.
//
// Every exported function takes and returns plain bytes/strings and never
// panics: a panic in any of them is recovered and turned into an error
// envelope, so a malformed input can never take down the long-lived Go runtime
// that backs an editor.
package wasmbinding

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/formancehq/numscript"
	"github.com/formancehq/numscript/internal/analysis"
)

// ---- Static analysis (check) --------------------------------------------------

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type rng struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type diagnostic struct {
	Range    rng    `json:"range"`
	Severity byte   `json:"severity"` // LSP-aligned: 1=Error 2=Warning 3=Info 4=Hint
	Message  string `json:"message"`
	Id       int32  `json:"id"`
}

type checkEnvelope struct {
	Diagnostics []diagnostic `json:"diagnostics"`
}

// Check runs the full static-analysis pass (parse + type checking) over the
// given source and returns a JSON envelope of diagnostics. Ranges are 0-based
// {line, character}, faithful to the Go contract; the TS binding converts them
// to 1-based Monaco markers. Severities are the LSP-aligned values returned by
// the analysis package verbatim.
func Check(source string) (out []byte) {
	defer func() {
		if r := recover(); r != nil {
			// Surface the internal failure as a single error diagnostic rather
			// than silently swallowing it, so the editor still shows something.
			out = mustMarshal(checkEnvelope{Diagnostics: []diagnostic{{
				Severity: byte(analysis.ErrorSeverity),
				Message:  fmt.Sprintf("internal error: %v", r),
			}}})
		}
	}()

	result := analysis.CheckSource(source)

	diagnostics := make([]diagnostic, 0, len(result.Diagnostics))
	for _, d := range result.Diagnostics {
		diagnostics = append(diagnostics, diagnostic{
			Range: rng{
				Start: position{Line: d.Range.Start.Line, Character: d.Range.Start.Character},
				End:   position{Line: d.Range.End.Line, Character: d.Range.End.Character},
			},
			Severity: d.Kind.Severity(),
			Message:  d.Kind.Message(),
			Id:       d.Id,
		})
	}

	return mustMarshal(checkEnvelope{Diagnostics: diagnostics})
}

// ---- Execution (run) ----------------------------------------------------------

// runInput mirrors the playground's existing request payload so the UI call
// site barely changes. balances/metadata use the nested-object shape the
// playground already produces (not the CLI's array shape).
type runInput struct {
	Script       string                            `json:"script"`
	Variables    map[string]string                 `json:"variables"`
	Balances     map[string]map[string]json.Number `json:"balances"`
	Metadata     map[string]map[string]string      `json:"metadata"`
	FeatureFlags []string                          `json:"featureFlags"`
}

// runValue is the success payload. Fields carry numscript's own JSON tags, and
// the slices/maps are guaranteed non-nil so the JSON never contains `null`
// where the UI expects an array/object.
type runValue struct {
	Postings     []numscript.Posting           `json:"postings"`
	TxMeta       map[string]string             `json:"txMeta"`
	AccountsMeta numscript.SetAccountsMetadata `json:"accountsMeta"`
}

type runEnvelope struct {
	Ok    bool      `json:"ok"`
	Value *runValue `json:"value,omitempty"`
	Error string    `json:"error,omitempty"`
}

// Run parses and executes the given script against an in-memory store built
// entirely from the input (no network). It returns a JSON envelope:
//
//	{ "ok": true,  "value": { "postings", "txMeta", "accountsMeta" } }
//	{ "ok": false, "error": "<message>\n<annotated source>" }
func Run(inputJSON []byte) (out []byte) {
	defer func() {
		if r := recover(); r != nil {
			out = mustMarshal(runEnvelope{Ok: false, Error: fmt.Sprintf("internal error: %v", r)})
		}
	}()

	var input runInput
	if err := json.Unmarshal(inputJSON, &input); err != nil {
		return mustMarshal(runEnvelope{Ok: false, Error: fmt.Sprintf("invalid input: %v", err)})
	}

	balances, err := toBalances(input.Balances)
	if err != nil {
		return mustMarshal(runEnvelope{Ok: false, Error: err.Error()})
	}

	store := numscript.StaticStore{
		Balances: balances,
		Meta:     toAccountsMetadata(input.Metadata),
	}

	flags := make(map[string]struct{}, len(input.FeatureFlags))
	for _, flag := range input.FeatureFlags {
		flags[flag] = struct{}{}
	}

	vars := numscript.VariablesMap(input.Variables)
	if vars == nil {
		vars = numscript.VariablesMap{}
	}

	parsed := numscript.Parse(input.Script)
	result, iErr := parsed.RunWithFeatureFlags(context.Background(), vars, store, flags)
	if iErr != nil {
		// Reuse the CLI's annotated-source formatting: the error message
		// followed by the offending source span. Useful for runtime errors
		// (missing funds, etc.) that the static pass can't catch.
		msg := iErr.Error() + "\n" + iErr.GetRange().ShowOnSource(input.Script)
		return mustMarshal(runEnvelope{Ok: false, Error: msg})
	}

	value := &runValue{
		Postings:     result.Postings,
		TxMeta:       result.Metadata,
		AccountsMeta: result.AccountsMetadata,
	}
	if value.Postings == nil {
		value.Postings = []numscript.Posting{}
	}
	if value.TxMeta == nil {
		value.TxMeta = map[string]string{}
	}
	if value.AccountsMeta == nil {
		value.AccountsMeta = numscript.SetAccountsMetadata{}
	}

	return mustMarshal(runEnvelope{Ok: true, Value: value})
}

// toBalances flattens the nested {account: {asset: amount}} object into the
// interpreter's row shape, preserving integer precision via big.Int (the
// amounts arrive as json.Number, never float64).
func toBalances(nested map[string]map[string]json.Number) (numscript.Balances, error) {
	balances := numscript.Balances{}
	for account, assets := range nested {
		for asset, amount := range assets {
			n, ok := parseIntAmount(amount.String())
			if !ok {
				return nil, fmt.Errorf("invalid balance amount for %s %s: %q", account, asset, amount.String())
			}
			balances = append(balances, numscript.BalanceRow{
				Account: account,
				Asset:   asset,
				Amount:  n,
			})
		}
	}
	return balances, nil
}

// parseIntAmount parses an integer balance amount. Plain base-10 integers are
// taken exactly. A JS number can also reach us in scientific notation (e.g.
// "1e+21" for a large integer that overflowed float64's exact range), so we fall
// back to high-precision float parsing and accept it only when it is exactly
// integral — genuine fractional amounts like "10.5" are still rejected.
func parseIntAmount(s string) (*big.Int, bool) {
	if n, ok := new(big.Int).SetString(s, 10); ok {
		return n, true
	}

	f, _, err := big.ParseFloat(s, 10, 256, big.ToNearestEven)
	if err != nil {
		return nil, false
	}
	n, acc := f.Int(nil)
	if acc != big.Exact {
		return nil, false
	}
	return n, true
}

// toAccountsMetadata flattens the nested {account: {key: value}} object into the
// interpreter's row shape.
func toAccountsMetadata(nested map[string]map[string]string) numscript.AccountsMetadata {
	meta := numscript.AccountsMetadata{}
	for account, kv := range nested {
		for key, value := range kv {
			meta = append(meta, numscript.AccountMetadataRow{
				Account: account,
				Key:     key,
				Value:   value,
			})
		}
	}
	return meta
}

// ---- Needed variables ---------------------------------------------------------

type neededVarsEnvelope struct {
	Ok        bool              `json:"ok"`
	Variables map[string]string `json:"variables,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// NeededVariables returns the variables a script declares but does not define
// inline (name -> type), so a host can pre-populate an inputs editor.
func NeededVariables(source string) (out []byte) {
	defer func() {
		if r := recover(); r != nil {
			out = mustMarshal(neededVarsEnvelope{Ok: false, Error: fmt.Sprintf("internal error: %v", r)})
		}
	}()

	vars := numscript.Parse(source).GetNeededVariables()
	if vars == nil {
		vars = map[string]string{}
	}
	return mustMarshal(neededVarsEnvelope{Ok: true, Variables: vars})
}

// ---- Version ------------------------------------------------------------------

// Version reports the binding's own ABI version. It is independent of the
// numscript library version and lets a host detect an incompatible wasm asset.
const Version = "1"

// ---- helpers ------------------------------------------------------------------

// mustMarshal marshals values whose shape is fully controlled here, so an error
// is a programmer mistake, not a runtime condition. If it somehow fails we still
// return a valid JSON error envelope rather than panicking.
func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Appendf(nil, `{"ok":false,"error":%q}`, "failed to encode result: "+err.Error())
	}
	return b
}
