package cmd

import (
	"os"
	"path/filepath"
	"testing"

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
	_, err = os.Stat(scriptPath + ".specs.json")
	require.NoError(t, err)
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
