package vm

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytecodeVersionCanRead(t *testing.T) {
	cases := []struct {
		reader, encoded BytecodeVersion
		ok              bool
	}{
		{BytecodeVersion{1, 0}, BytecodeVersion{1, 0}, true},
		{BytecodeVersion{1, 1}, BytecodeVersion{1, 0}, true},  // newer reader runs older minor
		{BytecodeVersion{1, 0}, BytecodeVersion{1, 1}, false}, // older reader does not assume it knows the new opcodes
		{BytecodeVersion{2, 0}, BytecodeVersion{1, 0}, false}, // major bump: meanings changed
		{BytecodeVersion{2, 0}, BytecodeVersion{1, 9}, false},
		{BytecodeVersion{1, 9}, BytecodeVersion{2, 0}, false},
		{BytecodeVersion{2, 0}, BytecodeVersion{2, 0}, true},
		{BytecodeVersion{1, 300}, BytecodeVersion{1, 299}, true}, // minor is a full u16, not a byte
		// 0.x is unstable: only the exact version is readable
		{BytecodeVersion{0, 1}, BytecodeVersion{0, 1}, true},
		{BytecodeVersion{0, 2}, BytecodeVersion{0, 1}, false},
		{BytecodeVersion{0, 1}, BytecodeVersion{0, 2}, false},
		{BytecodeVersion{1, 0}, BytecodeVersion{0, 9}, false},
		{BytecodeVersion{0, 9}, BytecodeVersion{1, 0}, false},
	}
	for _, tc := range cases {
		t.Run(tc.reader.String()+" reads "+tc.encoded.String(), func(t *testing.T) {
			require.Equal(t, tc.ok, tc.reader.CanRead(tc.encoded))
		})
	}
}

// The header layout is pinned: the 4-byte magic, then major and minor as two
// little-endian u16 fields, then the u16 section count — 10 bytes in all.
func TestHeaderLayout(t *testing.T) {
	require.Equal(t, 10, formatHeaderLen)

	program := Program{}.Encode()
	require.Equal(t, "NUMB", string(program[:4]))
	require.Equal(t, CurrentBytecodeVersion.Major, le.Uint16(program[4:]))
	require.Equal(t, CurrentBytecodeVersion.Minor, le.Uint16(program[6:]))
	require.Equal(t, uint16(4), le.Uint16(program[8:]), "Program.Encode writes four sections")

	vars := Vars{}.Encode()
	require.Equal(t, "NVAR", string(vars[:4]))
	require.Equal(t, CurrentBytecodeVersion.Major, le.Uint16(vars[4:]))
	require.Equal(t, CurrentBytecodeVersion.Minor, le.Uint16(vars[6:]))
	require.Equal(t, uint16(2), le.Uint16(vars[8:]), "Vars.Encode writes two sections")

	require.Equal(t, "2.1", BytecodeVersion{2, 1}.String())
}

// Encode stamps CurrentBytecodeVersion whatever the struct's own Version says,
// and both the peek and the decode read it back.
func TestEncodeWritesCurrentVersion(t *testing.T) {
	stale := BytecodeVersion{9, 9}

	programBytes := Program{Version: stale}.Encode()
	peeked, err := PeekProgramVersion(programBytes)
	require.NoError(t, err)
	require.Equal(t, CurrentBytecodeVersion, peeked)

	program, err := DecodeProgram(programBytes)
	require.NoError(t, err)
	require.Equal(t, CurrentBytecodeVersion, program.Version)

	varsBytes := Vars{Version: stale}.Encode()
	peeked, err = PeekVarsVersion(varsBytes)
	require.NoError(t, err)
	require.Equal(t, CurrentBytecodeVersion, peeked)

	vars, err := DecodeVars(varsBytes)
	require.NoError(t, err)
	require.Equal(t, CurrentBytecodeVersion, vars.Version)
}

func TestPeekVersionChecksMagicOnly(t *testing.T) {
	_, err := PeekProgramVersion(Vars{}.Encode())
	require.Error(t, err, "a vars blob is not a program")

	_, err = PeekVarsVersion(Program{}.Encode())
	require.Error(t, err, "a program blob is not vars")

	_, err = PeekProgramVersion([]byte("NUM"))
	require.Error(t, err)

	_, err = PeekProgramVersion(Program{}.Encode()[:formatHeaderLen-1])
	require.Error(t, err, "a header short of the section count is not a header")
}

// The decoders apply CanRead against CurrentBytecodeVersion and report an
// unreadable blob with the typed error carrying both versions, while the peek
// still returns the encoded version so a caller can say what it was.
func TestDecodeAppliesVersionRule(t *testing.T) {
	withVersion := func(buf []byte, v BytecodeVersion) []byte {
		out := bytes.Clone(buf)
		le.PutUint16(out[4:], v.Major)
		le.PutUint16(out[6:], v.Minor)
		return out
	}

	current := CurrentBytecodeVersion

	type version struct {
		v  BytecodeVersion
		ok bool
	}
	versions := map[string]version{
		"current":         {current, true},
		"newer minor":     {BytecodeVersion{current.Major, current.Minor + 1}, false},
		"far newer minor": {BytecodeVersion{current.Major, 0xFFFF}, false},
		"newer major":     {BytecodeVersion{current.Major + 1, 0}, false},
	}
	if current.Major > 0 {
		versions["older major"] = version{BytecodeVersion{current.Major - 1, current.Minor}, false}
	}
	if current.Minor > 0 {
		// an unstable reader rejects older minors too
		versions["older minor"] = version{BytecodeVersion{current.Major, current.Minor - 1}, current.Major > 0}
	}

	blobs := []struct {
		name   string
		magic  string
		buf    []byte
		decode func([]byte) (BytecodeVersion, error)
	}{
		{"program", "NUMB", Program{}.Encode(), func(b []byte) (BytecodeVersion, error) {
			p, err := DecodeProgram(b)
			return p.Version, err
		}},
		{"vars", "NVAR", Vars{}.Encode(), func(b []byte) (BytecodeVersion, error) {
			v, err := DecodeVars(b)
			return v.Version, err
		}},
	}

	for name, tc := range versions {
		for _, blob := range blobs {
			t.Run(name+"/"+blob.name, func(t *testing.T) {
				buf := withVersion(blob.buf, tc.v)

				peeked, err := peekVersion(blob.magic, buf)
				require.NoError(t, err)
				require.Equal(t, tc.v, peeked)

				got, err := blob.decode(buf)
				if tc.ok {
					require.NoError(t, err)
					require.Equal(t, tc.v, got)
					return
				}

				var unsupported UnsupportedBytecodeVersionError
				require.ErrorAs(t, err, &unsupported)
				require.Equal(t, tc.v, unsupported.Encoded)
				require.Equal(t, current, unsupported.Supported)
				require.Contains(t, err.Error(), tc.v.String())
			})
		}
	}
}
