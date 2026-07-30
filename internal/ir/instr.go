// Package ir is the compiler's intermediate representation, and everything that
// operates on it: Parse and Dump convert to and from the textual format (see
// ir-textual-format.md), Typecheck checks register types, Assemble lowers to a
// vm.Program. The grammar's AST is internal, so callers only see instructions.
package ir

import (
	"fmt"
	"math/big"
)

type Reg uint

type Label string

type BinKind interface {
	fmt.Stringer
	sig() binaryOpSig
}

type (
	OpMinInt    struct{}
	OpAddInt    struct{}
	OpSubInt    struct{}
	OpAddString struct{}
	// OpStrEq yields a bool: it is the one comparison that produces a value
	// rather than trapping, and what the jumps branch on.
	OpStrEq       struct{}
	OpSubPortion  struct{}
	OpMakePortion struct{}
	// OpMonetaryToString takes the asset (str) and the amount (int) of a monetary
	// and produces its "ASSET AMOUNT" form, the inverse of runtime.ParseMonetary.
	OpMonetaryToString struct{}
)

type UnKind interface {
	fmt.Stringer
	sig() unaryOpSig
}

type (
	OpIntCopy     struct{}
	OpPortionCopy struct{}
	OpNegInt      struct{}
	OpIntToString struct{}
	// OpIsZero projects an int onto a bool, which is how a quantity reaches a
	// jump: the jumps take a bool, so the projection has to be explicit.
	OpIsZero          struct{}
	OpPortionToString struct{}
)

type VarType interface {
	fmt.Stringer
	assembleLoad(a *assembler, Dest Reg, Index uint16) error
}

type (
	VarInt struct{}
	VarStr struct{}
)

type MetaType interface {
	fmt.Stringer
	assembleMeta(a *assembler, Dest, Account, Key Reg) error
}

type (
	MetaStr     struct{}
	MetaInt     struct{}
	MetaPortion struct{}
)

type (
	PullAccount struct {
		Dest                  Reg  // int: amount pulled
		Account               Reg  // str
		Cap, Overdraft, Color *Reg // int, int, str
	}
	SendToAccount struct {
		Account, Cap *Reg // str, int
	}
	Save struct {
		Account Reg  // str
		Asset   Reg  // str
		Amount  *Reg // int; nil = save all
	}
	MakeAllotment struct {
		Dest     []Reg // int, len N
		Amount   Reg   // int
		Portions []Reg // portion, len N
	}
	CheckEnoughFunds struct{ Got, Needed Reg } // int
	AssertLeftover   struct {
		Portion Reg  // the allotment leftover (1 - sum of the given Portions)
		Exact   bool // no `remaining` clause: leftover must be exactly 0, else >= 0
	}
	SetCurrentAsset          struct{ Asset Reg }               // str
	AssertSameAsset          struct{ Left, Right Reg }         // str, str
	AssertValidAccount       struct{ Account Reg }             // str
	AssertValidColor         struct{ Color Reg }               // str
	AssertNonNegativeBalance struct{ Balance, Account Reg }    // int (the amount), str
	SetTxMeta                struct{ Key, Value Reg }          // str, str
	SetAccountMeta           struct{ Account, Key, Value Reg } // str, str, str
	MetaVar                  struct {
		Dest         Reg
		Account, Key Reg // str, str
		Typ          MetaType
	}
	// MetaMonetary is meta<monetary>: one store read yields both halves, so it is
	// the only two-destination read and is not a MetaType.
	MetaMonetary struct {
		DestAsset  Reg // str
		DestAmount Reg // int
		Account    Reg // str
		Key        Reg // str
	}
	FetchBalance struct {
		Dest           Reg // int (the amount; the asset is the Asset operand)
		Account, Asset Reg // str, str
	} // reads the run-state (impure)
	LoadVar struct {
		Dest  Reg
		Typ   VarType
		Index uint16
	}
	// The two conditional jumps differ only in which edge of the bool jumps, so
	// either branch of a condition is one instruction and no negation is needed.
	JmpIfFalse struct {
		Cond   Reg // bool
		Target Label
	}
	JmpIfTrue struct {
		Cond   Reg // bool
		Target Label
	}
	Jmp struct {
		Target Label
	}
	LoadInt struct {
		Dest  Reg
		Value big.Int
	}
	LoadStr struct {
		Dest  Reg
		Value string
	}
	// ConstBool assembles to Op_ConstTrue or Op_ConstFalse: the value is in the
	// opcode, so there is no pool entry.
	ConstBool struct {
		Dest  Reg
		Value bool
	}
	BinaryOp struct {
		Op                BinKind
		Dest, Left, Right Reg
	}
	UnaryOp struct {
		Op        UnKind
		Dest, Arg Reg
	}
	LabelMarker struct{ Label Label }

	Snapshot struct{ Dest Reg } // int: the source-queue mark (for oneof backtracking)
	Restore  struct{ Mark Reg } // int: a mark produced by Snapshot
)

