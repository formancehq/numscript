package vm

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/runtime"
	"github.com/stretchr/testify/require"
)

func TestVarsRoundTrip(t *testing.T) {
	in := Vars{
		StringsPool: []string{"alice", "USD/2"},
		IntsPool:    []big.Int{*big.NewInt(1), *big.NewInt(4), *big.NewInt(-100)},
	}

	out, err := DecodeVars(in.Encode())
	require.NoError(t, err)
	require.Equal(t, in, out)
}

func TestVarsRoundTripEdgeValues(t *testing.T) {
	in := Vars{
		StringsPool: []string{"", "héllo", "x"},
		IntsPool:    []big.Int{*big.NewInt(0), *new(big.Int).Lsh(big.NewInt(1), 300), *big.NewInt(-1)},
	}
	out, err := DecodeVars(in.Encode())
	require.NoError(t, err)
	require.Equal(t, in.StringsPool, out.StringsPool)
	for i := range in.IntsPool {
		require.Zero(t, out.IntsPool[i].Cmp(&in.IntsPool[i]))
	}
}

func TestDecodeVarsMalformed(t *testing.T) {
	u32 := func(v uint32) []byte {
		b := make([]byte, 4)
		le.PutUint32(b, v)
		return b
	}
	oneSection := func(tag uint16, content []byte) []byte {
		var b []byte
		b = appendFormatHeader(b, "NVAR", 1)
		return appendSection(b, tag, content)
	}
	badMagic := Vars{}.Encode()
	badMagic[0] = 'X'

	newerVersion := Vars{}.Encode()
	le.PutUint16(newerVersion[4:], FormatVersion+1)

	cases := map[string][]byte{
		"bad magic":              badMagic,
		"short buffer":           {'N', 'V', 'A'},
		"newer version":          newerVersion,
		"string count truncated": oneSection(SectionStringsPool, []byte{0, 0}),
		"string count absurd":    oneSection(SectionStringsPool, u32(0xFFFFFFFF)),
		"string body oob":        oneSection(SectionStringsPool, append(u32(1), u32(5)...)),
		"int count absurd":       oneSection(SectionIntsPool, u32(0xFFFFFFFF)),
		"int magnitude oob":      oneSection(SectionIntsPool, append(append(u32(1), 0), u32(5)...)),
		"int invalid sign":       oneSection(SectionIntsPool, append(append(u32(1), 2), u32(0)...)),
	}
	for name, buf := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeVars(buf)
			require.Error(t, err)
		})
	}
}

func FuzzDecodeVars(f *testing.F) {
	f.Add(Vars{}.Encode())
	f.Add(Vars{
		StringsPool: []string{"x"},
		IntsPool:    []big.Int{*big.NewInt(1)},
	}.Encode())
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeVars(data) // must not panic on arbitrary input
	})
}

func TestLoadVarOpcodes(t *testing.T) {
	vars, err := DecodeVars(Vars{
		StringsPool: []string{"world", "dest"},
		IntsPool:    []big.Int{*big.NewInt(42)},
	}.Encode())
	require.NoError(t, err)

	prog := Program{
		Instructions: []Instruction{
			bc(Op_LoadStr, sUSD, 0), // r_s0 = "USD/2" (current asset)
			abc(Op_SetCurrentAsset, sUSD, 0, 0),
			bc(Op_LoadVarStr, 1, 0),                  // r_s1 = var strings[0] = "world"
			bc(Op_LoadVarStr, 2, 1),                  // r_s2 = var strings[1] = "dest"
			bc(Op_LoadVarInt, 0, 0),                  // r_i0 = var ints[0] = 42
			abc(Op_PullAccount, 1, 1, 0),             // r_i1 = pull(world, cap r_i0)
			abc(0, nilReg, nilReg, 0),                // ext: no overdraft, no color
			abc(Op_SendToAccount, 2, nilReg, nilReg), // send to dest
		},
		StringsPool: []string{"USD/2"},
	}

	res, execErr := Exec(context.Background(), NewVm(prog), &vars, mockStore{})
	require.Nil(t, execErr)

	want := []runtime.Posting{
		{Source: "world", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(42)},
	}
	require.Equal(t, want, res.Postings)
}
