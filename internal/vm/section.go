package vm

import "fmt"

// v2 dropped the monetary register bank: MK_MONETARY/GET_AMOUNT/GET_ASSET are
// gone, BALANCE/ASSERT_NON_NEGATIVE_BALANCE/MONETARY_TO_STRING changed operand
// banks, META_MONETARY grew an ext word, and SectionMaxRegisters lost its 4th byte.
const FormatVersion uint16 = 2

const (
	SectionInstructions uint16 = 0x01 // NUMB only
	SectionStringsPool  uint16 = 0x02
	SectionIntsPool     uint16 = 0x03
	SectionMaxRegisters uint16 = 0x04 // NUMB only; optional, absent => every bank defaults to maxRegDefault
)

// A section tag with this bit set must be understood by the decoder: an unknown
// such tag is a hard error rather than a skipped section.
const mustUnderstandBit uint16 = 0x8000

// magic(4) + version(2) + section count(2)
const formatHeaderLen = 4 + 2 + 2

func appendFormatHeader(buf []byte, magic string, sectionCount uint16) []byte {
	buf = append(buf, magic...)
	var h [4]byte
	le.PutUint16(h[0:], FormatVersion)
	le.PutUint16(h[2:], sectionCount)
	return append(buf, h[:]...)
}

func appendSection(buf []byte, tag uint16, content []byte) []byte {
	var h [6]byte
	le.PutUint16(h[0:], tag)
	le.PutUint32(h[2:], uint32(len(content)))
	buf = append(buf, h[:]...)
	return append(buf, content...)
}

// decodeSections validates the magic and version, then walks the section list
// into a tag -> content map. Missing sections are simply absent (callers treat
// them as empty). Unknown tags are skipped unless they carry mustUnderstandBit.
func decodeSections(magic string, buf []byte, knownTags ...uint16) (map[uint16][]byte, error) {
	if len(buf) < formatHeaderLen || string(buf[0:4]) != magic {
		return nil, fmt.Errorf("bad magic (expected %q)", magic)
	}
	version := le.Uint16(buf[4:])
	if version > FormatVersion {
		return nil, fmt.Errorf("encoded by a newer numscript version (format v%d, supported up to v%d)", version, FormatVersion)
	}

	known := make(map[uint16]bool, len(knownTags))
	for _, t := range knownTags {
		known[t] = true
	}

	count := le.Uint16(buf[6:])
	idx := formatHeaderLen
	sections := make(map[uint16][]byte, count)
	for i := range count {
		if idx+6 > len(buf) {
			return nil, fmt.Errorf("section %d: header truncated at offset %d", i, idx)
		}
		tag := le.Uint16(buf[idx:])
		length := le.Uint32(buf[idx+2:])
		idx += 6

		end := uint64(idx) + uint64(length)
		if end > uint64(len(buf)) {
			return nil, fmt.Errorf("section %d (tag 0x%x): content [%d:%d] exceeds buffer %d", i, tag, idx, end, len(buf))
		}

		if !known[tag] && tag&mustUnderstandBit != 0 {
			return nil, fmt.Errorf("unknown required section tag 0x%x", tag)
		}
		if _, dup := sections[tag]; dup {
			return nil, fmt.Errorf("duplicate section tag 0x%x", tag)
		}
		sections[tag] = buf[idx:end]
		idx = int(end)
	}
	return sections, nil
}
