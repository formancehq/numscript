package vm

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/runtime"
)

const nilReg byte = 0xFF

// The three ops a mark cannot survive, rejected by the arms below. save is on the
// list permanently, unlike the other two: it floors the balance at zero, and the
// clamp destroys the information needed to invert it.
var (
	errSendWhileMarkOpen     = errors.New("send while a mark is open")
	errSetAssetWhileMarkOpen = errors.New("set_current_asset while a mark is open")
	errSaveWhileMarkOpen     = errors.New("save while a mark is open")
)

type Vm struct {
	program  Program
	runstate *runtime.RunState

	// a monetary is not a bank of its own: it travels as a (str asset, int amount)
	// register pair
	stringsRegs  [256]string // asset,string,account
	intsRegs     [256]big.Int
	portionsRegs [256]big.Rat
	boolsRegs    [256]bool
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
	runtimeStore := runtimeStoreAdapter{
		ctx:   ctx,
		store: store,
	}
	// RunState fetches balances lazily through this store; a fetch error surfaces
	// from the RunState call that triggered it, wrapped in StoreError below.
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
	boolsRegs := &vm.boolsRegs
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
			// TODO crashes if this is the last instruction (the ext word is
			// missing): instrs[pc] reads past the end. e.g. a program ending in a
			// lone Op_PullAccount word.
			instrExt := instrs[pc]
			pc++

			account := stringsRegs[instr.B]

			var cap *big.Int
			if instr.C != nilReg {
				cap = &intsRegs[instr.C]
			}

			var overdraft *big.Int
			if instrExt.A != nilReg {
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
			// a send while a mark is open would consume the queue from the front and
			// leave that mark pointing at the wrong boundary. Compiled numscript never
			// emits it, since sources only pull.
			if runstate.HasOpenMark() {
				return runtime.ExecutionResult{}, InternalError{Err: errSendWhileMarkOpen}
			}

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
			// a save while a mark is open survives the rewind, which only repays queued
			// sources and reverses postings; its floor at zero is not invertible at all.
			if runstate.HasOpenMark() {
				return runtime.ExecutionResult{}, InternalError{Err: errSaveWhileMarkOpen}
			}

			account := stringsRegs[instr.A]
			asset := stringsRegs[instr.B]
			var amount *big.Int
			if instr.C != nilReg {
				amount = &intsRegs[instr.C]
			}
			if err := runstate.Save(account, "", asset, "", amount); err != nil {
				return runtime.ExecutionResult{}, StoreError{Wrapped: err}
			}

		// the mark ops take no register: the mark is a depth on a LIFO the run-state
		// owns, so nothing can name a depth it never marked
		case Op_MarkPush:
			runstate.MarkPush()

		case Op_MarkEnd:
			if err := runstate.MarkEnd(instr.A == 1); err != nil {
				return runtime.ExecutionResult{}, InternalError{Err: err}
			}

		case Op_AssertLeftover:
			leftover := &portionsRegs[instr.A]
			sign := leftover.Sign()
			if sign < 0 || (instr.B == 1 && sign != 0) {
				sum := new(big.Rat).Sub(big.NewRat(1, 1), leftover)
				return runtime.ExecutionResult{}, InvalidAllotmentSum{ActualSum: *sum}
			}

		case Op_SetCurrentAsset:
			// a rewind repays queued funds into the current asset's balance, so
			// changing the asset mid-region would repay the wrong one
			if runstate.HasOpenMark() {
				return runtime.ExecutionResult{}, InternalError{Err: errSetAssetWhileMarkOpen}
			}
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

		case Op_AssertValidColor:
			color := stringsRegs[instr.A]
			if !runtime.ValidateColor(color) {
				return runtime.ExecutionResult{}, InvalidColor{Color: color}
			}

		case Op_AssertNonNegativeBalance:
			amount := &intsRegs[instr.A]
			if amount.Sign() < 0 {
				return runtime.ExecutionResult{}, NegativeBalanceError{
					Account: stringsRegs[instr.B],
					Amount:  *amount,
				}
			}

		case Op_AssertNonNegativeAmount:
			amount := &intsRegs[instr.A]
			if amount.Sign() < 0 {
				return runtime.ExecutionResult{}, NegativeAmountError{Amount: *amount}
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
			// TODO crashes if this is the last instruction (the ext word carrying the
			// amount destination is missing), same as Op_PullAccount.
			instrExt := instrs[pc]
			pc++

			account, key := stringsRegs[instr.B], stringsRegs[instr.C]
			v, err := lookupMeta(ctx, store, account, key)
			if err != nil {
				return runtime.ExecutionResult{}, err
			}
			asset, amount, merr := runtime.ParseMonetary(v)
			if merr != nil {
				return runtime.ExecutionResult{}, BadMetaValueError{Account: account, Key: key, Raw: v}
			}
			stringsRegs[instr.A] = asset
			intsRegs[instrExt.A].Set(amount)

			// --- Vars
			// TODO both crash if vars is nil (Exec called with no vars for a
			// program that reads them), or if GetBC() >= len(vars pool) (caller
			// passed fewer vars than the program declares).
		case Op_LoadVarInt:
			intsRegs[instr.A].Set(&vars.IntsPool[instr.GetBC()])

		case Op_LoadVarStr:
			stringsRegs[instr.A] = vars.StringsPool[instr.GetBC()]

		// --- Jumps
		case Op_JmpIfFalse:
			if !boolsRegs[instr.A] {
				pc += int(instr.GetBC())
			}

		case Op_JmpIfTrue:
			if boolsRegs[instr.A] {
				pc += int(instr.GetBC())
			}

		case Op_Jmp:
			pc += int(instr.GetBC())

		// --- consts
		// TODO both crash if GetBC() >= len(pool), e.g. an Op_LoadInt referring to
		// pool index 5 in a program whose ints pool has 3 entries.
		case Op_LoadInt:
			const_ := &intsPool[instr.GetBC()]
			intsRegs[instr.A].Set(const_)

		case Op_LoadStr:
			const_ := stringsPool[instr.GetBC()]
			stringsRegs[instr.A] = const_

		case Op_ConstTrue:
			boolsRegs[instr.A] = true

		case Op_ConstFalse:
			boolsRegs[instr.A] = false

			// ---  Binary ops
		case Op_LtInt:
			boolsRegs[instr.A] = intsRegs[instr.B].Cmp(&intsRegs[instr.C]) < 0

		case Op_EqInt:
			boolsRegs[instr.A] = intsRegs[instr.B].Cmp(&intsRegs[instr.C]) == 0

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

		case Op_StrEq:
			boolsRegs[instr.A] = stringsRegs[instr.B] == stringsRegs[instr.C]

		// portion comparison is *value* comparison: big.Rat normalises on
		// construction, so 1/2 and 2/4 are the same rational and compare equal.
		// Comparing numerator/denominator pairs separately would be wrong.
		case Op_LtPortion:
			boolsRegs[instr.A] = portionsRegs[instr.B].Cmp(&portionsRegs[instr.C]) < 0

		case Op_EqPortion:
			boolsRegs[instr.A] = portionsRegs[instr.B].Cmp(&portionsRegs[instr.C]) == 0

		case Op_AddPortion:
			left := &portionsRegs[instr.B]
			right := &portionsRegs[instr.C]
			portionsRegs[instr.A].Add(left, right)

		case Op_SubPortion:
			left := &portionsRegs[instr.B]
			right := &portionsRegs[instr.C]
			portionsRegs[instr.A].Sub(left, right)

		case Op_MulPortion:
			left := &portionsRegs[instr.B]
			right := &portionsRegs[instr.C]
			portionsRegs[instr.A].Mul(left, right)

		case Op_IntToPortion:
			portionsRegs[instr.A].SetInt(&intsRegs[instr.B])

		// floor: big.Rat's denominator is always positive, so Div (Euclidean) is
		// the floor for negatives too
		case Op_PortionToInt:
			p := &portionsRegs[instr.B]
			intsRegs[instr.A].Div(p.Num(), p.Denom())

		case Op_MkPortion:
			num := &intsRegs[instr.B]
			den := &intsRegs[instr.C]
			if den.Sign() == 0 {
				return runtime.ExecutionResult{}, DivideByZeroError{Numerator: *num}
			}
			portionsRegs[instr.A].SetFrac(num, den)

		case Op_Balance:
			account := stringsRegs[instr.B]
			asset := stringsRegs[instr.C]

			bal, err := runstate.GetAccountBalance(account, "", asset, "")
			if err != nil {
				return runtime.ExecutionResult{}, StoreError{Wrapped: err}
			}
			// only the amount: the asset of the result is the asset operand, which
			// the caller already holds in reg C
			intsRegs[instr.A].Set(bal)

		// --- Unary ops
		case Op_IntCopy:
			arg := &intsRegs[instr.B]
			intsRegs[instr.A].Set(arg)

		case Op_PortionCopy:
			arg := &portionsRegs[instr.B]
			portionsRegs[instr.A].Set(arg)

		case Op_StrCopy:
			stringsRegs[instr.A] = stringsRegs[instr.B]

		case Op_BoolCopy:
			boolsRegs[instr.A] = boolsRegs[instr.B]

		case Op_NegInt:
			arg := &intsRegs[instr.B]
			intsRegs[instr.A].Neg(arg)

		case Op_IntToString:
			stringsRegs[instr.A] = intsRegs[instr.B].String()

		case Op_PortionToString:
			stringsRegs[instr.A] = portionsRegs[instr.B].String()

		case Op_MonetaryToString:
			stringsRegs[instr.A] = stringsRegs[instr.B] + " " + intsRegs[instr.C].String()

		case Op_IsZero:
			boolsRegs[instr.A] = intsRegs[instr.B].Sign() == 0

		case Op_Not:
			boolsRegs[instr.A] = !boolsRegs[instr.B]

		default:
			return runtime.ExecutionResult{}, InternalError{Err: fmt.Errorf("unknown opcode %d", instr.Opcode)}
		}
	}

	return runtime.ExecutionResult{
		Postings:         runstate.GetPostings(),
		Metadata:         txMeta,
		AccountsMetadata: accountsMeta,
	}, nil
}
