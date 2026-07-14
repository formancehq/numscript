package vm

import "fmt"

type regBank int

const (
	bankInt regBank = iota
	bankStr
	bankPortion
	bankMonetary
	// bankCurrentAsset is a pseudo-bank with a single slot, tracking whether the
	// current asset has been set. It is never allocated as a register.
	bankCurrentAsset
)

var currentAssetRef = regRef{bankCurrentAsset, 0}

type regRef struct {
	bank  regBank
	index int
}

// decoded describes the register/pool/jump operands an instruction touches, so
// the verifier can size register banks and check every access without the VM
// having to guard anything at run time.
type decoded struct {
	reads      []regRef
	writes     []regRef
	constInt   int // index into program IntsPool, or -1
	constStr   int
	varInt     int // index into vars IntsPool, or -1
	varStr     int
	jumpTarget int // instruction index, or -1
}

type regCounts struct {
	ints, strings, portions, monetaries int
}

type programInfo struct {
	regs       regCounts
	varIntsLen int
	varStrsLen int
}

// Verify statically checks that a program is safe to execute: a nil result
// guarantees the execution loop cannot read out of bounds or crash on it.
func Verify(p Program) error {
	_, err := verify(p)
	return err
}

func instrWords(op byte) int {
	switch Opcode(op) {
	case Op_PullAccount, Op_MkAllotment:
		return 2
	default:
		return 1
	}
}

// verify statically checks a program cannot make the VM read out of bounds and
// returns the exact register-bank sizes it needs. A verified program can be run
// with no bounds checks in the execution loop.
func verify(p Program) (programInfo, error) {
	instrs := p.Instructions
	n := len(instrs)

	boundary := make([]bool, n+1)
	boundary[n] = true // jumping past the last instruction halts, which is fine

	type step struct {
		at int
		d  decoded
	}
	var steps []step

	for i := 0; i < n; {
		op := instrs[i].Opcode
		w := instrWords(op)
		if i+w > n {
			return programInfo{}, fmt.Errorf("truncated instruction at %d: opcode %d needs %d words", i, op, w)
		}
		var ext Instruction
		if w == 2 {
			ext = instrs[i+1]
		}
		d, err := decodeInstr(instrs[i], ext)
		if err != nil {
			return programInfo{}, fmt.Errorf("at instruction %d: %w", i, err)
		}
		boundary[i] = true
		steps = append(steps, step{i, d})
		i += w
	}

	info := programInfo{}
	grow := func(r regRef) {
		s := r.index + 1
		switch r.bank {
		case bankInt:
			if s > info.regs.ints {
				info.regs.ints = s
			}
		case bankStr:
			if s > info.regs.strings {
				info.regs.strings = s
			}
		case bankPortion:
			if s > info.regs.portions {
				info.regs.portions = s
			}
		case bankMonetary:
			if s > info.regs.monetaries {
				info.regs.monetaries = s
			}
		}
	}

	for _, st := range steps {
		d := st.d
		for _, r := range d.reads {
			grow(r)
		}
		for _, r := range d.writes {
			grow(r)
		}
		if d.constInt >= 0 {
			if d.constInt >= len(p.IntsPool) {
				return programInfo{}, fmt.Errorf("at instruction %d: int constant %d out of range (pool size %d)", st.at, d.constInt, len(p.IntsPool))
			}
		}
		if d.constStr >= 0 {
			if d.constStr >= len(p.StringsPool) {
				return programInfo{}, fmt.Errorf("at instruction %d: string constant %d out of range (pool size %d)", st.at, d.constStr, len(p.StringsPool))
			}
		}
		if d.varInt >= 0 && d.varInt+1 > info.varIntsLen {
			info.varIntsLen = d.varInt + 1
		}
		if d.varStr >= 0 && d.varStr+1 > info.varStrsLen {
			info.varStrsLen = d.varStr + 1
		}
		if d.jumpTarget >= 0 {
			t := d.jumpTarget
			if t <= st.at {
				return programInfo{}, fmt.Errorf("at instruction %d: backward jump to %d", st.at, t)
			}
			if t > n || !boundary[t] {
				return programInfo{}, fmt.Errorf("at instruction %d: jump to %d is not an instruction boundary", st.at, t)
			}
		}
	}

	// definite assignment: a register (or the current asset) may only be read if
	// it was written on every path reaching that instruction. Jumps are forward
	// (checked above), so every predecessor is earlier and one ordered pass over
	// the intersection of predecessors' written-sets suffices.
	at2idx := make(map[int]int, len(steps))
	for k, st := range steps {
		at2idx[st.at] = k
	}
	preds := make([][]int, len(steps))
	for k, st := range steps {
		if j, ok := at2idx[st.at+instrWords(instrs[st.at].Opcode)]; ok {
			preds[j] = append(preds[j], k)
		}
		if st.d.jumpTarget >= 0 {
			if j, ok := at2idx[st.d.jumpTarget]; ok {
				preds[j] = append(preds[j], k)
			}
		}
	}
	assignedOut := make([]map[regRef]bool, len(steps))
	for k, st := range steps {
		in := intersectAssigned(assignedOut, preds[k])
		for _, r := range st.d.reads {
			if !in[r] {
				return programInfo{}, fmt.Errorf("at instruction %d: %s read before being assigned on all paths", st.at, describeRef(r))
			}
		}
		for _, r := range st.d.writes {
			in[r] = true
		}
		assignedOut[k] = in
	}

	return info, nil
}

