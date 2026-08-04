package compiler_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/formancehq/numscript"
	"github.com/stretchr/testify/require"
)

func TestCompilerExample(t *testing.T) {
	script := `
		vars {
			account $acc
		}

		send [USD/2 10] (
			source = $acc
			destination = @dest
		)
	`

	varsEncoder, compiledProgram, compilationErr := numscript.Compile(script)
	require.NoError(t, compilationErr) // e.g. parsing errors or type errors or any other kind of compile-time errors

	{
		// The compiledProgram represents the compiled version of the program.
		// We can serialize it into a []byte sequence and decode it back to the same data structure.
		// the serialised []byte format is meant to be used to send it over the wire
		bytecode := compiledProgram.Encode() // <- cast to []byte

		decodedCompiledProgram, decodingErr := numscript.DecodeCompiledProgram(bytecode) // <- decode it back
		require.NoError(t, decodingErr)
		require.Equal(t, decodedCompiledProgram, compiledProgram)
	}

	// the vars encoder must be stored by the leader, so that it can encode the vars payload
	// in a way that can be consumed by the vm
	vars, err := varsEncoder.Encode(map[string]string{
		"acc": "src_account",
	})
	require.NoError(t, err)

	{
		// just like the compiledProgram. the Vars can be serialised and deserialised into/from []byte
		serialisedVars := vars.Encode() // <- []byte to be sent over the wire from leader to nodes

		decodedVars, decodingErr := numscript.DecodeVars(serialisedVars) // <- turning []byte into Vars
		require.NoError(t, decodingErr)
		require.Equal(t, decodedVars, vars)
	}

	// We can initialise the vm by passing the numscript.CompiledProgram value.
	// Not only it's valid to re-use the same instance of the VM from many script runs,
	// it's actually best to keep that in memory instead of the keeping the program and re-creating the vm each time
	// this way we can avoid allocating/deallocating the registers and vm state each time
	vm := numscript.NewVm(compiledProgram)

	// mock store (repr'd as map)
	store := testStore{
		"src_account": 100,
	}

	result, execErr := numscript.ExecVm(context.Background(), vm, &vars, store)
	require.NoError(t, execErr) // e.g. missing funds, or any other runtime error
	require.Equal(t, []numscript.Posting{
		{
			Source:      "src_account",
			Destination: "dest",
			Asset:       "USD/2",
			Amount:      big.NewInt(10),
		},
	}, result.Postings)
}

type testStore map[string]int64

func (s testStore) GetBalance(ctx context.Context, account, scope, asset, color string) (*big.Int, error) {
	return big.NewInt(s[account]), nil
}

func (testStore) GetMetadata(ctx context.Context, account, scope, key string) (string, bool, error) {
	return "", false, nil
}
