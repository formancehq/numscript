package analysis

import (
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/ansi"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

type Severity = byte

// !important! keep in sync with LSP specs
// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#diagnosticSeverity
const (
	_ Severity = iota
	ErrorSeverity
	WarningSeverity
	InformationSeverity
	HintSeverity
)

func SeverityToAnsiString(s Severity) string {
	switch s {
	case ErrorSeverity:
		return ansi.ColorRed("Error")
	case WarningSeverity:
		return ansi.ColorYellow("Warning")
	case InformationSeverity:
		return "Info"
	case HintSeverity:
		return "Hint"
	default:
		return utils.NonExhaustiveMatchPanic[string](s)
	}
}

func SeverityToString(s Severity) string {
	switch s {
	case ErrorSeverity:
		return "Error"
	case WarningSeverity:
		return "Warning"
	case InformationSeverity:
		return "Info"
	case HintSeverity:
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

func (e Parsing) Message() string {
	return e.Description
}

func (Parsing) Severity() Severity {
	return ErrorSeverity
}

type InvalidType struct {
	Name string
}

// TODO evaluate suggestion using Levenshtein distance
func (e InvalidType) Message() string {
	allowedTypeList := ""
	for index, t := range AllowedTypes {
		if index != 0 {
			allowedTypeList += ", "
		}
		allowedTypeList += t
	}

	return fmt.Sprintf("'%s' is not a valid type. Allowed types are: %s", e.Name, allowedTypeList)
}

func (InvalidType) Severity() Severity {
	return ErrorSeverity
}

type DuplicateVariable struct {
	Name string
}

func (e DuplicateVariable) Message() string {
	return fmt.Sprintf("A variable with the name '$%s' was already declared", e.Name)
}

func (DuplicateVariable) Severity() Severity {
	return ErrorSeverity
}

type UnboundVariable struct {
	Name string
	Type string
}

// TODO evaluate suggestion using Levenshtein distance
func (e UnboundVariable) Message() string {
	return fmt.Sprintf("The variable '$%s' was not declared", e.Name)
}

func (UnboundVariable) Severity() Severity {
	return ErrorSeverity
}

type UnusedVar struct {
	Name string
}

func (e UnusedVar) Message() string {
	return fmt.Sprintf("The variable '$%s' is never used", e.Name)
}

func (UnusedVar) Severity() Severity {
	return WarningSeverity
}

type TypeMismatch struct {
	Expected string
	Got      string
}

func (e TypeMismatch) Message() string {
	return fmt.Sprintf("Type mismatch (expected '%s', got '%s' instead)", e.Expected, e.Got)
}

func (TypeMismatch) Severity() Severity {
	return ErrorSeverity
}

type AssetMismatch struct {
	Expected string
	Got      string
}

func (e AssetMismatch) Message() string {
	return fmt.Sprintf("Asset mismatch (expected '%s', got '%s' instead)", e.Expected, e.Got)
}

func (AssetMismatch) Severity() Severity {
	return ErrorSeverity
}

type RemainingIsNotLast struct{}

func (e RemainingIsNotLast) Message() string {
	return "A 'remaining' clause should be the last in an allotment expression"
}
func (RemainingIsNotLast) Severity() Severity {
	return ErrorSeverity
}

type BadAllotmentSum struct {
	Sum big.Rat
}

func (e BadAllotmentSum) Message() string {
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
func (BadAllotmentSum) Severity() Severity {
	return ErrorSeverity
}

type DivByZero struct {
	Sum big.Rat
}

func (e DivByZero) Message() string {
	return "Cannot divide by zero"
}

func (DivByZero) Severity() Severity {
	return ErrorSeverity
}

type FixedPortionVariable struct {
	Value big.Rat
}

func (e FixedPortionVariable) Message() string {
	return fmt.Sprintf("Using a variable expression can lead to a runtime error if the expression doesn't resolve to %s.\n\nConsider using a hard-coded value or adding a 'remaining' clause to prevent the error", e.Value.String())
}
func (FixedPortionVariable) Severity() Severity {
	return WarningSeverity
}

type RedundantRemaining struct{}

func (e RedundantRemaining) Message() string {
	return "Redundant 'remaining' clause (allotment already sums to 1)"
}
func (RedundantRemaining) Severity() Severity {
	return WarningSeverity
}

type UnknownFunction struct {
	Name string
}

func (e UnknownFunction) Message() string {
	res, exists := Builtins[e.Name]
	if exists {
		return fmt.Sprintf("You cannot use this function here (try to use it in a %s context)", res.ContextName())
	}
	// TODO suggest alternatives using Levenshtein distance
	return fmt.Sprintf("The function '%s' does not exist", e.Name)
}

func (UnknownFunction) Severity() Severity {
	return ErrorSeverity
}

type BadArity struct {
	Expected int
	Actual   int
}

func (e BadArity) Message() string {
	return fmt.Sprintf("Wrong number of arguments (expected %d, got %d instead)", e.Expected, e.Actual)
}

func (BadArity) Severity() Severity {
	return ErrorSeverity
}

type InvalidWorldOverdraft struct{}

func (e InvalidWorldOverdraft) Message() string {
	return "@world is already set to be ovedraft"
}

func (InvalidWorldOverdraft) Severity() Severity {
	return WarningSeverity
}

type NoAllotmentInSendAll struct{}

func (e NoAllotmentInSendAll) Message() string {
	return "Cannot take all balance of an allotment source"
}

func (NoAllotmentInSendAll) Severity() Severity {
	return WarningSeverity
}

type InvalidUnboundedAccount struct{}

func (e InvalidUnboundedAccount) Message() string {
	return "Cannot take all balance of an unbounded source"
}

func (InvalidUnboundedAccount) Severity() Severity {
	return ErrorSeverity
}

type EmptiedAccount struct {
	Name string
}

func (e EmptiedAccount) Message() string {
	return fmt.Sprintf("@%s is already empty at this point", e.Name)
}

func (EmptiedAccount) Severity() Severity {
	return WarningSeverity
}

type UnboundedAccountIsNotLast struct{}

func (e UnboundedAccountIsNotLast) Message() string {
	return "Inorder sources after an unbounded overdraft are never reached"
}

func (UnboundedAccountIsNotLast) Severity() Severity {
	return WarningSeverity
}

type VersionMismatch struct {
	RequiredVersion parser.Version
	GotVersion      parser.Version
}

func (e VersionMismatch) Message() string {
	return fmt.Sprintf("Version mismatch. Required version '%s' (but got '%s' instead)", e.RequiredVersion.String(), e.GotVersion.String())
}

func (VersionMismatch) Severity() Severity {
	return ErrorSeverity
}

type ExperimentalFeature struct {
	Name string
}

func (e ExperimentalFeature) Message() string {
	return fmt.Sprintf("This feature is experimental. Add this comment to your script:\n// @feature_flag %s", e.Name)
}

func (ExperimentalFeature) Severity() Severity {
	return ErrorSeverity
}

type BoundAssetType = uint

const (
	BoundAssetTypeMonetary BoundAssetType = iota
	BoundAssetTypeAsset
)

type BoundAsset struct {
	BoundAssetType BoundAssetType
	InferredAsset  string
}

func (e BoundAsset) Message() string {
	switch e.BoundAssetType {
	case BoundAssetTypeAsset:
		return fmt.Sprintf(`This asset is inferred to be '%s'. Receiving a different asset will cause a runtime error.
You may want to remove this variable and use the hardcoded value instead.`, e.InferredAsset)

	case BoundAssetTypeMonetary:
		return fmt.Sprintf(`This monetary is inferred to always have asset '%s'. Receiving a monetary of different asset will cause a runtime error.
You may want to use a variable of type number instead.`, e.InferredAsset)

	default:
		return utils.NonExhaustiveMatchPanic[string](e)
	}
}

func (BoundAsset) Severity() Severity {
	return WarningSeverity
}
