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
