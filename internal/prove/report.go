package prove

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/ansi"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/prove/encode"
	"github.com/formancehq/numscript/internal/prove/prop"
)

func yesNo(b *bool) string {
	switch {
	case b == nil:
		return "unknown"
	case *b:
		return "yes"
	}
	return "no"
}

func (r *Report) touched() string {
	accounts := make([]string, len(r.Problem.Accounts))
	for i, a := range r.Problem.Accounts {
		accounts[i] = "@" + a
	}
	return strings.Join(accounts, ", ") + " in " + strings.Join(r.Problem.Assets, ", ")
}

func statusIcon(s Status) string {
	switch s {
	case Valid:
		return ansi.ColorGreen("✓")
	case Counterexample:
		return ansi.ColorRed("×")
	case Unknown:
		return ansi.ColorYellow("?")
	}
	return ansi.ColorRed("!")
}

// Pretty renders the human report.
func (r *Report) Pretty(title string) string {
	w := &strings.Builder{}
	fmt.Fprintf(w, "%s: touches %s; success reachable: %s; failure reachable: %s\n\n",
		title, r.touched(), yesNo(r.SuccessReachable), yesNo(r.FailureReachable))

	width := 0
	for _, res := range r.Results {
		width = max(width, len(res.Property.Name))
	}
	for _, res := range r.Results {
		fmt.Fprintf(w, "  %s %-*s  %-14s %s\n", statusIcon(res.Status), width, res.Property.Name, res.Status, res.Duration.Round(1e6))
	}

	counts := map[Status]int{}
	for _, res := range r.Results {
		counts[res.Status]++
		if res.Status == Valid || res.Status == Unknown {
			continue
		}
		fmt.Fprintf(w, "\n%s %s: %s\n", statusIcon(res.Status), res.Property.Name, prop.Dump(res.Property.Formula))
		switch res.Status {
		case Counterexample, ToolBug:
			if res.Before != nil {
				fmt.Fprintln(w, "  initial balances:")
				fmt.Fprint(w, indent(balanceRows(res.Before).PrettyPrint(), "    "))
				fmt.Fprintf(w, "  outcome: %s\n", r.outcomeLine(res.Model))
				fmt.Fprintln(w, "  atoms:")
				for _, av := range res.AtomValues() {
					fmt.Fprintf(w, "    %s = %s\n", av.Atom, av.Value)
				}
			}
			if res.Status == ToolBug {
				fmt.Fprintln(w, "  the prover's model does not match the vm; this is a bug in the prover, not in the script:")
				for _, d := range res.Diffs {
					fmt.Fprintf(w, "    %s\n", d)
				}
				if res.Replayed {
					fmt.Fprintln(w, "    the property holds on the vm run")
				}
				if res.Err != nil {
					fmt.Fprintf(w, "    %s\n", res.Err)
				}
			}
		case SolverError:
			fmt.Fprintf(w, "  %s\n", res.Err)
		}
	}

	var parts []string
	for _, s := range []Status{Valid, Counterexample, Unknown, ToolBug, SolverError} {
		if counts[s] != 0 {
			parts = append(parts, fmt.Sprintf("%d %s", counts[s], s))
		}
	}
	fmt.Fprintf(w, "\n%s\n", strings.Join(parts, ", "))
	return w.String()
}

func (r *Report) outcomeLine(o encode.Outcome) string {
	if o.Success {
		return "success"
	}
	for _, s := range r.Problem.FailSites {
		if s.Index == o.FailIndex {
			return fmt.Sprintf("fails at instruction %d `%s`: %s", o.FailIndex, s.Text, o.FailKind)
		}
	}
	return fmt.Sprintf("fails: %s", o.FailKind)
}

func indent(s, prefix string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n") + "\n"
}

func balanceRows(before map[encode.Key]*big.Int) interpreter.Balances {
	keys := make([]encode.Key, 0, len(before))
	for k := range before {
		keys = append(keys, k)
	}
	rows := make(interpreter.Balances, 0, len(keys))
	for _, k := range sortKeys(keys) {
		rows = append(rows, interpreter.BalanceRow{Account: k.Account, Asset: k.Asset, Amount: before[k]})
	}
	return rows
}

func sortKeys(keys []encode.Key) []encode.Key {
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && less(keys[j], keys[j-1]); j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

func less(a, b encode.Key) bool {
	if a.Account != b.Account {
		return a.Account < b.Account
	}
	return a.Asset < b.Asset
}

type jsonReport struct {
	Script           string       `json:"script"`
	Accounts         []string     `json:"accounts"`
	Assets           []string     `json:"assets"`
	SuccessReachable *bool        `json:"successReachable"`
	FailureReachable *bool        `json:"failureReachable"`
	Results          []jsonResult `json:"results"`
}

type jsonResult struct {
	Name           string              `json:"name"`
	Formula        string              `json:"formula"`
	Status         Status              `json:"status"`
	DurationMs     int64               `json:"durationMs"`
	Error          string              `json:"error,omitempty"`
	Counterexample *jsonCounterexample `json:"counterexample,omitempty"`
}

type jsonCounterexample struct {
	Balances interpreter.Balances `json:"balances"`
	Outcome  string               `json:"outcome"`
	Failure  *jsonFailure         `json:"failure,omitempty"`
	Atoms    map[string]string    `json:"atoms"`
	Diffs    []string             `json:"modelVmDiffs,omitempty"`
}

type jsonFailure struct {
	Kind        string `json:"kind"`
	Instruction int    `json:"instruction"`
}

// JSON renders the report for machines. Balance amounts are numbers, atom
// values are strings.
func (r *Report) JSON(title string) ([]byte, error) {
	out := jsonReport{
		Script:           title,
		Accounts:         r.Problem.Accounts,
		Assets:           r.Problem.Assets,
		SuccessReachable: r.SuccessReachable,
		FailureReachable: r.FailureReachable,
		Results:          []jsonResult{},
	}
	for _, res := range r.Results {
		jr := jsonResult{
			Name:       res.Property.Name,
			Formula:    prop.Dump(res.Property.Formula),
			Status:     res.Status,
			DurationMs: res.Duration.Milliseconds(),
		}
		if res.Err != nil {
			jr.Error = res.Err.Error()
		}
		if res.Before != nil {
			ce := &jsonCounterexample{Balances: balanceRows(res.Before), Outcome: "success", Atoms: map[string]string{}, Diffs: res.Diffs}
			if !res.Model.Success {
				ce.Outcome = "fails"
				ce.Failure = &jsonFailure{Kind: res.Model.FailKind.String(), Instruction: res.Model.FailIndex}
			}
			for _, av := range res.AtomValues() {
				ce.Atoms[av.Atom.String()] = av.Value.String()
			}
			jr.Counterexample = ce
		}
		out.Results = append(out.Results, jr)
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
