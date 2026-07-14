package vm

import (
	"fmt"
	"math/big"
)

type Vars struct {
	StringsPool []string
	IntsPool    []big.Int
}

func DecodeVars(buf []byte) (Vars, error) {
	if len(buf) < 4 || string(buf[0:4]) != "NVAR" {
		return Vars{}, fmt.Errorf("bad vars magic")
	}

	idx := 4

	dataSegment, err := readArr("data segment", buf, &idx, id)
	if err != nil {
		return Vars{}, err
	}

	stringsPool, err := readArr("strings pool", buf, &idx, func(buf []byte) ([]string, error) {
		return parseStringsPool(buf, dataSegment)
	})
	if err != nil {
		return Vars{}, err
	}

	intsPool, err := readArr("ints", buf, &idx, func(buf []byte) ([]big.Int, error) {
		return parseIntsPool(buf, dataSegment)
	})
	if err != nil {
		return Vars{}, err
	}

	return Vars{
		StringsPool: stringsPool,
		IntsPool:    intsPool,
	}, nil
}

func (v Vars) Encode() []byte {
	data, strTable, intTable := encodePools(v.StringsPool, v.IntsPool)

	const headerLen = 4 + 3*8 // magic + 3 section pointers
	dataStart := uint32(headerLen)
	strTableStart := dataStart + uint32(len(data))
	intTableStart := strTableStart + uint32(len(strTable))

	buf := make([]byte, 0, int(intTableStart)+len(intTable))
	buf = append(buf, "NVAR"...)
	buf = appendSection(buf, dataStart, uint32(len(data)))
	buf = appendSection(buf, strTableStart, uint32(len(strTable)))
	buf = appendSection(buf, intTableStart, uint32(len(intTable)))
	buf = append(buf, data...)
	buf = append(buf, strTable...)
	buf = append(buf, intTable...)
	return buf
}

func offsetTable(offsets []uint32) []byte {
	table := make([]byte, 4*len(offsets))
	for i, off := range offsets {
		le.PutUint32(table[i*4:], off)
	}
	return table
}

func appendSection(buf []byte, start, length uint32) []byte {
	var b [8]byte
	le.PutUint32(b[0:], start)
	le.PutUint32(b[4:], length)
	return append(buf, b[:]...)
}
