package compiler

// mapSources implementations: each returns a copy of the instruction with its
// SOURCE (read) registers passed through f; destinations are untouched. These
// are the rewrite primitive used by peephole passes (register substitution).
//
// Every vInstr type must have a mapSources that maps exactly the registers its
// sources() method reports.

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

func (i takeAccount) mapSources(f func(reg) reg) vInstr {
	i.account = f(i.account)
	i.cap = mapOptReg(f, i.cap)
	i.overdraft = mapOptReg(f, i.overdraft)
	i.color = mapOptReg(f, i.color)
	return i
}

func (i postAccount) mapSources(f func(reg) reg) vInstr {
	i.srcAccount = f(i.srcAccount)
	i.dstAccount = f(i.dstAccount)
	i.amount = f(i.amount)
	i.color = mapOptReg(f, i.color)
	return i
}

func (i postFromUnbounded) mapSources(f func(reg) reg) vInstr {
	i.srcAccount = f(i.srcAccount)
	i.dstAccount = f(i.dstAccount)
	i.cap = f(i.cap)
	i.color = mapOptReg(f, i.color)
	return i
}

func (i save) mapSources(f func(reg) reg) vInstr {
	i.account = f(i.account)
	i.asset = f(i.asset)
	i.amount = mapOptReg(f, i.amount)
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

func (i assertLeftover) mapSources(f func(reg) reg) vInstr {
	i.portion = f(i.portion)
	return i
}

func (i setCurrentAsset) mapSources(f func(reg) reg) vInstr {
	i.asset = f(i.asset)
	return i
}

func (i assertSameAsset) mapSources(f func(reg) reg) vInstr {
	i.left = f(i.left)
	i.right = f(i.right)
	return i
}

func (i assertValidAccount) mapSources(f func(reg) reg) vInstr {
	i.account = f(i.account)
	return i
}

func (i assertNonNegativeBalance) mapSources(f func(reg) reg) vInstr {
	i.balance = f(i.balance)
	i.account = f(i.account)
	return i
}

func (i setTxMeta) mapSources(f func(reg) reg) vInstr {
	i.key = f(i.key)
	i.value = f(i.value)
	return i
}

func (i setAccountMeta) mapSources(f func(reg) reg) vInstr {
	i.account = f(i.account)
	i.key = f(i.key)
	i.value = f(i.value)
	return i
}

func (i metaVar) mapSources(f func(reg) reg) vInstr {
	i.account = f(i.account)
	i.key = f(i.key)
	return i
}

func (i fetchBalance) mapSources(f func(reg) reg) vInstr {
	i.account = f(i.account)
	i.asset = f(i.asset)
	return i
}

func (i loadVar) mapSources(func(reg) reg) vInstr { return i }

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
