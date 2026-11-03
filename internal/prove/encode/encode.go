package encode

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/funds"
	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/prove/smt"
	"github.com/formancehq/numscript/internal/vm"
)

type encoder struct {
	instrs []ir.Instr
	sc     *smt.Script

	keys     map[Key]struct{}
	moves    map[MoveKey]struct{}
	accounts map[string]struct{}
	assets   map[string]struct{}

	pending map[ir.Label][]*state
	fails   []FailSite
	cur     *state
}

// Encode models instrs. extraKeys and extraMoves are pairs a property names
// that the program may never touch; they get a free before input and zero
// movements.
func Encode(instrs []ir.Instr, extraKeys []Key, extraMoves []MoveKey) (*Problem, error) {
	if err := ir.Typecheck(instrs); err != nil {
		return nil, err
	}
	prog, err := ir.Assemble(instrs)
	if err != nil {
		if errors.Is(err, ir.ErrRegisterBankOverflow) || errors.Is(err, ir.ErrProgramTooLarge) {
			return nil, UnsupportedError{Index: -1, Reason: err.Error()}
		}
		return nil, err
	}
	if err := vm.Verify(prog); err != nil {
		return nil, err
	}

	e := &encoder{
		instrs:   instrs,
		sc:       smt.NewBareScript(),
		keys:     map[Key]struct{}{},
		moves:    map[MoveKey]struct{}{},
		accounts: map[string]struct{}{},
		assets:   map[string]struct{}{},
		pending:  map[ir.Label][]*state{},
		cur:      newState(),
	}
	if err := e.prescan(); err != nil {
		return nil, err
	}
	for pos, in := range instrs {
		if lm, ok := in.(ir.LabelMarker); ok {
			edges := append([]*state{e.cur}, e.pending[lm.Label]...)
			delete(e.pending, lm.Label)
			merged, err := e.merge(pos, edges)
			if err != nil {
				return nil, err
			}
			e.cur = merged
			continue
		}
		if e.cur.dead {
			continue
		}
		if err := e.step(pos, in); err != nil {
			return nil, err
		}
	}
	if len(e.pending) != 0 {
		return nil, fmt.Errorf("encode: a jump targets a label that is not in the program")
	}
	return e.finish(extraKeys, extraMoves), nil
}

func (e *encoder) text(pos int) string {
	if pos < 0 {
		return ""
	}
	return fmt.Sprint(e.instrs[pos])
}

func (e *encoder) unsupported(pos int, reason string) error {
	return UnsupportedError{Index: pos, Text: e.text(pos), Reason: reason}
}

// prescan rejects what v1 does not model, naming the first offending
// instruction whether or not it is reachable, and collects every (account,
// asset) pair a funds operand can name. Constants are followed in text order
// with jumps ignored, so an account on a path the model folds away (a run that
// always fails first) is still known to properties.
func (e *encoder) prescan() error {
	consts := map[ir.Reg]string{}
	asset, assetKnown := "", false
	touchPair := func(account ir.Reg, assetName string, known bool) {
		if a, ok := consts[account]; ok && known {
			e.touch(Key{a, assetName})
		}
	}
	for pos, in := range e.instrs {
		reason := ""
		switch in := in.(type) {
		case ir.LoadStr:
			consts[in.Dest] = in.Value
		case ir.UnaryOp:
			if _, ok := in.Op.(ir.OpStrCopy); ok {
				if v, ok := consts[in.Arg]; ok {
					consts[in.Dest] = v
				} else {
					delete(consts, in.Dest)
				}
			}
		case ir.BinaryOp:
			if _, ok := in.Op.(ir.OpAddString); ok {
				l, lok := consts[in.Left]
				r, rok := consts[in.Right]
				if lok && rok {
					consts[in.Dest] = l + r
				} else {
					delete(consts, in.Dest)
				}
			}
		case ir.SetCurrentAsset:
			asset, assetKnown = consts[in.Asset]
		case ir.LoadVar:
			reason = "script variables are not supported"
		case ir.MetaVar, ir.MetaMonetary:
			reason = "metadata reads are not supported"
		case ir.AssertValidColor:
			reason = "asset colors are not supported"
		case ir.AssertValidScope, ir.AssertUnscoped:
			reason = "scoped accounts are not supported"
		case ir.PullAccount:
			if in.Color != nil {
				reason = "asset colors are not supported"
			} else if in.Scope != nil {
				reason = "scoped accounts are not supported"
			}
			touchPair(in.Account, asset, assetKnown)
		case ir.SendToAccount:
			if in.Scope != nil {
				reason = "scoped accounts are not supported"
			}
			if in.Account != nil {
				touchPair(*in.Account, asset, assetKnown)
			}
		case ir.Save:
			if in.Scope != nil {
				reason = "scoped accounts are not supported"
			}
			a, ok := consts[in.Asset]
			touchPair(in.Account, a, ok)
		case ir.FetchBalance:
			if in.Scope != nil {
				reason = "scoped accounts are not supported"
			}
			a, ok := consts[in.Asset]
			touchPair(in.Account, a, ok)
		}
		if reason != "" {
			return e.unsupported(pos, reason)
		}
	}
	return nil
}

