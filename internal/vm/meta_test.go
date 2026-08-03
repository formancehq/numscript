package vm

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/runtime"
	"github.com/stretchr/testify/require"
)

func TestSetAccountMeta(t *testing.T) {
	prog := Program{
		Instructions: []Instruction{
			bc(Op_LoadStr, 0, 0),            // r_s0 = "acc"
			bc(Op_LoadStr, 1, 1),            // r_s1 = "k"
			bc(Op_LoadStr, 2, 2),            // r_s2 = "v"
			abc(Op_SetAccountMeta, 0, 1, 2), // set_account_meta(acc, k, v)
		},
		StringsPool: []string{"acc", "k", "v"},
	}

	res, execErr := Exec(context.Background(), NewVm(prog), nil, mockStore{})
	require.Nil(t, execErr)
	require.Equal(t, runtime.AccountsMetadata{"acc": {"k": "v"}}, res.AccountsMetadata)
}

func TestMetaStr(t *testing.T) {
	// meta("config", "beneficiary") == "alice"; then send [USD/2 100] from world to it.
	prog := Program{
		Instructions: []Instruction{
			bc(Op_LoadStr, 0, 0), // s0 = "USD/2"
			abc(Op_SetCurrentAsset, 0, 0, 0),
			bc(Op_LoadStr, 1, 1),                     // s1 = "config"
			bc(Op_LoadStr, 2, 2),                     // s2 = "beneficiary"
			abc(Op_MetaStr, 3, 1, 2),                 // s3 = meta(config, beneficiary) = "alice"
			bc(Op_LoadStr, 4, 3),                     // s4 = "world"
			bc(Op_LoadInt, 0, 0),                     // i0 = 100 (cap)
			abc(Op_PullAccount, 1, 4, 0),             // i1 = pull(world, cap i0)
			abc(0, nilReg, nilReg, 0),                // ext: no overdraft, no color
			abc(Op_SendToAccount, 3, nilReg, nilReg), // send to s3 (alice)
		},
		StringsPool: []string{"USD/2", "config", "beneficiary", "world"},
		IntsPool:    []big.Int{*big.NewInt(100)},
	}

	store := mockStore{meta: map[string]map[string]string{
		"config": {"beneficiary": "alice"},
	}}

	res, execErr := Exec(context.Background(), NewVm(prog), nil, store)
	require.Nil(t, execErr)
	require.Equal(t, []runtime.Posting{
		{Source: "world", Destination: "alice", Asset: "USD/2", Amount: big.NewInt(100)},
	}, res.Postings)
}
