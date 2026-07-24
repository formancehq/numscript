package vm

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProgramEncodeDecodeRoundTrip(t *testing.T) {
	prog := Program{
		Instructions: []Instruction{
			abc(Op_LoadStr, 0, 1, 2),
			bc(Op_LoadInt, 3, 1),
			abc(Op_AddInt, 4, 3, 3),
		},
		StringsPool: []string{"world", "dest", "USD/2"},
		IntsPool:    []big.Int{*big.NewInt(0), *big.NewInt(-42)},
	}
	got, err := DecodeProgram(prog.Encode())
	require.NoError(t, err)
	require.Equal(t, prog.Instructions, got.Instructions)
	require.Equal(t, prog.StringsPool, got.StringsPool)
	for i := range prog.IntsPool {
		require.Zero(t, got.IntsPool[i].Cmp(&prog.IntsPool[i]))
	}
}

func TestEmptyProgramRoundTrip(t *testing.T) {
	_, err := DecodeProgram(Program{}.Encode())
	require.NoError(t, err)
}

func TestDecodeSkipsUnknownSection(t *testing.T) {
	prog := Program{
		Instructions: []Instruction{bc(Op_LoadInt, 0, 0)},
		IntsPool:     []big.Int{*big.NewInt(7)},
	}
	buf := prog.Encode()
	// bump the section count and append an unknown (skippable) section
	le.PutUint16(buf[6:], le.Uint16(buf[6:])+1)
	buf = appendSection(buf, 0x0999, []byte("future"))

	got, err := DecodeProgram(buf)
	require.NoError(t, err)
	require.Equal(t, prog.Instructions, got.Instructions)
}

func TestDecodeRejectsUnknownRequiredSection(t *testing.T) {
	buf := Program{}.Encode()
	le.PutUint16(buf[6:], le.Uint16(buf[6:])+1)
	buf = appendSection(buf, mustUnderstandBit|0x0999, []byte("required"))

	_, err := DecodeProgram(buf)
	require.Error(t, err)
}

func TestDecodeRejectsNewerVersion(t *testing.T) {
	buf := Program{}.Encode()
	le.PutUint16(buf[4:], FormatVersion+1)

	_, err := DecodeProgram(buf)
	require.Error(t, err)
}

func TestDecodeRejectsTruncatedSection(t *testing.T) {
	buf := Program{StringsPool: []string{"abc"}}.Encode()
	_, err := DecodeProgram(buf[:len(buf)-2])
	require.Error(t, err)
}

func TestDecodeRejectsDuplicateSection(t *testing.T) {
	buf := Program{}.Encode()
	le.PutUint16(buf[6:], le.Uint16(buf[6:])+1)
	buf = appendSection(buf, SectionStringsPool, nil)

	_, err := DecodeProgram(buf)
	require.Error(t, err)
}

func TestRoundTripEdgeValues(t *testing.T) {
	prog := Program{
		StringsPool: []string{"", "héllo", "x"},
		IntsPool:    []big.Int{*big.NewInt(0), *new(big.Int).Lsh(big.NewInt(1), 300), *big.NewInt(-1)},
	}
	got, err := DecodeProgram(prog.Encode())
	require.NoError(t, err)
	require.Equal(t, prog.StringsPool, got.StringsPool)
	for i := range prog.IntsPool {
		require.Zero(t, got.IntsPool[i].Cmp(&prog.IntsPool[i]))
	}
}

func TestDecodeMalformed(t *testing.T) {
	u32 := func(v uint32) []byte {
		b := make([]byte, 4)
		le.PutUint32(b, v)
		return b
	}
	oneSection := func(tag uint16, content []byte) []byte {
		var b []byte
		b = appendFormatHeader(b, "NUMB", 1)
		return appendSection(b, tag, content)
	}
	badMagic := Program{}.Encode()
	badMagic[0] = 'X'

	cases := map[string][]byte{
		"bad magic":               badMagic,
		"short buffer":            {'N', 'U', 'M'},
		"instructions not mult 4": oneSection(SectionInstructions, []byte{1, 2, 3}),
		"string count truncated":  oneSection(SectionStringsPool, []byte{0, 0}),
		"string count absurd":     oneSection(SectionStringsPool, u32(0xFFFFFFFF)),
		"string body oob":         oneSection(SectionStringsPool, append(u32(1), u32(5)...)),
		"int count absurd":        oneSection(SectionIntsPool, u32(0xFFFFFFFF)),
		"int magnitude oob":       oneSection(SectionIntsPool, append(append(u32(1), 0), u32(5)...)),
		"int invalid sign":        oneSection(SectionIntsPool, append(append(u32(1), 2), u32(0)...)),
	}
	for name, buf := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeProgram(buf)
			require.Error(t, err)
		})
	}
}

func FuzzDecodeProgram(f *testing.F) {
	f.Add(Program{}.Encode())
	f.Add(Program{
		Instructions: []Instruction{bc(Op_LoadInt, 0, 0)},
		StringsPool:  []string{"x"},
		IntsPool:     []big.Int{*big.NewInt(1)},
	}.Encode())
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeProgram(data) // must not panic on arbitrary input
	})
}

func TestDecodeMissingPoolIsEmpty(t *testing.T) {
	// a program with only an instructions section
	var buf []byte
	buf = appendFormatHeader(buf, "NUMB", 1)
	buf = appendSection(buf, SectionInstructions, []byte{byte(Op_LoadInt), 0, 0, 0})

	got, err := DecodeProgram(buf)
	require.NoError(t, err)
	require.Empty(t, got.StringsPool)
	require.Empty(t, got.IntsPool)
}
