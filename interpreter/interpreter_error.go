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
