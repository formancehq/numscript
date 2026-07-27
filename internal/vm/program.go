package vm

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

type Program struct {
	Instructions []Instruction

	StringsPool []string
	IntsPool    []big.Int

	MaxRegString   byte
	MaxRegInt      byte
	MaxRegPortion  byte
	MaxRegMonetary byte
}

var le = binary.LittleEndian

// TODO review AI blob
func (p Program) Encode() []byte {
	instrs := make([]byte, 4*len(p.Instructions))
	for i, ins := range p.Instructions {
		instrs[i*4], instrs[i*4+1], instrs[i*4+2], instrs[i*4+3] = ins.Opcode, ins.A, ins.B, ins.C
	}

	strs := encodeStringsPool(p.StringsPool)
	ints := encodeIntsPool(p.IntsPool)
	maxRegs := encodeMaxRegs(p)

	buf := make([]byte, 0, formatHeaderLen+4*6+len(instrs)+len(strs)+len(ints)+len(maxRegs))
	buf = appendFormatHeader(buf, "NUMB", 4)
	buf = appendSection(buf, SectionInstructions, instrs)
	buf = appendSection(buf, SectionStringsPool, strs)
	buf = appendSection(buf, SectionIntsPool, ints)
	buf = appendSection(buf, SectionMaxRegisters, maxRegs)
	return buf
}

// These fields hold the per-bank register *count* (== highest index + 1), as
// emitted by the assembler. Real indices are 0..0xFE (0xFF is the nil sentinel),
// so the largest possible count is 255. When the max-registers section is absent
// we assume the bank uses every usable register, i.e. this default.
const maxRegDefault byte = 255

func encodeMaxRegs(p Program) []byte {
	return []byte{p.MaxRegString, p.MaxRegInt, p.MaxRegPortion, p.MaxRegMonetary}
}

// One byte per bank, positional, append-only order. The section length is the
// number of banks the writer knew.
//
//   - absent (len 0): no info, so every bank defaults to maxRegDefault (safe).
//   - present: bank i uses buf[i] when i < len; banks beyond len default to 0,
//     since a bank the (older) writer didn't know is a type the program predates
//     and provably uses none of.
//
// Extra trailing bytes (a newer writer) are ignored; a program that actually uses
// such a bank is rejected later via its unknown opcodes.
func parseMaxRegs(buf []byte) (str, i, portion, monetary byte) {
	if len(buf) == 0 {
		return maxRegDefault, maxRegDefault, maxRegDefault, maxRegDefault
	}
	at := func(idx int) byte {
		if idx < len(buf) {
			return buf[idx]
		}
		return 0
	}
	return at(0), at(1), at(2), at(3)
}

func encodeStringsPool(strings []string) []byte {
	buf := make([]byte, 4)
	le.PutUint32(buf, uint32(len(strings)))
	var lenBuf [4]byte
	for _, s := range strings {
		le.PutUint32(lenBuf[:], uint32(len(s)))
		buf = append(buf, lenBuf[:]...)
		buf = append(buf, s...)
	}
	return buf
}

func encodeIntsPool(ints []big.Int) []byte {
	buf := make([]byte, 4)
	le.PutUint32(buf, uint32(len(ints)))
	var lenBuf [4]byte
	for i := range ints {
		n := &ints[i]
		sign := byte(0)
		if n.Sign() < 0 {
			sign = 1
		}
		mag := n.Bytes() // absolute value, big-endian (big.Int's native form)
		buf = append(buf, sign)
		le.PutUint32(lenBuf[:], uint32(len(mag)))
		buf = append(buf, lenBuf[:]...)
		buf = append(buf, mag...)
	}
	return buf
}

func parseInstructions(buf []byte) ([]Instruction, error) {
	if len(buf)%4 != 0 {
		return nil, fmt.Errorf("instructions section size %d not a multiple of 4", len(buf))
	}
	instructions := make([]Instruction, len(buf)/4)
	for i := range instructions {
		off := i * 4
		instructions[i] = Instruction{
			buf[off],
			buf[off+1],
			buf[off+2],
			buf[off+3],
		}
	}
	return instructions, nil
}

func parseStringsPool(buf []byte) ([]string, error) {
	if len(buf) == 0 {
		return nil, nil
	}
	if len(buf) < 4 {
		return nil, fmt.Errorf("strings pool: count truncated")
	}
	n := le.Uint32(buf)
	total := uint64(len(buf))
	if uint64(n)*4 > total-4 { // every record is at least a 4B length prefix
		return nil, fmt.Errorf("strings pool: count %d exceeds buffer size %d", n, total)
	}
	idx := uint64(4)
	out := make([]string, n)
	for i := range out {
		if idx+4 > total {
			return nil, fmt.Errorf("string %d: length prefix out of bounds", i)
		}
		strLen := uint64(le.Uint32(buf[idx:]))
		idx += 4
		end := idx + strLen // operands <= ~4.3e9, sum fits in uint64
		if end > total {
			return nil, fmt.Errorf("string %d: body [%d:%d] out of bounds (%d)", i, idx, end, total)
		}
		out[i] = string(buf[idx:end]) // copies; Program no longer references buf
		idx = end
	}
	return out, nil
}

func parseIntsPool(buf []byte) ([]big.Int, error) {
	if len(buf) == 0 {
		return nil, nil
	}
	if len(buf) < 4 {
		return nil, fmt.Errorf("ints pool: count truncated")
	}
	n := le.Uint32(buf)
	total := uint64(len(buf))
	if uint64(n)*5 > total-4 { // every record is at least a 5B header
		return nil, fmt.Errorf("ints pool: count %d exceeds buffer size %d", n, total)
	}
	idx := uint64(4)
	out := make([]big.Int, n)
	for i := range out {
		if idx+5 > total {
			return nil, fmt.Errorf("int %d: header out of bounds", i)
		}
		sign := buf[idx]
		magLen := uint64(le.Uint32(buf[idx+1:]))
		idx += 5
		end := idx + magLen
		if end > total {
			return nil, fmt.Errorf("int %d: magnitude [%d:%d] out of bounds (%d)", i, idx, end, total)
		}
		out[i].SetBytes(buf[idx:end]) // big-endian, unsigned magnitude
		switch sign {
		case 0:
			// non-negative
		case 1:
			out[i].Neg(&out[i])
		default:
			return nil, fmt.Errorf("int %d: invalid sign byte %d", i, sign)
		}
		idx = end
	}
	return out, nil
}

func DecodeProgram(buf []byte) (Program, error) {
	sections, err := decodeSections("NUMB", buf, SectionInstructions, SectionStringsPool, SectionIntsPool, SectionMaxRegisters)
	if err != nil {
		return Program{}, err
	}

	instructions, err := parseInstructions(sections[SectionInstructions])
	if err != nil {
		return Program{}, err
	}

	stringsPool, err := parseStringsPool(sections[SectionStringsPool])
	if err != nil {
		return Program{}, err
	}

	intsPool, err := parseIntsPool(sections[SectionIntsPool])
	if err != nil {
		return Program{}, err
	}

	maxStr, maxInt, maxPortion, maxMonetary := parseMaxRegs(sections[SectionMaxRegisters])

	return Program{
		Instructions:   instructions,
		StringsPool:    stringsPool,
		IntsPool:       intsPool,
		MaxRegString:   maxStr,
		MaxRegInt:      maxInt,
		MaxRegPortion:  maxPortion,
		MaxRegMonetary: maxMonetary,
	}, nil
}
