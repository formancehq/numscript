// Package verify answers user questions about a numscript script by encoding
// the script's execution semantics as SMT-LIB2 (via the compiler's symbolic
// interpreter), appending a query written in a small DSL, and shelling out to a
// local `z3` binary for a definitive proof or a concrete counterexample.
package verify

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/formancehq/numscript/internal/compiler"
)

// Mode selects how the query's satisfiability answer maps to a verdict.
type Mode int

const (
	// Prove asks "does this always hold?" — we assert the negation and hope
	// for unsat.
	Prove Mode = iota
	// Find asks "can this happen?" — we assert the formula and hope for sat.
	Find
)

func (m Mode) String() string {
	if m == Find {
		return "find"
	}
	return "prove"
}

// Verdict is the user-facing outcome. Raw sat/unsat is never surfaced.
type Verdict int

const (
	Proved         Verdict = iota // prove: holds for every input
	Counterexample                // prove: a model falsifies it
	Impossible                    // find: no such case exists
	Witness                       // find: a model exhibits it
	Unknown                       // undecidable or timed out
)

func (v Verdict) String() string {
	switch v {
	case Proved:
		return "✅ Proved"
	case Counterexample:
		return "❌ Counterexample"
	case Impossible:
		return "❌ Impossible"
	case Witness:
		return "✅ Witness found"
	default:
		return "⚠️  Unknown"
	}
}

// Result is the outcome of a verification run.
type Result struct {
	Mode    Mode
	Verdict Verdict
	// Model is the z3 model text (only for Counterexample / Witness).
	Model string
	// SMT is the full script sent to z3 (useful for debugging).
	SMT string
	// Raw is z3's raw stdout.
	Raw string
}

// Options tunes a verification run.
type Options struct {
	// TimeoutMs is the per-query z3 timeout; 0 uses a default.
	TimeoutMs int
	// Z3Path overrides the z3 binary location ("z3" by default / PATH lookup).
	Z3Path string
}

const defaultTimeoutMs = 10000

// Verify encodes source, parses and resolves the query, runs z3, and maps the
// answer to a verdict. The query may be prefixed with `prove:` or `find:`;
// absent a prefix it defaults to prove (invariant-shaped queries).
func Verify(ctx context.Context, source, query string, opts Options) (*Result, error) {
	enc, err := compiler.SymbolicEncodeSource(source)
	if err != nil {
		return nil, err
	}

	mode, body := splitMode(query)

	ast, err := parseQuery(body)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	goal, err := resolve(ast, enc.Symbols)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	if goal.kind != kindBool {
		return nil, fmt.Errorf("query: top-level expression must be boolean, got %s", goal.kind)
	}

	timeout := opts.TimeoutMs
	if timeout == 0 {
		timeout = defaultTimeoutMs
	}

	var b strings.Builder
	b.WriteString("(set-option :produce-models true)\n")
	fmt.Fprintf(&b, "(set-option :timeout %d)\n", timeout)
	b.WriteString(enc.SMTLIB)
	if mode == Prove {
		fmt.Fprintf(&b, "(assert (not %s))\n", goal.smt)
	} else {
		fmt.Fprintf(&b, "(assert %s)\n", goal.smt)
	}
	b.WriteString("(check-sat)\n(get-model)\n")
	script := b.String()

	sat, model, raw, err := runZ3(ctx, opts.Z3Path, script)
	if err != nil {
		return nil, err
	}

	res := &Result{Mode: mode, SMT: script, Raw: raw}
	switch sat {
	case "unsat":
		if mode == Prove {
			res.Verdict = Proved
		} else {
			res.Verdict = Impossible
		}
	case "sat":
		res.Model = model
		if mode == Prove {
			res.Verdict = Counterexample
		} else {
			res.Verdict = Witness
		}
	default:
		res.Verdict = Unknown
	}
	return res, nil
}

func splitMode(query string) (Mode, string) {
	q := strings.TrimSpace(query)
	switch {
	case strings.HasPrefix(q, "prove:"):
		return Prove, strings.TrimPrefix(q, "prove:")
	case strings.HasPrefix(q, "find:"):
		return Find, strings.TrimPrefix(q, "find:")
	default:
		return Prove, q
	}
}

// runZ3 feeds the script to z3 on stdin and returns the sat line, the model
// block (everything after the sat line) and the raw stdout.
func runZ3(ctx context.Context, z3Path, script string) (sat, model, raw string, err error) {
	if z3Path == "" {
		z3Path = "z3"
	}
	cmd := exec.CommandContext(ctx, z3Path, "-in")
	cmd.Stdin = strings.NewReader(script)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	// z3 exits non-zero on some errors but still prints usable output; we key
	// off the parsed first line rather than the exit code.
	_ = cmd.Run()

	raw = out.String()
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", "", raw, fmt.Errorf("z3 produced no output (stderr: %s)", strings.TrimSpace(errBuf.String()))
	}

	// Walk the output line by line. The verdict is the first standalone
	// sat/unsat/unknown line. A `(error ...)` line *before* the verdict means
	// we emitted malformed SMT — a hard failure. A `(error ...)` *after* the
	// verdict (e.g. "model is not available" when get-model runs on an unsat
	// result) is benign and ignored.
	lines := strings.Split(trimmed, "\n")
	verdictIdx := -1
	for i, ln := range lines {
		s := strings.TrimSpace(ln)
		if s == "sat" || s == "unsat" || s == "unknown" {
			sat = s
			verdictIdx = i
			break
		}
		if strings.HasPrefix(s, "(error ") {
			return "", "", raw, fmt.Errorf("z3 rejected the SMT as invalid:\n%s", trimmed)
		}
	}
	if verdictIdx == -1 {
		return "", "", raw, fmt.Errorf("unexpected z3 output: %s", trimmed)
	}
	if verdictIdx+1 < len(lines) {
		model = strings.TrimSpace(strings.Join(lines[verdictIdx+1:], "\n"))
	}
	return sat, model, raw, nil
}