func (e *encoder) def(name string, t smt.Term) smt.Term {
	return e.sc.Define(name, t)
}

func (e *encoder) comment(pos int) {
	e.sc.Comment("%d: %s", pos, e.text(pos))
}

func (e *encoder) touch(k Key) {
	e.keys[k] = struct{}{}
	e.accounts[k.Account] = struct{}{}
	e.assets[k.Asset] = struct{}{}
}

// balanceOf is the running balance of k in s: the before input until the
// program writes it.
func (e *encoder) balanceOf(s *state, k Key) smt.Term {
	if t, ok := s.bal[k]; ok {
		return t
	}
	e.touch(k)
	return smt.Symbol(BeforeName(k), smt.SortInt)
}

func (e *encoder) reg(pos int, r ir.Reg) (regVal, error) {
	v, ok := e.cur.regs[r]
	if !ok {
		return regVal{}, fmt.Errorf("encode: instruction %d: register $r%d is read before being written", pos, r)
	}
	return v, nil
}

func (e *encoder) term(pos int, r ir.Reg) (smt.Term, error) {
	v, err := e.reg(pos, r)
	if err != nil {
		return nil, err
	}
	if v.t == nil {
		return nil, fmt.Errorf("encode: instruction %d: register $r%d is not an int or bool", pos, r)
	}
	return v.t, nil
}

func (e *encoder) optTerm(pos int, r *ir.Reg) (smt.Term, error) {
	if r == nil {
		return nil, nil
	}
	return e.term(pos, *r)
}

func (e *encoder) str(pos int, r ir.Reg) (string, error) {
	v, err := e.reg(pos, r)
	if err != nil {
		return "", err
	}
	if !v.s.known {
		return "", e.unsupported(pos, fmt.Sprintf("the value of string register $r%d is not a constant", r))
	}
	return v.s.s, nil
}

func (e *encoder) portion(pos int, r ir.Reg) (*linForm, error) {
	v, err := e.reg(pos, r)
	if err != nil {
		return nil, err
	}
	if v.p == nil {
		return nil, fmt.Errorf("encode: instruction %d: register $r%d is not a portion", pos, r)
	}
	return v.p, nil
}

func (e *encoder) currentAsset(pos int) (string, error) {
	if !e.cur.asset.known {
		return "", e.unsupported(pos, "the current asset is not a constant")
	}
	return e.cur.asset.s, nil
}

func (e *encoder) setInt(pos int, dest ir.Reg, t smt.Term) {
	e.cur.regs[dest] = regVal{t: e.def(fmt.Sprintf("$r%d@%d", dest, pos), t)}
}

func (e *encoder) setStr(dest ir.Reg, s string, known bool) {
	e.cur.regs[dest] = regVal{s: strVal{known: known, s: s}}
}

