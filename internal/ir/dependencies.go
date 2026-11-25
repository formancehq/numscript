package ir

import (
	"context"
	"fmt"
	"maps"
	"math/big"

	"github.com/formancehq/numscript/internal/funds"
	"github.com/formancehq/numscript/internal/vm"
)

type AccountDependency struct {
	Account string
	Scope   string
	Color   string
	Asset   string
}

type MetaDependency struct {
	Account string
	Scope   string
	Key     string
}

type Dependencies struct {
	AccountsReads  map[AccountDependency]struct{}
	AccountsWrites map[AccountDependency]struct{}
	MetaReads      map[MetaDependency]struct{}
	MetaWrites     map[MetaDependency]struct{}
	TxMetaWrites   map[string]struct{}
}

// UnresolvedOperand is returned when an operand that names a dependency (an
// account, asset, color, scope or metadata key) depends on the outcome of a
// pull, so no single value can be given for it.
type UnresolvedOperand struct {
	Instr   Instr
	Operand string
}

func (e UnresolvedOperand) Error() string {
	return fmt.Sprintf("cannot statically resolve the %s operand of %T", e.Operand, e.Instr)
}

// absVal is a register's value during the analysis: a *big.Int, string,
// *big.Rat or bool, or unknown when known is false. Known values are never
// mutated after being stored.
type absVal struct {
	known bool
	v     any
}

var unknown = absVal{}

func knownVal(v any) absVal { return absVal{known: true, v: v} }

func (a absVal) equal(b absVal) bool {
	if !a.known || !b.known {
		return false
	}
	switch av := a.v.(type) {
	case *big.Int:
		bv, ok := b.v.(*big.Int)
		return ok && av.Cmp(bv) == 0
	case *big.Rat:
		bv, ok := b.v.(*big.Rat)
		return ok && av.Cmp(bv) == 0
	case string:
		bv, ok := b.v.(string)
		return ok && av == bv
	case bool:
		bv, ok := b.v.(bool)
		return ok && av == bv
	default:
		return false
	}
}

type balanceKey struct {
	account, scope, asset string
}

type depState struct {
	regs  []absVal
	asset absVal
	// no branch on an unknown condition lies on any path reaching here, so
	// (barring a funds error) execution reaches this point
	certain bool
	// colors pulled since the last set_current_asset: the funds queue carries
	// them to whichever destination the send reaches
	colors map[string]struct{}
	// balances a pull or save may have changed on some path reaching here
	dirty map[balanceKey]struct{}
}

func (s *depState) clone() *depState {
	return &depState{
		regs:    append([]absVal(nil), s.regs...),
		asset:   s.asset,
		certain: s.certain,
		colors:  maps.Clone(s.colors),
		dirty:   maps.Clone(s.dirty),
	}
}

// join merges src into dst, which is nil when no path has reached it yet.
func join(dst *depState, src *depState) *depState {
	if src == nil {
		return dst
	}
	if dst == nil {
		return src.clone()
	}
	for i := range dst.regs {
		if !dst.regs[i].equal(src.regs[i]) {
			dst.regs[i] = unknown
		}
	}
	if !dst.asset.equal(src.asset) {
		dst.asset = unknown
	}
	dst.certain = dst.certain && src.certain
	maps.Copy(dst.colors, src.colors)
	maps.Copy(dst.dirty, src.dirty)
	return dst
}

type depAnalysis struct {
	ctx   context.Context
	vars  *vm.Vars
	store vm.Store
	deps  Dependencies
}

