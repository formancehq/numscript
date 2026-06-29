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
	opSubPortion   struct{}
	opMakePortion  struct{}
	opMakeMonetary struct{}
)

type unKind interface {
	fmt.Stringer
	sig() unaryOpSig
}

type (
	opIntCopy     struct{}
	opPortionCopy struct{}
	opGetAsset    struct{}
	opGetAmount   struct{}
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
	makeAllotment struct {
		dest     []reg // int, len N
		amount   reg   // int
		portions []reg // portion, len N
	}
	checkEnoughFunds    struct{ got, needed reg } // int
	setCurrentAsset     struct{ asset reg }       // str
	checkEqCurrentAsset struct{ got reg }         // str
	fetchVariable       struct {
		dest  reg
		index uint32
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

func (i makeAllotment) dests() []reg   { return i.dest }
func (i makeAllotment) sources() []reg { return append(append([]reg{}, i.portions...), i.amount) }

func (i checkEnoughFunds) dests() []reg   { return nil }
func (i checkEnoughFunds) sources() []reg { return []reg{i.got, i.needed} }

func (i setCurrentAsset) dests() []reg   { return nil }
func (i setCurrentAsset) sources() []reg { return []reg{i.asset} }

func (i checkEqCurrentAsset) dests() []reg   { return nil }
func (i checkEqCurrentAsset) sources() []reg { return []reg{i.got} }

func (i fetchVariable) dests() []reg   { return []reg{i.dest} }
func (i fetchVariable) sources() []reg { return nil }

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
