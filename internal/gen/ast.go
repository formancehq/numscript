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

// An "AsVar" field marks one value occurrence as written through a `vars {}`
// variable instead of inline. It is per occurrence, not per value: the same
// account address or asset can appear inline in one place and as a var in
// another within the same script. That is what makes two distinct resources
// resolve to the same account — the aliasing behind formancehq/ledger#2056,
// which this generator previously could not produce because every account it
// emitted was a literal.
//
// Only the rendering changes, never the value, so passes that reason about
// addresses (cleanup.go) are unaffected.
type Monetary struct {
	Asset string
	// Render Asset through a var rather than inline.
	AssetAsVar bool
	// Render the whole monetary through a var (`$monetary_N`) rather than as a
	// `[asset amount]` literal. Ignored for negative amounts, which have no
	// literal form to bind (see toBuilderMonetary).
	AsVar  bool
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
	// Render Account through a var rather than inline. See AsVar.
	AccountAsVar bool

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
	// Render Portion through a var rather than inline.
	PortionAsVar bool
	Source       Source
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
	// Render Account through a var rather than inline. See AsVar.
	AccountAsVar bool

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
	Portion *big.Rat
	// Render Portion through a var rather than inline.
	PortionAsVar bool
	KeptOrDest   KeptOrDest
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
	IsSendAll bool
	Amount    Monetary
	Asset     string
	// IsSendAll only: render Asset through a var rather than inline.
	AssetAsVar  bool
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
	// NumLit only: render Lit through a var rather than inline.
	LitAsVar bool

	// NumAdd, NumSub only
	Left, Right *NumExpr
}

// VarDeclKind is the origin of a `vars {}` declaration this generator can
// produce: either `monetary $name = balance(<account>, <asset>)` or
// `number $name = meta(<account>, "<key>")`.
type VarDeclKind int

const (
	VarFromBalance VarDeclKind = iota
	VarFromMeta
)

// VarDecl is a `vars {}` declaration whose value comes from the compiler,
// not a runtime-supplied binding. See builder.NewMonetaryVarFromBalance /
// builder.NewNumberVarFromMeta and internal/oracle's VisitVars
// (OriginAccountBalanceContext/OriginAccountMetaContext).
type VarDecl struct {
	Kind    VarDeclKind
	Account string
	// Render Account through a var rather than inline, so that
	// `balance($accountN, ...)` and a literal `@acc` elsewhere in the script
	// become two resources aliasing one account.
	AccountAsVar bool

	// VarFromBalance only
	Asset string
	// Render Asset through a var rather than inline.
	AssetAsVar bool

	// VarFromMeta only
	Key string
}

// BalanceKey identifies one (account, asset) starting balance.
type BalanceKey struct {
	Account string
	Asset   string
}

// MetaKey identifies one (account, key) starting metadata entry.
type MetaKey struct {
	Account string
	Key     string
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
	ExtraSetTxMetaVar
	ExtraSendFromAccountVar
	ExtraSendToAccountVar
)

// AccountVarDecl is a plain (runtime-fed) account-typed `vars {}`
// declaration — unlike VarDecl, its value isn't computed by the compiler
// from an origin call; it's a literal account address supplied via the vars
// binding map at run time, exactly like a real caller's `$var` would be.
// Value is occasionally "world" specifically to exercise an account
// variable that resolves to @world in source position — a real, understood
// asymmetry between the two engines (the oracle rejects it in
// ResolveBalances with "`@world` can only be used as a variable in the
// experimental interpreter, or if it is never used as a source"; the new
// interpreter accepts it transparently, same as a literal @world). Compare
// already tolerates that specific direction (oracle rejects, new interpreter
// accepts) as expected, not a mismatch.
//
// This is distinct from an account occurrence with AccountAsVar set: that
// routes an ordinary account reference through the pooled vars mechanism,
// whereas this declares a var whose bound value is itself interesting.
type AccountVarDecl struct {
	Value string
}

// ExtraStatement is one non-send statement. Only the fields relevant to Kind
// are populated.
type ExtraStatement struct {
	Kind ExtraStatementKind

	// ExtraSave: the amount to save, either a literal (Monetary) or a
	// reference to Script.Vars[*VarIdx] (a declared VarFromBalance var) —
	// exactly one of the two is set. ExtraSendFromAccountVar,
	// ExtraSendToAccountVar: the send amount (always a literal).
	Monetary *Monetary

	// ExtraSave, ExtraSendVar (must index a VarFromBalance decl),
	// ExtraSetTxMetaVar (must index a VarFromMeta decl)
	VarIdx *int

	// ExtraSendFromAccountVar, ExtraSendToAccountVar: index into
	// Script.AccountVars for the account-typed var used as source/
	// destination (respectively); the other side is the literal Account.
	AccountVarIdx *int

	// ExtraSave (source account), ExtraSetAccountMeta, ExtraSendVar (source),
	// ExtraSendFromAccountVar (destination), ExtraSendToAccountVar (source)
	Account string
	// Render Account through a var rather than inline.
	AccountAsVar bool

	// ExtraSaveAll
	Asset string
	// Render Asset through a var rather than inline.
	AssetAsVar bool

	// ExtraSetTxMeta, ExtraSetAccountMeta, ExtraSetTxMetaVar
	Key   string
	Value NumExpr
	// ExtraSetTxMeta, ExtraSetAccountMeta: when non-nil the meta value is this
	// string rather than the Value expression, exercising string-typed values.
	StringValue *string
	// Render StringValue through a var rather than inline.
	StringValueAsVar bool

	// ExtraSendVar: `send $<Vars[VarIdx]> (source = Account, destination = Destination)`
	Destination string
	// Render Destination through a var rather than inline.
	DestinationAsVar bool
}

// Script is the full output of one round of generation: a `vars {}` block,
// optional seed-funding statements, the core send-only program, extra
// non-send statements, and/or pre-set starting balances/metadata (populated
// instead of, or alongside, Seeds — see genBalances). Order records a
// riffle-interleaving of Program and Extra (true = take next from Program,
// false = take next from Extra), so extra statements land throughout the
// script instead of only after every send.
type Script struct {
	Vars        []VarDecl
	AccountVars []AccountVarDecl
	Seeds       Program
	Program     Program
	Extra       []ExtraStatement
	Order       []bool
	Balances    map[BalanceKey]*big.Int
	Metadata    map[MetaKey]string
}
