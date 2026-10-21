package vm

import "fmt"

// BytecodeVersion identifies the bytecode wire format a program or vars blob
// was encoded with, as major.minor — two 16-bit header fields, major first.
// It versions the bytecode format only: the compiler and the library are
// versioned independently of it, and a new compiler release does not imply a
// new bytecode version.
//
// The split carries the compatibility rule (see CanRead). A minor bump is
// additive — new opcodes, new sections, new optional operands — and leaves the
// meaning of everything an older writer could produce untouched, so a 1.1
// reader runs 1.0 bytecode. A 1.0 reader does not accept 1.1 bytecode: it may
// happen to know every opcode a given blob uses, but that is not assumed. A
// major bump changes the meaning of existing encodings, so a 2.0 reader
// accepts no 1.x blob at all.
//
// Major 0 is unstable: any 0.x change may change the meaning of existing
// encodings, so a 0.x reader accepts exactly its own version, older and newer
// minors alike. A host holding a 0.x blob that this build rejects should
// recompile the script from source rather than fail.
type BytecodeVersion struct {
	Major uint16
	Minor uint16
}

// CurrentBytecodeVersion is the version Encode writes and the newest one the
// decoders read.
var CurrentBytecodeVersion = BytecodeVersion{Major: 0, Minor: 1}

// CanRead reports whether a reader at version v accepts a blob encoded with
// version encoded: the same major, and a minor no newer than the reader's. An
// unstable (0.x) reader accepts only its exact version.
func (v BytecodeVersion) CanRead(encoded BytecodeVersion) bool {
	if v.Major == 0 {
		return encoded == v
	}
	return encoded.Major == v.Major && encoded.Minor <= v.Minor
}

func (v BytecodeVersion) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

// UnsupportedBytecodeVersionError is what the decoders return for a blob this
// build cannot read: another major, a minor newer than CurrentBytecodeVersion,
// or, while the current version is unstable, any other 0.x minor.
type UnsupportedBytecodeVersionError struct {
	Encoded   BytecodeVersion
	Supported BytecodeVersion
}

func (e UnsupportedBytecodeVersionError) Error() string {
	if e.Supported.Major == 0 {
		return fmt.Sprintf("bytecode version %s is not readable by this build, which reads only the unstable version %s",
			e.Encoded, e.Supported)
	}
	return fmt.Sprintf("bytecode version %s is not readable by this build, which reads %d.0 through %s",
		e.Encoded, e.Supported.Major, e.Supported)
}

const (
	SectionInstructions uint16 = 0x01 // NUMB only
	SectionStringsPool  uint16 = 0x02
	SectionIntsPool     uint16 = 0x03
	SectionMaxRegisters uint16 = 0x04 // NUMB only; optional, absent => every bank defaults to maxRegDefault
)

// A section tag with this bit set must be understood by the decoder: an unknown
// such tag is a hard error rather than a skipped section.
const mustUnderstandBit uint16 = 0x8000

// magic(4) + major(2) + minor(2) + section count(2)
const formatHeaderLen = 4 + 2 + 2 + 2

func appendFormatHeader(buf []byte, magic string, sectionCount uint16) []byte {
	buf = append(buf, magic...)
	var h [6]byte
	le.PutUint16(h[0:], CurrentBytecodeVersion.Major)
	le.PutUint16(h[2:], CurrentBytecodeVersion.Minor)
	le.PutUint16(h[4:], sectionCount)
	return append(buf, h[:]...)
}

func appendSection(buf []byte, tag uint16, content []byte) []byte {
	var h [6]byte
	le.PutUint16(h[0:], tag)
	le.PutUint32(h[2:], uint32(len(content)))
	buf = append(buf, h[:]...)
	return append(buf, content...)
}

// peekVersion validates the magic and returns the header's version without
// checking that this build can read it, so a caller can report which version
// a blob it cannot read was written with.
func peekVersion(magic string, buf []byte) (BytecodeVersion, error) {
	if len(buf) < formatHeaderLen || string(buf[0:4]) != magic {
		return BytecodeVersion{}, fmt.Errorf("bad magic (expected %q)", magic)
	}
	return BytecodeVersion{Major: le.Uint16(buf[4:]), Minor: le.Uint16(buf[6:])}, nil
}

// decodeSections validates the magic and version, then walks the section list
// into a tag -> content map. Missing sections are simply absent (callers treat
// them as empty). Unknown tags are skipped unless they carry mustUnderstandBit.
// The encoded version is returned alongside, for callers that want to record
// which version a decoded value was written by.
func decodeSections(magic string, buf []byte, knownTags ...uint16) (map[uint16][]byte, BytecodeVersion, error) {
	version, err := peekVersion(magic, buf)
	if err != nil {
		return nil, BytecodeVersion{}, err
	}
	if !CurrentBytecodeVersion.CanRead(version) {
		return nil, BytecodeVersion{}, UnsupportedBytecodeVersionError{Encoded: version, Supported: CurrentBytecodeVersion}
	}

	known := make(map[uint16]bool, len(knownTags))
	for _, t := range knownTags {
		known[t] = true
	}

	count := le.Uint16(buf[8:])
	idx := formatHeaderLen
	sections := make(map[uint16][]byte, count)
	for i := range count {
		if idx+6 > len(buf) {
			return nil, BytecodeVersion{}, fmt.Errorf("section %d: header truncated at offset %d", i, idx)
		}
		tag := le.Uint16(buf[idx:])
		length := le.Uint32(buf[idx+2:])
		idx += 6

		end := uint64(idx) + uint64(length)
		if end > uint64(len(buf)) {
			return nil, BytecodeVersion{}, fmt.Errorf("section %d (tag 0x%x): content [%d:%d] exceeds buffer %d", i, tag, idx, end, len(buf))
		}

		if !known[tag] && tag&mustUnderstandBit != 0 {
			return nil, BytecodeVersion{}, fmt.Errorf("unknown required section tag 0x%x", tag)
		}
		if _, dup := sections[tag]; dup {
			return nil, BytecodeVersion{}, fmt.Errorf("duplicate section tag 0x%x", tag)
		}
		sections[tag] = buf[idx:end]
		idx = int(end)
	}
	return sections, version, nil
}
