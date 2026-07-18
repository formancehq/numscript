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

	// errors if the account name in str reg A is not well-formed
	Op_AssertValidAccount

	// errors (NegativeBalanceError) if the monetary in reg A has a negative
	// amount; B = account str reg (for the error)
	Op_AssertNonNegativeBalance

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
	Op_LoadInt    // LoadConst (`Int)    -> b_c = const-pool index
	Op_LoadStr    // LoadConst (`String) -> b_c = const-pool index
	Op_LoadIntImm // LoadInt immediate   -> b_c = the (unsigned, u16) value itself

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

	// The most general form: account,cap,overdraft,color (2 words).
	// The 0xFF special register means NULL for cap,overdraft and color
	Op_PullAccount

	// Compact single-word form for the common plain-account pull:
	// cap=Some, overdraft=BoundedZero, no color. A=dest, B=account, C=cap.
	// (world is still treated as unbounded.)
	Op_PullAccountCapZero

	// // cap=None, overdraft=Bounded r
	// Op_PullAccountOverdraft
	// // cap=Some,  overdraft=Unbounded
	// Op_PullAccountUnboundedOverdraft

	// dest_start,inp_arr_start,inp_arr_size|amt
	Op_MkAllotment

	// account?, cap?, color?
	Op_SendToAccount

	// --- funds-bypass fast path (1-source/1-destination send) ---
	// Take is Pull without queuing: it computes the available amount and debits
	// the source, leaving the posting to a later Op_Post. Same operand layout and
	// overdraft/world handling as Op_PullAccount (2 words). Word1: A=dest(int),
	// B=src(str), C=cap(int or nil). Word2: A=overdraft(int or nil), B=color(str
	// or nil).
	Op_Take

	// Compact Take for the common plain-account case (cap present, overdraft
	// bounded-zero, no color; world stays unbounded). Mirrors
	// Op_PullAccountCapZero, single word: A=dest, B=src, C=cap.
	Op_TakeCapZero

	// Post emits a direct posting src->dst of the amount in reg C (currentAsset),
	// crediting dst, WITHOUT debiting src (Take already did). Single word:
	// A=src(str), B=dst(str), C=amount(int).
	Op_Post

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

	// Op_PostFromUnbounded is the aggressive single->single fast path for an
	// UNBOUNDED source (@world, or `allowing unbounded overdraft`). Because an
	// unbounded pull makes exactly `cap` available regardless of balance, and can
	// never be short, the source debit AND the enough-funds check are both
	// elided: this one op emits the posting src->dst of `cap` (current asset) and
	// credits dst, with no source debit and no funds queue. Single word:
	// A=src(str), B=dst(str), C=cap(int). Emitted by the postFromUnbounded
	// peephole only when the source balance is provably never observed afterward
	// (no `balance` read; non-world source not re-pulled), so dropping the debit
	// is unobservable.
	Op_PostFromUnbounded

	// Op_PostFromUnboundedLeaf is Op_PostFromUnbounded for a LEAF destination: it
	// additionally skips crediting dst's balance. Emitted only when the peephole
	// proves dst's balance is never observed (no `balance` read; dst never a later
	// funding source; dst never saved), so the credit — and the balance-map lookup
	// it entails, the dominant cost of the fused path — is dead. Same operand
	// layout as Op_PostFromUnbounded (A=src, B=dst, C=cap).
	Op_PostFromUnboundedLeaf

	// Op_TakeCapZeroSlot / Op_PostSlot are Op_TakeCapZero / Op_Post with a
	// compile-assigned balance SLOT: a two-word form whose second word carries the
	// slot index in ext.A. The slot is a fast path over the string-keyed balance
	// map for a constant (account, asset) access (see runtime.entryForSlot).
	// Word1 is identical to the base op; word2 = {_, A=slot, _, _}.
	Op_TakeCapZeroSlot
	Op_PostSlot
)
