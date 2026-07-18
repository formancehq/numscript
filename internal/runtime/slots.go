package runtime

import "math/big"

// DynamicSlot marks a balance access whose (account, asset) is not a
// compile-time constant (e.g. the account comes from a variable): it has no
// reserved slot and goes through the string-keyed balance map.
const DynamicSlot = -1

// entryForSlot returns the balance entry for key, using a compile-assigned slot
// index as a fast path over the string-keyed map (map lookup is ~half of a warm
// send; an array index is ~4× cheaper).
//
// Soundness with dynamic (variable) accounts: the slot caches the SAME
// *balanceEntry the map holds — it is populated by going through entryFor the
// first time — so a slotted (constant-account) access and a dynamic (var) access
// that resolve to the same (account, asset) share one entry, and balances stay
// coherent. The cache persists across Reset (entries are pooled, never freed);
// freshen re-stamps a stale entry on first touch in a new generation.
func (s *RunState) entryForSlot(slot int, key PairKey) *balanceEntry {
	if slot < 0 {
		return s.entryFor(key)
	}
	for slot >= len(s.balanceSlots) {
		s.balanceSlots = append(s.balanceSlots, nil)
	}
	if e := s.balanceSlots[slot]; e != nil {
		return s.freshen(e)
	}
	e := s.entryFor(key)
	s.balanceSlots[slot] = e
	return e
}

// TakeSlot is Take with a slot-keyed balance access. See Take for the amount
// math (identical) and entryForSlot for the slot semantics.
func (s *RunState) TakeSlot(out *big.Int, src, scope string, cap, overdraft *big.Int, color string, slot int) error {
	key := PairKey{src, scope, s.currentAsset, color}
	if overdraft == nil {
		// unbounded: available = max(0, cap), independent of the balance
		out.Set(cap)
		if out.Sign() < 0 {
			out.SetInt64(0)
		}
		e := s.entryForSlot(slot, key)
		e.amount.Sub(&e.amount, out)
		return nil
	}

	e := s.entryForSlot(slot, key)
	if err := s.loadBase(key, e); err != nil {
		return err
	}
	currentBal := &e.amount

	out.Set(currentBal)
	if overdraft.Sign() > 0 {
		out.Add(out, overdraft)
	}
	if out.Sign() < 0 {
		out.SetInt64(0)
	}
	if cap.Cmp(out) < 0 {
		out.Set(cap)
	}
	if out.Sign() < 0 {
		out.SetInt64(0)
	}
	currentBal.Sub(currentBal, out)
	return nil
}

// PostDirectSlot is PostDirect with a slot-keyed destination credit. Mirrors
// addPosting's body (the credit is the balance touch we accelerate).
func (s *RunState) PostDirectSlot(src, srcScope, dst, dstScope, color string, amount *big.Int, slot int) error {
	if amount.Sign() <= 0 {
		return nil
	}
	amt := s.takeBig()
	amt.Set(amount)
	s.postings = append(s.postings, Posting{
		Source:           src,
		SourceScope:      srcScope,
		Destination:      dst,
		DestinationScope: dstScope,
		Asset:            s.currentAsset,
		Color:            color,
		Amount:           amt,
	})
	e := s.entryForSlot(slot, PairKey{dst, dstScope, s.currentAsset, color})
	e.amount.Add(&e.amount, amount)
	return nil
}
