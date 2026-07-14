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

const (
	// --- misc / state ---
	Op_SetCurrentAsset Opcode = iota

	Op_AssertSameAsset

	// --- metadata ---
	// A = key (str reg), B = value (str reg)
	Op_SetTxMeta

	// A = account (str reg), B = key (str reg), C = value (str reg)
	Op_SetAccountMeta

	// meta(account, key) read, dispatched on the target type.
	// A = dest, B = account (str reg), C = key (str reg)
	Op_MetaStr
	Op_MetaInt
	Op_MetaPortion
	Op_MetaMonetary

	// --- variables ---
	Op_LoadVarInt // b_c = int-var index
	Op_LoadVarStr // b_c = string-var index

	// --- constants ---
	// may split into one opcode per expr_typ later
	Op_LoadInt // LoadConst (`Int)    -> b_c = const-pool index
	Op_LoadStr // LoadConst (`String) -> b_c = const-pool index

	// --- funds ---
	Op_CheckEnoughFunds

	// checks the allotment leftover portion in reg A: errors if it is negative
	// (portions summing to > 1), and — when B == 1 (no `remaining` clause) — if it
	// is non-zero (portions not summing to exactly 1)
	Op_AssertLeftover

	// save: reduce balance of account A for asset B by amount C (C == nilReg =>
	// save all), floored at 0
	Op_Save

	// --- PullAccount (cap? × overdraft) ---

	// The most general form:
	// account,cap,overdraft,color
	// The 0xFF special register means NULL for cap,overdraft and color
	Op_PullAccount

	// // cap=None, overdraft=BoundedZero
	// Op_PullAccountBoundedZero
	// // cap=None, overdraft=Bounded r
	// Op_PullAccountOverdraft
	// // cap=Some,  overdraft=BoundedZero
	// Op_PullAccountCap

	// // cap=Some,  overdraft=Unbounded
	// Op_PullAccountUnboundedOverdraft

	// dest_start,inp_arr_start,inp_arr_size|amt
	Op_MkAllotment

	// account?, cap?, color?
	Op_SendToAccount

	// --- control flow ---
	Op_JmpIfZero // b_c = resolved instruction offset
	// note: Label emits no instruction; it only feeds the symbol table at assemble time

	// --- UnaryOp ---
	Op_GetAmount
	Op_GetAsset
	Op_IntCopy
	Op_PortionCopy
	Op_NegInt
	Op_IntToString
	Op_PortionToString
	Op_MonetaryToString

	// --- BinaryOp ---
	Op_MinInt
	Op_AddInt
	Op_SubInt
	Op_AddString
	Op_SubPortion
	Op_MkPortion
	Op_MkMonetary
	Op_Balance // reads the account balance from the run-state
)
