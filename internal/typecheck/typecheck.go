// Package typecheck is the shared, side-effect-free type checker for numscript.
// It synthesizes the (base) type of every expression, resolves variable types
// from their declarations, and reports type/name/arity errors — fault-tolerantly
// (it collects all errors instead of bailing, using TypeAny to avoid cascades).
//
// It is meant to be the single source of truth for typing, consumed both by the
// analysis module (LSP/CI) and by the compiler. It intentionally does NOT do
// asset-identity inference, feature-version gating, or lint-style warnings —
// those stay in the analysis module.
package typecheck

import (
	"fmt"
	"slices"
	"strings"

	"github.com/formancehq/numscript/internal/builtins"
	"github.com/formancehq/numscript/internal/parser"
)

type Type = string

const (
	TypeNumber   Type = "number"
	TypeString   Type = "string"
	TypeAsset    Type = "asset"
	TypeMonetary Type = "monetary"
	TypeAccount  Type = "account"
	TypePortion  Type = "portion"

	// TypeAny is the type of an expression whose type couldn't be determined
	// (e.g. an unbound variable or an unknown function). It's compatible with
	// everything, so it suppresses cascading errors.
	TypeAny Type = "any"
)

// order mirrors analysis.AllowedTypes so the InvalidType message reads identically
var allowedTypes = []Type{TypeMonetary, TypeAccount, TypePortion, TypeAsset, TypeNumber, TypeString}

func isTypeAllowed(t string) bool { return slices.Contains(allowedTypes, t) }

// --- errors

// Severity mirrors the LSP DiagnosticSeverity spec (and analysis.Severity, an
// alias of byte too) so an ErrorKind directly satisfies analysis.DiagnosticKind.
type Severity = byte

const severityError Severity = 1

// ErrorKind is both a typecheck error and a renderable diagnostic (Message +
// Severity), so callers can push it as a diagnostic without a translation layer.
type ErrorKind interface {
	errorKind()
	Message() string
	Severity() Severity
}

type (
	TypeMismatch    struct{ Expected, Got string }
	UnboundVariable struct{ Name, Type string }
	InvalidType     struct{ Name string }
	BadArity        struct{ Expected, Actual int }
	// UnknownFunction is either a truly-unknown name (WrongContext == "") or a
	// known builtin used in the wrong context (WrongContext is the context it
	// belongs to, e.g. "statement"). typecheck only ever emits the former.
	UnknownFunction   struct{ Name, WrongContext string }
	DuplicateVariable struct{ Name string }
)

func (TypeMismatch) errorKind()      {}
func (UnboundVariable) errorKind()   {}
func (InvalidType) errorKind()       {}
func (BadArity) errorKind()          {}
func (UnknownFunction) errorKind()   {}
func (DuplicateVariable) errorKind() {}

func (e TypeMismatch) Message() string {
	return fmt.Sprintf("Type mismatch (expected '%s', got '%s' instead)", e.Expected, e.Got)
}

func (e UnboundVariable) Message() string {
	return fmt.Sprintf("The variable '$%s' was not declared", e.Name)
}

func (e InvalidType) Message() string {
	return fmt.Sprintf("'%s' is not a valid type. Allowed types are: %s", e.Name, strings.Join(allowedTypes, ", "))
}

func (e BadArity) Message() string {
	return fmt.Sprintf("Wrong number of arguments (expected %d, got %d instead)", e.Expected, e.Actual)
}

func (e UnknownFunction) Message() string {
	if e.WrongContext != "" {
		return fmt.Sprintf("You cannot use this function here (try to use it in a %s context)", e.WrongContext)
	}
	return fmt.Sprintf("The function '%s' does not exist", e.Name)
}

func (e DuplicateVariable) Message() string {
	return fmt.Sprintf("A variable with the name '$%s' was already declared", e.Name)
}

func (TypeMismatch) Severity() Severity      { return severityError }
func (UnboundVariable) Severity() Severity   { return severityError }
func (InvalidType) Severity() Severity       { return severityError }
func (BadArity) Severity() Severity          { return severityError }
func (UnknownFunction) Severity() Severity   { return severityError }
func (DuplicateVariable) Severity() Severity { return severityError }

type Error struct {
	Range parser.Range
	Kind  ErrorKind
}

// --- builtin function signatures

type fnSig struct {
	params []Type
	ret    Type // "" for statement functions (no return)
}

var builtinSigs = map[string]fnSig{
	builtins.SetTxMeta:      {params: []Type{TypeString, TypeAny}},
	builtins.SetAccountMeta: {params: []Type{TypeAccount, TypeString, TypeAny}},
	builtins.Meta:           {params: []Type{TypeAccount, TypeString}, ret: TypeAny},
	builtins.Balance:        {params: []Type{TypeAccount, TypeAsset}, ret: TypeMonetary},
	builtins.Overdraft:      {params: []Type{TypeAccount, TypeAsset}, ret: TypeMonetary},
	builtins.GetAsset:       {params: []Type{TypeMonetary}, ret: TypeAsset},
	builtins.GetAmount:      {params: []Type{TypeMonetary}, ret: TypeNumber},
}

