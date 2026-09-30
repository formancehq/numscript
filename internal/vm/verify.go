package vm

import "fmt"

// regBank identifies which of the VM's register banks an operand indexes. The
// bank is never encoded in the instruction — it is implied by the opcode — so the
// verifier has to reconstruct it from the same table Exec switches on.
type regBank int

const (
	bankInt regBank = iota
	bankStr
	bankPortion
	bankBool
	// bankCurrentAsset is a pseudo-bank with a single slot, tracking whether the
	// current asset has been set. It is never allocated as a register.
	bankCurrentAsset
)

func (b regBank) String() string {
	switch b {
	case bankInt:
		return "int"
	case bankStr:
		return "string"
	case bankPortion:
		return "portion"
	case bankBool:
		return "bool"
	case bankCurrentAsset:
		return "current asset"
	default:
		return "?"
	}
}

var currentAssetRef = regRef{bankCurrentAsset, 0}

type regRef struct {
	bank  regBank
	index int
}

func (r regRef) String() string {
	if r.bank == bankCurrentAsset {
		return "current asset"
	}
	return fmt.Sprintf("%s register %d", r.bank, r.index)
}

// decoded describes the register/pool/jump operands an instruction touches, so
// the verifier can check every access the execution loop will make without the
// VM having to guard anything at run time.
type decoded struct {
	reads    []regRef
	writes   []regRef
	constInt int // index into Program.IntsPool, or -1
	constStr int
	varInt   int // index into Vars.IntsPool, or -1
	varStr   int
	// jumpDelta is the unsigned forward offset from the instruction *after* this
	// one, or -1 when the instruction does not jump
	jumpDelta int
	// noFallThrough marks an unconditional jump, whose successor is its target
	// only. Giving it a fallthrough edge too would intersect the assigned-set
	// with a path that cannot be taken, rejecting valid programs.
	noFallThrough bool
	// markDelta is +1 for a mark push, -1 for a mark end, 0 otherwise
	markDelta int
	// notInMark marks an op Exec refuses while a mark is open
	notInMark bool
}

type programInfo struct {
	varIntsLen int
	varStrsLen int
}

// Verify statically checks that a program is safe to execute. It is opt-in: Exec
// assumes well-formed bytecode (the compiler's output always is) and does not
// call this. Run it on any program that did not come out of ir.Assemble in this
// process — anything decoded from a file or a wire.
//
// A nil result guarantees the execution loop cannot read out of bounds: no
// truncated instruction, no jump into the middle of one, no out-of-range pool or
// register index, and no read of a register that was not written on every path
// reaching it.
//
// It does not check that vars were supplied; for that see VerifyWithVars.
func Verify(p Program) error {
	_, err := verify(p)
	return err
}

// VerifiedVarsInfo records, opaquely, the variable-pool sizes a Program was
// found to require by a successful VerifyWithVars call. A caller that keeps
// the Program around — e.g. one compiled artifact reused across many calls
// with (normally) the same Vars shape — can hold on to this instead of the
// Program's raw pool-size requirements, and later ask CheckVars whether a
// new Vars value would still satisfy VerifyWithVars, without paying for the
// static pass again. The fields are deliberately unexported: what "shape"
// means is this package's business, not a caller's.
type VerifiedVarsInfo struct {
	varIntsLen int
	varStrsLen int
}

// CheckVars reports, in O(1) and without re-running verification, whether
// vars is guaranteed to satisfy the Program this VerifiedVarsInfo was
// obtained from — i.e. whether VerifyWithVars(thatProgram, vars) would
// succeed, without calling it. A false result is not itself a rejection: it
// only means the caller must call VerifyWithVars to get an authoritative
// answer (and, on failure, a precise error).
func (info VerifiedVarsInfo) CheckVars(vars *Vars) bool {
	if vars == nil {
		return info.varIntsLen == 0 && info.varStrsLen == 0
	}
	return len(vars.IntsPool) >= info.varIntsLen && len(vars.StringsPool) >= info.varStrsLen
}

