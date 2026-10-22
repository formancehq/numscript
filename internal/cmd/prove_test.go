package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/formancehq/numscript/internal/prove/smt"
	"github.com/stretchr/testify/require"
)

func requireZ3(t *testing.T) {
	t.Helper()
	if err := (smt.Solver{}).Available(); err != nil {
		if os.Getenv("NUMSCRIPT_REQUIRE_Z3") != "" {
			t.Fatal(err)
		}
		t.Skip(err)
	}
}

func writeProveFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "x.num")
	require.NoError(t, os.WriteFile(scriptPath, []byte(`send [USD/2 10] (
  source = @src
  destination = @dest
)`), 0o644))
	return scriptPath
}

func TestProveExitCodes(t *testing.T) {
	requireZ3(t)
	scriptPath := writeProveFixture(t)
	base := ProveArgs{Solver: "z3", Timeout: 5 * time.Second, OutFormatOpt: OutputFormatJson}

	opts := base
	opts.Props = []string{"success => received(@dest, USD/2) == 10"}
	require.NoError(t, proveScript(scriptPath, opts))

	opts = base
	opts.Props = []string{"holds: success => received(@dest, USD/2) == 10", "refuted: success"}
	require.Equal(t, exitCodeError{1}, proveScript(scriptPath, opts))

	// no --prop and no sidecar file
	require.ErrorContains(t, proveScript(scriptPath, base), "no properties")

	opts = base
	opts.Props = []string{"after(@nobody, USD/2) == 0"}
	require.ErrorContains(t, proveScript(scriptPath, opts), "invalid property")

	opts = base
	opts.Props = []string{"success =>"}
	require.ErrorContains(t, proveScript(scriptPath, opts), "invalid property")
}

func TestProvePropsFileAndEmitSMT(t *testing.T) {
	requireZ3(t)
	scriptPath := writeProveFixture(t)
	require.NoError(t, os.WriteFile(scriptPath+".props", []byte(`prop ok: before(@src, USD/2) >= 10 => success`), 0o644))
	smtDir := filepath.Join(filepath.Dir(scriptPath), "smt")
	require.NoError(t, proveScript(scriptPath, ProveArgs{Solver: "z3", Timeout: 5 * time.Second, EmitSMT: smtDir, OutFormatOpt: OutputFormatPretty}))
	emitted, err := os.ReadFile(filepath.Join(smtDir, "ok.smt2"))
	require.NoError(t, err)
	require.Contains(t, string(emitted), "(assert (not property))")

	// an unsupported script is a hard error
	require.NoError(t, os.WriteFile(scriptPath, []byte(`vars { account $d }
send [USD/2 10] (source = @src destination = $d)`), 0o644))
	require.ErrorContains(t, proveScript(scriptPath, ProveArgs{Solver: "z3", OutFormatOpt: OutputFormatPretty}), "script variables are not supported")
}
