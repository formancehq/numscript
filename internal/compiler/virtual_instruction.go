package compiler

import (
	"fmt"
	"math/big"
)

type reg int

type label string

type binKind interface {
	fmt.Stringer
	sig() binaryOpSig
}

type (
	opMinInt       struct{}
	opAddInt       struct{}
	opSubInt       struct{}
	opAddString    struct{}
	opSubPortion   struct{}
	opMakePortion  struct{}
	opMakeMonetary struct{}
)

type unKind interface {
	fmt.Stringer
	sig() unaryOpSig
}

type (
	opIntCopy          struct{}
	opPortionCopy      struct{}
	opGetAsset         struct{}
	opGetAmount        struct{}
	opNegInt           struct{}
	opIntToString      struct{}
	opPortionToString  struct{}
	opMonetaryToString struct{}
)

type varType interface {
	fmt.Stringer
	assembleLoad(a *assembler, dest reg, index uint16) error
}

type (
	varInt struct{}
	varStr struct{}
)

type metaType interface {
	fmt.Stringer
	assembleMeta(a *assembler, dest, account, key reg) error
}

type (
	metaStr      struct{}
	metaInt      struct{}
	metaPortion  struct{}
	metaMonetary struct{}
)

type (
	pullAccount struct {
		dest                  reg  // int: amount pulled
		account               reg  // str
		cap, overdraft, color *reg // int, int, str
		// boundedZero means "overdraft of exactly 0" without a register (the
		// plain-account case). It is mutually exclusive with overdraft != nil;
		// when set, the assembler can emit the compact single-word pull op.
		boundedZero bool
	}
	sendToAccount struct {
		account, cap *reg // str, int
	}
	// takeAccount is a pullAccount that does NOT queue: it computes the pulled
	// amount and debits the source, leaving the matching posting to a postAccount.
	// Same operands as pullAccount. Produced only by the fundsBypass peephole (the
	// source half of a fused 1-source/1-destination send); the compiler never
	// emits it directly.
	takeAccount struct {
		dest                  reg  // int: amount taken
		account               reg  // str
		cap, overdraft, color *reg // int, int, str
		boundedZero           bool
	}
	// postAccount emits a direct posting src->dst of the (already-taken) amount,
	// crediting dst without debiting src. The destination half of a fused
	// 1-source/1-destination send (see fundsBypass).
	postAccount struct {
		srcAccount, dstAccount reg  // str, str
		amount                 reg  // int
		color                  *reg // str
	}
	save struct {
		account reg  // str
		asset   reg  // str
		amount  *reg // int; nil = save all
	}
	makeAllotment struct {
		dest     []reg // int, len N
		amount   reg   // int
		portions []reg // portion, len N
	}
	checkEnoughFunds struct{ got, needed reg } // int
	assertLeftover   struct {
		portion reg  // the allotment leftover (1 - sum of the given portions)
		exact   bool // no `remaining` clause: leftover must be exactly 0, else >= 0
	}
	setCurrentAsset          struct{ asset reg }               // str
	assertSameAsset          struct{ left, right reg }         // str, str
	assertValidAccount       struct{ account reg }             // str
	assertNonNegativeBalance struct{ balance, account reg }    // monetary, str
	setTxMeta                struct{ key, value reg }          // str, str
	setAccountMeta           struct{ account, key, value reg } // str, str, str
	metaVar                  struct {
		dest         reg
		account, key reg // str, str
		typ          metaType
	}
	fetchBalance struct {
		dest           reg // monetary
		account, asset reg // str, str
	} // reads the run-state (impure)
	loadVar struct {
		dest  reg
		typ   varType
		index uint16
	}
	jmpIfZero struct {
		cond   reg // int
		target label
	}
	loadInt struct {
		dest  reg
		value big.Int
	}
	loadStr struct {
		dest  reg
		value string
	}
	binaryOp struct {
		op                binKind
		dest, left, right reg
	}
	unaryOp struct {
		op        unKind
		dest, arg reg
	}
	labelMarker struct{ label label }
)

