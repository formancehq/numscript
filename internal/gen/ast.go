// Package gen is a Go port of numscript_gen's Haskell generator
// (Gen.hs/Numscript.hs/Utils.hs), used to produce random numscript programs
// for differential testing against the legacy ledger "machine" (see
// internal/oracle). It matches the Haskell generator's scope 1:1: only
// `send`/`send *` statements, no vars/print/save/set_tx_meta/conditionals.
//
// Generation happens against this package's own intermediate AST (mirroring
// Numscript.hs) rather than directly against builder/, because the cleanup
// pass (cleanup.go, porting Utils.hs) needs a real, rewritable tree to walk
// — builder's Source/Destination are opaque render closures. The final
// program is converted to builder.Statement values only once cleanup has
// settled (see convert.go).
package gen

import "math/big"

type Monetary struct {
	Asset  string
	Amount *big.Int
}

type SourceKind int

const (
	SrcAccount SourceKind = iota
	SrcAccountOverdraft
	SrcCapped
	SrcInorder
	SrcAllotment
)

type Source struct {
	Kind SourceKind

	// SrcAccount, SrcAccountOverdraft
	Account string

	// SrcAccountOverdraft only: nil means unbounded overdraft
	Overdraft *Monetary

	// SrcCapped only
	Cap   *Monetary
	Inner *Source

	// SrcInorder only
	Sources []Source

	// SrcAllotment only
	Clauses []SourceAllotmentClause
}

type SourceAllotmentClause struct {
	Portion *big.Rat
	Source  Source
}

type DestKind int

const (
	DestAccount DestKind = iota
	DestInorder
	DestAllotment
)

type Destination struct {
	Kind DestKind

	// DestAccount only
	Account string

	// DestInorder only
	InorderClauses []DestInorderClause
	Remaining      *KeptOrDest

	// DestAllotment only
	AllotClauses []DestAllotmentClause
}

type DestInorderClause struct {
	Max        Monetary
	KeptOrDest KeptOrDest
}

type DestAllotmentClause struct {
	Portion    *big.Rat
	KeptOrDest KeptOrDest
}

type KeptOrDestKind int

const (
	Kept KeptOrDestKind = iota
	To
)

type KeptOrDest struct {
	Kind KeptOrDestKind
	Dest *Destination // valid only when Kind == To
}

type Statement struct {
	// If IsSendAll, Asset is used (unbounded send). Otherwise Amount is used.
	IsSendAll   bool
	Amount      Monetary
	Asset       string
	Source      Source
	Destination Destination
}

type Program []Statement

// NumExprKind is a small arithmetic-expression AST used for
// set_tx_meta/set_account_meta values: literals combined with +/-, the only
// runtime operators the oracle's grammar supports (see NumScript.g4's
// `expression` rule: ExprAddSub over literals/variables, nothing else).
type NumExprKind int

const (
	NumLit NumExprKind = iota
	NumAdd
	NumSub
)

type NumExpr struct {
	Kind NumExprKind

	// NumLit only
	Lit *big.Int

	// NumAdd, NumSub only
	Left, Right *NumExpr
}

// VarDecl is a `vars {}` declaration whose value comes from the compiler
// (currently only `monetary $name = balance(<account>, <asset>)` — the only
// origin kind this generator produces; see builder.NewMonetaryVarFromBalance
// and internal/oracle's VisitVars/OriginAccountBalanceContext).
type VarDecl struct {
	Account string
	Asset   string
}

// BalanceKey identifies one (account, asset) starting balance.
type BalanceKey struct {
	Account string
	Asset   string
}

// ExtraStatementKind enumerates the non-send statement kinds this generator
// can produce, on top of Program's send-only statements.
type ExtraStatementKind int

const (
	ExtraSave ExtraStatementKind = iota
	ExtraSaveAll
	ExtraSetTxMeta
	ExtraSetAccountMeta
	ExtraSendVar
)

// ExtraStatement is one non-send statement. Only the fields relevant to Kind
// are populated.
type ExtraStatement struct {
	Kind ExtraStatementKind

	// ExtraSave: the amount to save, either a literal (Monetary) or a
	// reference to Script.Vars[*VarIdx] (a declared balance()-origin var) —
	// exactly one of the two is set.
	Monetary *Monetary
	VarIdx   *int

	// ExtraSave (source account), ExtraSetAccountMeta, ExtraSendVar (source)
	Account string

	// ExtraSaveAll
	Asset string

	// ExtraSetTxMeta, ExtraSetAccountMeta
	Key   string
	Value NumExpr

	// ExtraSendVar: `send $<Vars[VarIdx]> (source = Account, destination = Destination)`
	Destination string
}

// Script is the full output of one round of generation: a `vars {}` block,
// optional seed-funding statements, the core send-only program, extra
// non-send statements, and/or pre-set starting balances (populated instead
// of, or alongside, Seeds — see genBalances).
type Script struct {
	Vars     []VarDecl
	Seeds    Program
	Program  Program
	Extra    []ExtraStatement
	Balances map[BalanceKey]*big.Int
}
