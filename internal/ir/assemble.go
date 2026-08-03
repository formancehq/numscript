package ir

import (
	"fmt"
	"math"
	"math/big"

	"github.com/formancehq/numscript/internal/vm"
)

const maxReg = 0xFF

type regPool struct {
	indexByReg map[Reg]byte
	next       int
}

func newRegPool() regPool {
	return regPool{
		indexByReg: map[Reg]byte{},
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

func (b *regPool) Index(r Reg) (byte, error) {
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
	Label Label
	index int
	// delta is the forward jump offset, relative to the instruction following index
	getInstruction func(delta uint16) vm.Instruction
}

// assembler lowers IR instructions into a vm.Program.
type assembler struct {
	instructions []vm.Instruction

	patches []patch
	labels  map[Label]uint16

	// one register bank per VM register bank
	ints     regPool
	strings  regPool
	Portions regPool
	bools    regPool

	intsPool    constPool[big.Int]
	stringsPool constPool[string]
}

func Assemble(instrs []Instr) (vm.Program, error) {
	a := &assembler{
		ints:     newRegPool(),
		strings:  newRegPool(),
		Portions: newRegPool(),
		bools:    newRegPool(),

		labels: map[Label]uint16{},

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
		labelIndex, ok := a.labels[patch.Label]
		if !ok {
			return vm.Program{}, fmt.Errorf("missing label declaration of `%s`", string(patch.Label))
		}

		next := patch.index + 1
		if int(labelIndex) < next {
			return vm.Program{}, fmt.Errorf("backward jump to label `%s`: jumps must go forward", string(patch.Label))
		}

		a.instructions[patch.index] = patch.getInstruction(uint16(int(labelIndex) - next))
	}

	return vm.Program{
		Instructions: a.instructions,
		StringsPool:  a.stringsPool.items,
		IntsPool:     a.intsPool.items,

		MaxRegString:  byte(a.strings.next),
		MaxRegPortion: byte(a.Portions.next),
		MaxRegInt:     byte(a.ints.next),
		MaxRegBool:    byte(a.bools.next),
	}, nil
}

func (as *assembler) intReg(r Reg) (byte, error)     { return as.ints.Index(r) }
func (as *assembler) strReg(r Reg) (byte, error)     { return as.strings.Index(r) }
func (as *assembler) portionReg(r Reg) (byte, error) { return as.Portions.Index(r) }
func (as *assembler) boolReg(r Reg) (byte, error)    { return as.bools.Index(r) }

func (as *assembler) optionalReg(
	regPool func(*assembler, Reg) (byte, error),
	Reg *Reg,
) (byte, error) {
	if Reg == nil {
		return maxReg, nil
	} else {
		reg_, err := regPool(as, *Reg)
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

// regResolver maps a virtual register to a concrete bank index. op sigs hold
// these as method expressions ((*assembler).intReg, ...) so that a sig is a
// static description of an op, independent of any assembler instance.
type regResolver = func(*assembler, Reg) (byte, error)

type unaryOpSig struct {
	opcode vm.Opcode
	dest   regResolver
	arg    regResolver
}

// copySig is the shape every bank copy shares: dest and src in the same bank.
func copySig(opcode vm.Opcode, bank regResolver) unaryOpSig {
	return unaryOpSig{opcode: opcode, dest: bank, arg: bank}
}

func (OpIntCopy) sig() unaryOpSig { return copySig(vm.Op_IntCopy, (*assembler).intReg) }
func (OpPortionCopy) sig() unaryOpSig {
	return copySig(vm.Op_PortionCopy, (*assembler).portionReg)
}
func (OpStrCopy) sig() unaryOpSig  { return copySig(vm.Op_StrCopy, (*assembler).strReg) }
func (OpBoolCopy) sig() unaryOpSig { return copySig(vm.Op_BoolCopy, (*assembler).boolReg) }
func (OpNegInt) sig() unaryOpSig {
	return unaryOpSig{
		opcode: vm.Op_NegInt,
		dest:   (*assembler).intReg,
		arg:    (*assembler).intReg,
	}
}
func (OpIntToString) sig() unaryOpSig {
	return unaryOpSig{
		opcode: vm.Op_IntToString,
		dest:   (*assembler).strReg,
		arg:    (*assembler).intReg,
	}
}
func (OpIsZero) sig() unaryOpSig {
	return unaryOpSig{
		opcode: vm.Op_IsZero,
		dest:   (*assembler).boolReg,
		arg:    (*assembler).intReg,
	}
}
func (OpNot) sig() unaryOpSig {
	return unaryOpSig{
		opcode: vm.Op_Not,
		dest:   (*assembler).boolReg,
		arg:    (*assembler).boolReg,
	}
}
func (OpPortionToString) sig() unaryOpSig {
	return unaryOpSig{
		opcode: vm.Op_PortionToString,
		dest:   (*assembler).strReg,
		arg:    (*assembler).portionReg,
	}
}
func (OpIntToPortion) sig() unaryOpSig {
	return unaryOpSig{
		opcode: vm.Op_IntToPortion,
		dest:   (*assembler).portionReg,
		arg:    (*assembler).intReg,
	}
}
func (OpPortionToInt) sig() unaryOpSig {
	return unaryOpSig{
		opcode: vm.Op_PortionToInt,
		dest:   (*assembler).intReg,
		arg:    (*assembler).portionReg,
	}
}
func (i UnaryOp) assemble(a *assembler) error {
	sig := i.Op.sig()

	dest, err := sig.dest(a, i.Dest)
	if err != nil {
		return err
	}
	arg, err := sig.arg(a, i.Arg)
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

// comparisonSig is the shape every binary comparison shares: two operands of one
// bank, a bool dest.
func comparisonSig(opcode vm.Opcode, operand regResolver) binaryOpSig {
	return binaryOpSig{
		opcode: opcode,
		dest:   (*assembler).boolReg,
		left:   operand,
		right:  operand,
	}
}

func (OpLtInt) sig() binaryOpSig { return comparisonSig(vm.Op_LtInt, (*assembler).intReg) }
func (OpEqInt) sig() binaryOpSig { return comparisonSig(vm.Op_EqInt, (*assembler).intReg) }
func (OpLtPortion) sig() binaryOpSig {
	return comparisonSig(vm.Op_LtPortion, (*assembler).portionReg)
}
func (OpEqPortion) sig() binaryOpSig {
	return comparisonSig(vm.Op_EqPortion, (*assembler).portionReg)
}
func (OpAddInt) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_AddInt,
		dest:   (*assembler).intReg,
		left:   (*assembler).intReg,
		right:  (*assembler).intReg,
	}
}
func (OpSubInt) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_SubInt,
		dest:   (*assembler).intReg,
		left:   (*assembler).intReg,
		right:  (*assembler).intReg,
	}
}
func (OpAddString) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_AddString,
		dest:   (*assembler).strReg,
		left:   (*assembler).strReg,
		right:  (*assembler).strReg,
	}
}
func (OpStrEq) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_StrEq,
		dest:   (*assembler).boolReg,
		left:   (*assembler).strReg,
		right:  (*assembler).strReg,
	}
}