// VerifyWithVars is Verify plus the check that vars carries every variable the
// program loads. A nil *Vars is only legal for a program that reads none. On
// success it also returns the VerifiedVarsInfo backing that check; see its
// doc for why a caller would want to keep it.
func VerifyWithVars(p Program, vars *Vars) (VerifiedVarsInfo, error) {
	info, err := verify(p)
	if err != nil {
		return VerifiedVarsInfo{}, err
	}

	result := VerifiedVarsInfo(info)
	if result.CheckVars(vars) {
		return result, nil
	}

	if vars == nil {
		return VerifiedVarsInfo{}, fmt.Errorf("program reads variables but none were provided")
	}
	if len(vars.IntsPool) < info.varIntsLen {
		return VerifiedVarsInfo{}, fmt.Errorf("program reads int var %d but only %d were provided", info.varIntsLen-1, len(vars.IntsPool))
	}
	return VerifiedVarsInfo{}, fmt.Errorf("program reads string var %d but only %d were provided", info.varStrsLen-1, len(vars.StringsPool))
}

// instrWords is how many 4-byte words an opcode occupies. The ones returning 2
// carry an "ext" word whose opcode byte is ignored; Exec reads it with
// `instrs[pc]; pc++`, which is what makes a truncated tail or a jump landing on
// an ext word a crash rather than an error.
func instrWords(op byte) int {
	switch Opcode(op) {
	case Op_PullAccount,
		Op_Save,
		Op_SetAccountMeta,
		Op_MetaStr,
		Op_MetaInt,
		Op_MetaPortion,
		Op_MetaMonetary,
		Op_Balance:
		return 2
	default:
		return 1
	}
}

func (p Program) maxReg(bank regBank) int {
	switch bank {
	case bankInt:
		return int(p.MaxRegInt)
	case bankStr:
		return int(p.MaxRegString)
	case bankPortion:
		return int(p.MaxRegPortion)
	case bankBool:
		return int(p.MaxRegBool)
	default:
		return 1 // the current-asset pseudo-bank has exactly one slot
	}
}

func verify(p Program) (programInfo, error) {
	instrs := p.Instructions
	n := len(instrs)

	// boundary[i] is true when i starts an instruction; the ext word of a
	// two-word instruction is deliberately not a boundary
	boundary := make([]bool, n+1)
	boundary[n] = true // jumping past the last instruction halts, which is fine

	var steps []step

	for i := 0; i < n; {
		op := instrs[i].Opcode
		w := instrWords(op)
		if i+w > n {
			return programInfo{}, fmt.Errorf("truncated instruction at %d: opcode 0x%02X needs %d words", i, op, w)
		}
		var ext Instruction
		if w == 2 {
			ext = instrs[i+1]
		}
		d, err := decodeInstr(instrs[i], ext)
		if err != nil {
			return programInfo{}, fmt.Errorf("at instruction %d: %w", i, err)
		}
		boundary[i] = true
		steps = append(steps, step{at: i, words: w, d: d})
		i += w
	}

	info := programInfo{}
	for _, st := range steps {
		d := st.d

		for _, r := range append(append([]regRef{}, d.reads...), d.writes...) {
			if r.bank == bankCurrentAsset {
				continue
			}
			if r.index >= p.maxReg(r.bank) {
				return programInfo{}, fmt.Errorf(
					"at instruction %d: %s is beyond the program's declared %d %s registers",
					st.at, r, p.maxReg(r.bank), r.bank)
			}
		}

		if d.constInt >= len(p.IntsPool) {
			return programInfo{}, fmt.Errorf("at instruction %d: int constant %d out of range (pool size %d)", st.at, d.constInt, len(p.IntsPool))
		}
		if d.constStr >= len(p.StringsPool) {
			return programInfo{}, fmt.Errorf("at instruction %d: string constant %d out of range (pool size %d)", st.at, d.constStr, len(p.StringsPool))
		}
		if d.varInt >= info.varIntsLen {
			info.varIntsLen = d.varInt + 1
		}
		if d.varStr >= info.varStrsLen {
			info.varStrsLen = d.varStr + 1
		}

		// deltas are unsigned and relative to the following instruction, so a
		// backward jump cannot be encoded; the hazard is landing on an ext word
		if t := st.target(); t >= 0 && t < n && !boundary[t] {
			return programInfo{}, fmt.Errorf("at instruction %d: jump to %d lands inside an instruction", st.at, t)
		}
	}

	g := buildCFG(steps)
	if err := checkDefiniteAssignment(steps, g); err != nil {
		return programInfo{}, err
	}
	if err := checkMarkBalance(steps, g); err != nil {
		return programInfo{}, err
	}

	return info, nil
}

