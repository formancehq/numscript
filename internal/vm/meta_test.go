package vm

import (
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

	res, execErr := Exec(NewVm(prog), nil, mockStore{})
	require.Nil(t, execErr)
	require.Equal(t, runtime.AccountsMetadata{"acc": {"k": "v"}}, res.AccountsMetadata)
}