// portionArithSig is the shape of portion addition and subtraction: three
// portion operands.
func portionArithSig(opcode vm.Opcode) binaryOpSig {
	return binaryOpSig{
		opcode: opcode,
		dest:   (*assembler).portionReg,
		left:   (*assembler).portionReg,
		right:  (*assembler).portionReg,
	}
}

func (OpAddPortion) sig() binaryOpSig { return portionArithSig(vm.Op_AddPortion) }
func (OpSubPortion) sig() binaryOpSig { return portionArithSig(vm.Op_SubPortion) }
func (OpMulPortion) sig() binaryOpSig { return portionArithSig(vm.Op_MulPortion) }
func (OpMakePortion) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_MkPortion,
		dest:   (*assembler).portionReg,
		left:   (*assembler).intReg,
		right:  (*assembler).intReg,
	}
}
func (OpMonetaryToString) sig() binaryOpSig {
	return binaryOpSig{
		opcode: vm.Op_MonetaryToString,
		dest:   (*assembler).strReg,
		left:   (*assembler).strReg,
		right:  (*assembler).intReg,
	}
}

func (i BinaryOp) assemble(a *assembler) error {
	sig := i.Op.sig()

	dest, err := sig.dest(a, i.Dest)
	if err != nil {
		return err
	}
	left, err := sig.left(a, i.Left)
	if err != nil {
		return err
	}
	right, err := sig.right(a, i.Right)
	if err != nil {
		return err
	}

	a.emit(sig.opcode, dest, left, right)
	return nil
}

func (i LoadInt) assemble(a *assembler) error {
	dest, err := a.intReg(i.Dest)
	if err != nil {
		return err
	}

	poolIndex, err := a.intsPool.alloc(i.Value)
	if err != nil {
		return err
	}

	a.emitBC(vm.Op_LoadInt, dest, poolIndex)
	return nil
}

func (i LoadStr) assemble(a *assembler) error {
	dest, err := a.strReg(i.Dest)
	if err != nil {
		return err
	}

	poolIndex, err := a.stringsPool.alloc(i.Value)
	if err != nil {
		return err
	}

	a.emitBC(vm.Op_LoadStr, dest, poolIndex)
	return nil
}

