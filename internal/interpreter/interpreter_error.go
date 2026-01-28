package interpreter

import (
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/parser"
)

type InternalError struct {
	parser.Range
	Posting Posting
}

func (e InternalError) Error() string {
	return fmt.Sprintf("The script produced a posting with invalid values: %v", e.Posting)
}

type MissingFundsErr struct {
	parser.Range
	Asset     string
	Needed    big.Int
	Available big.Int
}

func (e MissingFundsErr) Error() string {
	return fmt.Sprintf("Not enough funds. Needed [%s %s] (only [%s %s] available)", e.Asset, e.Needed.String(), e.Asset, e.Available.String())
}

func (e MissingFundsErr) Is(target error) bool {
	_, ok := target.(MissingFundsErr)
	return ok
}

type InvalidMonetaryLiteral struct {
	parser.Range
	Source string
}

func (e InvalidMonetaryLiteral) Error() string {
	return fmt.Sprintf("invalid monetary literal: '%s'", e.Source)
}

type InvalidNumberLiteral struct {
	parser.Range
	Source string
}

func (e InvalidNumberLiteral) Error() string {
	return fmt.Sprintf("invalid number literal: '%s'", e.Source)
}

type MetadataNotFound struct {
	parser.Range
	Account string
	Key     string
}

func (e MetadataNotFound) Error() string {
	return fmt.Sprintf("account '@%s' doesn't have metadata associated to the '%s' key", e.Account, e.Key)
}

type TypeError struct {
	parser.Range
	Expected string
	Value    Value
}

func (e TypeError) Error() string {
	return fmt.Sprintf("Invalid value received. Expecting value of type %s (got %s instead)", e.Expected, e.Value.String())
}

type UnboundVariableErr struct {
	parser.Range
	Name string
}

func (e UnboundVariableErr) Error() string {
	return fmt.Sprintf("Unbound variable: $%s", e.Name)
}

type BadPortionParsingErr struct {
	parser.Range
	Source string
	Reason string
}

func (e BadPortionParsingErr) Error() string {
	return fmt.Sprintf("Bad portion: %s", e.Reason)
}

type MissingVariableErr struct {
	parser.Range
	Name string
}

func (e MissingVariableErr) Error() string {
	return fmt.Sprintf("Variable is missing in json: %s", e.Name)
}

type UnboundFunctionErr struct {
	parser.Range
	Name string
}

func (e UnboundFunctionErr) Error() string {
	return fmt.Sprintf("Invalid function: %s", e.Name)
}

type BadArityErr struct {
	parser.Range
	ExpectedArity  int
	GivenArguments int
}

func (e BadArityErr) Error() string {
	return fmt.Sprintf("Bad arity: expected %d arguments (got %d instead)", e.ExpectedArity, e.GivenArguments)
}

type InvalidTypeErr struct {
	parser.Range
	Name string
}

func (e InvalidTypeErr) Error() string {
	return fmt.Sprintf("This type does not exist: %s", e.Name)
}

type NegativeBalanceError struct {
	parser.Range
	Account string
	Amount  big.Int
}

func (e NegativeBalanceError) Error() string {
	return fmt.Sprintf("Cannot fetch negative balance from account @%s", e.Account)
}

type NegativeAmountErr struct {
	parser.Range
	Amount MonetaryInt
}

func (e NegativeAmountErr) Error() string {
	return fmt.Sprintf("Cannot send negative amount: %s", e.Amount.String())
}

type InvalidAllotmentInSendAll struct {
	parser.Range
}

func (e InvalidAllotmentInSendAll) Error() string {
	return "cannot take all balance of an allotment source"
}

type DivideByZero struct {
	parser.Range
	Numerator *big.Int
}

func (e DivideByZero) Error() string {
	return fmt.Sprintf("cannot divide by zero (in %s/0)", e.Numerator.String())
}

type InvalidUnboundedInSendAll struct {
	parser.Range
	Name string
}

func (e InvalidUnboundedInSendAll) Error() string {
	return "cannot take all balance from an unbounded source"
}

type MismatchedCurrencyError struct {
	parser.Range
	Expected string
	Got      string
}

func (e MismatchedCurrencyError) Error() string {
	return fmt.Sprintf("Mismatched currency (expected '%s', got '%s' instead)", e.Expected, e.Got)
}

type InvalidAllotmentSum struct {
	parser.Range
	ActualSum big.Rat
}

func (e InvalidAllotmentSum) Error() string {
	return fmt.Sprintf("Invalid allotment: portions sum should be 1 (got %s instead)", e.ActualSum.String())
}

type QueryBalanceError struct {
	parser.Range
	WrappedError error
}

func (e QueryBalanceError) Error() string {
	return e.WrappedError.Error()
}

type QueryMetadataError struct {
	parser.Range
	WrappedError error
}

func (e QueryMetadataError) Error() string {
	return e.WrappedError.Error()
}

type ExperimentalFeature struct {
	parser.Range
	FlagName string
}

func (e ExperimentalFeature) Error() string {
	return fmt.Sprintf("this feature is experimental. You need the '%s' feature flag to enable it", e.FlagName)
}

type CannotCastToString struct {
	parser.Range
	Value Value
}

func (e CannotCastToString) Error() string {
	return fmt.Sprintf("Cannot cast this value to string: %s", e.Value)
}

type InvalidAccountName struct {
	parser.Range
	Name string
}

func (e InvalidAccountName) Error() string {
	return fmt.Sprintf("Invalid account name: @%s", e.Name)
}

type InvalidAsset struct {
	parser.Range
	Name string
}

func (e InvalidAsset) Error() string {
	return fmt.Sprintf("Invalid asset name: %s", e.Name)
}

type InvalidNestedMeta struct {
	parser.Range
}

func (InvalidNestedMeta) Error() string {
	return "Invalid usage of meta() function: the meta function cannot be nested in a sub-expression."
}

type InvalidColor struct {
	parser.Range
	Color string
}

func (e InvalidColor) Error() string {
	return fmt.Sprintf("Invalid color name: '%s'. Only uppercase letters are allowed.", e.Color)
}
