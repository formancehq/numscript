// Package prove checks properties of an IR program for every initial balance:
// the program is encoded into SMT, each property's negation is asserted, and
// a model (if any) is replayed on the real VM before being reported as a
// counterexample.
package prove

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/prove/encode"
	"github.com/formancehq/numscript/internal/prove/prop"
	"github.com/formancehq/numscript/internal/prove/smt"
)

type Options struct {
	Solver smt.Solver
}

// PropertyError is a property that does not parse or typecheck.
type PropertyError struct {
	Property prop.Property
	Errors   []prop.Error
}

func (e PropertyError) Error() string {
	msgs := make([]string, len(e.Errors))
	for i, err := range e.Errors {
		msgs[i] = err.Error()
	}
	return fmt.Sprintf("property %q: %s", e.Property.Name, strings.Join(msgs, "; "))
}

// Check encodes instrs once and checks every property against it.
func Check(ctx context.Context, instrs []ir.Instr, props []prop.Property, opts Options) (*Report, error) {
	p, err := encode.Encode(instrs, nil, nil)
	if err != nil {
		return nil, err
	}
	known := prop.Known{Accounts: map[string]struct{}{}, Assets: map[string]struct{}{}}
	for _, a := range p.Accounts {
		known.Accounts[a] = struct{}{}
	}
	for _, a := range p.Assets {
		known.Assets[a] = struct{}{}
	}

	var extraKeys []encode.Key
	var extraMoves []encode.MoveKey
	for _, pr := range props {
		if errs := prop.Typecheck(pr.Formula, known); len(errs) != 0 {
			return nil, PropertyError{Property: pr, Errors: errs}
		}
		for _, a := range prop.Atoms(pr.Formula) {
			if a.Kind == prop.Moved {
				extraMoves = append(extraMoves, encode.MoveKey{From: a.Account, To: a.Account2, Asset: a.Asset})
			} else {
				extraKeys = append(extraKeys, encode.Key{Account: a.Account, Asset: a.Asset})
			}
		}
	}
	p.AddKeys(extraKeys, extraMoves)

	report := &Report{Problem: p}
	report.SuccessReachable = reachable(ctx, p, p.Success, opts.Solver)
	report.FailureReachable = reachable(ctx, p, p.Fails, opts.Solver)
	for _, pr := range props {
		report.Results = append(report.Results, checkOne(ctx, instrs, p, pr, opts.Solver))
	}
	return report, nil
}

func reachable(ctx context.Context, p *encode.Problem, flag smt.Term, solver smt.Solver) *bool {
	res, err := solver.CheckSat(ctx, p.Text()+"(assert "+smt.Print(flag)+")\n", nil)
	if err != nil || res.Status == smt.Unknown {
		return nil
	}
	v := res.Status == smt.Sat
	return &v
}

func resolver(p *encode.Problem) func(prop.Atom) smt.Term {
	return func(a prop.Atom) smt.Term {
		k := encode.Key{Account: a.Account, Asset: a.Asset}
		switch a.Kind {
		case prop.Before:
			return p.Before[k]
		case prop.After:
			return p.After[k]
		case prop.Sent:
			return p.Sent[k]
		case prop.Received:
			return p.Received[k]
		case prop.Moved:
			return p.Moved[encode.MoveKey{From: a.Account, To: a.Account2, Asset: a.Asset}]
		}
		panic("unknown atom kind")
	}
}

// Problem returns the SMT-LIB2 text of a property's problem: the model, the
// property, and the assertion of its negation.
func Problem(p *encode.Problem, pr prop.Property) string {
	property := prop.Lower(pr.Formula, resolver(p), p.Success)
	var sb strings.Builder
	sb.WriteString(p.Text())
	sb.WriteString("\n; property ")
	sb.WriteString(pr.Name)
	sb.WriteString(": ")
	sb.WriteString(prop.Dump(pr.Formula))
	sb.WriteString("\n(define-fun property () Bool ")
	sb.WriteString(smt.Print(property))
	sb.WriteString(")\n(assert (not property))\n")
	return sb.String()
}

func checkOne(ctx context.Context, instrs []ir.Instr, p *encode.Problem, pr prop.Property, solver smt.Solver) (result Result) {
	start := time.Now()
	result = Result{Property: pr, SMT: Problem(p, pr)}
	defer func() { result.Duration = time.Since(start) }()

	res, err := solver.CheckSat(ctx, result.SMT, append(p.InputNames(), p.OutputNames()...))
	if err != nil {
		result.Status, result.Err = SolverError, err
		return result
	}
	switch res.Status {
	case smt.Unsat:
		result.Status = Valid
		return result
	case smt.Unknown:
		result.Status = Unknown
		return result
	}

	before, model, err := p.OutcomeFromValues(res.Values)
	if err != nil {
		result.Status, result.Err = SolverError, err
		return result
	}
	result.Before, result.Model = before, model
	replay, err := encode.Replay(ctx, instrs, before, p.Keys, p.MoveKeys)
	if err != nil {
		result.Status, result.Err = ToolBug, err
		return result
	}
	result.Replay = replay
	result.Diffs = encode.Compare(model, replay)
	result.Replayed = prop.Eval(pr.Formula, lookup(before, replay), replay.Success)
	if len(result.Diffs) != 0 || result.Replayed {
		result.Status = ToolBug
		return result
	}
	result.Status = Counterexample
	return result
}

func lookup(before map[encode.Key]*big.Int, out encode.Outcome) func(prop.Atom) *big.Int {
	orZero := func(v *big.Int) *big.Int {
		if v == nil {
			return new(big.Int)
		}
		return v
	}
	return func(a prop.Atom) *big.Int {
		k := encode.Key{Account: a.Account, Asset: a.Asset}
		switch a.Kind {
		case prop.Before:
			return orZero(before[k])
		case prop.After:
			if v, ok := out.After[k]; ok {
				return v
			}
			return orZero(before[k])
		case prop.Sent:
			return orZero(out.Sent[k])
		case prop.Received:
			return orZero(out.Received[k])
		case prop.Moved:
			return orZero(out.Moved[encode.MoveKey{From: a.Account, To: a.Account2, Asset: a.Asset}])
		}
		panic("unknown atom kind")
	}
}

// AtomValues lists every atom of the property with its value on the
// counterexample, in source order of first appearance.
func (r Result) AtomValues() []AtomValue {
	look := lookup(r.Before, r.Replay)
	atoms := prop.Atoms(r.Property.Formula)
	out := make([]AtomValue, len(atoms))
	for i, a := range atoms {
		out[i] = AtomValue{Atom: a, Value: look(a)}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Atom.String() < out[j].Atom.String() })
	return out
}

type AtomValue struct {
	Atom  prop.Atom
	Value *big.Int
}