// step is one decoded instruction, tagged with where it sits in the stream.
type step struct {
	at    int
	words int
	d     decoded
}

// target is the absolute instruction index this step jumps to, or -1 when it
// does not jump. A target of len(instrs) or beyond halts, which is legal.
func (s step) target() int {
	if s.d.jumpDelta < 0 {
		return -1
	}
	return s.at + s.words + s.d.jumpDelta
}

// cfg is the control-flow graph over steps: edges are fallthroughs and jump
// targets, indexed by position in steps.
type cfg struct {
	preds     [][]int
	reachable []bool
	// exits[k] is true when step k can end the run: by falling off the last
	// instruction or jumping to len(instrs) or beyond
	exits []bool
}

// Jumps are forward-only, so every predecessor of a step is earlier in the
// stream: one ordered pass over the steps sees each step after all of its
// predecessors.
func buildCFG(steps []step) cfg {
	at2idx := make(map[int]int, len(steps))
	for k, st := range steps {
		at2idx[st.at] = k
	}

	g := cfg{
		preds:     make([][]int, len(steps)),
		reachable: make([]bool, len(steps)),
		exits:     make([]bool, len(steps)),
	}
	edge := func(k, at int) {
		if j, ok := at2idx[at]; ok {
			g.preds[j] = append(g.preds[j], k)
		} else {
			g.exits[k] = true
		}
	}
	for k, st := range steps {
		if !st.d.noFallThrough {
			edge(k, st.at+st.words)
		}
		if t := st.target(); t >= 0 {
			edge(k, t)
		}
	}

	// An unreachable step never executes, so its reads cannot crash and its
	// (empty) assigned-set must not poison the intersection at a later join.
	if len(steps) > 0 {
		g.reachable[0] = true
	}
	for k := range steps {
		for _, p := range g.preds[k] {
			if g.reachable[p] {
				g.reachable[k] = true
				break
			}
		}
	}
	return g
}

// checkDefiniteAssignment rejects a read of a register (or of the current asset)
// that was not written on every path reaching that instruction.
//
// This subsumes bytecode-level type confusion: a bank is part of a regRef, so a
// slot written as an int and later read as a string is a read of a register that
// was never written.
func checkDefiniteAssignment(steps []step, g cfg) error {
	assignedOut := make([]map[regRef]bool, len(steps))
	for k, st := range steps {
		if !g.reachable[k] {
			continue
		}
		in := intersectAssigned(assignedOut, filterReachable(g.preds[k], g.reachable))
		for _, r := range st.d.reads {
			if !in[r] {
				return fmt.Errorf("at instruction %d: %s read before being assigned on all paths", st.at, r)
			}
		}
		for _, r := range st.d.writes {
			in[r] = true
		}
		assignedOut[k] = in
	}
	return nil
}

// checkMarkBalance rejects a program in which the number of open marks is not
// the same on every path reaching an instruction, a mark end runs with no open
// mark, a send, save or set_current_asset runs with a mark open, or the run can
// end with a mark open.
func checkMarkBalance(steps []step, g cfg) error {
	depthOut := make([]int, len(steps))
	for k, st := range steps {
		if !g.reachable[k] {
			continue
		}
		in := 0
		preds := filterReachable(g.preds[k], g.reachable)
		for i, p := range preds {
			if i == 0 {
				in = depthOut[p]
			} else if depthOut[p] != in {
				return fmt.Errorf("at instruction %d: paths join with different numbers of open marks (%d and %d)", st.at, in, depthOut[p])
			}
		}
		if st.d.notInMark && in > 0 {
			return fmt.Errorf("at instruction %d: send, save or set_current_asset while a mark is open", st.at)
		}
		out := in + st.d.markDelta
		if out < 0 {
			return fmt.Errorf("at instruction %d: mark end with no open mark", st.at)
		}
		if g.exits[k] && out != 0 {
			return fmt.Errorf("at instruction %d: the run can end with %d open marks", st.at, out)
		}
		depthOut[k] = out
	}
	return nil
}