// ResolveDependencies computes the accounts and metadata a program can touch,
// without running the funds engine: registers are propagated as constants, the
// amount a pull yields is unknown, and a jump on an unknown condition follows
// both edges. IR jumps only go forward, so one pass in instruction order reaches
// the fixpoint.
//
// Balances and metadata are read from the store, as execution would, so the
// result is exact for the given store and vars. It over-approximates where a
// branch depends on pulled amounts.
func ResolveDependencies(ctx context.Context, instrs []Instr, vars *vm.Vars, store vm.Store) (Dependencies, error) {
	a := &depAnalysis{
		ctx:   ctx,
		vars:  vars,
		store: store,
		deps: Dependencies{
			AccountsReads:  map[AccountDependency]struct{}{},
			AccountsWrites: map[AccountDependency]struct{}{},
			MetaReads:      map[MetaDependency]struct{}{},
			MetaWrites:     map[MetaDependency]struct{}{},
			TxMetaWrites:   map[string]struct{}{},
		},
	}

	nRegs := 0
	for _, instr := range instrs {
		for _, r := range append(instr.dests(), instr.sources()...) {
			nRegs = max(nRegs, int(r)+1)
		}
	}

	cur := &depState{
		regs:    make([]absVal, nRegs),
		certain: true,
		colors:  map[string]struct{}{},
		dirty:   map[balanceKey]struct{}{},
	}
	pending := map[Label]*depState{}

	for _, instr := range instrs {
		if lbl, ok := instr.(LabelMarker); ok {
			cur = join(cur, pending[lbl.Label])
			delete(pending, lbl.Label)
			continue
		}
		if cur == nil {
			continue
		}

		switch instr := instr.(type) {
		case Jmp:
			pending[instr.Target] = join(pending[instr.Target], cur)
			cur = nil

		case JmpIfFalse:
			cur = a.branch(pending, cur, instr.Cond, false, instr.Target)

		case JmpIfTrue:
			cur = a.branch(pending, cur, instr.Cond, true, instr.Target)

		default:
			err := a.step(cur, instr)
			if err == nil {
				continue
			}
			// an execution error raised on known operands is raised by every run
			// that gets here; on an uncertain path no run may, so the path ends
			if _, isExecErr := err.(vm.ExecutionError); isExecErr && !cur.certain {
				cur = nil
				continue
			}
			return Dependencies{}, err
		}
	}

	return a.deps, nil
}

// branch handles a conditional jump taken when cond == jumpWhen, and returns the
// fallthrough state (nil when the fallthrough is unreachable).
func (a *depAnalysis) branch(pending map[Label]*depState, cur *depState, cond Reg, jumpWhen bool, target Label) *depState {
	c := cur.regs[cond]
	if !c.known {
		cur.certain = false
		pending[target] = join(pending[target], cur)
		return cur
	}
	if c.v.(bool) == jumpWhen {
		pending[target] = join(pending[target], cur)
		return nil
	}
	return cur
}

