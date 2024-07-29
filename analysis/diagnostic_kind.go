package analysis

import (
	"fmt"
	"math/big"
	"numscript/utils"
)

type Severity = byte

// !important! keep in sync with LSP specs
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#diagnosticSeverity
const (
	_ Severity = iota
	ErrorSeverity
	WarningSeverity
	Information
	Hint
)

func SeverityToAnsiString(s Severity) string {
	switch s {
	case ErrorSeverity:
		return "\033[31mError\033[0m"
	case WarningSeverity:
		return "\033[33mWarning\033[0m"
	case Information:
		return "Info"
	case Hint:
		return "Hint"
	default:
		return utils.NonExhaustiveMatchPanic[string](s)
	}
}

type DiagnosticKind interface {
	Message() string
	Severity() Severity
}

// ###### Diagnostics

type Parsing struct {
	Description string
}

func (e *Parsing) Message() string {
	return e.Description
}

func (*Parsing) Severity() Severity {
	return ErrorSeverity
}

type InvalidType struct {
	Name string
}

// TODO evaluate suggestion using Levenshtein distance
func (e *InvalidType) Message() string {
	allowedTypeList := ""
	for index, t := range AllowedTypes {
		if index != 0 {
			allowedTypeList += ", "
		}
		allowedTypeList += t
	}

	return fmt.Sprintf("'%s' is not a valid type. Allowed types are: %s", e.Name, allowedTypeList)
}

func (*InvalidType) Severity() Severity {
	return ErrorSeverity
}

type DuplicateVariable struct {
	Name string
}

func (e *DuplicateVariable) Message() string {
	return fmt.Sprintf("A variable with the name '$%s' was already declared", e.Name)
}

func (*DuplicateVariable) Severity() Severity {
	return ErrorSeverity
}

type UnboundVariable struct {
	Name string
}

// TODO evaluate suggestion using Levenshtein distance
func (e *UnboundVariable) Message() string {
	return fmt.Sprintf("The variable '$%s' was not declared", e.Name)
}

func (*UnboundVariable) Severity() Severity {
	return ErrorSeverity
}

type UnusedVar struct {
	Name string
}

func (e *UnusedVar) Message() string {
	return fmt.Sprintf("The variable '$%s' is never used", e.Name)
}

func (*UnusedVar) Severity() Severity {
	return WarningSeverity
}

type TypeMismatch struct {
	Expected string
	Got      string
}

func (e *TypeMismatch) Message() string {
	return fmt.Sprintf("Type mismatch (expected '%s', got '%s' instead)", e.Expected, e.Got)
}

func (*TypeMismatch) Severity() Severity {
	return ErrorSeverity
}

type RemainingIsNotLast struct{}

func (e *RemainingIsNotLast) Message() string {
	return "A 'remaining' clause should be the last in an allotment expression"
}
func (*RemainingIsNotLast) Severity() Severity {
	return ErrorSeverity
}

type BadAllotmentSum struct {
	Sum big.Rat
}

func (e *BadAllotmentSum) Message() string {
	one := big.NewRat(1, 1)

	switch e.Sum.Cmp(one) {
	// sum > 1
	case 1:
		return fmt.Sprintf("Allotment portions are greater than one (Got %s)", e.Sum.String())

	// sum < 1
	case -1:
		return fmt.Sprintf("Allotment portions are lesser than one (Got %s). Maybe try adding a 'remaining' clause?", e.Sum.String())
	}

	panic(fmt.Sprintf("unreachable state: allotment=%s", e.Sum.String()))
}
func (*BadAllotmentSum) Severity() Severity {
	return ErrorSeverity
}

type FixedPortionVariable struct {
	Value big.Rat
}

func (e *FixedPortionVariable) Message() string {
	return fmt.Sprintf("This variable always has the same value (%s)", e.Value.String())
}
func (*FixedPortionVariable) Severity() Severity {
	return WarningSeverity
}

type RedundantRemaining struct{}

func (e *RedundantRemaining) Message() string {
	return "Redundant 'remaining' clause (allotment already sums to 1)"
}
func (*RedundantRemaining) Severity() Severity {
	return WarningSeverity
}

type UnknownFunction struct {
	Name string
}

func (e *UnknownFunction) Message() string {
	res, exists := Builtins[e.Name]
	if exists {
		return fmt.Sprintf("You cannot use this function here (try to use it in a %s context)", res.ContextName())
	}
	// TODO suggest alternatives using Levenshtein distance
	return fmt.Sprintf("The function '%s' does not exist", e.Name)
}

func (*UnknownFunction) Severity() Severity {
	return ErrorSeverity
}

type BadArity struct {
	Expected int
	Actual   int
}

func (e *BadArity) Message() string {
	return fmt.Sprintf("Wrong number of arguments (expected %d, got %d instead)", e.Expected, e.Actual)
}

func (*BadArity) Severity() Severity {
	return ErrorSeverity
}
