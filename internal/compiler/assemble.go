package compiler

import (
	"fmt"
	"math"
	"math/big"

	"github.com/formancehq/numscript/internal/vm"
)

const maxReg = 0xFF

type regPool struct {
	indexByReg map[reg]byte
	next       int
}

func newRegPool() regPool {
	return regPool{
		indexByReg: map[reg]byte{},
	}
}

type constPool[T any] struct {
	indexByValue map[string]uint16
	items        []T
	toString     func(T) string
}

func newConstPool[T any](toString func(T) string) constPool[T] {
	return constPool[T]{
		indexByValue: map[string]uint16{},
		toString:     toString,
	}
}

func (p *constPool[T]) alloc(item T) (uint16, error) {
	strValue := p.toString(item)
	index, ok := p.indexByValue[strValue]
	if !ok {
		l := len(p.items)
		if l > math.MaxUint16 {
			return 0, fmt.Errorf("error: too many consts (overflowed the u16 len)")
		}
		index = uint16(l)
		p.indexByValue[strValue] = index
		p.items = append(p.items, item)
	}

	return index, nil
}

func (b *regPool) index(r reg) (byte, error) {
	if idx, ok := b.indexByReg[r]; ok {
		return idx, nil
	}
	if b.next >= maxReg {
		return 0, fmt.Errorf("register bank overflow: more than %d registers in one bank (register allocation not implemented yet)", maxReg)
	}
	idx := byte(b.next)
	b.next++
	b.indexByReg[r] = idx
	return idx, nil
}

type patch struct {
	label          label
	index          int
	getInstruction func(labelIndex uint16) vm.Instruction
}

// assembler lowers virtual instructions into a vm.Program.
type assembler struct {
	instructions []vm.Instruction

	patches []patch
	labels  map[label]uint16

	// one register bank per VM register bank
	ints       regPool
	strings    regPool
	portions   regPool
	monetaries regPool

	intsPool    constPool[big.Int]
	stringsPool constPool[string]
}

func Assemble(instrs []vInstr) (vm.Program, error) {
	a := &assembler{
		ints:       newRegPool(),
		strings:    newRegPool(),
		portions:   newRegPool(),
		monetaries: newRegPool(),

		labels: map[label]uint16{},

		intsPool: newConstPool(func(i big.Int) string {
			return i.String()
		}),
		stringsPool: newConstPool(func(s string) string {
			return s
		}),
	}
	for _, instr := range instrs {
		if err := instr.assemble(a); err != nil {
			return vm.Program{}, err
		}
	}

	// now we run the patches
	for _, patch := range a.patches {
		labelIndex, ok := a.labels[patch.label]
		if !ok {
			return vm.Program{}, fmt.Errorf("Missing label declaration of `%s`", string(patch.label))
		}

		a.instructions[patch.index] = patch.getInstruction(labelIndex)
	}

	return vm.Program{
		Instructions: a.instructions,
		StringsPool:  a.stringsPool.items,
		IntsPool:     a.intsPool.items,
	}, nil
}

func (as *assembler) intReg(r reg) (byte, error)      { return as.ints.index(r) }
func (as *assembler) strReg(r reg) (byte, error)      { return as.strings.index(r) }
func (as *assembler) portionReg(r reg) (byte, error)  { return as.portions.index(r) }
func (as *assembler) monetaryReg(r reg) (byte, error) { return as.monetaries.index(r) }

func (as *assembler) optionalReg(
	regPool func(*assembler, reg) (byte, error),
	reg *reg,
) (byte, error) {
	if reg == nil {
		return maxReg, nil
	} else {
		reg_, err := regPool(as, *reg)
		if err != nil {
			return 0, err
		}
		return reg_, nil
	}

}

func (as *assembler) emit(op vm.Opcode, a, b, c byte) {
	as.instructions = append(as.instructions, vm.Instruction{
		Opcode: byte(op),
		A:      a,
		B:      b,
		C:      c,
	})
}

