package encode

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/formancehq/numscript/internal/funds"
	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/prove/smt"
	"github.com/formancehq/numscript/internal/vm"
)

// Outcome is what a run did: the outcome flag and the aggregates, either
// read off a solver model or computed from a real VM run.
type Outcome struct {
	Success bool
	// FailKind and FailIndex describe the abort. FailIndex is -1 when the
	// outcome comes from the VM, which does not report positions.
	FailKind  FailKind
	FailIndex int

	Sent     map[Key]*big.Int
	Received map[Key]*big.Int
	After    map[Key]*big.Int
	Moved    map[MoveKey]*big.Int
}

func newOutcome() Outcome {
	return Outcome{
		FailIndex: -1,
		Sent:      map[Key]*big.Int{},
		Received:  map[Key]*big.Int{},
		After:     map[Key]*big.Int{},
		Moved:     map[MoveKey]*big.Int{},
	}
}

// OutcomeFromValues reads a solver model: the before inputs and the outcome.
func (p *Problem) OutcomeFromValues(vals map[string]smt.Value) (map[Key]*big.Int, Outcome, error) {
	get := func(name string) (smt.Value, error) {
		v, ok := vals[name]
		if !ok {
			return smt.Value{}, fmt.Errorf("encode: the model has no value for %s", name)
		}
		return v, nil
	}
	before := map[Key]*big.Int{}
	for _, k := range p.Keys {
		v, err := get(BeforeName(k))
		if err != nil {
			return nil, Outcome{}, err
		}
		before[k] = v.Int
	}
	return before, p.outcome(func(t smt.Term) (smt.Value, error) {
		if n, ok := smt.SymbolName(t); ok {
			return get(n)
		}
		if b, ok := smt.AsBool(t); ok {
			return smt.BoolValue(b), nil
		}
		if n, ok := smt.AsInt(t); ok {
			return smt.IntValue(n), nil
		}
		return smt.Value{}, fmt.Errorf("encode: output %s is not a named symbol", smt.Print(t))
	}), nil
}

// OutcomeFromBalances evaluates the model under concrete initial balances,
// without a solver.
func (p *Problem) OutcomeFromBalances(before map[Key]*big.Int) Outcome {
	ev := p.Evaluator(before)
	return p.outcome(ev.Eval)
}

func (p *Problem) outcome(eval func(smt.Term) (smt.Value, error)) Outcome {
	must := func(t smt.Term) smt.Value {
		v, err := eval(t)
		if err != nil {
			panic(err)
		}
		return v
	}
	out := newOutcome()
	out.Success = must(p.Success).Bool
	if !out.Success {
		for _, s := range p.FailSites {
			if must(s.Flag).Bool {
				out.FailKind = s.Kind
				out.FailIndex = s.Index
				break
			}
		}
	}
	for _, k := range p.Keys {
		out.Sent[k] = must(p.Sent[k]).Int
		out.Received[k] = must(p.Received[k]).Int
		out.After[k] = must(p.After[k]).Int
	}
	for _, m := range p.MoveKeys {
		out.Moved[m] = must(p.Moved[m]).Int
	}
	return out
}

type staticStore map[funds.PairKey]*big.Int

func (s staticStore) GetBalance(_ context.Context, account, scope, asset, color string) (*big.Int, error) {
	if v, ok := s[funds.PairKey{Account: account, Scope: scope, Asset: asset, Color: color}]; ok {
		return new(big.Int).Set(v), nil
	}
	return new(big.Int), nil
}

func (s staticStore) GetMetadata(context.Context, string, string, string) (string, bool, error) {
	return "", false, nil
}

