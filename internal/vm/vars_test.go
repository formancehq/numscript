package vm

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/formancehq/numscript/internal/runtime"
)

func TestVarsRoundTrip(t *testing.T) {
	in := Vars{
		StringsPool: []string{"alice", "USD/2"},
		IntsPool:    []big.Int{*big.NewInt(1), *big.NewInt(4), *big.NewInt(-100)},
	}

	out, err := DecodeVars(in.Encode())
	if err != nil {
		t.Fatalf("DecodeVars: %v", err)
	}

	if !reflect.DeepEqual(out, in) {
		t.Errorf("round trip mismatch\n got: %+v\nwant: %+v", out, in)
	}
}

func TestLoadVarOpcodes(t *testing.T) {
	vars, err := DecodeVars(Vars{
		StringsPool: []string{"world", "dest"},
		IntsPool:    []big.Int{*big.NewInt(42)},
	}.Encode())
	if err != nil {
		t.Fatalf("DecodeVars: %v", err)
	}

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
