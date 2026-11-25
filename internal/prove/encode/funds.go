package encode

import (
	"fmt"
	"sort"

	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/prove/smt"
)

func (e *encoder) pull(pos int, in ir.PullAccount) error {
	account, err := e.str(pos, in.Account)
	if err != nil {
		return err
	}
	asset, err := e.currentAsset(pos)
	if err != nil {
		return err
	}
	cap, err := e.optTerm(pos, in.Cap)
	if err != nil {
		return err
	}
	overdraft, err := e.optTerm(pos, in.Overdraft)
	if err != nil {
		return err
	}
	if cap == nil && overdraft == nil {
		e.comment(pos)
		e.fail(pos, UncappedSource, smt.Bool(true))
		return nil
	}

	k := Key{account, asset}
	e.comment(pos)
	var out smt.Term
	switch overdraft {
	case nil:
		// unbounded: the balance is never read
		out = smt.Max0(cap)
	default:
		bal := e.balanceOf(e.cur, k)
		available := e.def(fmt.Sprintf("available@%d", pos), smt.Max0(smt.Add(bal, smt.Max0(overdraft))))
		if cap == nil {
			out = available
		} else {
			out = smt.Ite(smt.Lt(cap, available), smt.Max0(cap), available)
		}
	}
	e.touch(k)
	e.setInt(pos, in.Dest, out)
	out = e.cur.regs[in.Dest].t
	e.cur.slots[pos] = &slot{pos: pos, account: account, amount: out}
	e.cur.bal[k] = e.def(fmt.Sprintf("bal(%s)@%d", k, pos), smt.Sub(e.balanceOf(e.cur, k), out))
	return nil
}

func (e *encoder) send(pos int, in ir.SendToAccount) error {
	asset, err := e.currentAsset(pos)
	if err != nil {
		return err
	}
	var dest *string
	if in.Account != nil {
		d, err := e.str(pos, *in.Account)
		if err != nil {
			return err
		}
		dest = &d
		e.touch(Key{d, asset})
	}
	rem, err := e.optTerm(pos, in.Cap)
	if err != nil {
		return err
	}

	e.comment(pos)
	zero := smt.Int64(0)
	sentAdd := map[Key]smt.Term{}
	refundAdd := map[Key]smt.Term{}
	movedAdd := map[MoveKey]smt.Term{}
	var received = zero
	for _, sl := range e.cur.sortedSlots() {
		if n, ok := smt.AsInt(sl.amount); ok && n.Sign() == 0 {
			continue
		}
		var take smt.Term
		if rem == nil {
			take = sl.amount
			sl.amount = zero
		} else {
			take = e.def(fmt.Sprintf("take#%d@%d", sl.pos, pos), smt.Ite(smt.Lt(zero, rem), smt.Min(sl.amount, rem), zero))
			rem = e.def(fmt.Sprintf("rem#%d@%d", sl.pos, pos), smt.Sub(rem, take))
			sl.amount = e.def(fmt.Sprintf("slot#%d@%d", sl.pos, pos), smt.Sub(sl.amount, take))
		}
		src := Key{sl.account, asset}
		if dest == nil {
			refundAdd[src] = smt.Add(orZero(refundAdd[src]), take)
			continue
		}
		e.touch(src)
		sentAdd[src] = smt.Add(orZero(sentAdd[src]), take)
		m := MoveKey{sl.account, *dest, asset}
		e.moves[m] = struct{}{}
		movedAdd[m] = smt.Add(orZero(movedAdd[m]), take)
		received = smt.Add(received, take)
	}

	rset := map[Key]struct{}{}
	for k := range refundAdd {
		rset[k] = struct{}{}
	}
	for _, k := range sortedKeys(rset) {
		e.cur.bal[k] = e.def(fmt.Sprintf("bal(%s)@%d", k, pos), smt.Add(e.balanceOf(e.cur, k), refundAdd[k]))
	}
	set := map[Key]struct{}{}
	for k := range sentAdd {
		set[k] = struct{}{}
	}
	for _, k := range sortedKeys(set) {
		e.cur.sent[k] = e.def(fmt.Sprintf("sent(%s)@%d", k, pos), smt.Add(orZero(e.cur.sent[k]), sentAdd[k]))
	}
	mset := map[MoveKey]struct{}{}
	for m := range movedAdd {
		mset[m] = struct{}{}
	}
	for _, m := range sortedMoveKeys(mset) {
		e.cur.moved[m] = e.def(fmt.Sprintf("moved(%s)@%d", m, pos), smt.Add(orZero(e.cur.moved[m]), movedAdd[m]))
	}
	if dest != nil {
		dk := Key{*dest, asset}
		e.cur.recv[dk] = e.def(fmt.Sprintf("received(%s)@%d", dk, pos), smt.Add(orZero(e.cur.recv[dk]), received))
		e.cur.bal[dk] = e.def(fmt.Sprintf("bal(%s)@%d", dk, pos), smt.Add(e.balanceOf(e.cur, dk), received))
	}
	return nil
}

