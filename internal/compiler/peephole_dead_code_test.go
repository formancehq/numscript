package compiler

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeadCode_RemovesUnusedPure(t *testing.T) {
	instrs := []vInstr{
		loadInt{dest: 0, value: *big.NewInt(5)}, // unused -> removed
		loadInt{dest: 1, value: *big.NewInt(7)}, // used below -> kept
		checkEnoughFunds{got: 1, needed: 1},
	}
	out, changed := deadCode{}.run(instrs)
	require.True(t, changed)
	require.Len(t, out, 2)
	// the survivor load is r1
	li, ok := out[0].(loadInt)
	require.True(t, ok)
	require.Equal(t, reg(1), li.dest)
}

func TestDeadCode_KeepsImpureWithUnusedDest(t *testing.T) {
	// pull_account's dest is unused, but the pull has side effects (debit +
	// queue), so it must be kept.
	cap := reg(1)
	instrs := []vInstr{
		loadStr{dest: 0, value: "a"},
		loadInt{dest: 1, value: *big.NewInt(10)},
		pullAccount{dest: 9, account: 0, cap: &cap}, // dest r9 unused
	}
	out, changed := deadCode{}.run(instrs)
	require.False(t, changed)
	require.Len(t, out, 3)
}

func TestDeadCode_NoOpWhenAllUsed(t *testing.T) {
	instrs := []vInstr{
		loadInt{dest: 0, value: *big.NewInt(1)},
		checkEnoughFunds{got: 0, needed: 0},
	}
	_, changed := deadCode{}.run(instrs)
	require.False(t, changed)
}
