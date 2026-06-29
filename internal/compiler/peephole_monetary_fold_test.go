package compiler

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// monetaryProgram is the canonical `send [USD 10]` preamble shape.
func monetaryProgram() []vInstr {
	return []vInstr{
		loadStr{dest: 0, value: "USD"},
		loadInt{dest: 1, value: *big.NewInt(10)},
		binaryOp{op: opMakeMonetary{}, dest: 2, left: 0, right: 1},
		unaryOp{op: opGetAsset{}, dest: 3, arg: 2},
		setCurrentAsset{asset: 3},
		unaryOp{op: opGetAmount{}, dest: 4, arg: 2},
		checkEnoughFunds{got: 4, needed: 4},
	}
}

func TestMonetaryFold(t *testing.T) {
	out, changed := monetaryFold{}.run(monetaryProgram())
	require.True(t, changed)

	// get_asset/get_amount are dropped; consumers now read the asset/amount
	// registers directly. mk_monetary is left for the dead-code pass.
	require.Equal(t, strings.TrimSpace(`
  $r0 <- load_const("USD")
  $r1 <- load_const(10)
  $r2 <- mk_monetary($r0, $r1)
  set_current_asset($r0)
  check_enough_funds($r1, $r1)`), strings.TrimSpace(dump(out)))
}

func TestMonetaryFold_NoOpWhenNoMonetary(t *testing.T) {
	instrs := []vInstr{
		loadInt{dest: 0, value: *big.NewInt(1)},
		checkEnoughFunds{got: 0, needed: 0},
	}
	_, changed := monetaryFold{}.run(instrs)
	require.False(t, changed)
}

func TestMonetaryFold_SkipsReassignedRegisters(t *testing.T) {
	// r2 (the monetary) is reassigned, so it is NOT single-def -> don't fold.
	instrs := []vInstr{
		loadStr{dest: 0, value: "USD"},
		loadInt{dest: 1, value: *big.NewInt(10)},
		binaryOp{op: opMakeMonetary{}, dest: 2, left: 0, right: 1},
		binaryOp{op: opMakeMonetary{}, dest: 2, left: 0, right: 1}, // 2nd def of r2
		unaryOp{op: opGetAsset{}, dest: 3, arg: 2},
	}
	_, changed := monetaryFold{}.run(instrs)
	require.False(t, changed)
}
