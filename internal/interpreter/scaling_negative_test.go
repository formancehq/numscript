package interpreter_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

// A scaling solution can include a negative conversion when the source holds a
// negative balance at one scale. Posting only its positive legs would debit the
// source without the compensating credit: here acc1 would give up EUR/1 20
// (2 EUR) to move 1 EUR, with the -100 EUR/2 leg silently discarded, and the
// swap account would keep the difference.
//
// The engine rejects the negative conversion instead, which surfaces as the
// same invalid-posting error the script produced before the funds engine
// generated these postings.
func TestScalingNegativeConversionIsRejected(t *testing.T) {
	parsed := parser.Parse(`send [EUR 1] (
  source = @acc1 with scaling through @swap
  destination = @dest
)`)
	require.Empty(t, parsed.Errors)

	store := interpreter.StaticStore{Balances: interpreter.Balances{
		{Account: "acc1", Asset: "EUR/2", Amount: big.NewInt(-100)},
		{Account: "acc1", Asset: "EUR/1", Amount: big.NewInt(20)},
	}}

	_, err := interpreter.RunProgram(
		context.Background(), parsed.Value, nil, store,
		map[string]struct{}{flags.AssetScaling: {}},
	)
	require.NotNil(t, err, "expected the script to fail rather than overpay")
	require.IsType(t, interpreter.InternalError{}, err)
}