func filterReachable(preds []int, reachable []bool) []int {
	out := preds[:0:0]
	for _, p := range preds {
		if reachable[p] {
			out = append(out, p)
		}
	}
	return out
}

func intersectAssigned(out []map[regRef]bool, preds []int) map[regRef]bool {
	res := map[regRef]bool{}
	if len(preds) == 0 {
		return res
	}
	for r := range out[preds[0]] {
		res[r] = true
	}
	for _, p := range preds[1:] {
		for r := range res {
			if !out[p][r] {
				delete(res, r)
			}
		}
	}
	return res
}

// decodeInstr mirrors, operand for operand, what the matching arm of Exec reads
// and writes. Every opcode in instruction.go must have a case here: an opcode
// missing from this switch is rejected as unknown, which is what keeps the two
// in step.
func decodeInstr(instr, ext Instruction) (decoded, error) {
	d := decoded{constInt: -1, constStr: -1, varInt: -1, varStr: -1, jumpDelta: -1}

	var err error
	// read/write name an operand Exec dereferences unconditionally, so nilReg
	// there is a malformed instruction rather than an absent operand
	read := func(bank regBank, idx byte) {
		if idx == nilReg {
			err = fmt.Errorf("%s operand is the nil register, but this operand is not optional", bank)
			return
		}
		d.reads = append(d.reads, regRef{bank, int(idx)})
	}
	write := func(bank regBank, idx byte) {
		if idx == nilReg {
			err = fmt.Errorf("%s destination is the nil register", bank)
			return
		}
		d.writes = append(d.writes, regRef{bank, int(idx)})
	}
	readOpt := func(bank regBank, idx byte) {
		if idx != nilReg {
			d.reads = append(d.reads, regRef{bank, int(idx)})
		}
	}
	flag := func(v byte) {
		if v > 1 {
			err = fmt.Errorf("flag operand is %d, expected 0 or 1", v)
		}
	}

	switch Opcode(instr.Opcode) {
	// --- state & assertions
	case Op_SetCurrentAsset:
		read(bankStr, instr.A)
		d.writes = append(d.writes, currentAssetRef)
		d.notInMark = true
	case Op_AssertSameAsset:
		read(bankStr, instr.A)
		read(bankStr, instr.B)
	case Op_AssertValidAccount, Op_AssertValidColor, Op_AssertValidScope:
		read(bankStr, instr.A)
	case Op_AssertNonNegativeBalance:
		read(bankInt, instr.A)
		read(bankStr, instr.B)
	case Op_AssertNonNegativeAmount:
		read(bankInt, instr.A)
	case Op_AssertNonNegativePortion:
		read(bankPortion, instr.A)
	case Op_AssertLeftover:
		read(bankPortion, instr.A)
		flag(instr.B)
	case Op_CheckEnoughFunds:
		read(bankInt, instr.A)
		read(bankInt, instr.B)
		// the asset names the MissingFundsError this may raise
		d.reads = append(d.reads, currentAssetRef)

	// --- constants & variables
	case Op_LoadInt:
		d.constInt = int(instr.GetBC())
		write(bankInt, instr.A)
	case Op_LoadStr:
		d.constStr = int(instr.GetBC())
		write(bankStr, instr.A)
	case Op_LoadVarInt:
		d.varInt = int(instr.GetBC())
		write(bankInt, instr.A)
	case Op_LoadVarStr:
		d.varStr = int(instr.GetBC())
		write(bankStr, instr.A)
	case Op_ConstTrue, Op_ConstFalse:
		write(bankBool, instr.A)

	// --- metadata
	case Op_SetTxMeta:
		read(bankStr, instr.A)
		read(bankStr, instr.B)
	case Op_SetAccountMeta:
		read(bankStr, instr.A)
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		readOpt(bankStr, ext.A) // scope
	case Op_MetaStr:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		readOpt(bankStr, ext.A)
		write(bankStr, instr.A)
	case Op_MetaInt:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		readOpt(bankStr, ext.A)
		write(bankInt, instr.A)
	case Op_MetaPortion:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		readOpt(bankStr, ext.A)
		write(bankPortion, instr.A)
	case Op_MetaMonetary:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		readOpt(bankStr, ext.B) // scope
		write(bankStr, instr.A) // asset
		write(bankInt, ext.A)   // amount
	case Op_Balance:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		readOpt(bankStr, ext.A) // scope
		write(bankInt, instr.A)

	// --- arithmetic & constructors
	case Op_AddInt, Op_SubInt:
		read(bankInt, instr.B)
		read(bankInt, instr.C)
		write(bankInt, instr.A)
	case Op_AddPortion, Op_SubPortion, Op_MulPortion:
		read(bankPortion, instr.B)
		read(bankPortion, instr.C)
		write(bankPortion, instr.A)
	case Op_MkPortion:
		read(bankInt, instr.B)
		read(bankInt, instr.C)
		write(bankPortion, instr.A)
	case Op_AddString:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		write(bankStr, instr.A)

	// --- unary & conversions
	case Op_IntCopy, Op_NegInt:
		read(bankInt, instr.B)
		write(bankInt, instr.A)
	case Op_PortionCopy:
		read(bankPortion, instr.B)
		write(bankPortion, instr.A)
	case Op_StrCopy:
		read(bankStr, instr.B)
		write(bankStr, instr.A)
	case Op_BoolCopy:
		read(bankBool, instr.B)
		write(bankBool, instr.A)
	case Op_IntToString:
		read(bankInt, instr.B)
		write(bankStr, instr.A)
	case Op_PortionToString:
		read(bankPortion, instr.B)
		write(bankStr, instr.A)
	case Op_MonetaryToString:
		read(bankStr, instr.B)
		read(bankInt, instr.C)
		write(bankStr, instr.A)
	case Op_IntToPortion:
		read(bankInt, instr.B)
		write(bankPortion, instr.A)
	case Op_PortionToInt:
		read(bankPortion, instr.B)
		write(bankInt, instr.A)

	// --- funds & postings
	case Op_PullAccount:
		read(bankStr, instr.B)    // account
		readOpt(bankInt, instr.C) // cap
		readOpt(bankInt, ext.A)   // overdraft
		readOpt(bankStr, ext.B)   // color
		readOpt(bankStr, ext.C)   // scope
		d.reads = append(d.reads, currentAssetRef)
		write(bankInt, instr.A)
	case Op_SendToAccount:
		readOpt(bankStr, instr.A) // destination
		readOpt(bankInt, instr.B) // cap
		readOpt(bankStr, instr.C) // scope
		d.reads = append(d.reads, currentAssetRef)
		d.notInMark = true
	case Op_Save:
		read(bankStr, instr.A)    // account
		read(bankStr, instr.B)    // asset
		readOpt(bankInt, instr.C) // amount, nil = save all
		readOpt(bankStr, ext.A)   // scope
		d.notInMark = true

	// --- marks
	case Op_MarkPush:
		d.markDelta = 1
	case Op_MarkEnd:
		flag(instr.A)
		d.markDelta = -1
		// a rewind repays queued sources into the current asset's balance, but it
		// is not listed as a reader: the repay loop only runs over sources queued
		// inside the region, and queueing one takes an Op_PullAccount, which
		// already requires the asset. A region that pulled nothing reads nothing.

	// --- comparisons
	case Op_LtInt, Op_EqInt:
		read(bankInt, instr.B)
		read(bankInt, instr.C)
		write(bankBool, instr.A)
	case Op_LtPortion, Op_EqPortion:
		read(bankPortion, instr.B)
		read(bankPortion, instr.C)
		write(bankBool, instr.A)
	case Op_StrEq:
		read(bankStr, instr.B)
		read(bankStr, instr.C)
		write(bankBool, instr.A)
	case Op_IsZero:
		read(bankInt, instr.B)
		write(bankBool, instr.A)

	// --- bool ops
	case Op_Not:
		read(bankBool, instr.B)
		write(bankBool, instr.A)

	// --- control flow
	case Op_JmpIfFalse, Op_JmpIfTrue:
		read(bankBool, instr.A)
		d.jumpDelta = int(instr.GetBC())
	case Op_Jmp:
		d.jumpDelta = int(instr.GetBC())
		d.noFallThrough = true

	default:
		return decoded{}, fmt.Errorf("unknown opcode 0x%02X", instr.Opcode)
	}

	if err != nil {
		return decoded{}, err
	}
	return d, nil
}