func (a *depAnalysis) step(st *depState, instr Instr) error {
	regs := st.regs

	str := func(r Reg) (string, bool) {
		v := regs[r]
		if !v.known {
			return "", false
		}
		return v.v.(string), true
	}
	optStr := func(r *Reg) (string, bool) {
		if r == nil {
			return "", true
		}
		return str(*r)
	}
	need := func(ok bool, operand string) error {
		if ok {
			return nil
		}
		return UnresolvedOperand{Instr: instr, Operand: operand}
	}

	switch instr := instr.(type) {
	case PullAccount:
		account, ok := str(instr.Account)
		if err := need(ok, "account"); err != nil {
			return err
		}
		scope, ok := optStr(instr.Scope)
		if err := need(ok, "scope"); err != nil {
			return err
		}
		color, ok := optStr(instr.Color)
		if err := need(ok, "color"); err != nil {
			return err
		}
		if err := need(st.asset.known, "current asset"); err != nil {
			return err
		}
		if instr.Cap == nil && instr.Overdraft == nil {
			return vm.InvalidUncappedSource{Account: account}
		}
		asset := st.asset.v.(string)
		dep := AccountDependency{Account: account, Scope: scope, Color: color, Asset: asset}
		// no overdraft operand means unbounded: the balance is never consulted
		if instr.Overdraft != nil {
			a.deps.AccountsReads[dep] = struct{}{}
		}
		a.deps.AccountsWrites[dep] = struct{}{}
		st.colors[color] = struct{}{}
		st.dirty[balanceKey{account, scope, asset}] = struct{}{}
		regs[instr.Dest] = unknown

	case SendToAccount:
		if instr.Account == nil {
			return nil
		}
		account, ok := str(*instr.Account)
		if err := need(ok, "account"); err != nil {
			return err
		}
		scope, ok := optStr(instr.Scope)
		if err := need(ok, "scope"); err != nil {
			return err
		}
		if err := need(st.asset.known, "current asset"); err != nil {
			return err
		}
		asset := st.asset.v.(string)
		for color := range st.colors {
			a.deps.AccountsWrites[AccountDependency{Account: account, Scope: scope, Color: color, Asset: asset}] = struct{}{}
		}

	case Save:
		account, ok := str(instr.Account)
		if err := need(ok, "account"); err != nil {
			return err
		}
		asset, ok := str(instr.Asset)
		if err := need(ok, "asset"); err != nil {
			return err
		}
		scope, ok := optStr(instr.Scope)
		if err := need(ok, "scope"); err != nil {
			return err
		}
		a.deps.AccountsReads[AccountDependency{Account: account, Scope: scope, Asset: asset}] = struct{}{}
		st.dirty[balanceKey{account, scope, asset}] = struct{}{}

	case SetCurrentAsset:
		st.asset = regs[instr.Asset]
		st.colors = map[string]struct{}{}

	case FetchBalance:
		account, ok := str(instr.Account)
		if err := need(ok, "account"); err != nil {
			return err
		}
		asset, ok := str(instr.Asset)
		if err := need(ok, "asset"); err != nil {
			return err
		}
		scope, ok := optStr(instr.Scope)
		if err := need(ok, "scope"); err != nil {
			return err
		}
		a.deps.AccountsReads[AccountDependency{Account: account, Scope: scope, Asset: asset}] = struct{}{}
		if _, dirty := st.dirty[balanceKey{account, scope, asset}]; dirty {
			regs[instr.Dest] = unknown
			return nil
		}
		bal, err := a.store.GetBalance(a.ctx, account, scope, asset, "")
		if err != nil {
			return vm.StoreError{Wrapped: err}
		}
		if bal == nil {
			bal = new(big.Int)
		}
		regs[instr.Dest] = knownVal(new(big.Int).Set(bal))

	case MetaVar:
		account, scope, key, err := a.metaOperands(instr, regs, instr.Account, instr.Key, instr.Scope)
		if err != nil {
			return err
		}
		raw, err := a.lookupMeta(account, scope, key)
		if err != nil {
			return err
		}
		switch instr.Typ.(type) {
		case MetaStr:
			regs[instr.Dest] = knownVal(raw)
		case MetaInt:
			n, ok := funds.ParseNumber(raw)
			if !ok {
				return vm.BadMetaValueError{Account: account, Key: key, Raw: raw}
			}
			regs[instr.Dest] = knownVal(n)
		case MetaPortion:
			r, err := funds.ParsePortion(raw)
			if err != nil {
				return vm.BadMetaValueError{Account: account, Key: key, Raw: raw}
			}
			regs[instr.Dest] = knownVal(r)
		default:
			return fmt.Errorf("unhandled meta type %T", instr.Typ)
		}

	case MetaMonetary:
		account, scope, key, err := a.metaOperands(instr, regs, instr.Account, instr.Key, instr.Scope)
		if err != nil {
			return err
		}
		raw, err := a.lookupMeta(account, scope, key)
		if err != nil {
			return err
		}
		asset, amount, merr := funds.ParseMonetary(raw)
		if merr != nil {
			return vm.BadMetaValueError{Account: account, Key: key, Raw: raw}
		}
		regs[instr.DestAsset] = knownVal(asset)
		regs[instr.DestAmount] = knownVal(amount)

	case SetTxMeta:
		key, ok := str(instr.Key)
		if err := need(ok, "key"); err != nil {
			return err
		}
		a.deps.TxMetaWrites[key] = struct{}{}

	case SetAccountMeta:
		account, scope, key, err := a.metaOperands(instr, regs, instr.Account, instr.Key, instr.Scope)
		if err != nil {
			return err
		}
		a.deps.MetaWrites[MetaDependency{Account: account, Scope: scope, Key: key}] = struct{}{}

	case LoadVar:
		if a.vars == nil {
			return fmt.Errorf("load_var %d: no vars given", instr.Index)
		}
		switch instr.Typ.(type) {
		case VarInt:
			if int(instr.Index) >= len(a.vars.IntsPool) {
				return fmt.Errorf("load_var %d: out of range", instr.Index)
			}
			regs[instr.Dest] = knownVal(new(big.Int).Set(&a.vars.IntsPool[instr.Index]))
		case VarStr:
			if int(instr.Index) >= len(a.vars.StringsPool) {
				return fmt.Errorf("load_var %d: out of range", instr.Index)
			}
			regs[instr.Dest] = knownVal(a.vars.StringsPool[instr.Index])
		default:
			return fmt.Errorf("unhandled var type %T", instr.Typ)
		}

	case LoadInt:
		regs[instr.Dest] = knownVal(new(big.Int).Set(&instr.Value))

	case LoadStr:
		regs[instr.Dest] = knownVal(instr.Value)

	case ConstBool:
		regs[instr.Dest] = knownVal(instr.Value)

	case BinaryOp:
		l, r := regs[instr.Left], regs[instr.Right]
		if !l.known || !r.known {
			regs[instr.Dest] = unknown
			return nil
		}
		v, err := evalBinary(instr.Op, l.v, r.v)
		if err != nil {
			return err
		}
		regs[instr.Dest] = knownVal(v)

	case UnaryOp:
		arg := regs[instr.Arg]
		if !arg.known {
			regs[instr.Dest] = unknown
			return nil
		}
		v, err := evalUnary(instr.Op, arg.v)
		if err != nil {
			return err
		}
		regs[instr.Dest] = knownVal(v)

	// the assertions fail the analysis only when every operand is known
	case AssertLeftover:
		if p := regs[instr.Portion]; p.known {
			leftover := p.v.(*big.Rat)
			if sign := leftover.Sign(); sign < 0 || (instr.Exact && sign != 0) {
				return vm.InvalidAllotmentSum{ActualSum: *new(big.Rat).Sub(big.NewRat(1, 1), leftover)}
			}
		}

	case AssertSameAsset:
		left, lok := str(instr.Left)
		right, rok := str(instr.Right)
		if lok && rok && left != right {
			return vm.AssetMismatchError{Expected: left, Got: right}
		}

	case AssertValidAccount:
		if account, ok := str(instr.Account); ok && !funds.ValidateAccount(account) {
			return vm.InvalidAccountName{Name: account}
		}

	case AssertValidColor:
		if color, ok := str(instr.Color); ok && !funds.ValidateColor(color) {
			return vm.InvalidColor{Color: color}
		}

	case AssertValidScope:
		if scope, ok := str(instr.Scope); ok && !funds.ValidateScope(scope) {
			return vm.InvalidScope{Scope: scope}
		}

	case AssertUnscoped:
		if scope, ok := str(instr.Scope); ok && scope != "" {
			account, _ := str(instr.Account)
			return vm.CannotCastScopedAccountToString{Account: account, Scope: scope}
		}

	case AssertNonNegativeBalance:
		if b := regs[instr.Balance]; b.known && b.v.(*big.Int).Sign() < 0 {
			account, _ := str(instr.Account)
			return vm.NegativeBalanceError{Account: account, Amount: *new(big.Int).Set(b.v.(*big.Int))}
		}

	case AssertNonNegativeAmount:
		if b := regs[instr.Amount]; b.known && b.v.(*big.Int).Sign() < 0 {
			return vm.NegativeAmountError{Amount: *new(big.Int).Set(b.v.(*big.Int))}
		}

	case AssertNonNegativePortion:
		if p := regs[instr.Portion]; p.known && p.v.(*big.Rat).Sign() < 0 {
			return vm.NegativePortionError{Portion: *new(big.Rat).Set(p.v.(*big.Rat))}
		}

	// got is always a pulled amount
	case CheckEnoughFunds:
	case MarkPush, MarkEnd:

	default:
		return fmt.Errorf("unhandled instruction %T", instr)
	}

	return nil
}

