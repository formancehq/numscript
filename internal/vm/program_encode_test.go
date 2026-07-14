package vm

import (
	"math/big"
	"reflect"
	"testing"
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
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(got.Instructions, prog.Instructions) {
		t.Fatalf("instructions mismatch:\n got %+v\nwant %+v", got.Instructions, prog.Instructions)
	}
	if !reflect.DeepEqual(got.StringsPool, prog.StringsPool) {
		t.Fatalf("strings mismatch: %+v vs %+v", got.StringsPool, prog.StringsPool)
	}
	for i := range prog.IntsPool {
		if got.IntsPool[i].Cmp(&prog.IntsPool[i]) != 0 {
			t.Fatalf("int %d: %s vs %s", i, got.IntsPool[i].String(), prog.IntsPool[i].String())
		}
	}
}

func TestEmptyProgramRoundTrip(t *testing.T) {
	if _, err := DecodeProgram(Program{}.Encode()); err != nil {
		t.Fatalf("empty round-trip: %v", err)
	}
}
