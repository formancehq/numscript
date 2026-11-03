// Package encode turns an IR program into an SMT model of its run over every
// initial balance: free Int inputs for the balances the program can read, one
// define-fun per state write, and named outputs for the run outcome and the
// amounts moved per (account, asset) pair.
package encode

import (
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/formancehq/numscript/internal/prove/smt"
)

// Key is a balance slot. Scope and color are "" in v1.
type Key struct{ Account, Asset string }

// MoveKey is a (source, destination, asset) posting aggregate.
type MoveKey struct{ From, To, Asset string }

func (k Key) String() string     { return "@" + k.Account + "," + k.Asset }
func (m MoveKey) String() string { return "@" + m.From + ",@" + m.To + "," + m.Asset }

func BeforeName(k Key) string    { return "before(" + k.String() + ")" }
func AfterName(k Key) string     { return "after(" + k.String() + ")" }
func SentName(k Key) string      { return "sent(" + k.String() + ")" }
func ReceivedName(k Key) string  { return "received(" + k.String() + ")" }
func MovedName(m MoveKey) string { return "moved(" + m.String() + ")" }

type FailKind int

const (
	MissingFunds FailKind = iota
	NegativeAmount
	NegativeBalance
	NegativePortion
	AllotmentSum
	AssetMismatch
	InvalidAccount
	DivideByZero
	UncappedSource
)

func (k FailKind) String() string {
	switch k {
	case MissingFunds:
		return "missing funds"
	case NegativeAmount:
		return "negative amount"
	case NegativeBalance:
		return "negative balance"
	case NegativePortion:
		return "negative portion"
	case AllotmentSum:
		return "invalid allotment sum"
	case AssetMismatch:
		return "asset mismatch"
	case InvalidAccount:
		return "invalid account name"
	case DivideByZero:
		return "division by zero"
	case UncappedSource:
		return "uncapped source"
	}
	return "?"
}

// FailSite is one instruction that can abort the run. Flags of different
// sites are never true together: each is guarded by the absence of every
// earlier failure on its path.
type FailSite struct {
	Index int
	Text  string
	Kind  FailKind
	Flag  smt.Term
}

type Problem struct {
	body     *smt.Script
	Keys     []Key
	MoveKeys []MoveKey
	// Accounts and Assets are the names the program can touch, sorted.
	Accounts []string
	Assets   []string

	Before   map[Key]smt.Term
	After    map[Key]smt.Term
	Sent     map[Key]smt.Term
	Received map[Key]smt.Term
	Moved    map[MoveKey]smt.Term
	Success  smt.Term
	Fails    smt.Term

	FailSites []FailSite
}

// Text is the complete SMT-LIB2 model: header, the before inputs, the body.
func (p *Problem) Text() string {
	var sb strings.Builder
	sb.WriteString(smt.Header)
	for _, k := range p.Keys {
		fmt.Fprintf(&sb, "(declare-const %s Int)\n", smt.Quote(BeforeName(k)))
	}
	for _, line := range p.body.Lines() {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	return sb.String()
}

// InputNames lists the before symbols, in Keys order.
func (p *Problem) InputNames() []string {
	out := make([]string, len(p.Keys))
	for i, k := range p.Keys {
		out[i] = BeforeName(k)
	}
	return out
}

// OutputNames lists every named output: outcome flags, fail sites, and the
// after/sent/received/moved aggregates.
func (p *Problem) OutputNames() []string {
	names := []string{"success", "fails"}
	for _, s := range p.FailSites {
		if n, ok := smt.SymbolName(s.Flag); ok {
			names = append(names, n)
		}
	}
	for _, k := range p.Keys {
		names = append(names, AfterName(k), SentName(k), ReceivedName(k))
	}
	for _, m := range p.MoveKeys {
		names = append(names, MovedName(m))
	}
	return names
}

// Evaluator evaluates the model's terms under concrete initial balances
// (missing keys are 0).
func (p *Problem) Evaluator(before map[Key]*big.Int) *smt.Evaluator {
	inputs := make(map[string]smt.Value, len(p.Keys))
	for _, k := range p.Keys {
		v := before[k]
		if v == nil {
			v = new(big.Int)
		}
		inputs[BeforeName(k)] = smt.IntValue(v)
	}
	return smt.NewEvaluator(p.body, inputs)
}

// UnsupportedError reports an instruction the encoder does not model.
type UnsupportedError struct {
	Index  int
	Text   string
	Reason string
}

func (e UnsupportedError) Error() string {
	return fmt.Sprintf("unsupported instruction %d `%s`: %s", e.Index, e.Text, e.Reason)
}

func sortedKeys(set map[Key]struct{}) []Key {
	out := make([]Key, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Account != out[j].Account {
			return out[i].Account < out[j].Account
		}
		return out[i].Asset < out[j].Asset
	})
	return out
}

func sortedMoveKeys(set map[MoveKey]struct{}) []MoveKey {
	out := make([]MoveKey, 0, len(set))
	for m := range set {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.From != b.From {
			return a.From < b.From
		}
		if a.To != b.To {
			return a.To < b.To
		}
		return a.Asset < b.Asset
	})
	return out
}

func sortedStrings(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// AddKeys declares pairs the program never touches, so a property may name
// them: a free before, after equal to it, and zero movements.
func (p *Problem) AddKeys(keys []Key, moves []MoveKey) {
	set := map[Key]struct{}{}
	for _, k := range p.Keys {
		set[k] = struct{}{}
	}
	added := false
	for _, k := range keys {
		if _, ok := set[k]; ok {
			continue
		}
		set[k] = struct{}{}
		added = true
		before := smt.Symbol(BeforeName(k), smt.SortInt)
		p.Before[k] = before
		p.Sent[k] = p.body.DefineAlways(SentName(k), smt.Int64(0))
		p.Received[k] = p.body.DefineAlways(ReceivedName(k), smt.Int64(0))
		p.After[k] = p.body.DefineAlways(AfterName(k), before)
	}
	if added {
		p.Keys = sortedKeys(set)
	}

	mset := map[MoveKey]struct{}{}
	for _, m := range p.MoveKeys {
		mset[m] = struct{}{}
	}
	added = false
	for _, m := range moves {
		if _, ok := mset[m]; ok {
			continue
		}
		mset[m] = struct{}{}
		added = true
		p.Moved[m] = p.body.DefineAlways(MovedName(m), smt.Int64(0))
	}
	if added {
		p.MoveKeys = sortedMoveKeys(mset)
	}
}