// fail records a site that aborts the run when cond holds. A path on which
// the abort is certain is dead from here on: nothing it would compute is
// applied, and the outputs are masked by fails.
func (e *encoder) fail(pos int, kind FailKind, cond smt.Term) {
	flagTerm := smt.And(e.cur.alive, cond)
	if smt.IsFalse(flagTerm) {
		return
	}
	flag := e.sc.DefineAlways(fmt.Sprintf("fail@%d", pos), flagTerm)
	e.fails = append(e.fails, FailSite{Index: pos, Text: e.text(pos), Kind: kind, Flag: flag})
	alive := smt.And(e.cur.alive, smt.Not(cond))
	if smt.IsFalse(alive) {
		e.cur = deadState()
		return
	}
	e.cur.alive = e.def(fmt.Sprintf("alive@%d", pos), alive)
}

func (e *encoder) branch(pos int, cond smt.Term, target ir.Label) {
	switch {
	case smt.IsTrue(cond):
		e.pending[target] = append(e.pending[target], e.cur)
		e.cur = deadState()
	case smt.IsFalse(cond):
	default:
		taken := e.cur.clone()
		taken.reach = e.def(fmt.Sprintf("reach@%d.t", pos), smt.And(e.cur.reach, cond))
		taken.alive = e.def(fmt.Sprintf("alive@%d.t", pos), smt.And(e.cur.alive, cond))
		e.pending[target] = append(e.pending[target], taken)
		e.cur.reach = e.def(fmt.Sprintf("reach@%d.f", pos), smt.And(e.cur.reach, smt.Not(cond)))
		e.cur.alive = e.def(fmt.Sprintf("alive@%d.f", pos), smt.And(e.cur.alive, smt.Not(cond)))
	}
}

func (e *encoder) step(pos int, in ir.Instr) error {
	switch in := in.(type) {
	case ir.LoadInt:
		e.cur.regs[in.Dest] = regVal{t: smt.Int(&in.Value)}
	case ir.LoadStr:
		e.setStr(in.Dest, in.Value, true)
	case ir.ConstBool:
		e.cur.regs[in.Dest] = regVal{t: smt.Bool(in.Value)}
	case ir.BinaryOp:
		return e.binary(pos, in)
	case ir.UnaryOp:
		return e.unary(pos, in)
	case ir.JmpIfTrue:
		c, err := e.term(pos, in.Cond)
		if err != nil {
			return err
		}
		e.branch(pos, c, in.Target)
	case ir.JmpIfFalse:
		c, err := e.term(pos, in.Cond)
		if err != nil {
			return err
		}
		e.branch(pos, smt.Not(c), in.Target)
	case ir.Jmp:
		e.branch(pos, smt.Bool(true), in.Target)
	case ir.SetCurrentAsset:
		v, err := e.reg(pos, in.Asset)
		if err != nil {
			return err
		}
		e.cur.asset = v.s
	case ir.PullAccount:
		return e.pull(pos, in)
	case ir.SendToAccount:
		return e.send(pos, in)
	case ir.Save:
		return e.save(pos, in)
	case ir.FetchBalance:
		return e.fetchBalance(pos, in)
	case ir.MarkPush:
		e.cur.marks = append(e.cur.marks, frame{pos: smt.Bool(true)})
	case ir.MarkEnd:
		return e.markEnd(pos, in)
	case ir.CheckEnoughFunds:
		got, err := e.term(pos, in.Got)
		if err != nil {
			return err
		}
		needed, err := e.term(pos, in.Needed)
		if err != nil {
			return err
		}
		e.fail(pos, MissingFunds, smt.Ne(got, needed))
	case ir.AssertNonNegativeAmount:
		t, err := e.term(pos, in.Amount)
		if err != nil {
			return err
		}
		e.fail(pos, NegativeAmount, smt.Lt(t, smt.Int64(0)))
	case ir.AssertNonNegativeBalance:
		t, err := e.term(pos, in.Balance)
		if err != nil {
			return err
		}
		e.fail(pos, NegativeBalance, smt.Lt(t, smt.Int64(0)))
	case ir.AssertNonNegativePortion:
		p, err := e.portion(pos, in.Portion)
		if err != nil {
			return err
		}
		e.fail(pos, NegativePortion, smt.Lt(p.sign(), smt.Int64(0)))
	case ir.AssertLeftover:
		p, err := e.portion(pos, in.Portion)
		if err != nil {
			return err
		}
		if in.Exact {
			e.fail(pos, AllotmentSum, smt.Ne(p.sign(), smt.Int64(0)))
		} else {
			e.fail(pos, AllotmentSum, smt.Lt(p.sign(), smt.Int64(0)))
		}
	case ir.AssertSameAsset:
		l, err := e.str(pos, in.Left)
		if err != nil {
			return err
		}
		r, err := e.str(pos, in.Right)
		if err != nil {
			return err
		}
		if l != r {
			e.fail(pos, AssetMismatch, smt.Bool(true))
		}
	case ir.AssertValidAccount:
		a, err := e.str(pos, in.Account)
		if err != nil {
			return err
		}
		if !funds.ValidateAccount(a) {
			e.fail(pos, InvalidAccount, smt.Bool(true))
		}
	case ir.SetTxMeta, ir.SetAccountMeta:
	default:
		return e.unsupported(pos, "not modeled")
	}
	return nil
}

