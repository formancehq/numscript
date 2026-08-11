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

func TestRunTestInitCmdPropagatesWriteError(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "main.num")
	require.NoError(t, os.WriteFile(scriptPath, []byte(testInitNumscript), 0644))

	// create a directory where the specs file should be written, so that
	// os.WriteFile fails deterministically
	require.NoError(t, os.Mkdir(scriptPath+".specs.json", 0755))

	err := runTestInitCmd(testInitArgs{path: scriptPath})

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to write specs file")
}
