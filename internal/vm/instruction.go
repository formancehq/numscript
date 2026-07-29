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

	// errors (NegativeBalanceError) if the monetary in reg A has a negative
	// amount; B = account str reg (for the error)
	Op_AssertNonNegativeBalance Opcode = 0x03

	// checks the allotment leftover portion in reg A: errors if it is negative
	// (portions summing to > 1), and — when B == 1 (no `remaining` clause) — if it
	// is non-zero (portions not summing to exactly 1)
	Op_AssertLeftover Opcode = 0x04

	Op_CheckEnoughFunds Opcode = 0x05

	// errors if the color in str reg A is not well-formed
	Op_AssertValidColor Opcode = 0x06

	// --- constants & variables (0x10) ---
	// may split into one opcode per expr_typ later
	Op_LoadInt Opcode = 0x10 // LoadConst (`Int)    -> b_c = const-pool index
	Op_LoadStr Opcode = 0x11 // LoadConst (`String) -> b_c = const-pool index

	Op_LoadVarInt Opcode = 0x12 // b_c = int-var index
	Op_LoadVarStr Opcode = 0x13 // b_c = string-var index

	// 0x14 Op_LoadIntImmediate: inline i16 literal in b_c. NOT IMPLEMENTED (reserved)

	// --- metadata (0x20) ---
	// A = key (str reg), B = value (str reg)
	Op_SetTxMeta Opcode = 0x20

	// A = account (str reg), B = key (str reg), C = value (str reg)
	Op_SetAccountMeta Opcode = 0x21

	// meta(account, key) read, dispatched on the target type.
	// A = dest, B = account (str reg), C = key (str reg)
	Op_MetaStr      Opcode = 0x22
	Op_MetaInt      Opcode = 0x23
	Op_MetaPortion  Opcode = 0x24
	Op_MetaMonetary Opcode = 0x25

	// --- arithmetic & constructors (0x30) ---
	Op_AddInt     Opcode = 0x30
	Op_SubInt     Opcode = 0x31
	Op_MinInt     Opcode = 0x32
	Op_SubPortion Opcode = 0x33
	Op_MkPortion  Opcode = 0x34
	Op_MkMonetary Opcode = 0x35
	Op_AddString  Opcode = 0x36

	// --- unary & conversions (0x40) ---
	Op_GetAmount        Opcode = 0x40
	Op_GetAsset         Opcode = 0x41
	Op_IntCopy          Opcode = 0x42
	Op_PortionCopy      Opcode = 0x43
	Op_NegInt           Opcode = 0x44
	Op_IntToString      Opcode = 0x45
	Op_PortionToString  Opcode = 0x46
	Op_MonetaryToString Opcode = 0x47

	// --- funds & postings (0x50) ---

	// The most general form: account,cap,overdraft,color
	// The 0xFF special register means NULL for cap,overdraft and color
	Op_PullAccount Opcode = 0x50

	// account?, cap?, color?
	Op_SendToAccount Opcode = 0x51

	// save: reduce balance of account A for asset B by amount C (C == nilReg =>
	// save all), floored at 0
	Op_Save Opcode = 0x52

	// dest_start,inp_arr_start,inp_arr_size|amt
	Op_MkAllotment Opcode = 0x53

	// reads the account balance from the run-state
	Op_Balance Opcode = 0x54

	// A = dest int reg; int_regs[A] = current source-queue mark (len(sources)).
	// Used for oneof backtracking.
	Op_Snapshot Opcode = 0x55

	// A = int reg holding a mark; rolls the source queue back to it.
	Op_Restore Opcode = 0x56

	// reserved (0x57..0x8F) for PullAccount specializations, e.g.:
	// // cap=None, overdraft=BoundedZero
	// Op_PullAccountBoundedZero
	// // cap=None, overdraft=Bounded r
	// Op_PullAccountOverdraft
	// // cap=Some,  overdraft=BoundedZero
	// Op_PullAccountCap
	// // cap=Some,  overdraft=Unbounded
	// Op_PullAccountUnboundedOverdraft

	// --- control flow (0x90) ---
	Op_JmpIfZero Opcode = 0x90 // b_c = resolved instruction offset
	// note: Label emits no instruction; it only feeds the symbol table at assemble time
)