func (e *encoder) binary(pos int, in ir.BinaryOp) error {
	switch in.Op.(type) {
	case ir.OpAddInt, ir.OpSubInt, ir.OpLtInt, ir.OpEqInt:
		l, err := e.term(pos, in.Left)
		if err != nil {
			return err
		}
		r, err := e.term(pos, in.Right)
		if err != nil {
			return err
		}
		var t smt.Term
		switch in.Op.(type) {
		case ir.OpAddInt:
			t = smt.Add(l, r)
		case ir.OpSubInt:
			t = smt.Sub(l, r)
		case ir.OpLtInt:
			t = smt.Lt(l, r)
		case ir.OpEqInt:
			t = smt.Eq(l, r)
		}
		e.setInt(pos, in.Dest, t)

	case ir.OpAddPortion, ir.OpSubPortion, ir.OpMulPortion, ir.OpLtPortion, ir.OpEqPortion:
		l, err := e.portion(pos, in.Left)
		if err != nil {
			return err
		}
		r, err := e.portion(pos, in.Right)
		if err != nil {
			return err
		}
		switch in.Op.(type) {
		case ir.OpAddPortion:
			e.cur.regs[in.Dest] = regVal{p: l.add(r)}
		case ir.OpSubPortion:
			e.cur.regs[in.Dest] = regVal{p: l.sub(r)}
		case ir.OpMulPortion:
			p, ok := l.mul(r)
			if !ok {
				return e.unsupported(pos, "nonlinear portion arithmetic")
			}
			e.cur.regs[in.Dest] = regVal{p: p}
		case ir.OpLtPortion:
			e.setInt(pos, in.Dest, l.lt(r))
		case ir.OpEqPortion:
			e.setInt(pos, in.Dest, l.eq(r))
		}

	case ir.OpMakePortion:
		num, err := e.term(pos, in.Left)
		if err != nil {
			return err
		}
		den, err := e.term(pos, in.Right)
		if err != nil {
			return err
		}
		d, ok := smt.AsInt(den)
		if !ok {
			return e.unsupported(pos, "nonlinear portion arithmetic: the denominator is not a constant")
		}
		if d.Sign() == 0 {
			e.fail(pos, DivideByZero, smt.Bool(true))
			return nil
		}
		e.cur.regs[in.Dest] = regVal{p: linOfInt(num).scale(new(big.Rat).SetFrac(big.NewInt(1), d))}

	case ir.OpAddString:
		l, lerr := e.reg(pos, in.Left)
		r, rerr := e.reg(pos, in.Right)
		if lerr != nil {
			return lerr
		}
		if rerr != nil {
			return rerr
		}
		e.setStr(in.Dest, l.s.s+r.s.s, l.s.known && r.s.known)

	case ir.OpStrEq:
		l, err := e.str(pos, in.Left)
		if err != nil {
			return err
		}
		r, err := e.str(pos, in.Right)
		if err != nil {
			return err
		}
		e.cur.regs[in.Dest] = regVal{t: smt.Bool(l == r)}

	case ir.OpMonetaryToString:
		asset, err := e.reg(pos, in.Left)
		if err != nil {
			return err
		}
		amount, err := e.term(pos, in.Right)
		if err != nil {
			return err
		}
		n, ok := smt.AsInt(amount)
		e.setStr(in.Dest, asset.s.s+" "+n.String(), asset.s.known && ok)

	default:
		return e.unsupported(pos, "not modeled")
	}
	return nil
}