func (e *encoder) save(pos int, in ir.Save) error {
	account, err := e.str(pos, in.Account)
	if err != nil {
		return err
	}
	asset, err := e.str(pos, in.Asset)
	if err != nil {
		return err
	}
	amount, err := e.optTerm(pos, in.Amount)
	if err != nil {
		return err
	}
	k := Key{account, asset}
	e.comment(pos)
	bal := e.balanceOf(e.cur, k)
	e.touch(k)
	if amount == nil {
		e.cur.bal[k] = e.def(fmt.Sprintf("bal(%s)@%d", k, pos), smt.Min(bal, smt.Int64(0)))
	} else {
		e.cur.bal[k] = e.def(fmt.Sprintf("bal(%s)@%d", k, pos), smt.Max0(smt.Sub(bal, amount)))
	}
	return nil
}

func (e *encoder) fetchBalance(pos int, in ir.FetchBalance) error {
	account, err := e.str(pos, in.Account)
	if err != nil {
		return err
	}
	asset, err := e.str(pos, in.Asset)
	if err != nil {
		return err
	}
	k := Key{account, asset}
	e.touch(k)
	e.setInt(pos, in.Dest, e.balanceOf(e.cur, k))
	return nil
}

func (e *encoder) markEnd(pos int, in ir.MarkEnd) error {
	if len(e.cur.marks) == 0 {
		return fmt.Errorf("encode: instruction %d: mark end with no open mark", pos)
	}
	f := e.cur.marks[len(e.cur.marks)-1]
	e.cur.marks = e.cur.marks[:len(e.cur.marks)-1]
	if !in.Rewind {
		return nil
	}
	asset, err := e.currentAsset(pos)
	if err != nil {
		return err
	}
	e.comment(pos)
	for _, sl := range e.cur.sortedSlots() {
		if n, ok := smt.AsInt(sl.amount); ok && n.Sign() == 0 {
			continue
		}
		// the slot is repaid when the region was opened before its pull
		cands := make([]int, 0, len(f))
		for c := range f {
			if c < sl.pos {
				cands = append(cands, c)
			}
		}
		sort.Ints(cands)
		guards := make([]smt.Term, len(cands))
		for i, c := range cands {
			guards[i] = f[c]
		}
		repay := smt.Or(guards...)
		if smt.IsFalse(repay) {
			continue
		}
		k := Key{sl.account, asset}
		e.cur.bal[k] = e.def(fmt.Sprintf("bal(%s)@%d", k, pos), smt.Add(e.balanceOf(e.cur, k), smt.Ite(repay, sl.amount, smt.Int64(0))))
		sl.amount = e.def(fmt.Sprintf("slot#%d@%d", sl.pos, pos), smt.Ite(repay, smt.Int64(0), sl.amount))
	}
	return nil
}