type Instr interface {
	dests() []Reg   // registers written
	sources() []Reg // registers read
	assemble(a *assembler) error
}

func (i PullAccount) dests() []Reg   { return []Reg{i.Dest} }
func (i PullAccount) sources() []Reg { return present(&i.Account, i.Cap, i.Overdraft, i.Color) }

func (i SendToAccount) dests() []Reg   { return nil }
func (i SendToAccount) sources() []Reg { return present(i.Account, i.Cap) }

func (i MakeAllotment) dests() []Reg   { return i.Dest }
func (i MakeAllotment) sources() []Reg { return append(append([]Reg{}, i.Portions...), i.Amount) }

func (i CheckEnoughFunds) dests() []Reg   { return nil }
func (i CheckEnoughFunds) sources() []Reg { return []Reg{i.Got, i.Needed} }

func (i Save) dests() []Reg { return nil }
func (i Save) sources() []Reg {
	regs := []Reg{i.Account, i.Asset}
	if i.Amount != nil {
		regs = append(regs, *i.Amount)
	}
	return regs
}

func (i AssertLeftover) dests() []Reg   { return nil }
func (i AssertLeftover) sources() []Reg { return []Reg{i.Portion} }

func (i SetCurrentAsset) dests() []Reg   { return nil }
func (i SetCurrentAsset) sources() []Reg { return []Reg{i.Asset} }

func (i AssertSameAsset) dests() []Reg   { return nil }
func (i AssertSameAsset) sources() []Reg { return []Reg{i.Left, i.Right} }

func (i AssertValidAccount) dests() []Reg   { return nil }
func (i AssertValidAccount) sources() []Reg { return []Reg{i.Account} }

func (i AssertValidColor) dests() []Reg   { return nil }
func (i AssertValidColor) sources() []Reg { return []Reg{i.Color} }

func (i AssertNonNegativeBalance) dests() []Reg   { return nil }
func (i AssertNonNegativeBalance) sources() []Reg { return []Reg{i.Balance, i.Account} }

func (i SetTxMeta) dests() []Reg   { return nil }
func (i SetTxMeta) sources() []Reg { return []Reg{i.Key, i.Value} }

func (i SetAccountMeta) dests() []Reg   { return nil }
func (i SetAccountMeta) sources() []Reg { return []Reg{i.Account, i.Key, i.Value} }

func (i MetaVar) dests() []Reg   { return []Reg{i.Dest} }
func (i MetaVar) sources() []Reg { return []Reg{i.Account, i.Key} }

func (i MetaMonetary) dests() []Reg   { return []Reg{i.DestAsset, i.DestAmount} }
func (i MetaMonetary) sources() []Reg { return []Reg{i.Account, i.Key} }

func (i FetchBalance) dests() []Reg   { return []Reg{i.Dest} }
func (i FetchBalance) sources() []Reg { return []Reg{i.Account, i.Asset} }

func (i LoadVar) dests() []Reg   { return []Reg{i.Dest} }
func (i LoadVar) sources() []Reg { return nil }

func (i JmpIfFalse) dests() []Reg   { return nil }
func (i JmpIfFalse) sources() []Reg { return []Reg{i.Cond} }

func (i JmpIfTrue) dests() []Reg   { return nil }
func (i JmpIfTrue) sources() []Reg { return []Reg{i.Cond} }

func (i Jmp) dests() []Reg   { return nil }
func (i Jmp) sources() []Reg { return nil }

func (i LoadInt) dests() []Reg   { return []Reg{i.Dest} }
func (i LoadInt) sources() []Reg { return nil }

func (i LoadStr) dests() []Reg   { return []Reg{i.Dest} }
func (i LoadStr) sources() []Reg { return nil }

func (i ConstBool) dests() []Reg   { return []Reg{i.Dest} }
func (i ConstBool) sources() []Reg { return nil }

func (i BinaryOp) dests() []Reg   { return []Reg{i.Dest} }
func (i BinaryOp) sources() []Reg { return []Reg{i.Left, i.Right} }

func (i UnaryOp) dests() []Reg   { return []Reg{i.Dest} }
func (i UnaryOp) sources() []Reg { return []Reg{i.Arg} }

func (i LabelMarker) dests() []Reg   { return nil }
func (i LabelMarker) sources() []Reg { return nil }

func (i Snapshot) dests() []Reg   { return []Reg{i.Dest} }
func (i Snapshot) sources() []Reg { return nil }

func (i Restore) dests() []Reg   { return nil }
func (i Restore) sources() []Reg { return []Reg{i.Mark} }

func present(regs ...*Reg) []Reg {
	out := make([]Reg, 0, len(regs))
	for _, r := range regs {
		if r != nil {
			out = append(out, *r)
		}
	}
	return out
}