func (i ConstBool) assemble(a *assembler) error {
	dest, err := a.boolReg(i.Dest)
	if err != nil {
		return err
	}

	opcode := vm.Op_ConstFalse
	if i.Value {
		opcode = vm.Op_ConstTrue
	}

	a.emit(opcode, dest, maxReg, maxReg)
	return nil
}

func (i CheckEnoughFunds) assemble(a *assembler) error {
	got, err := a.intReg(i.Got)
	if err != nil {
		return err
	}

	needed, err := a.intReg(i.Needed)
	if err != nil {
		return err
	}

	a.emit(vm.Op_CheckEnoughFunds, got, needed, maxReg)
	return nil
}

func (i Save) assemble(a *assembler) error {
	account, err := a.strReg(i.Account)
	if err != nil {
		return err
	}
	asset, err := a.strReg(i.Asset)
	if err != nil {
		return err
	}
	amount, err := a.optionalReg((*assembler).intReg, i.Amount)
	if err != nil {
		return err
	}
	a.emit(vm.Op_Save, account, asset, amount)
	return nil
}

func (i AssertLeftover) assemble(a *assembler) error {
	portion, err := a.portionReg(i.Portion)
	if err != nil {
		return err
	}
	var exact byte
	if i.Exact {
		exact = 1
	}
	a.emit(vm.Op_AssertLeftover, portion, exact, maxReg)
	return nil
}

func (i SetCurrentAsset) assemble(a *assembler) error {
	assetReg, err := a.strReg(i.Asset)
	if err != nil {
		return err
	}

	a.emit(vm.Op_SetCurrentAsset, assetReg, maxReg, maxReg)
	return nil
}

