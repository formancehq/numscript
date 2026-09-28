package vm

import (
	"math/big"
)

type Vars struct {
	StringsPool []string
	IntsPool    []big.Int

	// Version is the wire format version these vars were built for, or (after
	// DecodeVars) the version they were actually encoded with. Encode always
	// writes the current FormatVersion regardless of this field.
	Version uint16
}

func DecodeVars(buf []byte) (Vars, error) {
	sections, version, err := decodeSections("NVAR", buf, SectionStringsPool, SectionIntsPool)
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
		Version:     version,
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
