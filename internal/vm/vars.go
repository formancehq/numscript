package vm

import (
	"math/big"
)

type Vars struct {
	StringsPool []string
	IntsPool    []big.Int
}

func DecodeVars(buf []byte) (Vars, error) {
	sections, err := decodeSections("NVAR", buf, SectionStringsPool, SectionIntsPool)
	if err != nil {
		return Vars{}, err
	}

	stringsPool, err := parseStringsPool(sections[SectionStringsPool])
	if err != nil {
		return Vars{}, err
	}

	intsPool, err := parseIntsPool(sections[SectionIntsPool])
	if err != nil {
		return Vars{}, err
	}

	return Vars{
		StringsPool: stringsPool,
		IntsPool:    intsPool,
	}, nil
}

func (v Vars) Encode() []byte {
	strs := encodeStringsPool(v.StringsPool)
	ints := encodeIntsPool(v.IntsPool)

	buf := make([]byte, 0, formatHeaderLen+2*6+len(strs)+len(ints))
	buf = appendFormatHeader(buf, "NVAR", 2)
	buf = appendSection(buf, SectionStringsPool, strs)
	buf = appendSection(buf, SectionIntsPool, ints)
	return buf
}