func (a *depAnalysis) metaOperands(instr Instr, regs []absVal, accountReg, keyReg Reg, scopeReg *Reg) (account, scope, key string, err error) {
	acc, key2 := regs[accountReg], regs[keyReg]
	if !acc.known {
		return "", "", "", UnresolvedOperand{Instr: instr, Operand: "account"}
	}
	if !key2.known {
		return "", "", "", UnresolvedOperand{Instr: instr, Operand: "key"}
	}
	if scopeReg != nil {
		s := regs[*scopeReg]
		if !s.known {
			return "", "", "", UnresolvedOperand{Instr: instr, Operand: "scope"}
		}
		scope = s.v.(string)
	}
	account, key = acc.v.(string), key2.v.(string)
	if _, isWrite := instr.(SetAccountMeta); !isWrite {
		a.deps.MetaReads[MetaDependency{Account: account, Scope: scope, Key: key}] = struct{}{}
	}
	return account, scope, key, nil
}

func (a *depAnalysis) lookupMeta(account, scope, key string) (string, error) {
	v, ok, err := a.store.GetMetadata(a.ctx, account, scope, key)
	if err != nil {
		return "", vm.StoreError{Wrapped: err}
	}
	if !ok {
		return "", vm.MetadataNotFoundError{Account: account, Key: key}
	}
	return v, nil
}

