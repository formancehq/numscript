package parser_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestPercentageLiteralToRatio(t *testing.T) {
	assert.Equal(t, big.NewRat(123, 100), parser.PercentageLiteral{
		Amount:         big.NewInt(123),
		FloatingDigits: 0,
	}.ToRatio())

	assert.Equal(t, big.NewRat(123, 1000), parser.PercentageLiteral{
		Amount:         big.NewInt(123),
		FloatingDigits: 1,
	}.ToRatio())

	assert.Equal(t, big.NewRat(123, 10000), parser.PercentageLiteral{
		Amount:         big.NewInt(123),
		FloatingDigits: 2,
	}.ToRatio())
}
