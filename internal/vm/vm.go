package vm

import (
	"context"
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
	verified bool
	info     programInfo
	runstate *runtime.RunState

	stringsRegs    []string // asset,string,account
	intsRegs       []big.Int
	portionsRegs   []big.Rat
	monetariesRegs []monetary
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
		ctx context.Context,
		account string,
		asset string,
		color string,
	) (*big.Int, error)

	GetMetadata(
		ctx context.Context,
		account,
		key string,
	) (string, bool, error)
}

func lookupMeta(ctx context.Context, store Store, account, key string) (string, ExecutionError) {
	v, ok, err := store.GetMetadata(ctx, account, key)
	if err != nil {
		return "", StoreError{Wrapped: err}
	}
	if !ok {
		return "", MetadataNotFoundError{Account: account, Key: key}
	}
	return v, nil
}

type runtimeStoreAdapter struct {
	ctx   context.Context
	store Store
}

func (s runtimeStoreAdapter) GetBalance(
	account string,
	asset string,
	color string,
) (*big.Int, error) {
	return s.store.GetBalance(s.ctx, account, asset, color)
}

func Exec[S Store](
	ctx context.Context,
	vm *Vm,
	vars *Vars,
	store S, // a generic S should allow monomorphisation of the Store
) (runtime.ExecutionResult, ExecutionError) {
	runtimeStore := runtimeStoreAdapter{store: store}
	// RunState fetches balances lazily through this store; a fetch error surfaces
	// from the RunState call that triggered it, wrapped in StoreError below.
	if !vm.verified {
		info, err := verify(vm.program)
		if err != nil {
			return runtime.ExecutionResult{}, MalformedProgramError{Reason: err.Error()}
		}
		vm.info = info
		vm.intsRegs = make([]big.Int, info.regs.ints)
		vm.stringsRegs = make([]string, info.regs.strings)
		vm.portionsRegs = make([]big.Rat, info.regs.portions)
		vm.monetariesRegs = make([]monetary, info.regs.monetaries)
		vm.verified = true
	}

	if vm.info.varIntsLen > 0 || vm.info.varStrsLen > 0 {
		if vars == nil || len(vars.IntsPool) < vm.info.varIntsLen || len(vars.StringsPool) < vm.info.varStrsLen {
			return runtime.ExecutionResult{}, MalformedProgramError{Reason: "program reads more variables than were provided"}
		}
	}

	if vm.runstate == nil {
		vm.runstate = runtime.New(runtimeStore)
	} else {
		vm.runstate.Reset(runtimeStore)
	}
	runstate := vm.runstate

	var txMeta map[string]string
	var accountsMeta runtime.AccountsMetadata

	// Hoist register banks and constant pools into locals so the hot loop indexes
	// them directly instead of reloading the header off *vm / vm.program on every
	// access.
	intsRegs := &vm.intsRegs
	stringsRegs := &vm.stringsRegs
	portionsRegs := &vm.portionsRegs
	monetariesRegs := &vm.monetariesRegs
	intsPool := vm.program.IntsPool
	stringsPool := vm.program.StringsPool

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

			account := stringsRegs[instr.B]

			var cap *big.Int
			if instr.C != nilReg {
				cap = &intsRegs[instr.C]
			}

			var overdraft *big.Int
			if account != worldAccount && instrExt.A != nilReg {
				overdraft = &intsRegs[instrExt.A]
			}

			var color string
			if instrExt.B != nilReg {
				color = stringsRegs[instrExt.B]
			}

			out := &intsRegs[instr.A]
			switch {
			case cap != nil:
				if err := runstate.Pull(out, account, "", cap, overdraft, color); err != nil {
					return runtime.ExecutionResult{}, StoreError{Wrapped: err}
				}
			case overdraft != nil:
				if err := runstate.PullUncapped(out, account, "", overdraft, color); err != nil {
					return runtime.ExecutionResult{}, StoreError{Wrapped: err}
				}
			default:
				return runtime.ExecutionResult{}, InvalidUncappedSource{Account: account}
			}

		case Op_SendToAccount:
			var dest *string
			if instr.A != nilReg {
				s := stringsRegs[instr.A]
				dest = &s
			}

			var cap *big.Int
			if instr.B != nilReg {
				cap = &intsRegs[instr.B]
			}

			var color *string
			if instr.C != nilReg {
				color = &stringsRegs[instr.C]
			}

			if cap == nil {
				if err := runstate.SendUncapped(dest, "", color); err != nil {
					return runtime.ExecutionResult{}, StoreError{Wrapped: err}
				}
			} else {
				if err := runstate.Send(dest, "", cap, color); err != nil {
					return runtime.ExecutionResult{}, StoreError{Wrapped: err}
				}
			}

		case Op_MkAllotment:
			instrExt := instrs[pc]
			pc++

			destArrStartReg := intsRegs[instr.A : int(instr.A)+int(instr.C)]
			inpArrStartReg := portionsRegs[instr.B : int(instr.B)+int(instr.C)]

			amt := &intsRegs[instrExt.A]

			runtime.MakeAllotment(
				destArrStartReg,
				amt,
				inpArrStartReg,
			)

		case Op_CheckEnoughFunds:
			got := &intsRegs[instr.A]
			needed := &intsRegs[instr.B]
			if got.Cmp(needed) == -1 {
				return runtime.ExecutionResult{}, MissingFundsError{
					Asset:  currentAsset,
					Got:    got,
					Needed: needed,
				}
			}

		case Op_Save:
			account := stringsRegs[instr.A]
			asset := stringsRegs[instr.B]
			var amount *big.Int
			if instr.C != nilReg {
				amount = &intsRegs[instr.C]
			}
			if err := runstate.Save(account, "", asset, "", amount); err != nil {
				return runtime.ExecutionResult{}, StoreError{Wrapped: err}
			}

		case Op_AssertLeftover:
			leftover := &portionsRegs[instr.A]
			sign := leftover.Sign()
			if sign < 0 || (instr.B == 1 && sign != 0) {
				sum := new(big.Rat).Sub(big.NewRat(1, 1), leftover)
				return runtime.ExecutionResult{}, InvalidAllotmentSum{ActualSum: *sum}
			}

		case Op_SetCurrentAsset:
			currentAsset = stringsRegs[instr.A]
			runstate.SetCurrentAsset(currentAsset)

		case Op_AssertSameAsset:
			left := stringsRegs[instr.A]
			right := stringsRegs[instr.B]
			if left != right {
				return runtime.ExecutionResult{}, AssetMismatchError{
					Expected: left,
					Got:      right,
				}
			}

		case Op_AssertValidAccount:
			account := stringsRegs[instr.A]
			if !runtime.ValidateAccount(account) {
				return runtime.ExecutionResult{}, InvalidAccountName{Name: account}
			}

		case Op_AssertNonNegativeBalance:
			m := &monetariesRegs[instr.A]
			if m.amount.Sign() < 0 {
				return runtime.ExecutionResult{}, NegativeBalanceError{
					Account: stringsRegs[instr.B],
					Amount:  m.amount,
				}
			}

		case Op_SetTxMeta:
			if txMeta == nil {
				txMeta = map[string]string{}
			}
			txMeta[stringsRegs[instr.A]] = stringsRegs[instr.B]

		case Op_SetAccountMeta:
			if accountsMeta == nil {
				accountsMeta = runtime.AccountsMetadata{}
			}
			account := stringsRegs[instr.A]
			accMeta := accountsMeta[account]
			if accMeta == nil {
				accMeta = runtime.AccountMetadata{}
				accountsMeta[account] = accMeta
			}
			accMeta[stringsRegs[instr.B]] = stringsRegs[instr.C]

		case Op_MetaStr:
			v, err := lookupMeta(ctx, store, stringsRegs[instr.B], stringsRegs[instr.C])
			if err != nil {
				return runtime.ExecutionResult{}, err
			}
			stringsRegs[instr.A] = v

		case Op_MetaInt:
			account, key := stringsRegs[instr.B], stringsRegs[instr.C]
			v, err := lookupMeta(ctx, store, account, key)
			if err != nil {
				return runtime.ExecutionResult{}, err
			}
			n, ok := runtime.ParseNumber(v)
			if !ok {
				return runtime.ExecutionResult{}, BadMetaValueError{Account: account, Key: key, Raw: v}
			}
			intsRegs[instr.A].Set(n)

		case Op_MetaPortion:
			account, key := stringsRegs[instr.B], stringsRegs[instr.C]
			v, err := lookupMeta(ctx, store, account, key)
			if err != nil {
				return runtime.ExecutionResult{}, err
			}
			r, perr := runtime.ParsePortion(v)
			if perr != nil {
				return runtime.ExecutionResult{}, BadMetaValueError{Account: account, Key: key, Raw: v}
			}
			portionsRegs[instr.A].Set(r)

		case Op_MetaMonetary:
			account, key := stringsRegs[instr.B], stringsRegs[instr.C]
			v, err := lookupMeta(ctx, store, account, key)
			if err != nil {
				return runtime.ExecutionResult{}, err
			}
			asset, amount, merr := runtime.ParseMonetary(v)
			if merr != nil {
				return runtime.ExecutionResult{}, BadMetaValueError{Account: account, Key: key, Raw: v}
			}
			dest := &monetariesRegs[instr.A]
			dest.asset = asset
			dest.amount.Set(amount)

			// --- Vars
		case Op_LoadVarInt:
			intsRegs[instr.A].Set(&vars.IntsPool[instr.GetBC()])

		case Op_LoadVarStr:
			stringsRegs[instr.A] = vars.StringsPool[instr.GetBC()]

		// --- Jumps
		case Op_JmpIfZero:
			arg := &intsRegs[instr.A]
			if arg.Sign() == 0 {
				pc = int(instr.GetBC())
			}

		// --- consts
		case Op_LoadInt:
			const_ := &intsPool[instr.GetBC()]
			intsRegs[instr.A].Set(const_)

		case Op_LoadStr:
			const_ := stringsPool[instr.GetBC()]
			stringsRegs[instr.A] = const_

			// ---  Binary ops
		case Op_MinInt:
			left := &intsRegs[instr.B]
			right := &intsRegs[instr.C]
			if left.Cmp(right) == -1 {
				intsRegs[instr.A].Set(left)
			} else {
				intsRegs[instr.A].Set(right)
			}

		case Op_AddInt:
			left := &intsRegs[instr.B]
			right := &intsRegs[instr.C]
			intsRegs[instr.A].Add(left, right)

		case Op_SubInt:
			left := &intsRegs[instr.B]
			right := &intsRegs[instr.C]
			intsRegs[instr.A].Sub(left, right)

		case Op_AddString:
			stringsRegs[instr.A] = stringsRegs[instr.B] + stringsRegs[instr.C]

		case Op_SubPortion:
			left := &portionsRegs[instr.B]
			right := &portionsRegs[instr.C]
			portionsRegs[instr.A].Sub(left, right)

		case Op_MkPortion:
			num := &intsRegs[instr.B]
			den := &intsRegs[instr.C]
			if den.Sign() == 0 {
				return runtime.ExecutionResult{}, DivideByZeroError{Numerator: *num}
			}
			portionsRegs[instr.A].SetFrac(num, den)

		case Op_MkMonetary:
			asset := stringsRegs[instr.B]
			amt := &intsRegs[instr.C]

			dest := &monetariesRegs[instr.A]
			dest.asset = asset
			dest.amount.Set(amt)

		case Op_Balance:
			account := stringsRegs[instr.B]
			asset := stringsRegs[instr.C]

			bal, err := runstate.GetAccountBalance(account, "", asset, "")
			if err != nil {
				return runtime.ExecutionResult{}, StoreError{Wrapped: err}
			}
			dest := &monetariesRegs[instr.A]
			dest.asset = asset
			dest.amount.Set(bal)

		// --- Unary ops
		case Op_IntCopy:
			arg := &intsRegs[instr.B]
			intsRegs[instr.A].Set(arg)

		case Op_PortionCopy:
			arg := &portionsRegs[instr.B]
			portionsRegs[instr.A].Set(arg)

		case Op_GetAsset:
			arg := &monetariesRegs[instr.B]
			stringsRegs[instr.A] = arg.asset

		case Op_GetAmount:
			arg := &monetariesRegs[instr.B]
			intsRegs[instr.A].Set(&arg.amount)

		case Op_NegInt:
			arg := &intsRegs[instr.B]
			intsRegs[instr.A].Neg(arg)

		case Op_IntToString:
			stringsRegs[instr.A] = intsRegs[instr.B].String()

		case Op_PortionToString:
			stringsRegs[instr.A] = portionsRegs[instr.B].String()

		case Op_MonetaryToString:
			mon := &monetariesRegs[instr.B]
			stringsRegs[instr.A] = mon.asset + " " + mon.amount.String()

		default:
			return runtime.ExecutionResult{}, InternalError{Opcode: instr.Opcode}
		}
	}

	return runtime.ExecutionResult{
		Postings:         runstate.GetPostings(),
		Metadata:         txMeta,
		AccountsMetadata: accountsMeta,
	}, nil
}
