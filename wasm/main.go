//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"
	"sync"
	"syscall/js"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
)

type diagnostic struct {
	StartLineNumber int    `json:"startLineNumber"`
	StartColumn     int    `json:"startColumn"`
	EndLineNumber   int    `json:"endLineNumber"`
	EndColumn       int    `json:"endColumn"`
	Message         string `json:"message"`
	Severity        string `json:"severity"`
}

type runArgs struct {
	Variables    map[string]string              `json:"variables"`
	Balances     map[string]map[string]*big.Int `json:"balances"`
	Metadata     map[string]map[string]string   `json:"metadata"`
	FeatureFlags []string                       `json:"featureFlags"`
}

type errResult struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type okResult struct {
	Ok    bool    `json:"ok"`
	Value okValue `json:"value"`
}

type okValue struct {
	Postings     []interpreter.Posting        `json:"postings"`
	TxMeta       map[string]string            `json:"txMeta"`
	AccountsMeta map[string]map[string]string `json:"accountsMeta"`
}

// parseCache memoizes parses keyed by source, so repeated ops against an
// unchanged buffer (diagnostics, then run, or LSP calls later) parse once.
// It is a content-addressed LRU: a hit is byte-identical source, a miss just
// re-parses, so there is no lifecycle for the caller to manage.
type parseCache struct {
	mu      sync.Mutex
	max     int
	order   []string
	entries map[string]parser.ParseResult
}

var cache = &parseCache{max: 1, entries: map[string]parser.ParseResult{}}

func (c *parseCache) getOrParse(source string) parser.ParseResult {
	c.mu.Lock()
	defer c.mu.Unlock()

	if pr, ok := c.entries[source]; ok {
		c.touch(source)
		return pr
	}

	pr := parser.Parse(source)
	c.entries[source] = pr
	c.order = append(c.order, source)
	c.evict()
	return pr
}

func (c *parseCache) touch(source string) {
	for i, s := range c.order {
		if s == source {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
	c.order = append(c.order, source)
}

// max <= 0 means unbounded.
func (c *parseCache) evict() {
	for c.max > 0 && len(c.order) > c.max {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.entries, oldest)
	}
}

func (c *parseCache) setMax(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.max = n
	c.evict()
}

func toDiagnostic(r parser.Range, msg string, severity string) diagnostic {
	return diagnostic{
		StartLineNumber: r.Start.Line + 1,
		StartColumn:     r.Start.Character + 1,
		EndLineNumber:   r.End.Line + 1,
		EndColumn:       r.End.Character + 1,
		Message:         msg,
		Severity:        severity,
	}
}

// buildDiagnostics mirrors analysis.CheckSource but reuses the cached parse:
// parse errors become diagnostics, then static analysis is appended.
func buildDiagnostics(pr parser.ParseResult) []diagnostic {
	out := []diagnostic{}

	for _, e := range pr.Errors {
		out = append(out, toDiagnostic(e.Range, e.Msg, "error"))
	}

	check := analysis.CheckProgram(pr.Value)
	for _, d := range check.Diagnostics {
		severity := "warning"
		if d.Kind.Severity() == analysis.ErrorSeverity {
			severity = "error"
		}
		out = append(out, toDiagnostic(d.Range, d.Kind.Message(), severity))
	}

	return out
}

func analyze(_ js.Value, args []js.Value) any {
	source := args[0].String()
	pr := cache.getOrParse(source)
	b, err := json.Marshal(buildDiagnostics(pr))
	if err != nil {
		return "[]"
	}
	return string(b)
}

func run(_ js.Value, args []js.Value) any {
	b, err := json.Marshal(runImpl(args[0].String(), args[1].String()))
	if err != nil {
		errBytes, _ := json.Marshal(errResult{Ok: false, Error: err.Error()})
		return string(errBytes)
	}
	return string(b)
}

func runImpl(source string, argsJSON string) any {
	pr := cache.getOrParse(source)

	if len(pr.Errors) != 0 {
		return errResult{Ok: false, Error: parser.ParseErrorsToString(pr.Errors, source)}
	}

	check := analysis.CheckProgram(pr.Value)
	var sb strings.Builder
	hasError := false
	for _, d := range check.Diagnostics {
		if d.Kind.Severity() != analysis.ErrorSeverity {
			continue
		}
		hasError = true
		sb.WriteString(analysis.SeverityToString(d.Kind.Severity()))
		sb.WriteString(": ")
		sb.WriteString(d.Kind.Message())
		sb.WriteString("\n")
		sb.WriteString(d.Range.ShowOnSource(source))
		sb.WriteString("\n")
	}
	if hasError {
		return errResult{Ok: false, Error: sb.String()}
	}

	var a runArgs
	if err := json.Unmarshal([]byte(argsJSON), &a); err != nil {
		return errResult{Ok: false, Error: err.Error()}
	}

	store := interpreter.StaticStore{}
	for account, assets := range a.Balances {
		for asset, amount := range assets {
			store.Balances = append(store.Balances, interpreter.BalanceRow{
				Account: account,
				Asset:   asset,
				Amount:  amount,
			})
		}
	}
	for account, kv := range a.Metadata {
		for key, value := range kv {
			store.Meta = append(store.Meta, interpreter.AccountMetadataRow{
				Account: account,
				Key:     key,
				Value:   value,
			})
		}
	}

	flags := map[string]struct{}{}
	for _, f := range a.FeatureFlags {
		flags[f] = struct{}{}
	}

	res, ierr := interpreter.RunProgram(context.Background(), pr.Value, a.Variables, store, flags)
	if ierr != nil {
		return errResult{Ok: false, Error: ierr.Error()}
	}

	// The frontend renders metadata as plain strings, so flatten the
	// interpreter's typed values back to their string form.
	txMeta := map[string]string{}
	for k, v := range res.Metadata {
		txMeta[k] = valueToString(v)
	}
	accountsMeta := map[string]map[string]string{}
	for _, row := range res.AccountsMetadata {
		if accountsMeta[row.Account] == nil {
			accountsMeta[row.Account] = map[string]string{}
		}
		accountsMeta[row.Account][row.Key] = valueToString(row.Value)
	}

	return okResult{
		Ok: true,
		Value: okValue{
			Postings:     res.Postings,
			TxMeta:       txMeta,
			AccountsMeta: accountsMeta,
		},
	}
}

// valueToString renders a metadata value for display: plain strings are
// unwrapped (String.String() would quote them), other types keep their
// canonical form (e.g. "@bob", "USD/2 100").
func valueToString(v interpreter.Value) string {
	if s, ok := v.(interpreter.String); ok {
		return string(s)
	}
	return v.String()
}

func configureCache(_ js.Value, args []js.Value) any {
	cache.setMax(args[0].Int())
	return nil
}

func main() {
	ns := js.Global().Get("Object").New()
	ns.Set("analyze", js.FuncOf(analyze))
	ns.Set("run", js.FuncOf(run))
	ns.Set("configureCache", js.FuncOf(configureCache))
	js.Global().Set("__numscript", ns)

	if ready := js.Global().Get("__numscriptReady"); ready.Type() == js.TypeFunction {
		ready.Invoke()
	}

	select {}
}
