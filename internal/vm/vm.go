package vm

import "math/big"

type monetary struct {
	asset  string
	amount big.Int
}

type Vm struct {
	program Program

	stringsRegs    [256]string // asset,string,account
	intsRegs       [256]big.Int
	portionsRegs   [256]big.Rat
	monetariesRegs [256]monetary
}

func NewVm(
	program Program,
) *Vm {
	return &Vm{
		program: program,
	}
}

type Store interface {
	GetBalance(
		account string,
		asset string,
	) *big.Int
}

func Exec[S Store](
	vm *Vm,
	vars any,
	store S, // a generic S should allow monomorphisation of the Store
) ExecutionError {
	instrs := vm.program.instructions
	instructionsLen := len(instrs)

	var currentAsset string
	pc := 0

	for pc < instructionsLen {
		instr := instrs[pc]
		pc++

		switch Opcode(instr.Opcode) {
		// --- Domain-specific ops
		case Op_PullAccount,
			Op_PullAccountCap,
			Op_PullAccountCapOverdraft,
			Op_PullAccountUnboundedOverdraft,
			Op_PullAccountOverdraft:
			panic("TODO impl pullAccount*")

		case Op_MkAllotment:
			panic("TODO mk allotment")

		case Op_SendToAccount,
			Op_SendToAccountAcc,
			Op_SendToAccountCap,
			Op_SendToAccountAccCap:
			panic("TODO send to account")

		case Op_CheckEnoughFunds:
			panic("TODO Op_CheckEnoughFunds")

		case Op_SetCurrentAsset:
			currentAsset = vm.stringsRegs[instr.A]

		case Op_CheckEqCurrentAsset:
			got := vm.stringsRegs[instr.A]
			if got != currentAsset {
				return AssetMismatchError{
					Got:      got,
					Expected: currentAsset,
				}
			}

			// --- Vars
		case Op_FetchVariable:
			// TODO we need to check if we'll use FetchVarNumber, FetchVarString, ..
			// or if we have a vars table that has this info
			panic("TODO fetch vars")

		// --- Jumps
		case Op_JmpIfZero:
			arg := &vm.intsRegs[instr.A]
			if arg.Sign() == 0 {
				pc = int(instr.GetBC())
			}

		// --- consts
		case Op_LoadInt:
			const_ := &vm.program.intsPool[instr.GetBC()]
			vm.intsRegs[instr.A].Set(const_)

		case Op_LoadStr:
			const_ := vm.program.stringsPool[instr.GetBC()]
			vm.stringsRegs[instr.A] = const_

			// ---  Binary ops
		case Op_MinInt:
			left := &vm.intsRegs[instr.B]
			right := &vm.intsRegs[instr.C]
			if left.Cmp(right) == -1 {
				vm.intsRegs[instr.A].Set(left)
			} else {
				vm.intsRegs[instr.A].Set(right)
			}

		case Op_AddInt:
			left := &vm.intsRegs[instr.B]
			right := &vm.intsRegs[instr.C]
			vm.intsRegs[instr.A].Add(left, right)

		case Op_SubInt:
			left := &vm.intsRegs[instr.B]
			right := &vm.intsRegs[instr.C]
			vm.intsRegs[instr.A].Sub(left, right)

		case Op_SubPortion:
			left := &vm.portionsRegs[instr.B]
			right := &vm.portionsRegs[instr.C]
			vm.portionsRegs[instr.A].Sub(left, right)

		case Op_MkPortion:
			num := &vm.intsRegs[instr.B]
			den := &vm.intsRegs[instr.C]
			vm.portionsRegs[instr.A].SetFrac(num, den)

		case Op_MkMonetary:
			asset := vm.stringsRegs[instr.B]
			amt := &vm.intsRegs[instr.C]

			dest := &vm.monetariesRegs[instr.A]
			dest.asset = asset
			dest.amount.Set(amt)

		// --- Unary ops
		case Op_IntCopy:
			arg := &vm.intsRegs[instr.B]
			vm.intsRegs[instr.A].Set(arg)

		case Op_PortionCopy:
			arg := &vm.portionsRegs[instr.B]
			vm.portionsRegs[instr.A].Set(arg)

		case Op_GetAsset:
			arg := &vm.monetariesRegs[instr.B]
			vm.stringsRegs[instr.A] = arg.asset

		case Op_GetAmount:
			arg := &vm.monetariesRegs[instr.B]
			vm.intsRegs[instr.A].Set(&arg.amount)

		default:
			panic("Invalid operation")
		}
	}

	return nil
}
