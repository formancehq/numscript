package vm

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/formancehq/numscript/internal/runtime"
)

// encodeVars builds an "NVAR" payload with the same layout as the const pool:
// a shared data segment plus a per-type offset table indexing into it.
func encodeVars(strs []string, ints []*big.Int) []byte {
	var data []byte
	strOffsets := make([]uint32, len(strs))
	for i, s := range strs {
		strOffsets[i] = uint32(len(data))
		var lenBuf [4]byte
		le.PutUint32(lenBuf[:], uint32(len(s)))
		data = append(data, lenBuf[:]...)
		data = append(data, s...)
	}
	intOffsets := make([]uint32, len(ints))
	for i, n := range ints {
		intOffsets[i] = uint32(len(data))
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

	strTable := make([]byte, 4*len(strs))
	for i, off := range strOffsets {
		le.PutUint32(strTable[i*4:], off)
	}
	intTable := make([]byte, 4*len(ints))
	for i, off := range intOffsets {
		le.PutUint32(intTable[i*4:], off)
	}

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

func appendSection(buf []byte, start, length uint32) []byte {
	var b [8]byte
	le.PutUint32(b[0:], start)
	le.PutUint32(b[4:], length)
	return append(buf, b[:]...)
}

func TestDecodeVars(t *testing.T) {
	buf := encodeVars(
		[]string{"alice", "USD/2"},
		[]*big.Int{big.NewInt(1), big.NewInt(4), big.NewInt(-100)},
	)

	vars, err := DecodeVars(buf)
	if err != nil {
		t.Fatalf("DecodeVars: %v", err)
	}

	wantStrs := []string{"alice", "USD/2"}
	if !reflect.DeepEqual(vars.StringsPool, wantStrs) {
		t.Errorf("strings: got %v want %v", vars.StringsPool, wantStrs)
	}

	wantInts := []big.Int{*big.NewInt(1), *big.NewInt(4), *big.NewInt(-100)}
	if !reflect.DeepEqual(vars.IntsPool, wantInts) {
		t.Errorf("ints: got %v want %v", vars.IntsPool, wantInts)
	}
}

func TestLoadVarOpcodes(t *testing.T) {
	vars, err := DecodeVars(encodeVars(
		[]string{"world", "dest"},
		[]*big.Int{big.NewInt(42)},
	))
	if err != nil {
		t.Fatalf("DecodeVars: %v", err)
	}

	prog := Program{
		Instructions: []Instruction{
			bc(Op_LoadStr, sUSD, 0),        // r_s0 = "USD/2" (current asset)
			abc(Op_SetCurrentAsset, sUSD, 0, 0),
			bc(Op_LoadVarStr, 1, 0),        // r_s1 = var strings[0] = "world"
			bc(Op_LoadVarStr, 2, 1),        // r_s2 = var strings[1] = "dest"
			bc(Op_LoadVarInt, 0, 0),        // r_i0 = var ints[0] = 42
			abc(Op_PullAccount, 1, 1, 0),   // r_i1 = pull(world, cap r_i0)
			abc(0, nilReg, nilReg, 0),      // ext: no overdraft, no color
			abc(Op_SendToAccount, 2, nilReg, nilReg), // send to dest
		},
		StringsPool: []string{"USD/2"},
	}

	got, execErr := Exec(NewVm(prog), &vars, mockStore{})
	if execErr != nil {
		t.Fatalf("Exec: %v", execErr)
	}

	want := []runtime.Posting{
		{Source: "world", Destination: "dest", Asset: "USD/2", Amount: big.NewInt(42)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("postings mismatch\n got: %+v\nwant: %+v", got, want)
	}
}