// --- Result / entrypoint

type Result struct {
	ExprTypes map[parser.ValueExpr]Type
	VarTypes  map[string]Type
	Errors    []Error
}

func Check(program parser.Program) Result {
	c := checker{
		exprTypes: map[parser.ValueExpr]Type{},
		varTypes:  map[string]Type{},
		declared:  map[string]struct{}{},
	}
	c.checkProgram(program)
	return Result{ExprTypes: c.exprTypes, VarTypes: c.varTypes, Errors: c.errors}
}

type checker struct {
	exprTypes map[parser.ValueExpr]Type
	varTypes  map[string]Type
	declared  map[string]struct{}
	errors    []Error
}

func (c *checker) push(rng parser.Range, kind ErrorKind) {
	c.errors = append(c.errors, Error{Range: rng, Kind: kind})
}

func (c *checker) checkProgram(program parser.Program) {
	if program.Vars != nil {
		for _, varDecl := range program.Vars.Declarations {
			if varDecl.Type != nil && !isTypeAllowed(varDecl.Type.Name) {
				c.push(varDecl.Type.Range, InvalidType{Name: varDecl.Type.Name})
			}

			if varDecl.Name != nil {
				if _, dup := c.declared[varDecl.Name.Name]; dup {
					c.push(varDecl.Name.Range, DuplicateVariable{Name: varDecl.Name.Name})
				} else {
					c.declared[varDecl.Name.Name] = struct{}{}
					if varDecl.Type != nil && isTypeAllowed(varDecl.Type.Name) {
						c.varTypes[varDecl.Name.Name] = varDecl.Type.Name
					}
				}
			}

			if varDecl.Origin != nil && varDecl.Type != nil {
				c.checkExpr(*varDecl.Origin, varDecl.Type.Name)
			}
		}
	}

	for _, statement := range program.Statements {
		c.checkStatement(statement)
	}
}

func (c *checker) checkStatement(statement parser.Statement) {
	switch statement := statement.(type) {
	case *parser.SaveStatement:
		c.checkSentValue(statement.SentValue)
		c.checkExpr(statement.Account, TypeAccount)

	case *parser.SendStatement:
		c.checkSentValue(statement.SentValue)
		c.checkSource(statement.Source)
		c.checkDestination(statement.Destination)

	case *parser.FnCall:
		c.checkFnCallArity(statement)
	}
}

func (c *checker) checkSentValue(sentValue parser.SentValue) {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueAll:
		c.checkExpr(sentValue.Asset, TypeAsset)
	case *parser.SentValueLiteral:
		c.checkExpr(sentValue.Monetary, TypeMonetary)
	}
}

func (c *checker) checkSource(source parser.Source) {
	if source == nil {
		return
	}
	switch source := source.(type) {
	case *parser.SourceAccount:
		c.checkExpr(source.ValueExpr, TypeAccount)
		c.checkExpr(source.Color, TypeString)

	case *parser.SourceOverdraft:
		c.checkExpr(source.Address, TypeAccount)
		c.checkExpr(source.Color, TypeString)
		if source.Bounded != nil {
			c.checkExpr(*source.Bounded, TypeMonetary)
		}

	case *parser.SourceWithScaling:
		c.checkExpr(source.Address, TypeAccount)
		c.checkExpr(source.Through, TypeAccount)

	case *parser.SourceInorder:
		for _, sub := range source.Sources {
			c.checkSource(sub)
		}

	case *parser.SourceOneof:
		for _, sub := range source.Sources {
			c.checkSource(sub)
		}

	case *parser.SourceCapped:
		c.checkExpr(source.Cap, TypeMonetary)
		c.checkSource(source.From)

	case *parser.SourceAllotment:
		for _, item := range source.Items {
			if al, ok := item.Allotment.(*parser.ValueExprAllotment); ok {
				c.checkExpr(al.Value, TypePortion)
			}
			c.checkSource(item.From)
		}
	}
}

func (c *checker) checkDestination(destination parser.Destination) {
	if destination == nil {
		return
	}
	switch destination := destination.(type) {
	case *parser.DestinationAccount:
		c.checkExpr(destination.ValueExpr, TypeAccount)

	case *parser.DestinationInorder:
		for _, clause := range destination.Clauses {
			c.checkExpr(clause.Cap, TypeMonetary)
			c.checkKeptOrDestination(clause.To)
		}
		c.checkKeptOrDestination(destination.Remaining)

	case *parser.DestinationOneof:
		for _, clause := range destination.Clauses {
			c.checkExpr(clause.Cap, TypeMonetary)
			c.checkKeptOrDestination(clause.To)
		}
		c.checkKeptOrDestination(destination.Remaining)

	case *parser.DestinationAllotment:
		for _, item := range destination.Items {
			if al, ok := item.Allotment.(*parser.ValueExprAllotment); ok {
				c.checkExpr(al.Value, TypePortion)
			}
			c.checkKeptOrDestination(item.To)
		}
	}
}

func (c *checker) checkKeptOrDestination(keptOrDest parser.KeptOrDestination) {
	if dest, ok := keptOrDest.(*parser.DestinationTo); ok {
		c.checkDestination(dest.Destination)
	}
}