// Replay runs instrs on the real VM from the given initial balances and
// aggregates its postings over keys and moves (plus any pair a posting
// touches).
func Replay(ctx context.Context, instrs []ir.Instr, before map[Key]*big.Int, keys []Key, moves []MoveKey) (Outcome, error) {
	prog, err := ir.Assemble(instrs)
	if err != nil {
		return Outcome{}, err
	}
	store := staticStore{}
	for k, v := range before {
		store[funds.PairKey{Account: k.Account, Asset: k.Asset}] = v
	}

	out := newOutcome()
	for _, k := range keys {
		out.Sent[k] = new(big.Int)
		out.Received[k] = new(big.Int)
	}
	for _, m := range moves {
		out.Moved[m] = new(big.Int)
	}

	result, execErr := vm.Exec(ctx, vm.NewVm(prog), nil, store)
	if execErr != nil {
		kind, ok := failKindOf(execErr)
		if !ok {
			return Outcome{}, fmt.Errorf("encode: replay: %w", execErr)
		}
		out.FailKind = kind
	} else {
		out.Success = true
		for _, p := range result.Postings {
			src, dst := Key{p.Source, p.Asset}, Key{p.Destination, p.Asset}
			addTo(out.Sent, src, p.Amount)
			addTo(out.Received, dst, p.Amount)
			m := MoveKey{p.Source, p.Destination, p.Asset}
			if out.Moved[m] == nil {
				out.Moved[m] = new(big.Int)
			}
			out.Moved[m].Add(out.Moved[m], p.Amount)
		}
	}
	for k := range out.Sent {
		if out.Received[k] == nil {
			out.Received[k] = new(big.Int)
		}
	}
	for k := range out.Received {
		if out.Sent[k] == nil {
			out.Sent[k] = new(big.Int)
		}
		b := before[k]
		if b == nil {
			b = new(big.Int)
		}
		out.After[k] = new(big.Int).Add(b, new(big.Int).Sub(out.Received[k], out.Sent[k]))
	}
	return out, nil
}

func addTo(m map[Key]*big.Int, k Key, v *big.Int) {
	if m[k] == nil {
		m[k] = new(big.Int)
	}
	m[k].Add(m[k], v)
}

func failKindOf(err error) (FailKind, bool) {
	var (
		missing   vm.MissingFundsError
		negAmount vm.NegativeAmountError
		negBal    vm.NegativeBalanceError
		negPor    vm.NegativePortionError
		allot     vm.InvalidAllotmentSum
		mismatch  vm.AssetMismatchError
		account   vm.InvalidAccountName
		divZero   vm.DivideByZeroError
		uncapped  vm.InvalidUncappedSource
	)
	switch {
	case errors.As(err, &missing):
		return MissingFunds, true
	case errors.As(err, &negAmount):
		return NegativeAmount, true
	case errors.As(err, &negBal):
		return NegativeBalance, true
	case errors.As(err, &negPor):
		return NegativePortion, true
	case errors.As(err, &allot):
		return AllotmentSum, true
	case errors.As(err, &mismatch):
		return AssetMismatch, true
	case errors.As(err, &account):
		return InvalidAccount, true
	case errors.As(err, &divZero):
		return DivideByZero, true
	case errors.As(err, &uncapped):
		return UncappedSource, true
	}
	return 0, false
}

// Compare lists every way model and replay disagree. Pairs missing on one
// side count as zero.
func Compare(model, replay Outcome) []string {
	var diffs []string
	if model.Success != replay.Success {
		diffs = append(diffs, fmt.Sprintf("outcome: model %s, vm %s", outcomeWord(model), outcomeWord(replay)))
		return diffs
	}
	if !model.Success && model.FailKind != replay.FailKind {
		diffs = append(diffs, fmt.Sprintf("failure: model %s, vm %s", model.FailKind, replay.FailKind))
	}
	cmpKeyed := func(name string, a, b map[Key]*big.Int) {
		set := map[Key]struct{}{}
		for k := range a {
			set[k] = struct{}{}
		}
		for k := range b {
			set[k] = struct{}{}
		}
		for _, k := range sortedKeys(set) {
			x, y := orZeroInt(a[k]), orZeroInt(b[k])
			if x.Cmp(y) != 0 {
				diffs = append(diffs, fmt.Sprintf("%s(%s): model %s, vm %s", name, k, x, y))
			}
		}
	}
	cmpKeyed("sent", model.Sent, replay.Sent)
	cmpKeyed("received", model.Received, replay.Received)
	cmpKeyed("after", model.After, replay.After)
	mset := map[MoveKey]struct{}{}
	for m := range model.Moved {
		mset[m] = struct{}{}
	}
	for m := range replay.Moved {
		mset[m] = struct{}{}
	}
	for _, m := range sortedMoveKeys(mset) {
		x, y := orZeroInt(model.Moved[m]), orZeroInt(replay.Moved[m])
		if x.Cmp(y) != 0 {
			diffs = append(diffs, fmt.Sprintf("moved(%s): model %s, vm %s", m, x, y))
		}
	}
	sort.Strings(diffs)
	return diffs
}

func outcomeWord(o Outcome) string {
	if o.Success {
		return "success"
	}
	return "fails (" + o.FailKind.String() + ")"
}

func orZeroInt(v *big.Int) *big.Int {
	if v == nil {
		return new(big.Int)
	}
	return v
}
