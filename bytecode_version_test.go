package numscript_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/formancehq/numscript"
	"github.com/stretchr/testify/require"
)

// The public surface a host needs to keep stored bytecode and the executing
// build in step: the current version, the version stamped on what Compile and
// Encode produce, the predicates that tell whether this build can run stored
// bytes, a peek that reads the version off them, and the typed rejection.
func TestBytecodeVersionPublicAPI(t *testing.T) {
	varsEncoder, program, err := numscript.Compile(`vars {
  monetary $amt
}

send $amt (
  source = @src
  destination = @dst
)`)
	require.NoError(t, err)
	require.Equal(t, numscript.CurrentBytecodeVersion, program.Version)

	programBytes := program.Encode()
	peeked, err := numscript.PeekCompiledProgramVersion(programBytes)
	require.NoError(t, err)
	require.Equal(t, numscript.CurrentBytecodeVersion, peeked)

	decoded, err := numscript.DecodeCompiledProgram(programBytes)
	require.NoError(t, err)
	require.Equal(t, numscript.CurrentBytecodeVersion, decoded.Version)

	vars, err := varsEncoder.Encode(map[string]string{"amt": "USD/2 100"})
	require.NoError(t, err)
	require.Equal(t, numscript.CurrentBytecodeVersion, vars.Version)

	varsBytes := vars.Encode()
	require.True(t, numscript.CanReadCompiledProgram(programBytes))
	require.True(t, numscript.CanReadVars(varsBytes))

	peeked, err = numscript.PeekVarsVersion(varsBytes)
	require.NoError(t, err)
	require.Equal(t, numscript.CurrentBytecodeVersion, peeked)

	// A blob from the next major is reported, not misread: the peek still
	// tells which version it is, the decoders reject it with the typed error.
	next := numscript.BytecodeVersion{Major: numscript.CurrentBytecodeVersion.Major + 1}
	foreign := bytes.Clone(programBytes)
	binary.LittleEndian.PutUint16(foreign[4:], next.Major)
	binary.LittleEndian.PutUint16(foreign[6:], next.Minor)

	peeked, err = numscript.PeekCompiledProgramVersion(foreign)
	require.NoError(t, err)
	require.Equal(t, next, peeked)
	require.False(t, numscript.CurrentBytecodeVersion.CanRead(peeked))
	require.False(t, numscript.CanReadCompiledProgram(foreign))

	_, err = numscript.DecodeCompiledProgram(foreign)
	var unsupported numscript.UnsupportedBytecodeVersionError
	require.ErrorAs(t, err, &unsupported)
	require.Equal(t, next, unsupported.Encoded)
	require.Equal(t, numscript.CurrentBytecodeVersion, unsupported.Supported)
}
