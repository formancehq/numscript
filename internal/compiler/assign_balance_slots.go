package compiler

// assignBalanceSlots gives each compile-time-constant (account, asset) balance
// access a dense integer slot, so the VM can index a small array instead of
// hashing a 4-string PairKey into the balance map (the map lookup is ~half a warm
// send). It runs after the peepholes, on the final instruction stream.
//
// Only the two ops with a slot form are annotated: the compact bounded-zero
// take_account (source debit) and post_account (destination credit) — the pair a
// single->single bypass lowers to. An access is slottable only when both the
// account (a loadStr constant) and the current asset (from the enclosing
// set_current_asset, also a constant) are known; a variable account or asset
// leaves it dynamic (map), and — because a slot caches the SAME entry the map
// holds — a mix of slotted and dynamic accesses to one account stays coherent
// (see runtime.entryForSlot). Slots must fit a byte; past 254 distinct pairs the
// rest stay dynamic.
func assignBalanceSlots(instrs []vInstr) []vInstr {
	ids := map[[2]string]int{}
	nextID := 0
	slotFor := func(account, asset string) (int, bool) {
		if asset == "" {
			return 0, false // asset not statically known
		}
		k := [2]string{account, asset}
		if id, ok := ids[k]; ok {
			return id, true
		}
		if nextID > 254 {
			return 0, false // out of byte-range slots; leave dynamic
		}
		id := nextID
		nextID++
		ids[k] = id
		return id, true
	}

	out := make([]vInstr, len(instrs))
	copy(out, instrs)

	currentAsset := "" // "" = unknown / not a constant
	for i, in := range out {
		switch v := in.(type) {
		case setCurrentAsset:
			if a, ok := constStrOf(instrs, v.asset); ok {
				currentAsset = a
			} else {
				currentAsset = ""
			}
		case takeAccount:
			// only the compact bounded-zero form has a slot op.
			if !v.boundedZero || v.color != nil || v.cap == nil {
				continue
			}
			if acc, ok := constStrOf(instrs, v.account); ok {
				if id, ok := slotFor(acc, currentAsset); ok {
					v.slot, v.hasSlot = id, true
					out[i] = v
				}
			}
		case postAccount:
			if v.color != nil {
				continue
			}
			if acc, ok := constStrOf(instrs, v.dstAccount); ok {
				if id, ok := slotFor(acc, currentAsset); ok {
					v.slot, v.hasSlot = id, true
					out[i] = v
				}
			}
		}
	}
	return out
}
