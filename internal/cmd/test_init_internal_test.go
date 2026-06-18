package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/stretchr/testify/require"
)

const testInitNumscript = `send [USD/2 100] (
	source = @world
	destination = @bob
)
`

func TestRunTestInitCmdCreatesSpecsFile(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "main.num")
	require.NoError(t, os.WriteFile(scriptPath, []byte(testInitNumscript), 0644))

	err := runTestInitCmd(testInitArgs{path: scriptPath})

	require.NoError(t, err)

	content, err := os.ReadFile(scriptPath + ".specs.json")
	require.NoError(t, err)

	var specs specs_format.Specs
	require.NoError(t, json.Unmarshal(content, &specs))
	require.Equal(t, "https://raw.githubusercontent.com/formancehq/numscript/main/specs.schema.json", specs.Schema)
	require.Len(t, specs.TestCases, 1)
	require.Equal(t, "example spec", specs.TestCases[0].It)
}

func TestRunTestInitCmdReturnsErrorWhenScriptMissing(t *testing.T) {
	err := runTestInitCmd(testInitArgs{path: filepath.Join(t.TempDir(), "missing.num")})

	require.Error(t, err)
}

func TestRunTestInitCmdReturnsErrorOnInvalidScript(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "main.num")
	require.NoError(t, os.WriteFile(scriptPath, []byte("send [USD/2 100] ("), 0644))

	err := runTestInitCmd(testInitArgs{path: scriptPath})

	require.Error(t, err)
}

func TestRunTestInitCmdReturnsMarshalError(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "main.num")
	require.NoError(t, os.WriteFile(scriptPath, []byte(testInitNumscript), 0644))

	original := jsonMarshalIndent
	jsonMarshalIndent = func(v any, prefix, indent string) ([]byte, error) {
		return nil, errors.New("marshal failure")
	}
	t.Cleanup(func() { jsonMarshalIndent = original })

	err := runTestInitCmd(testInitArgs{path: scriptPath})

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to marshal specs file")
	require.ErrorContains(t, err, "marshal failure")
}

func TestRunTestInitCmdReturnsWriteError(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "main.num")
	require.NoError(t, os.WriteFile(scriptPath, []byte(testInitNumscript), 0644))

	// Create a directory where the specs file should be written,
	// so that os.WriteFile fails deterministically.
	require.NoError(t, os.Mkdir(scriptPath+".specs.json", 0755))

	err := runTestInitCmd(testInitArgs{path: scriptPath})

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to write specs file")
}
