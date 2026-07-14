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

	data, strTable, intTable := encodePools(p.StringsPool, p.IntsPool)

	const headerLen = 4 + 4*8 // magic + 4 section pointers
	instrStart := uint32(headerLen)
	dataStart := instrStart + uint32(len(instrs))
	strTableStart := dataStart + uint32(len(data))
	intTableStart := strTableStart + uint32(len(strTable))

	buf := make([]byte, 0, int(intTableStart)+len(intTable))
	buf = append(buf, "NUMB"...)
	buf = appendSection(buf, instrStart, uint32(len(instrs)))
	buf = appendSection(buf, dataStart, uint32(len(data)))
	buf = appendSection(buf, strTableStart, uint32(len(strTable)))
	buf = appendSection(buf, intTableStart, uint32(len(intTable)))
	buf = append(buf, instrs...)
	buf = append(buf, data...)
	buf = append(buf, strTable...)
	buf = append(buf, intTable...)
	return buf
}

// TODO review AI blob
func encodePools(strings []string, ints []big.Int) (data, strTable, intTable []byte) {
	strOffsets := make([]uint32, len(strings))
	for i, s := range strings {
		strOffsets[i] = uint32(len(data))
		var lenBuf [4]byte
		le.PutUint32(lenBuf[:], uint32(len(s)))
		data = append(data, lenBuf[:]...)
		data = append(data, s...)
	}
	intOffsets := make([]uint32, len(ints))
	for i := range ints {
		intOffsets[i] = uint32(len(data))
		n := &ints[i]
		sign := byte(0)
		if n.Sign() < 0 {
			sign = 1
		}
		mag := new(big.Int).Abs(n).Bytes()
		var hdr [5]byte
		hdr[0] = sign
		le.PutUint32(hdr[1:], uint32(len(mag)))
		data = append(data, hdr[:]...)
		data = append(data, mag...)
	}
	return data, offsetTable(strOffsets), offsetTable(intOffsets)
}

func readArr[T any](
	segmentName string,
	buf []byte,
	idx *int,

	parse func(buf []byte) (T, error),
) (T, error) {
	if *idx+8 > len(buf) {
		var def_ T
		return def_, fmt.Errorf("header truncated at offset %d (while reading %s)", *idx, segmentName)
	}

	arrStart := le.Uint32(buf[*idx:])
	*idx += 4
	arrLen := le.Uint32(buf[*idx:])
	*idx += 4

	arrEnd := uint64(arrStart) + uint64(arrLen)
	if arrEnd > uint64(len(buf)) {
		var def_ T
		return def_, fmt.Errorf("section [%d:%d] exceeds buffer %d (while reading %s)", arrStart, arrEnd, len(buf), segmentName)
	}

	return parse(buf[arrStart:arrEnd])
}

func id(buf []byte) ([]byte, error) {
	return buf, nil
}

func parseInstructions(buf []byte) ([]Instruction, error) {
	// TODO check len is multiple of 4
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

// TODO this is claude-generated: double check
func parseStringsPool(poolBuf []byte, dataSegment []byte) ([]string, error) {
	if len(poolBuf)%4 != 0 {
		return nil, fmt.Errorf("string pool size %d not a multiple of 4", len(poolBuf))
	}
	dsLen := uint64(len(dataSegment))
	n := len(poolBuf) / 4
	out := make([]string, n)

	for i := range n {
		offset := uint64(le.Uint32(poolBuf[i*4:]))

		// length prefix
		if offset+4 > dsLen {
			return nil, fmt.Errorf("string %d: length prefix at %d out of bounds (data %d)", i, offset, dsLen)
		}
		strLen := uint64(le.Uint32(dataSegment[offset:]))

		start := offset + 4
		end := start + strLen // safe: each operand <= ~4.3e9, sum fits in uint64
		if end > dsLen {
			return nil, fmt.Errorf("string %d: body [%d:%d] out of bounds (data %d)", i, start, end, dsLen)
		}

		out[i] = string(dataSegment[start:end]) // copies; Program no longer references buf
	}
	return out, nil
}

// TODO this is claude-generated: double check
func parseIntsPool(poolBuf []byte, dataSegment []byte) ([]big.Int, error) {
	if len(poolBuf)%4 != 0 {
		return nil, fmt.Errorf("int pool size %d not a multiple of 4", len(poolBuf))
	}
	dsLen := uint64(len(dataSegment))
	n := len(poolBuf) / 4
	out := make([]big.Int, n)

	for i := range n {
		offset := uint64(le.Uint32(poolBuf[i*4:]))

		// header: sign byte + u32 magnitude length
		if offset+5 > dsLen {
			return nil, fmt.Errorf("int %d: header at %d out of bounds (data %d)", i, offset, dsLen)
		}
		sign := dataSegment[offset]
		magLen := uint64(le.Uint32(dataSegment[offset+1:]))

		start := offset + 5
		end := start + magLen
		if end > dsLen {
			return nil, fmt.Errorf("int %d: magnitude [%d:%d] out of bounds (data %d)", i, start, end, dsLen)
		}

		out[i].SetBytes(dataSegment[start:end]) // big-endian, unsigned magnitude
		switch sign {
		case 0:
			// non-negative
		case 1:
			out[i].Neg(&out[i])
		default:
			return nil, fmt.Errorf("int %d: invalid sign byte %d", i, sign)
		}
	}
	return out, nil
}

func DecodeProgram(buf []byte) (Program, error) {
	// 0..4 reserved for magic word
	if len(buf) < 4 || string(buf[0:4]) != "NUMB" {
		return Program{}, fmt.Errorf("bad magic")
	}

	idx := 4

	instructions, err := readArr("instructions", buf, &idx, parseInstructions) // <- TODO copy into instructions
	if err != nil {
		return Program{}, err
	}

	dataSegment, err := readArr("data segment", buf, &idx, id)
	if err != nil {
		return Program{}, err
	}

	stringsPool, err := readArr("strings pool", buf, &idx, func(buf []byte) ([]string, error) {
		return parseStringsPool(buf, dataSegment)
	})
	if err != nil {
		return Program{}, err
	}

	intsPool, err := readArr("ints", buf, &idx, func(buf []byte) ([]big.Int, error) {
		return parseIntsPool(buf, dataSegment)
	})
	if err != nil {
		return Program{}, err
	}

	return Program{
		Instructions: instructions,
		StringsPool:  stringsPool,
		IntsPool:     intsPool,
	}, nil
}