func (as *assembler) emitBC(op vm.Opcode, a byte, bc uint16) {
	as.instructions = append(as.instructions, vm.NewBC(op, a, bc))
}

// regResolver maps a virtual register to a concrete bank index. Op sigs hold
// these as method expressions ((*assembler).intReg, ...) so that a sig is a
// static description of an op, independent of any assembler instance.
type regResolver = func(*assembler, reg) (byte, error)

type unaryOpSig struct {
	opcode vm.Opcode
	dest   regResolver
	arg    regResolver
}

func (opIntCopy) sig() unaryOpSig {
	return unaryOpSig{
		opcode: vm.Op_IntCopy,
		dest:   (*assembler).intReg,
		arg:    (*assembler).intReg,
	}
}
func (opPortionCopy) sig() unaryOpSig {
	return unaryOpSig{
		opcode: vm.Op_PortionCopy,
		dest:   (*assembler).portionReg,
		arg:    (*assembler).portionReg,
	}
}
func (opGetAsset) sig() unaryOpSig {
	return unaryOpSig{
		opcode: vm.Op_GetAsset,
		dest:   (*assembler).strReg,
		arg:    (*assembler).monetaryReg,
	}
}
func (opGetAmount) sig() unaryOpSig {
	return unaryOpSig{
		opcode: vm.Op_GetAmount,
		dest:   (*assembler).intReg,
		arg:    (*assembler).monetaryReg,
	}
}

func (i unaryOp) assemble(a *assembler) error {
	sig := i.op.sig()

	dest, err := sig.dest(a, i.dest)
	if err != nil {
		return err
	}
	arg, err := sig.arg(a, i.arg)
	if err != nil {
		return err
	}

	a.emit(sig.opcode, dest, arg, maxReg)
	return nil
}

type binaryOpSig struct {
	opcode vm.Opcode
	dest   regResolver
	left   regResolver
	right  regResolver
}

func (opMinInt) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_MinInt,
		dest:   (*assembler).intReg,
		left:   (*assembler).intReg,
		right:  (*assembler).intReg,
	}
}
func (opAddInt) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_AddInt,
		dest:   (*assembler).intReg,
		left:   (*assembler).intReg,
		right:  (*assembler).intReg,
	}
}
func (opSubInt) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_SubInt,
		dest:   (*assembler).intReg,
		left:   (*assembler).intReg,
		right:  (*assembler).intReg,
	}
}
func (opSubPortion) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_SubPortion,
		dest:   (*assembler).portionReg,
		left:   (*assembler).portionReg,
		right:  (*assembler).portionReg,
	}
}
func (opMakePortion) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_MkPortion,
		dest:   (*assembler).portionReg,
		left:   (*assembler).intReg,
		right:  (*assembler).intReg,
	}
}
func (opMakeMonetary) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_MkMonetary,
		dest:   (*assembler).monetaryReg,
		left:   (*assembler).strReg,
		right:  (*assembler).intReg,
	}
}

func (i binaryOp) assemble(a *assembler) error {
	sig := i.op.sig()

	dest, err := sig.dest(a, i.dest)
	if err != nil {
		return err
	}
	left, err := sig.left(a, i.left)
	if err != nil {
		return err
	}
	right, err := sig.right(a, i.right)
	if err != nil {
		return err
	}

	a.emit(sig.opcode, dest, left, right)
	return nil
}

func (i loadInt) assemble(a *assembler) error {
	dest, err := a.intReg(i.dest)
	if err != nil {
		return err
	}

	// Small unsigned constants are encoded directly in the instruction (no const
	// pool entry, no big.Int copy on load) — numscript constants are unsigned in
	// the common case (overdraft 0, small caps, allotment num/den).
	if i.value.IsUint64() {
		if v := i.value.Uint64(); v <= math.MaxUint16 {
			a.emitBC(vm.Op_LoadIntImm, dest, uint16(v))
			return nil
		}
	}

	poolIndex, err := a.intsPool.alloc(i.value)
	if err != nil {
		return err
	}

	a.emitBC(vm.Op_LoadInt, dest, poolIndex)
	return nil
}

