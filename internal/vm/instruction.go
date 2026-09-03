package vm

import "encoding/binary"

type Instruction struct {
	Opcode byte
	A      byte
	B      byte
	C      byte
}

// Little endian view of the b and c fields
func (i Instruction) GetBC() uint16 {
	return uint16(i.B) | uint16(i.C)<<8
}

func NewBC(
	opcode Opcode,
	a byte,
	bc uint16,
) Instruction {
	var bcBytes [2]byte
	binary.LittleEndian.PutUint16(bcBytes[:], bc)

	return Instruction{
		Opcode: byte(opcode),
		A:      a,
		B:      bcBytes[0],
		C:      bcBytes[1],
	}
}

type Opcode byte

// Opcodes are grouped by category with gaps, so new instructions can be added to
// a category without renumbering. See instruction-encoding.md.
const (
	// --- state & assertions (0x00) ---
	Op_SetCurrentAsset Opcode = 0x00

	Op_AssertSameAsset Opcode = 0x01

	// errors if the account name in str reg A is not well-formed
	Op_AssertValidAccount Opcode = 0x02

	// errors (NegativeBalanceError) if the amount in int reg A is negative;
	// B = account str reg (for the error)
	Op_AssertNonNegativeBalance Opcode = 0x03

	// checks the allotment leftover portion in reg A: errors if negative (portions
	// summing to > 1), and — when B == 1 (no `remaining` clause) — if non-zero
	Op_AssertLeftover Opcode = 0x04

	Op_CheckEnoughFunds Opcode = 0x05

	// errors if the color in str reg A is not well-formed
	Op_AssertValidColor Opcode = 0x06

	// errors (NegativeAmountError) if the amount in int reg A is negative —
	// a sent/saved amount, not tied to any account (unlike
	// Op_AssertNonNegativeBalance, which is a balance() read)
	Op_AssertNonNegativeAmount Opcode = 0x07

	// --- constants & variables (0x10) ---
	// may split into one opcode per expr_typ later
	Op_LoadInt Opcode = 0x10 // LoadConst (`Int)    -> b_c = const-pool index
	Op_LoadStr Opcode = 0x11 // LoadConst (`String) -> b_c = const-pool index

	Op_LoadVarInt Opcode = 0x12 // b_c = int-var index
	Op_LoadVarStr Opcode = 0x13 // b_c = string-var index

	// 0x14 Op_LoadIntImmediate: inline i16 literal in b_c. NOT IMPLEMENTED (reserved)

	// A = dest (bool reg); one opcode per constant, so there is no operand to decode
	Op_ConstTrue  Opcode = 0x15
	Op_ConstFalse Opcode = 0x16

	// --- metadata (0x20) ---
	// A = key (str reg), B = value (str reg)
	Op_SetTxMeta Opcode = 0x20

	// A = account (str reg), B = key (str reg), C = value (str reg)
	Op_SetAccountMeta Opcode = 0x21

	// meta(account, key) read, dispatched on the target type.
	// A = dest, B = account (str reg), C = key (str reg)
	Op_MetaStr     Opcode = 0x22
	Op_MetaInt     Opcode = 0x23
	Op_MetaPortion Opcode = 0x24

	// as above, but a monetary needs two destinations, so the amount's goes in an
	// ext word: A = dest asset (str reg), ext.A = dest amount (int reg)
	Op_MetaMonetary Opcode = 0x25

	// --- arithmetic & constructors (0x30) ---
	Op_AddInt Opcode = 0x30
	Op_SubInt Opcode = 0x31
	// 0x32 was Op_MinInt: a min is a comparison and a copy, so it is Op_LtInt
	// plus a branch. Reserved, do not reuse.
	Op_SubPortion Opcode = 0x33
	Op_MkPortion  Opcode = 0x34
	// 0x35 was Op_MkMonetary: a monetary is a (str asset, int amount) register
	// pair, so there is nothing to construct. Reserved, do not reuse.
	Op_AddString Opcode = 0x36

	// 0x37 was Op_StrEq: moved to the comparison group, now 0x62. Reserved, do
	// not reuse.

	// not adjacent to Op_SubPortion (0x33) because 0x32 is burned and 0x34..0x37
	// are taken
	Op_AddPortion Opcode = 0x38

	// an allotment share is a mul plus a floor (Op_PortionToInt)
	Op_MulPortion Opcode = 0x39

	// 0x3A..0x3F reserved

	// --- unary & conversions (0x40) ---
	// 0x40 was Op_GetAmount and 0x41 was Op_GetAsset: projecting a monetary is now
	// just naming one of its two registers. Reserved, do not reuse.
	//
	// One copy per register bank: A = dest, B = src, both in that bank. There is no
	// monetary copy — a monetary is a (str asset, int amount) pair, so copy the two
	// halves. The family is split across 0x42..0x43 and 0x4A..0x4B because
	// 0x44..0x49 were already spoken for.
	Op_IntCopy     Opcode = 0x42
	Op_PortionCopy Opcode = 0x43

	Op_NegInt          Opcode = 0x44
	Op_IntToString     Opcode = 0x45
	Op_PortionToString Opcode = 0x46

	// A = dest (str reg), B = asset (str reg), C = amount (int reg)
	Op_MonetaryToString Opcode = 0x47

	// 0x48 was Op_IsZero: moved to the comparison group, now 0x63. Reserved, do
	// not reuse.

	// 0x49 was Op_Not: moved to the bool-ops group, now 0x70. Reserved, do not
	// reuse.

	// the other two bank copies; see Op_IntCopy above
	Op_StrCopy  Opcode = 0x4A
	Op_BoolCopy Opcode = 0x4B

	// Op_IntToPortion is exact; Op_PortionToInt floors.
	Op_IntToPortion Opcode = 0x4C
	Op_PortionToInt Opcode = 0x4D

	// 0x4E..0x4F reserved

	// --- funds & postings (0x50) ---

	// The most general form: account,cap,overdraft,color
	// The 0xFF special register means NULL for cap,overdraft and color
	Op_PullAccount Opcode = 0x50

	// account?, cap?, color?
	Op_SendToAccount Opcode = 0x51

	// save: reduce balance of account A for asset B by amount C (C == nilReg =>
	// save all), floored at 0
	Op_Save Opcode = 0x52

	// 0x53 was Op_MkAllotment: an allotment share is now built out of pure ops
	// (Op_IntToPortion, Op_MulPortion, Op_PortionToInt plus the leftover fixup),
	// so there is no variadic domain instruction. Reserved, do not reuse.

	// reads the account balance from the run-state
	Op_Balance Opcode = 0x54

	// --- marks (oneof backtracking) ---
	//
	// 0x55 was Op_Snapshot and 0x56 was Op_Restore: the source-queue mark used to
	// travel through an int register, which let any int be passed to a restore. The
	// pair below takes no register; the mark lives on a LIFO owned by the run-state.
	//
	// There is no "rewind but keep the mark" opcode: a retry is Op_MarkEnd with the
	// rewind flag followed by a fresh Op_MarkPush, so pushes and ends match strictly
	// and mark depth is a function of position in the instruction stream. A future
	// IR verifier can therefore prove pushes and ends balance and that no
	// Op_SendToAccount / Op_SetCurrentAsset / Op_Save sits inside a region; until it
	// exists the VM enforces that at execution time.

	// opens a region at the current source-queue depth and posting count
	Op_MarkPush Opcode = 0x55

	// A = rewind flag. Always pops the innermost mark; A == 1 additionally repays
	// everything pulled and reverses everything posted since the matching
	// Op_MarkPush, while A == 0 commits it.
	Op_MarkEnd Opcode = 0x56

	// reserved (0x57..0x5F) for PullAccount specializations, e.g.:
	// // cap=None, overdraft=BoundedZero
	// Op_PullAccountBoundedZero
	// // cap=None, overdraft=Bounded r
	// Op_PullAccountOverdraft
	// // cap=Some,  overdraft=BoundedZero
	// Op_PullAccountCap
	// // cap=Some,  overdraft=Unbounded
	// Op_PullAccountUnboundedOverdraft
	//
	// This block used to run to 0x8F; the comparison and bool-ops groups below took
	// 0x60..0x7F out of it.

	// --- comparisons (0x60) ---
	// A = dest (bool reg) for all of them; the operand banks are what the opcode
	// implies. Op_IsZero is unary and the rest binary, but they are one group
	// because they are the whole set of bool *producers*.
	//
	// Only `<` and `==` exist, per type. The other surface operators are normalised
	// by the front end:
	//
	//	a <  b   ->  Lt(a, b)
	//	a >  b   ->  Lt(b, a)            operands swapped
	//	a <= b   ->  Not(Lt(b, a))
	//	a >= b   ->  Not(Lt(a, b))
	//	a == b   ->  Eq(a, b)
	//	a != b   ->  Not(Eq(a, b))
	Op_LtInt     Opcode = 0x60
	Op_EqInt     Opcode = 0x61
	Op_StrEq     Opcode = 0x62 // was 0x37
	Op_IsZero    Opcode = 0x63 // was 0x48
	Op_LtPortion Opcode = 0x64
	Op_EqPortion Opcode = 0x65

	// reserved (0x66..0x6F) for `<` and `==` on types that don't exist yet. Str gets
	// equality only, never ordering. Bool equality and structural comparison of
	// tuples/arrays are front-end expansions rather than opcodes.

	// --- bool ops (0x70) ---
	// A = dest (bool reg), B = src (bool reg).
	Op_Not Opcode = 0x70 // was 0x49

	// reserved (0x71..0x7F) for and/or; both are expressible as branches, so
	// neither is needed for completeness

	// --- control flow (0x90) ---
	// A = cond (bool reg); b_c = unsigned forward delta, added to the pc of the
	// next instruction. A quantity is not a condition: project it with Op_IsZero.
	Op_JmpIfFalse Opcode = 0x90
	// unconditional; b_c = unsigned forward delta, as above
	Op_Jmp Opcode = 0x91
	// the dual of Op_JmpIfFalse, so either edge of a bool can be the branch without
	// a negation instruction
	Op_JmpIfTrue Opcode = 0x92
	// Label emits no instruction; it only feeds the symbol table at assemble time
)