func (e *encoder) unary(pos int, in ir.UnaryOp) error {
	switch in.Op.(type) {
	case ir.OpIntCopy, ir.OpBoolCopy, ir.OpNegInt, ir.OpIsZero, ir.OpNot:
		t, err := e.term(pos, in.Arg)
		if err != nil {
			return err
		}
		switch in.Op.(type) {
		case ir.OpNegInt:
			t = smt.Neg(t)
		case ir.OpIsZero:
			t = smt.Eq(t, smt.Int64(0))
		case ir.OpNot:
			t = smt.Not(t)
		}
		e.setInt(pos, in.Dest, t)

	case ir.OpStrCopy:
		v, err := e.reg(pos, in.Arg)
		if err != nil {
			return err
		}
		e.cur.regs[in.Dest] = regVal{s: v.s}

	case ir.OpPortionCopy:
		p, err := e.portion(pos, in.Arg)
		if err != nil {
			return err
		}
		e.cur.regs[in.Dest] = regVal{p: p}

	case ir.OpIntToPortion:
		t, err := e.term(pos, in.Arg)
		if err != nil {
			return err
		}
		e.cur.regs[in.Dest] = regVal{p: linOfInt(t)}

	case ir.OpPortionToInt:
		p, err := e.portion(pos, in.Arg)
		if err != nil {
			return err
		}
		e.setInt(pos, in.Dest, p.toInt())

	case ir.OpIntToString:
		t, err := e.term(pos, in.Arg)
		if err != nil {
			return err
		}
		n, ok := smt.AsInt(t)
		if ok {
			e.setStr(in.Dest, n.String(), true)
		} else {
			e.setStr(in.Dest, "", false)
		}

	case ir.OpPortionToString:
		p, err := e.portion(pos, in.Arg)
		if err != nil {
			return err
		}
		if p.isConst() {
			e.setStr(in.Dest, p.konst.String(), true)
		} else {
			e.setStr(in.Dest, "", false)
		}

	default:
		return e.unsupported(pos, "not modeled")
	}
	return nil
}

func (e *encoder) finish(extraKeys []Key, extraMoves []MoveKey) *Problem {
	fin := e.cur
	for _, k := range extraKeys {
		e.keys[k] = struct{}{}
	}
	for _, m := range extraMoves {
		e.moves[m] = struct{}{}
	}

	e.sc.Blank()
	e.sc.Comment("outputs: the ledger applies nothing on failure")
	flags := make([]smt.Term, len(e.fails))
	for i, s := range e.fails {
		flags[i] = s.Flag
	}
	p := &Problem{
		body:      e.sc,
		Keys:      sortedKeys(e.keys),
		MoveKeys:  sortedMoveKeys(e.moves),
		Accounts:  sortedStrings(e.accounts),
		Assets:    sortedStrings(e.assets),
		Before:    map[Key]smt.Term{},
		After:     map[Key]smt.Term{},
		Sent:      map[Key]smt.Term{},
		Received:  map[Key]smt.Term{},
		Moved:     map[MoveKey]smt.Term{},
		FailSites: e.fails,
	}
	p.Fails = e.sc.DefineAlways("fails", smt.Or(flags...))
	p.Success = e.sc.DefineAlways("success", smt.Not(p.Fails))
	for _, k := range p.Keys {
		before := smt.Symbol(BeforeName(k), smt.SortInt)
		p.Before[k] = before
		sent := e.sc.DefineAlways(SentName(k), smt.Ite(p.Fails, smt.Int64(0), orZero(fin.sent[k])))
		recv := e.sc.DefineAlways(ReceivedName(k), smt.Ite(p.Fails, smt.Int64(0), orZero(fin.recv[k])))
		p.Sent[k] = sent
		p.Received[k] = recv
		p.After[k] = e.sc.DefineAlways(AfterName(k), smt.Add(before, smt.Sub(recv, sent)))
	}
	for _, m := range p.MoveKeys {
		p.Moved[m] = e.sc.DefineAlways(MovedName(m), smt.Ite(p.Fails, smt.Int64(0), orZero(fin.moved[m])))
	}
	return p
}