func (i loadStr) assemble(a *assembler) error {
	dest, err := a.strReg(i.dest)
	if err != nil {
		return err
	}

	poolIndex, err := a.stringsPool.alloc(i.value)
	if err != nil {
		return err
	}

	a.emitBC(vm.Op_LoadStr, dest, poolIndex)
	return nil
}

func (i checkEnoughFunds) assemble(a *assembler) error {
	got, err := a.intReg(i.got)
	if err != nil {
		return err
	}

	needed, err := a.intReg(i.needed)
	if err != nil {
		return err
	}

	a.emit(vm.Op_CheckEnoughFunds, got, needed, maxReg)
	return nil
}

func (i setCurrentAsset) assemble(a *assembler) error {
	assetReg, err := a.strReg(i.asset)
	if err != nil {
		return err
	}

	a.emit(vm.Op_SetCurrentAsset, assetReg, maxReg, maxReg)
	return nil
}

func (i pullAccount) assemble(a *assembler) error {
	dest, err := a.intReg(i.dest)
	if err != nil {
		return err
	}

	account, err := a.strReg(i.account)
	if err != nil {
		return err
	}

	// compact single-word form: bounded-zero overdraft, no color, cap present
	if i.boundedZero && i.color == nil && i.cap != nil {
		cap, err := a.intReg(*i.cap)
		if err != nil {
			return err
		}
		a.emit(vm.Op_PullAccountCapZero, dest, account, cap)
		return nil
	}

	cap, err := a.optionalReg((*assembler).intReg, i.cap)
	if err != nil {
		return err
	}

	overdraft, err := a.optionalReg((*assembler).intReg, i.overdraft)
	if err != nil {
		return err
	}

	color, err := a.optionalReg((*assembler).strReg, i.color)
	if err != nil {
		return err
	}

	a.emit(vm.Op_PullAccount, dest, account, cap)

	a.instructions = append(a.instructions, vm.Instruction{
		Opcode: maxReg,    // <- UNUSED
		A:      overdraft, // overdraft (int)
		B:      color,     // color (str)
		C:      maxReg,    // <- UNUSED
	})

	return nil
}

func (i sendToAccount) assemble(a *assembler) error {
	account, err := a.optionalReg((*assembler).strReg, i.account)
	if err != nil {
		return err
	}

	cap, err := a.optionalReg((*assembler).intReg, i.cap)
	if err != nil {
		return err
	}

	a.emit(vm.Op_SendToAccount, account, cap, maxReg)
	return nil
}

func (i checkEqCurrentAsset) assemble(a *assembler) error {
	asset, err := a.strReg(i.got)
	if err != nil {
		return err
	}

	a.emit(vm.Op_CheckEqCurrentAsset, asset, maxReg, maxReg)

	return nil
}

func (i jmpIfZero) assemble(a *assembler) error {
	cond, err := a.intReg(i.cond)
	if err != nil {
		return err
	}

	a.patches = append(a.patches, patch{
		label: i.target,
		index: len(a.instructions),
		getInstruction: func(labelIndex uint16) vm.Instruction {
			return vm.NewBC(vm.Op_JmpIfZero, cond, labelIndex)
		},
	})

	// Emit dummy instruction
	a.emit(0, 0, 0, 0)

	return nil
}

func (i makeAllotment) assemble(a *assembler) error { panic("TODO assemble makeAllotment") }
func (i fetchVariable) assemble(a *assembler) error { panic("TODO assemble fetchVariable") }

func (i labelMarker) assemble(a *assembler) error {
	l := len(a.instructions)
	if l > math.MaxUint16 {
		return fmt.Errorf("too many labels: overflown max safe uint16")
	}

	a.labels[i.label] = uint16(l)

	return nil
}
