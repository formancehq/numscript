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

	GetMetadata(account, key string) (string, bool)
}

func lookupMeta(store Store, account, key string) (string, ExecutionError) {
	v, ok := store.GetMetadata(account, key)
	if !ok {
		return "", MetadataNotFoundError{Account: account, Key: key}
	}
	return v, nil
}

func Exec[S Store](
	vm *Vm,
	vars *Vars,
	store S, // a generic S should allow monomorphisation of the Store
) (runtime.ExecutionResult, ExecutionError) {
	if vm.runstate == nil {
		vm.runstate = runtime.New(store)
	} else {
		vm.runstate.Reset(store)
	}
	runstate := vm.runstate

	var txMeta map[string]string
	var accountsMeta runtime.AccountsMetadata

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

			var color string
			if instrExt.B != nilReg {
				color = vm.stringsRegs[instrExt.B]
			}

			out := &vm.intsRegs[instr.A]
			switch {
			case cap != nil:
				runstate.Pull(out, account, cap, overdraft, color)
			case overdraft != nil:
				runstate.PullUncapped(out, account, overdraft, color)
			default:
				return runtime.ExecutionResult{}, InvalidUncappedSource{Account: account}
			}

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
				return runtime.ExecutionResult{}, MissingFundsError{
					Asset:  currentAsset,
					Got:    got,
					Needed: needed,
				}
			}

		case Op_Save:
			account := vm.stringsRegs[instr.A]
			asset := vm.stringsRegs[instr.B]
			var amount *big.Int
			if instr.C != nilReg {
				amount = &vm.intsRegs[instr.C]
			}
			runstate.Save(account, asset, "", amount)

		case Op_AssertLeftover:
			leftover := &vm.portionsRegs[instr.A]
			sign := leftover.Sign()
			if sign < 0 || (instr.B == 1 && sign != 0) {
				sum := new(big.Rat).Sub(big.NewRat(1, 1), leftover)
				return runtime.ExecutionResult{}, InvalidAllotmentSum{ActualSum: *sum}
			}

		case Op_SetCurrentAsset:
			currentAsset = vm.stringsRegs[instr.A]
			runstate.SetCurrentAsset(currentAsset)

		case Op_AssertSameAsset:
			left := vm.stringsRegs[instr.A]
			right := vm.stringsRegs[instr.B]
			if left != right {
				return runtime.ExecutionResult{}, AssetMismatchError{
					Expected: left,
					Got:      right,
				}
			}

		case Op_AssertValidAccount:
			account := vm.stringsRegs[instr.A]
			if !runtime.ValidateAccount(account) {
				return runtime.ExecutionResult{}, InvalidAccountName{Name: account}
			}

		case Op_AssertNonNegativeBalance:
			m := &vm.monetariesRegs[instr.A]
			if m.amount.Sign() < 0 {
				return runtime.ExecutionResult{}, NegativeBalanceError{
					Account: vm.stringsRegs[instr.B],
					Amount:  m.amount,
				}
			}

		case Op_SetTxMeta:
			if txMeta == nil {
				txMeta = map[string]string{}
			}
			txMeta[vm.stringsRegs[instr.A]] = vm.stringsRegs[instr.B]

		case Op_SetAccountMeta:
			if accountsMeta == nil {
				accountsMeta = runtime.AccountsMetadata{}
			}
			account := vm.stringsRegs[instr.A]
			accMeta := accountsMeta[account]
			if accMeta == nil {
				accMeta = runtime.AccountMetadata{}
				accountsMeta[account] = accMeta
			}
			accMeta[vm.stringsRegs[instr.B]] = vm.stringsRegs[instr.C]

		case Op_MetaStr:
			v, err := lookupMeta(store, vm.stringsRegs[instr.B], vm.stringsRegs[instr.C])
			if err != nil {
				return runtime.ExecutionResult{}, err
			}
			vm.stringsRegs[instr.A] = v

		case Op_MetaInt:
			account, key := vm.stringsRegs[instr.B], vm.stringsRegs[instr.C]
			v, err := lookupMeta(store, account, key)
			if err != nil {
				return runtime.ExecutionResult{}, err
			}
			n, ok := runtime.ParseNumber(v)
			if !ok {
				return runtime.ExecutionResult{}, BadMetaValueError{Account: account, Key: key, Raw: v}
			}
			vm.intsRegs[instr.A].Set(n)

		case Op_MetaPortion:
			account, key := vm.stringsRegs[instr.B], vm.stringsRegs[instr.C]
			v, err := lookupMeta(store, account, key)
			if err != nil {
				return runtime.ExecutionResult{}, err
			}
			r, perr := runtime.ParsePortion(v)
			if perr != nil {
				return runtime.ExecutionResult{}, BadMetaValueError{Account: account, Key: key, Raw: v}
			}
			vm.portionsRegs[instr.A].Set(r)

		case Op_MetaMonetary:
			account, key := vm.stringsRegs[instr.B], vm.stringsRegs[instr.C]
			v, err := lookupMeta(store, account, key)
			if err != nil {
				return runtime.ExecutionResult{}, err
			}
			asset, amount, merr := runtime.ParseMonetary(v)
			if merr != nil {
				return runtime.ExecutionResult{}, BadMetaValueError{Account: account, Key: key, Raw: v}
			}
			dest := &vm.monetariesRegs[instr.A]
			dest.asset = asset
			dest.amount.Set(amount)

			// --- Vars
		case Op_LoadVarInt:
			vm.intsRegs[instr.A].Set(&vars.IntsPool[instr.GetBC()])

		case Op_LoadVarStr:
			vm.stringsRegs[instr.A] = vars.StringsPool[instr.GetBC()]

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

		case Op_AddString:
			vm.stringsRegs[instr.A] = vm.stringsRegs[instr.B] + vm.stringsRegs[instr.C]

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

		case Op_Balance:
			account := vm.stringsRegs[instr.B]
			asset := vm.stringsRegs[instr.C]

			dest := &vm.monetariesRegs[instr.A]
			dest.asset = asset
			dest.amount.Set(runstate.GetAccountBalance(account, asset, ""))

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

		case Op_NegInt:
			arg := &vm.intsRegs[instr.B]
			vm.intsRegs[instr.A].Neg(arg)

		case Op_IntToString:
			vm.stringsRegs[instr.A] = vm.intsRegs[instr.B].String()

		case Op_PortionToString:
			vm.stringsRegs[instr.A] = vm.portionsRegs[instr.B].String()

		case Op_MonetaryToString:
			mon := &vm.monetariesRegs[instr.B]
			vm.stringsRegs[instr.A] = mon.asset + " " + mon.amount.String()

		default:
			panic("Invalid operation")
		}
	}

	return runtime.ExecutionResult{
		Postings:         runstate.GetPostings(),
		Metadata:         txMeta,
		AccountsMetadata: accountsMeta,
	}, nil
}