func intersectAssigned(out []map[regRef]bool, preds []int) map[regRef]bool {
	res := map[regRef]bool{}
	if len(preds) == 0 {
		return res
	}
	for r := range out[preds[0]] {
		res[r] = true
	}
	for _, p := range preds[1:] {
		for r := range res {
			if !out[p][r] {
				delete(res, r)
			}
		}
	}
	return res
}

func describeRef(r regRef) string {
	if r.bank == bankCurrentAsset {
		return "current asset"
	}
	return fmt.Sprintf("register (bank %d, index %d)", r.bank, r.index)
}

func decodeInstr(instr, ext Instruction) (decoded, error) {
	d := decoded{constInt: -1, constStr: -1, varInt: -1, varStr: -1, jumpTarget: -1}
	read := func(bank regBank, idx byte) { d.reads = append(d.reads, regRef{bank, int(idx)}) }
	write := func(bank regBank, idx byte) { d.writes = append(d.writes, regRef{bank, int(idx)}) }
	readOpt := func(bank regBank, idx byte) {
		if idx != nilReg {
			d.reads = append(d.reads, regRef{bank, int(idx)})
		}
	}

	switch Opcode(instr.Opcode) {
	case Op_PullAccount:
		read(bankStr, instr.B)
		readOpt(bankInt, instr.C)
		readOpt(bankInt, ext.A)
		readOpt(bankStr, ext.B)
		d.reads = append(d.reads, currentAssetRef)
		write(bankInt, instr.A)
	case Op_SendToAccount:
		readOpt(bankStr, instr.A)
		readOpt(bankInt, instr.B)
		readOpt(bankStr, instr.C)
		d.reads = append(d.reads, currentAssetRef)
	case Op_MkAllotment:
		for j := int(instr.A); j < int(instr.A)+int(instr.C); j++ {
			d.writes = append(d.writes, regRef{bankInt, j})
		}
		for j := int(instr.B); j < int(instr.B)+int(instr.C); j++ {
			d.reads = append(d.reads, regRef{bankPortion, j})
		}
		read(bankInt, ext.A)
	case Op_CheckEnoughFunds:
		read(bankInt, instr.A)
		read(bankInt, instr.B)
	case Op_Save:
		read(bankStr, instr.A)
		read(bankStr, instr.B)
		readOpt(bankInt, instr.C)
	case Op_AssertLeftover:
		read(bankPortion, instr.A)
	case Op_SetCurrentAsset:
		read(bankStr, instr.A)
		d.writes = append(d.writes, currentAssetRef)
	case Op_AssertSameAsset:
		read(bankStr, instr.A)
		read(bankStr, instr.B)
	case Op_AssertValidAccount:
		read(bankStr, instr.A)
	case Op_AssertNonNegativeBalance:
		read(bankMonetary, instr.A)
		read(bankStr, instr.B)
	case Op_SetTxMeta:
		read(bankStr, instr.A)
		read(bankStr, instr.B)
	case Op_SetAccountMeta:
		read(bankStr, instr.A)
		read(bankStr, instr.B)
		read(bankStr, instr.C)
	case Op_MetaStr:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		write(bankStr, instr.A)
	case Op_MetaInt:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		write(bankInt, instr.A)
	case Op_MetaPortion:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		write(bankPortion, instr.A)
	case Op_MetaMonetary:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		write(bankMonetary, instr.A)
	case Op_LoadVarInt:
		d.varInt = int(instr.GetBC())
		write(bankInt, instr.A)
	case Op_LoadVarStr:
		d.varStr = int(instr.GetBC())
		write(bankStr, instr.A)
	case Op_LoadInt:
		d.constInt = int(instr.GetBC())
		write(bankInt, instr.A)
	case Op_LoadStr:
		d.constStr = int(instr.GetBC())
		write(bankStr, instr.A)
	case Op_JmpIfZero:
		read(bankInt, instr.A)
		d.jumpTarget = int(instr.GetBC())
	case Op_MinInt, Op_AddInt, Op_SubInt:
		read(bankInt, instr.B)
		read(bankInt, instr.C)
		write(bankInt, instr.A)
	case Op_AddString:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		write(bankStr, instr.A)
	case Op_SubPortion:
		read(bankPortion, instr.B)
		read(bankPortion, instr.C)
		write(bankPortion, instr.A)
	case Op_MkPortion:
		read(bankInt, instr.B)
		read(bankInt, instr.C)
		write(bankPortion, instr.A)
	case Op_MkMonetary:
		read(bankStr, instr.B)
		read(bankInt, instr.C)
		write(bankMonetary, instr.A)
	case Op_Balance:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		write(bankMonetary, instr.A)
	case Op_IntCopy:
		read(bankInt, instr.B)
		write(bankInt, instr.A)
	case Op_PortionCopy:
		read(bankPortion, instr.B)
		write(bankPortion, instr.A)
	case Op_GetAsset:
		read(bankMonetary, instr.B)
		write(bankStr, instr.A)
	case Op_GetAmount:
		read(bankMonetary, instr.B)
		write(bankInt, instr.A)
	case Op_NegInt:
		read(bankInt, instr.B)
		write(bankInt, instr.A)
	case Op_IntToString:
		read(bankInt, instr.B)
		write(bankStr, instr.A)
	case Op_PortionToString:
		read(bankPortion, instr.B)
		write(bankStr, instr.A)
	case Op_MonetaryToString:
		read(bankMonetary, instr.B)
		write(bankStr, instr.A)
	default:
		return decoded{}, fmt.Errorf("unknown opcode %d", instr.Opcode)
	}

	return d, nil
}
