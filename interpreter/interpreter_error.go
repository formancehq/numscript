package interpreter

import (
	"fmt"
	"math/big"
)

type MissingFundsErr struct {
	Missing big.Int
	Sent    big.Int
}

func (e MissingFundsErr) Error() string {
	return fmt.Sprintf("Not enough funds. Missing %s (sent %s)", e.Missing.String(), e.Sent.String())
}

type TypeError struct {
	Expected string
	Value    Value
}

func (e TypeError) Error() string {
	return fmt.Sprintf("Invalid value received. Expecting value of type %s (got %s instead)", e.Expected, e.Value.String())
}

type UnboundVariableErr struct {
	Name string
}

func (e UnboundVariableErr) Error() string {
	return fmt.Sprintf("Unbound variable: %s", e.Name)
}

type MissingVariableErr struct {
	Name string
}

func (e MissingVariableErr) Error() string {
	return fmt.Sprintf("Variable is missing in json: %s", e.Name)
}

type UnboundFunctionErr struct {
	Name string
}

func (e UnboundFunctionErr) Error() string {
	return fmt.Sprintf("Invalid function: %s", e.Name)
}

type BadArityErr struct {
	ExpectedArity  int
	GivenArguments int
}

func (e BadArityErr) Error() string {
	return fmt.Sprintf("Bad arity: expected %d arguments (got %d instead)", e.ExpectedArity, e.GivenArguments)
}

type InvalidTypeErr struct {
	Name string
}

func (e InvalidTypeErr) Error() string {
	return fmt.Sprintf("This type does not exist: %s", e.Name)
}

type NegativeBalanceError struct {
	Account string
	Amount  big.Int
}

func (e NegativeBalanceError) Error() string {
	return fmt.Sprintf("Cannot fetch negative balance from account @%s", e.Account)
}

type NegativeAmountErr struct{ Amount MonetaryInt }

func (e NegativeAmountErr) Error() string {
	return fmt.Sprintf("Cannot send negative amount: %s", e.Amount.String())
}

type InvalidAllotmentInSendAll struct {
}

func (e InvalidAllotmentInSendAll) Error() string {
	return "cannot take all balance of an allotment source"
}

type InvalidUnboundedInSendAll struct{ Name string }

func (e InvalidUnboundedInSendAll) Error() string {
	return "cannot take all balance from an unbounded source"
}

type MismatchedCurrencyError struct {
	Expected string
	Got      string
}

func (e MismatchedCurrencyError) Error() string {
	return fmt.Sprintf("Mismatched currency (expected '%s', got '%s' instead)", e.Expected, e.Got)
}
