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

	buf := make([]byte, 0, formatHeaderLen+3*6+len(instrs)+len(strs)+len(ints))
	buf = appendFormatHeader(buf, "NUMB", 3)
	buf = appendSection(buf, SectionInstructions, instrs)
	buf = appendSection(buf, SectionStringsPool, strs)
	buf = appendSection(buf, SectionIntsPool, ints)
	return buf
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
	sections, err := decodeSections("NUMB", buf, SectionInstructions, SectionStringsPool, SectionIntsPool)
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

	return Program{
		Instructions: instructions,
		StringsPool:  stringsPool,
		IntsPool:     intsPool,
	}, nil
}
