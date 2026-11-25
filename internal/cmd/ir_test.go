package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/formancehq/numscript/internal/ir"
	"github.com/stretchr/testify/require"
)

func TestDumpIRRoundTrips(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "x.num")
	require.NoError(t, os.WriteFile(scriptPath, []byte(`send [USD/2 10] (
  source = @src
  destination = @dest
)`), 0o644))

	outPath := filepath.Join(dir, "x.ir")
	require.NoError(t, dumpIR(scriptPath, IRArgs{OutputPath: outPath}))

	text, err := os.ReadFile(outPath)
	require.NoError(t, err)

	instrs, errs := ir.Parse(string(text))
	require.Empty(t, errs)
	require.Equal(t, string(text), ir.Dump(instrs))
	require.NoError(t, ir.Typecheck(instrs))
}

func TestDumpIRReportsCompileErrors(t *testing.T) {
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "x.num")
	require.NoError(t, os.WriteFile(scriptPath, []byte(`send [USD/2 10] (
  source = @src
  destination = $dest
)`), 0o644))

	err := dumpIR(scriptPath, IRArgs{OutputPath: filepath.Join(dir, "x.ir")})
	require.EqualError(t, err, "compilation failed")
}
