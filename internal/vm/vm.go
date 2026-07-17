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

// bigZero is the implicit overdraft bound for Op_PullAccountCapZero. The runtime
// only reads the overdraft (never mutates it), so a shared instance is safe.
var bigZero = new(big.Int)

type Vm struct {
	program  Program
	runstate *runtime.RunState

	stringsRegs    [256]string // asset,string,account
	intsRegs       [256]big.Int
	portionsRegs   [256]runtime.Portion
	monetariesRegs [256]monetary

	// portScratch is reusable scratch for the integer sub_portion cross-multiply,
	// so portion arithmetic allocates nothing. Not reentrant (single-threaded VM).
	portScratch [2]big.Int
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

// Exec runs the compiled program on the (reused) VM and returns the resulting
// postings.
//
// LIFETIME: the returned slice and its posting Amounts are owned by the VM and
// valid only until the next Exec on this Vm (which recycles them). This suits the
// hot loop, where the caller consumes/persists the postings before the next run;
// a caller that needs to retain them must deep-copy first.
func Exec[S Store](
	ctx context.Context,
	vm *Vm,
	vars *Vars,
	store S, // a generic S should allow monomorphisation of the Store
) (runtime.ExecutionResult, ExecutionError) {
	runtimeStore := runtimeStoreAdapter{store: store}
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

		case Op_PullAccountCapZero:
			// compact pull: cap present, overdraft bounded-zero, no color
			account := vm.stringsRegs[instr.B]
			overdraft := bigZero // bounded by 0...
			if account == worldAccount {
				overdraft = nil // ...except world, which is unbounded
			}
			if err := runstate.Pull(&vm.intsRegs[instr.A], account, "", &vm.intsRegs[instr.C], overdraft, ""); err != nil {
				return runtime.ExecutionResult{}, StoreError{Wrapped: err}
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
				if err := runstate.SendUncapped(dest, "", color); err != nil {
					return runtime.ExecutionResult{}, StoreError{Wrapped: err}
				}
			} else {
				if err := runstate.Send(dest, "", cap, color); err != nil {
					return runtime.ExecutionResult{}, StoreError{Wrapped: err}
				}
			}

		case Op_Take:
			instrExt := instrs[pc]
			pc++

			src := vm.stringsRegs[instr.B]

			var cap *big.Int
			if instr.C != nilReg {
				cap = &vm.intsRegs[instr.C]
			}

			var overdraft *big.Int
			if src != worldAccount && instrExt.A != nilReg {
				overdraft = &vm.intsRegs[instrExt.A]
			}

			var color string
			if instrExt.B != nilReg {
				color = vm.stringsRegs[instrExt.B]
			}

			out := &vm.intsRegs[instr.A]
			switch {
			case cap != nil:
				if err := runstate.Take(out, src, "", cap, overdraft, color); err != nil {
					return runtime.ExecutionResult{}, StoreError{Wrapped: err}
				}
			case overdraft != nil:
				if err := runstate.TakeUncapped(out, src, "", overdraft, color); err != nil {
					return runtime.ExecutionResult{}, StoreError{Wrapped: err}
				}
			default:
				return runtime.ExecutionResult{}, InvalidUncappedSource{Account: src}
			}

		case Op_TakeCapZero:
			src := vm.stringsRegs[instr.B]

			overdraft := bigZero // bounded by 0...
			if src == worldAccount {
				overdraft = nil // ...except world, which is unbounded
			}

			if err := runstate.Take(&vm.intsRegs[instr.A], src, "", &vm.intsRegs[instr.C], overdraft, ""); err != nil {
				return runtime.ExecutionResult{}, StoreError{Wrapped: err}
			}

		case Op_Post:
			src := vm.stringsRegs[instr.A]
			dst := vm.stringsRegs[instr.B]
			amount := &vm.intsRegs[instr.C]
			if err := runstate.PostDirect(src, "", dst, "", "", amount); err != nil {
				return runtime.ExecutionResult{}, StoreError{Wrapped: err}
			}

		case Op_MkAllotment:
			// TODO crashes if this is the last instruction (missing ext word),
			// same as Op_PullAccount.
			instrExt := instrs[pc]
			pc++

			// TODO crashes when instr.A+instr.C > 256: the slice runs past the
			// register bank. Both are bytes, so A+C can be up to 510.
			destArrStartReg := vm.intsRegs[instr.A : instr.A+instr.C]
			inpArrStartReg := vm.portionsRegs[instr.B : instr.B+instr.C]

			amt := &vm.intsRegs[instrExt.A]

			runstate.MakeAllotment(
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
			if err := runstate.Save(account, "", asset, "", amount); err != nil {
				return runtime.ExecutionResult{}, StoreError{Wrapped: err}
			}

		case Op_AssertLeftover:
			leftover := &vm.portionsRegs[instr.A]
			sign := leftover.Num.Sign() // Den > 0, so the fraction's sign is Num's
			if sign < 0 || (instr.B == 1 && sign != 0) {
				// sum of given portions = 1 - leftover = (Den - Num) / Den
				num := new(big.Int).Sub(&leftover.Den, &leftover.Num)
				sum := new(big.Rat).SetFrac(num, &leftover.Den)
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
			v, err := lookupMeta(ctx, store, vm.stringsRegs[instr.B], vm.stringsRegs[instr.C])
			if err != nil {
				return runtime.ExecutionResult{}, err
			}
			vm.stringsRegs[instr.A] = v

		case Op_MetaInt:
			account, key := vm.stringsRegs[instr.B], vm.stringsRegs[instr.C]
			v, err := lookupMeta(ctx, store, account, key)
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
			v, err := lookupMeta(ctx, store, account, key)
			if err != nil {
				return runtime.ExecutionResult{}, err
			}
			r, perr := runtime.ParsePortion(v)
			if perr != nil {
				return runtime.ExecutionResult{}, BadMetaValueError{Account: account, Key: key, Raw: v}
			}
			// ParsePortion returns a normalized *big.Rat (Denom > 0); copy its
			// num/den into the integer Portion register.
			dest := &vm.portionsRegs[instr.A]
			dest.Num.Set(r.Num())
			dest.Den.Set(r.Denom())

		case Op_MetaMonetary:
			account, key := vm.stringsRegs[instr.B], vm.stringsRegs[instr.C]
			v, err := lookupMeta(ctx, store, account, key)
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
			// TODO both crash if vars is nil (Exec called with no vars for a
			// program that reads them), or if GetBC() >= len(vars pool) (caller
			// passed fewer vars than the program declares).
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
		// TODO both crash if GetBC() >= len(pool), e.g. an Op_LoadInt referring to
		// pool index 5 in a program whose ints pool has 3 entries.
		case Op_LoadInt:
			const_ := &vm.program.IntsPool[instr.GetBC()]
			vm.intsRegs[instr.A].Set(const_)

		case Op_LoadIntImm:
			vm.intsRegs[instr.A].SetUint64(uint64(instr.GetBC()))

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
			// (l.Num/l.Den) - (r.Num/r.Den) = (l.Num*r.Den - r.Num*l.Den)/(l.Den*r.Den)
			// Compute into scratch first so dest may alias left or right; both dens
			// are > 0 so the product den stays > 0.
			t0, t1 := &vm.portScratch[0], &vm.portScratch[1]
			t0.Mul(&left.Num, &right.Den)
			t1.Mul(&right.Num, &left.Den)
			t0.Sub(t0, t1)
			t1.Mul(&left.Den, &right.Den)
			dest := &vm.portionsRegs[instr.A]
			dest.Num.Set(t0)
			dest.Den.Set(t1)

		case Op_MkPortion:
			num := &vm.intsRegs[instr.B]
			den := &vm.intsRegs[instr.C]
			if den.Sign() == 0 {
				return runtime.ExecutionResult{}, DivideByZeroError{Numerator: *num}
			}
			dest := &vm.portionsRegs[instr.A]
			dest.Num.Set(num)
			dest.Den.Set(den)
			if den.Sign() < 0 { // keep Den > 0 (invariant relied on elsewhere)
				dest.Num.Neg(&dest.Num)
				dest.Den.Neg(&dest.Den)
			}

		case Op_MkMonetary:
			asset := vm.stringsRegs[instr.B]
			amt := &vm.intsRegs[instr.C]

			dest := &vm.monetariesRegs[instr.A]
			dest.asset = asset
			dest.amount.Set(amt)

		case Op_Balance:
			account := vm.stringsRegs[instr.B]
			asset := vm.stringsRegs[instr.C]

			bal, err := runstate.GetAccountBalance(account, "", asset, "")
			if err != nil {
				return runtime.ExecutionResult{}, StoreError{Wrapped: err}
			}
			dest := &vm.monetariesRegs[instr.A]
			dest.asset = asset
			dest.amount.Set(bal)

		// --- Unary ops
		case Op_IntCopy:
			arg := &vm.intsRegs[instr.B]
			vm.intsRegs[instr.A].Set(arg)

		case Op_PortionCopy:
			arg := &vm.portionsRegs[instr.B]
			dest := &vm.portionsRegs[instr.A]
			dest.Num.Set(&arg.Num)
			dest.Den.Set(&arg.Den)

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
			// rare (metadata) path: reduce via big.Rat so the string matches the
			// canonical lowest-terms form (e.g. "1/2", not an unreduced "2/4").
			p := &vm.portionsRegs[instr.B]
			vm.stringsRegs[instr.A] = new(big.Rat).SetFrac(&p.Num, &p.Den).String()

		case Op_MonetaryToString:
			mon := &vm.monetariesRegs[instr.B]
			vm.stringsRegs[instr.A] = mon.asset + " " + mon.amount.String()

		default:
			return runtime.ExecutionResult{}, InternalError{Opcode: instr.Opcode}
		}
	}

	return runtime.ExecutionResult{
		Postings:         runstate.PostingsRef(),
		Metadata:         txMeta,
		AccountsMetadata: accountsMeta,
	}, nil
}