type vInstr interface {
	dests() []reg   // registers written
	sources() []reg // registers read
	assemble(a *assembler) error

	// mapSources returns a copy of the instruction with every source (read)
	// register replaced by f(r). Destinations are left untouched. It is the
	// rewrite primitive peephole passes use for register substitution. (See
	// virtual_instruction_map.go for the implementations.)
	mapSources(f func(reg) reg) vInstr
}

func (i pullAccount) dests() []reg   { return []reg{i.dest} }
func (i pullAccount) sources() []reg { return present(&i.account, i.cap, i.overdraft, i.color) }

func (i sendToAccount) dests() []reg   { return nil }
func (i sendToAccount) sources() []reg { return present(i.account, i.cap) }

func (i takeAccount) dests() []reg   { return []reg{i.dest} }
func (i takeAccount) sources() []reg { return present(&i.account, i.cap, i.overdraft, i.color) }

func (i postAccount) dests() []reg { return nil }
func (i postAccount) sources() []reg {
	return present(&i.srcAccount, &i.dstAccount, &i.amount, i.color)
}

func (i makeAllotment) dests() []reg   { return i.dest }
func (i makeAllotment) sources() []reg { return append(append([]reg{}, i.portions...), i.amount) }

func (i checkEnoughFunds) dests() []reg   { return nil }
func (i checkEnoughFunds) sources() []reg { return []reg{i.got, i.needed} }

func (i save) dests() []reg { return nil }
func (i save) sources() []reg {
	regs := []reg{i.account, i.asset}
	if i.amount != nil {
		regs = append(regs, *i.amount)
	}
	return regs
}

func (i assertLeftover) dests() []reg   { return nil }
func (i assertLeftover) sources() []reg { return []reg{i.portion} }

func (i setCurrentAsset) dests() []reg   { return nil }
func (i setCurrentAsset) sources() []reg { return []reg{i.asset} }

func (i assertSameAsset) dests() []reg   { return nil }
func (i assertSameAsset) sources() []reg { return []reg{i.left, i.right} }

func (i assertValidAccount) dests() []reg   { return nil }
func (i assertValidAccount) sources() []reg { return []reg{i.account} }

func (i assertNonNegativeBalance) dests() []reg   { return nil }
func (i assertNonNegativeBalance) sources() []reg { return []reg{i.balance, i.account} }

func (i setTxMeta) dests() []reg   { return nil }
func (i setTxMeta) sources() []reg { return []reg{i.key, i.value} }

func (i setAccountMeta) dests() []reg   { return nil }
func (i setAccountMeta) sources() []reg { return []reg{i.account, i.key, i.value} }

func (i metaVar) dests() []reg   { return []reg{i.dest} }
func (i metaVar) sources() []reg { return []reg{i.account, i.key} }

func (i fetchBalance) dests() []reg   { return []reg{i.dest} }
func (i fetchBalance) sources() []reg { return []reg{i.account, i.asset} }

func (i loadVar) dests() []reg   { return []reg{i.dest} }
func (i loadVar) sources() []reg { return nil }

func (i jmpIfZero) dests() []reg   { return nil }
func (i jmpIfZero) sources() []reg { return []reg{i.cond} }

func (i loadInt) dests() []reg   { return []reg{i.dest} }
func (i loadInt) sources() []reg { return nil }

func (i loadStr) dests() []reg   { return []reg{i.dest} }
func (i loadStr) sources() []reg { return nil }

func (i binaryOp) dests() []reg   { return []reg{i.dest} }
func (i binaryOp) sources() []reg { return []reg{i.left, i.right} }

func (i unaryOp) dests() []reg   { return []reg{i.dest} }
func (i unaryOp) sources() []reg { return []reg{i.arg} }

func (i labelMarker) dests() []reg   { return nil }
func (i labelMarker) sources() []reg { return nil }

func present(regs ...*reg) []reg {
	out := make([]reg, 0, len(regs))
	for _, r := range regs {
		if r != nil {
			out = append(out, *r)
		}
	}
	return out
}
