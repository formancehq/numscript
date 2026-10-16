package ir

// pruneDeadBranches removes control flow that can be decided statically:
//
//   - a conditional jump on a constant bool becomes an unconditional jmp, or
//     nothing;
//   - the instructions between an unconditional jmp and the next label some
//     reachable jump targets can never run, and go;
//   - a jump whose target is the next instruction is a no-op, and goes;
//   - a label no jump targets any more goes.
//
// Jumps only go forward, so one pass in program order knows at every label
// whether a reachable jump targets it: every jump to a label sits before it.
func pruneDeadBranches(instrs []Instr) []Instr {
	defs := regDefs(instrs)
	changed := false

	// jumps on a constant condition
	folded := make([]Instr, 0, len(instrs))
	for _, in := range instrs {
		switch j := in.(type) {
		case JmpIfFalse:
			if v, ok := constBoolOf(instrs, defs, j.Cond); ok {
				changed = true
				if !v {
					folded = append(folded, Jmp{Target: j.Target})
				}
				continue
			}
		case JmpIfTrue:
			if v, ok := constBoolOf(instrs, defs, j.Cond); ok {
				changed = true
				if v {
					folded = append(folded, Jmp{Target: j.Target})
				}
				continue
			}
		}
		folded = append(folded, in)
	}

	// unreachable instructions. Labels are all kept here: the ones left with no
	// jump are dropped at the end.
	live := map[Label]bool{}
	reachable := true
	kept := make([]Instr, 0, len(folded))
	for _, in := range folded {
		if l, ok := in.(LabelMarker); ok {
			if live[l.Label] {
				reachable = true
			}
			kept = append(kept, in)
			continue
		}
		if !reachable {
			changed = true
			continue
		}
		kept = append(kept, in)
		switch j := in.(type) {
		case Jmp:
			live[j.Target] = true
			reachable = false
		case JmpIfFalse:
			live[j.Target] = true
		case JmpIfTrue:
			live[j.Target] = true
		}
	}

	// jumps to the next instruction
	out := make([]Instr, 0, len(kept))
	for i, in := range kept {
		if target, ok := jumpTarget(in); ok && jumpsToNext(kept, i, target) {
			changed = true
			continue
		}
		out = append(out, in)
	}

	// labels no jump targets
	referenced := map[Label]bool{}
	for _, in := range out {
		if target, ok := jumpTarget(in); ok {
			referenced[target] = true
		}
	}
	n := 0
	for _, in := range out {
		if l, ok := in.(LabelMarker); ok && !referenced[l.Label] {
			changed = true
			continue
		}
		out[n] = in
		n++
	}
	out = out[:n]

	if !changed {
		return nil
	}
	return out
}

func jumpTarget(in Instr) (Label, bool) {
	switch j := in.(type) {
	case Jmp:
		return j.Target, true
	case JmpIfFalse:
		return j.Target, true
	case JmpIfTrue:
		return j.Target, true
	default:
		return "", false
	}
}

// jumpsToNext reports whether the jump at instrs[i] targets a label that sits
// before the next instruction, i.e. with only labels in between.
func jumpsToNext(instrs []Instr, i int, target Label) bool {
	for _, in := range instrs[i+1:] {
		l, ok := in.(LabelMarker)
		if !ok {
			return false
		}
		if l.Label == target {
			return true
		}
	}
	return false
}
