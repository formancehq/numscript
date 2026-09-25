package ir

import (
	"testing"

	"github.com/formancehq/numscript/internal/vm"
	"github.com/stretchr/testify/require"
)

// The mark ops carry no register, so there is nothing for the typechecker to get
// wrong — and, more to the point, no operand a caller could use to name a queue
// depth the run-state never marked. What does need checking about them (that pushes
// and ends balance, and that no send or asset change sits inside a region) is a
// control-flow property, left to run time until a verifier pass exists.
func TestMark_Typecheck(t *testing.T) {
	require.NoError(t, Typecheck([]Instr{
		MarkPush{},
		MarkEnd{Rewind: true},
		MarkPush{},
		MarkEnd{Rewind: false},
	}))

	// with no operand there is no such thing as an ill-typed mark op: even an
	// unbalanced stream is well-typed, and fails at run time instead
	require.NoError(t, Typecheck([]Instr{MarkEnd{Rewind: false}}))
	require.NoError(t, Typecheck([]Instr{
		LoadStr{Dest: 0, Value: "USD/2"},
		MarkEnd{Rewind: true},
	}))
}

// One instruction, two textual names — the same shape as
// assert_leftover / assert_leftover_exact.
func TestMark_Dump(t *testing.T) {
	out := Dump([]Instr{
		MarkPush{},
		MarkEnd{Rewind: true},
		MarkEnd{Rewind: false},
	})
	require.Contains(t, out, "mark_push()")
	require.Contains(t, out, "mark_rewind()")
	require.Contains(t, out, "mark_commit()")
}

func TestMark_Assemble(t *testing.T) {
	prog, err := Assemble([]Instr{
		MarkPush{},
		MarkEnd{Rewind: true},
		MarkEnd{Rewind: false},
	})
	require.NoError(t, err)
	// the rewind flag rides in A; there is no register to allocate
	require.Equal(t, []vm.Instruction{
		{Opcode: byte(vm.Op_MarkPush), A: 0xFF, B: 0xFF, C: 0xFF},
		{Opcode: byte(vm.Op_MarkEnd), A: 1, B: 0xFF, C: 0xFF},
		{Opcode: byte(vm.Op_MarkEnd), A: 0, B: 0xFF, C: 0xFF},
	}, prog.Instructions)

	// no register is consumed in any bank — the old Op_Snapshot spent a big.Int one
	require.Zero(t, prog.MaxRegInt)
}

func TestMark_ParseRoundTrip(t *testing.T) {
	instrs, errs := Parse(`
  mark_push()
  mark_rewind()
  mark_push()
  mark_commit()
`)
	require.Empty(t, errs)
	require.Equal(t, []Instr{
		MarkPush{},
		MarkEnd{Rewind: true},
		MarkPush{},
		MarkEnd{Rewind: false},
	}, instrs)

	// a Dump of the parsed stream parses back to the same program, so the two names
	// round-trip through the single instruction
	reparsed, errs := Parse(Dump(instrs))
	require.Empty(t, errs)
	require.Equal(t, instrs, reparsed)
}
