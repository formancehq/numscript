package parser

import (
	"math/big"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/stretchr/testify/require"
)

// These inputs should be unreachable through Parse() (the lexer only emits
// well-formed tokens), but the parsing helpers must degrade gracefully to a
// zero-value literal instead of panicking and killing the process.

func TestParsePercentageRatioInvalidSourceDoesNotPanic(t *testing.T) {
	require.NotPanics(t, func() {
		lit := parsePercentageRatio("not a number%", Range{})
		require.Equal(t, big.NewInt(0), lit.Amount)
		require.Equal(t, 0, lit.FloatingDigits)
	})
}

func TestParseNumberLiteralInvalidNumberDoesNotPanic(t *testing.T) {
	tk := antlr.NewCommonToken(&antlr.TokenSourceCharStreamPair{}, 0, antlr.TokenDefaultChannel, 0, 0)
	tk.SetText("_")
	node := antlr.NewTerminalNodeImpl(tk)

	require.NotPanics(t, func() {
		lit := parseNumberLiteral(node)
		require.Equal(t, big.NewInt(0), lit.Number)
	})
}
