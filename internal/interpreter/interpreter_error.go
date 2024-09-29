package interpreter

import (
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/parser"
)

type MissingFundsErr struct {
	parser.Range
	Asset     string
	Needed    big.Int
	Available big.Int
}

func (e MissingFundsErr) Error() string {
	return fmt.Sprintf("Not enough funds. Needed [%s %s] (only [%s %s] available)", e.Asset, e.Needed.String(), e.Asset, e.Available.String())
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