func (i PullAccount) assemble(a *assembler) error {
	dest, err := a.intReg(i.Dest)
	if err != nil {
		return err
	}

	account, err := a.strReg(i.Account)
	if err != nil {
		return err
	}

	cap, err := a.optionalReg((*assembler).intReg, i.Cap)
	if err != nil {
		return err
	}

	overdraft, err := a.optionalReg((*assembler).intReg, i.Overdraft)
	if err != nil {
		return err
	}

	color, err := a.optionalReg((*assembler).strReg, i.Color)
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

func (i SendToAccount) assemble(a *assembler) error {
	account, err := a.optionalReg((*assembler).strReg, i.Account)
	if err != nil {
		return err
	}

	cap, err := a.optionalReg((*assembler).intReg, i.Cap)
	if err != nil {
		return err
	}

	a.emit(vm.Op_SendToAccount, account, cap, maxReg)
	return nil
}

func (i AssertSameAsset) assemble(a *assembler) error {
	left, err := a.strReg(i.Left)
	if err != nil {
		return err
	}
	right, err := a.strReg(i.Right)
	if err != nil {
		return err
	}

	a.emit(vm.Op_AssertSameAsset, left, right, maxReg)

	return nil
}

func (i AssertValidAccount) assemble(a *assembler) error {
	account, err := a.strReg(i.Account)
	if err != nil {
		return err
	}

	a.emit(vm.Op_AssertValidAccount, account, maxReg, maxReg)

	return nil
}

func (i AssertValidColor) assemble(a *assembler) error {
	color, err := a.strReg(i.Color)
	if err != nil {
		return err
	}

	a.emit(vm.Op_AssertValidColor, color, maxReg, maxReg)

	return nil
}

func (i AssertNonNegativeBalance) assemble(a *assembler) error {
	balance, err := a.intReg(i.Balance)
	if err != nil {
		return err
	}
	account, err := a.strReg(i.Account)
	if err != nil {
		return err
	}

	a.emit(vm.Op_AssertNonNegativeBalance, balance, account, maxReg)

	return nil
}

func (i SetTxMeta) assemble(a *assembler) error {
	key, err := a.strReg(i.Key)
	if err != nil {
		return err
	}
	value, err := a.strReg(i.Value)
	if err != nil {
		return err
	}

	a.emit(vm.Op_SetTxMeta, key, value, maxReg)

	return nil
}

func (a *assembler) emitMeta(opcode vm.Opcode, dest byte, account, key Reg) error {
	acc, err := a.strReg(account)
	if err != nil {
		return err
	}
	k, err := a.strReg(key)
	if err != nil {
		return err
	}
	a.emit(opcode, dest, acc, k)
	return nil
}

func (MetaStr) assembleMeta(a *assembler, dest, account, key Reg) error {
	d, err := a.strReg(dest)
	if err != nil {
		return err
	}
	return a.emitMeta(vm.Op_MetaStr, d, account, key)
}

func (MetaInt) assembleMeta(a *assembler, dest, account, key Reg) error {
	d, err := a.intReg(dest)
	if err != nil {
		return err
	}
	return a.emitMeta(vm.Op_MetaInt, d, account, key)
}

func (MetaPortion) assembleMeta(a *assembler, dest, account, key Reg) error {
	d, err := a.portionReg(dest)
	if err != nil {
		return err
	}
	return a.emitMeta(vm.Op_MetaPortion, d, account, key)
}

func (i MetaVar) assemble(a *assembler) error {
	return i.Typ.assembleMeta(a, i.Dest, i.Account, i.Key)
}

// MetaMonetary needs four operands, so it spills the second destination into an
// ext word, like PullAccount.
func (i MetaMonetary) assemble(a *assembler) error {
	destAsset, err := a.strReg(i.DestAsset)
	if err != nil {
		return err
	}
	destAmount, err := a.intReg(i.DestAmount)
	if err != nil {
		return err
	}
	if err := a.emitMeta(vm.Op_MetaMonetary, destAsset, i.Account, i.Key); err != nil {
		return err
	}
	a.instructions = append(a.instructions, vm.Instruction{
		Opcode: maxReg,
		A:      destAmount,
		B:      maxReg,
		C:      maxReg,
	})
	return nil
}

func (i SetAccountMeta) assemble(a *assembler) error {
	account, err := a.strReg(i.Account)
	if err != nil {
		return err
	}
	key, err := a.strReg(i.Key)
	if err != nil {
		return err
	}
	value, err := a.strReg(i.Value)
	if err != nil {
		return err
	}

	a.emit(vm.Op_SetAccountMeta, account, key, value)

	return nil
}

func (i FetchBalance) assemble(a *assembler) error {
	dest, err := a.intReg(i.Dest)
	if err != nil {
		return err
	}
	account, err := a.strReg(i.Account)
	if err != nil {
		return err
	}
	asset, err := a.strReg(i.Asset)
	if err != nil {
		return err
	}

	a.emit(vm.Op_Balance, dest, account, asset)

	return nil
}

// assembleCondJmp emits either conditional jump: they differ only in the opcode.
func (a *assembler) assembleCondJmp(opcode vm.Opcode, cond Reg, target Label) error {
	condReg, err := a.boolReg(cond)
	if err != nil {
		return err
	}

	a.patches = append(a.patches, patch{
		Label: target,
		index: len(a.instructions),
		getInstruction: func(delta uint16) vm.Instruction {
			return vm.NewBC(opcode, condReg, delta)
		},
	})

	// Emit dummy instruction
	a.emit(0, 0, 0, 0)

	return nil
}

func (i JmpIfFalse) assemble(a *assembler) error {
	return a.assembleCondJmp(vm.Op_JmpIfFalse, i.Cond, i.Target)
}

func (i JmpIfTrue) assemble(a *assembler) error {
	return a.assembleCondJmp(vm.Op_JmpIfTrue, i.Cond, i.Target)
}

func (i Jmp) assemble(a *assembler) error {
	a.patches = append(a.patches, patch{
		Label: i.Target,
		index: len(a.instructions),
		getInstruction: func(delta uint16) vm.Instruction {
			return vm.NewBC(vm.Op_Jmp, 0, delta)
		},
	})

	// Emit dummy instruction
	a.emit(0, 0, 0, 0)

	return nil
}

func (VarInt) assembleLoad(a *assembler, dest Reg, index uint16) error {
	d, err := a.intReg(dest)
	if err != nil {
		return err
	}
	a.emitBC(vm.Op_LoadVarInt, d, index)
	return nil
}

func (VarStr) assembleLoad(a *assembler, dest Reg, index uint16) error {
	d, err := a.strReg(dest)
	if err != nil {
		return err
	}
	a.emitBC(vm.Op_LoadVarStr, d, index)
	return nil
}

func (i LoadVar) assemble(a *assembler) error {
	return i.Typ.assembleLoad(a, i.Dest, i.Index)
}

func (i LabelMarker) assemble(a *assembler) error {
	l := len(a.instructions)
	if l > math.MaxUint16 {
		return fmt.Errorf("too many labels: overflown max safe uint16")
	}

	a.labels[i.Label] = uint16(l)

	return nil
}

func (i Snapshot) assemble(a *assembler) error {
	dest, err := a.intReg(i.Dest)
	if err != nil {
		return err
	}
	a.emit(vm.Op_Snapshot, dest, maxReg, maxReg)
	return nil
}

func (i Restore) assemble(a *assembler) error {
	mark, err := a.intReg(i.Mark)
	if err != nil {
		return err
	}
	a.emit(vm.Op_Restore, mark, maxReg, maxReg)
	return nil
}