// checkExpr synthesizes lit's type, records it, and asserts it matches want.
func (c *checker) checkExpr(lit parser.ValueExpr, want Type) {
	got := c.synthType(lit, want)
	if want != TypeAny && got != TypeAny && want != got {
		c.push(lit.GetRange(), TypeMismatch{Expected: want, Got: got})
	}
}

// synthType synthesizes lit's type. hint is the type expected by the context; it
// is only used to annotate an unbound-variable error (matching the interpreter's
// diagnostic), never to influence the synthesized type.
func (c *checker) synthType(lit parser.ValueExpr, hint Type) Type {
	if lit == nil {
		return TypeAny
	}
	t := c.synthTypeInner(lit, hint)
	c.exprTypes[lit] = t
	return t
}

func (c *checker) synthTypeInner(lit parser.ValueExpr, hint Type) Type {
	switch lit := lit.(type) {
	case *parser.Variable:
		t, ok := c.varTypes[lit.Name]
		if !ok {
			if _, declared := c.declared[lit.Name]; !declared {
				c.push(lit.Range, UnboundVariable{Name: lit.Name, Type: hint})
			}
			return TypeAny
		}
		return t

	case *parser.MonetaryLiteral:
		c.checkExpr(lit.Asset, TypeAsset)
		c.checkExpr(lit.Amount, TypeNumber)
		return TypeMonetary

	case *parser.BinaryInfix:
		switch lit.Operator {
		case parser.InfixOperatorPlus, parser.InfixOperatorMinus:
			return c.checkInfixOverload(lit, []Type{TypeNumber, TypeMonetary})
		case parser.InfixOperatorDiv:
			c.checkExpr(lit.Left, TypeNumber)
			c.checkExpr(lit.Right, TypeNumber)
			return TypePortion
		default:
			c.checkExpr(lit.Left, TypeAny)
			c.checkExpr(lit.Right, TypeAny)
			return TypeAny
		}

	case *parser.Prefix:
		switch lit.Operator {
		case parser.PrefixOperatorMinus:
			return c.checkHasOneOfTypes(lit.Expr, []Type{TypeNumber, TypeMonetary})
		default:
			return TypeAny
		}

	case *parser.AccountInterpLiteral:
		for _, part := range lit.Parts {
			if v, ok := part.(*parser.Variable); ok {
				c.checkExpr(v, TypeAny)
			}
		}
		return TypeAccount

	case *parser.PercentageLiteral:
		return TypePortion
	case *parser.AssetLiteral:
		return TypeAsset
	case *parser.NumberLiteral:
		return TypeNumber
	case *parser.StringLiteral:
		return TypeString

	case *parser.FnCall:
		return c.checkFnCall(lit)

	default:
		return TypeAny
	}
}

func (c *checker) checkInfixOverload(bin *parser.BinaryInfix, allowed []Type) Type {
	leftType := c.synthType(bin.Left, allowed[0])
	if leftType == TypeAny || slices.Contains(allowed, leftType) {
		c.checkExpr(bin.Right, leftType)
		return leftType
	}
	c.push(bin.Left.GetRange(), TypeMismatch{Expected: strings.Join(allowed, "|"), Got: leftType})
	return TypeAny
}

func (c *checker) checkHasOneOfTypes(expr parser.ValueExpr, allowed []Type) Type {
	exprType := c.synthType(expr, allowed[0])
	if exprType == TypeAny || slices.Contains(allowed, exprType) {
		return exprType
	}
	c.push(expr.GetRange(), TypeMismatch{Expected: strings.Join(allowed, "|"), Got: exprType})
	return TypeAny
}

func (c *checker) checkFnCall(fnCall *parser.FnCall) Type {
	ret := TypeAny
	if sig, ok := builtinSigs[fnCall.Caller.Name]; ok {
		ret = sig.ret
		if ret == "" {
			ret = TypeAny
		}
	}
	c.checkFnCallArity(fnCall)
	return ret
}

func (c *checker) checkFnCallArity(fnCall *parser.FnCall) {
	var validArgs []parser.ValueExpr
	for _, arg := range fnCall.Args {
		if arg != nil {
			validArgs = append(validArgs, arg)
		}
	}

	sig, resolved := builtinSigs[fnCall.Caller.Name]
	if !resolved {
		for _, arg := range validArgs {
			c.checkExpr(arg, TypeAny)
		}
		c.push(fnCall.Caller.Range, UnknownFunction{Name: fnCall.Caller.Name})
		return
	}

	expected := len(sig.params)
	actual := len(validArgs)
	if actual < expected {
		c.push(fnCall.Range, BadArity{Expected: expected, Actual: actual})
	} else if actual > expected {
		first := validArgs[expected]
		last := validArgs[len(validArgs)-1]
		c.push(parser.Range{Start: first.GetRange().Start, End: last.GetRange().End},
			BadArity{Expected: expected, Actual: actual})
	}

	for i, arg := range validArgs {
		if i >= len(sig.params) {
			break
		}
		c.checkExpr(arg, sig.params[i])
	}
}