func evalBinary(op BinKind, l, r any) (any, error) {
	switch op.(type) {
	case OpAddInt:
		return new(big.Int).Add(l.(*big.Int), r.(*big.Int)), nil
	case OpSubInt:
		return new(big.Int).Sub(l.(*big.Int), r.(*big.Int)), nil
	case OpLtInt:
		return l.(*big.Int).Cmp(r.(*big.Int)) < 0, nil
	case OpEqInt:
		return l.(*big.Int).Cmp(r.(*big.Int)) == 0, nil
	case OpAddString:
		return l.(string) + r.(string), nil
	case OpStrEq:
		return l.(string) == r.(string), nil
	case OpLtPortion:
		return l.(*big.Rat).Cmp(r.(*big.Rat)) < 0, nil
	case OpEqPortion:
		return l.(*big.Rat).Cmp(r.(*big.Rat)) == 0, nil
	case OpAddPortion:
		return new(big.Rat).Add(l.(*big.Rat), r.(*big.Rat)), nil
	case OpSubPortion:
		return new(big.Rat).Sub(l.(*big.Rat), r.(*big.Rat)), nil
	case OpMulPortion:
		return new(big.Rat).Mul(l.(*big.Rat), r.(*big.Rat)), nil
	case OpMakePortion:
		num, den := l.(*big.Int), r.(*big.Int)
		if den.Sign() == 0 {
			return nil, vm.DivideByZeroError{Numerator: *new(big.Int).Set(num)}
		}
		return new(big.Rat).SetFrac(num, den), nil
	case OpMonetaryToString:
		return l.(string) + " " + r.(*big.Int).String(), nil
	default:
		return nil, fmt.Errorf("unhandled binary op %T", op)
	}
}

func evalUnary(op UnKind, arg any) (any, error) {
	switch op.(type) {
	case OpIntCopy, OpPortionCopy, OpStrCopy, OpBoolCopy:
		return arg, nil
	case OpNegInt:
		return new(big.Int).Neg(arg.(*big.Int)), nil
	case OpIntToString:
		return arg.(*big.Int).String(), nil
	case OpIsZero:
		return arg.(*big.Int).Sign() == 0, nil
	case OpNot:
		return !arg.(bool), nil
	case OpPortionToString:
		return arg.(*big.Rat).String(), nil
	case OpIntToPortion:
		return new(big.Rat).SetInt(arg.(*big.Int)), nil
	case OpPortionToInt:
		p := arg.(*big.Rat)
		return new(big.Int).Div(p.Num(), p.Denom()), nil
	default:
		return nil, fmt.Errorf("unhandled unary op %T", op)
	}
}
