package compiler

// mapSources implementations: each returns a copy of the instruction with its
// SOURCE (read) registers passed through f; destinations are untouched. These
// are the rewrite primitive used by peephole passes (register substitution).

func mapOptReg(f func(reg) reg, r *reg) *reg {
	if r == nil {
		return nil
	}
	nr := f(*r)
	return &nr
}

func (i pullAccount) mapSources(f func(reg) reg) vInstr {
	i.account = f(i.account)
	i.cap = mapOptReg(f, i.cap)
	i.overdraft = mapOptReg(f, i.overdraft)
	i.color = mapOptReg(f, i.color)
	return i
}

func (i sendToAccount) mapSources(f func(reg) reg) vInstr {
	i.account = mapOptReg(f, i.account)
	i.cap = mapOptReg(f, i.cap)
	return i
}

func (i makeAllotment) mapSources(f func(reg) reg) vInstr {
	portions := make([]reg, len(i.portions))
	for j, r := range i.portions {
		portions[j] = f(r)
	}
	i.portions = portions
	i.amount = f(i.amount)
	return i
}

func (i checkEnoughFunds) mapSources(f func(reg) reg) vInstr {
	i.got = f(i.got)
	i.needed = f(i.needed)
	return i
}

func (i setCurrentAsset) mapSources(f func(reg) reg) vInstr {
	i.asset = f(i.asset)
	return i
}

func (i checkEqCurrentAsset) mapSources(f func(reg) reg) vInstr {
	i.got = f(i.got)
	return i
}

func (i fetchVariable) mapSources(func(reg) reg) vInstr { return i }

func (i jmpIfZero) mapSources(f func(reg) reg) vInstr {
	i.cond = f(i.cond)
	return i
}

func (i loadInt) mapSources(func(reg) reg) vInstr { return i }

func (i loadStr) mapSources(func(reg) reg) vInstr { return i }

func (i binaryOp) mapSources(f func(reg) reg) vInstr {
	i.left = f(i.left)
	i.right = f(i.right)
	return i
}

func (i unaryOp) mapSources(f func(reg) reg) vInstr {
	i.arg = f(i.arg)
	return i
}

func (i labelMarker) mapSources(func(reg) reg) vInstr { return i }
