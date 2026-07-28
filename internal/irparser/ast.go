package irparser

import (
	"github.com/formancehq/numscript/internal/parser"
)

// ---- AST types for the IR textual format ----

// Program is the root of the parsed IR text.
type Program struct {
	Stmts []Stmt
}

// Stmt is either a LabelStmt or an InstrStmt.
type Stmt interface {
	stmt()
}

// LabelStmt represents a label marker line, e.g. "#inorder_end_0".
type LabelStmt struct {
	Range parser.Range
	Name  string
}

func (*LabelStmt) stmt() {}

// InstrStmt represents one instruction line.
type InstrStmt struct {
	Range parser.Range

	// Dest is the destination; nil when the instruction has no dest (e.g. "set_current_asset($r3)").
	Dest *Dest

	// One of these is set:
	Call           *InstrCall // e.g. "mk_monetary($r0, $r1)"
	Const          *Const     // e.g. "$r0 = \"USD/2\"" or "$r0 = 42"
	Infix          *Infix     // e.g. "$r3 = $r1 + $r2"
	CompoundAssign *Infix     // e.g. "$r5 += $r9"  (Left is implicit from Dest)
}

func (*InstrStmt) stmt() {}

// Dest is the LHS of an assignment.
type Dest struct {
	Range parser.Range
	Kind  DestKind
	Regs  []RegRef // non-empty for DestReg and DestList
}

type DestKind int

const (
	DestReg     DestKind = iota // single register
	DestDiscard                 // _
	DestList                    // [$r0, $r1]
)

// RegRef is a parsed register reference, e.g. "$r0".
type RegRef struct {
	Range parser.Range
	Name  string
}

// InstrCall is a function-call instruction: name(args).
type InstrCall struct {
	Range     parser.Range
	Name      string   // e.g. "mk_monetary", "pull_account"
	TypeParam string   // "" if none, else "int", "str", etc.
	Args      []Arg    // may be empty
}

// Arg is a single argument to an instruction.
type Arg struct {
	Range parser.Range
	Label string // empty for positional args; "account", "cap", etc. for labeled
	Value Value
}

// Value is a value that can appear as an argument.
type Value struct {
	Range parser.Range
	Kind  ValueKind

	// Exactly one of these is set, depending on Kind:
	Reg   *RegRef   // ValReg
	Label *string   // ValLabel: the label name without '#'
	Int   *string   // ValInt: raw numeric string
	Bool  *bool     // ValBool
	Regs  *[]RegRef // ValRegList
}

type ValueKind int

const (
	ValReg     ValueKind = iota // $r0
	ValLabel                    // #my_label
	ValInt                      // 42
	ValBool                     // true / false
	ValRegList                  // [$r0, $r1]
)

// Const is a constant literal: a string or an integer.
type Const struct {
	Range parser.Range
	Kind  ConstKind

	// Exactly one is set:
	StrVal *string
	IntVal *string // raw numeric string
}

type ConstKind int

const (
	ConstString ConstKind = iota
	ConstInt
)

// Infix is a binary operation with infix syntax.
type Infix struct {
	Range parser.Range
	Op    string // "+" or "-"
	Left  RegRef
	Right RegRef
}
