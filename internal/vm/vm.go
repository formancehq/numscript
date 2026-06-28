package vm

import (
	"math/big"

	"github.com/formancehq/numscript/internal/runtime"
)

type monetary struct {
	asset  string
	amount big.Int
}

const nilReg byte = 0xFF
const worldAccount = "world"

type Vm struct {
	program  Program
	runstate *runtime.RunState

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
		color string,
	) *big.Int
}

func Exec[S Store](
	vm *Vm,
	vars any,
	store S, // a generic S should allow monomorphisation of the Store
) ([]runtime.Posting, ExecutionError) {
	if vm.runstate == nil {
		vm.runstate = runtime.New(store)
	} else {
		vm.runstate.Reset(store)
	}
	runstate := vm.runstate

	instrs := vm.program.Instructions
	instructionsLen := len(instrs)

	var currentAsset string
	pc := 0

	for pc < instructionsLen {
		instr := instrs[pc]
		pc++

		switch Opcode(instr.Opcode) {
		// --- Domain-specific ops
		case Op_PullAccount:
			instrExt := instrs[pc]
			pc++

			account := vm.stringsRegs[instr.B]

			var cap *big.Int
			if instr.C != nilReg {
				cap = &vm.intsRegs[instr.C]
			}

			var overdraft *big.Int
			if account != worldAccount && instrExt.A != nilReg {
				overdraft = &vm.intsRegs[instrExt.A]
			}

			if overdraft == nil && cap == nil {
				return nil, InvalidUncappedSource{
					Account: account,
				}
			}

			var color string
			if instrExt.B != nilReg {
				color = vm.stringsRegs[instrExt.B]
			}

			runstate.Pull(
				&vm.intsRegs[instr.A],
				account,
				cap,
				overdraft,
				color,
			)

		case Op_SendToAccount:
			var dest *string
			if instr.A != nilReg {
				s := vm.stringsRegs[instr.A]
				dest = &s
			}

			var cap *big.Int
			if instr.B != nilReg {
				cap = &vm.intsRegs[instr.B]
			}

			var color *string
			if instr.C != nilReg {
				color = &vm.stringsRegs[instr.C]
			}

			if cap == nil {
				runstate.SendUncapped(dest, color)
			} else {
				runstate.Send(dest, cap, color)
			}

		case Op_MkAllotment:
			instrExt := instrs[pc]
			pc++

			destArrStartReg := vm.intsRegs[instr.A : instr.A+instr.C]
			inpArrStartReg := vm.portionsRegs[instr.B : instr.B+instr.C]

			amt := &vm.intsRegs[instrExt.A]

			runtime.MakeAllotment(
				destArrStartReg,
				amt,
				inpArrStartReg,
			)

		case Op_CheckEnoughFunds:
			got := &vm.intsRegs[instr.A]
			needed := &vm.intsRegs[instr.B]
			if got.Cmp(needed) == -1 {
				return nil, MissingFundsError{
					Asset:  currentAsset,
					Got:    got,
					Needed: needed,
				}
			}

		case Op_SetCurrentAsset:
			currentAsset = vm.stringsRegs[instr.A]
			runstate.SetCurrentAsset(currentAsset)

		case Op_CheckEqCurrentAsset:
			got := vm.stringsRegs[instr.A]
			if got != currentAsset {
				return nil, AssetMismatchError{
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
			const_ := &vm.program.IntsPool[instr.GetBC()]
			vm.intsRegs[instr.A].Set(const_)

		case Op_LoadStr:
			const_ := vm.program.StringsPool[instr.GetBC()]
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

	return runstate.GetPostings(), nil
}
