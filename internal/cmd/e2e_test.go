package cmd_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// e2eBinaryPath is the real numscript binary, built once in TestMain. Unlike
// the rest of this package's tests, which call cmd/specs_format functions
// in-process, these tests exec the actual CLI: runTestCmd calls os.Exit
// directly, so it can only be observed as a subprocess, and this is also the
// only place that exercises cobra's flag wiring (e.g. --migrate) end to end.
var e2eBinaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "numscript-e2e-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	e2eBinaryPath = filepath.Join(dir, "numscript")
	buildCmd := exec.Command("go", "build", "-o", e2eBinaryPath, "github.com/formancehq/numscript/cmd/numscript")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to build numscript binary: %s\n%s\n", err, out)
		_ = os.RemoveAll(dir)
		os.Exit(1)
	}

	// os.Exit below bypasses defers, so clean up explicitly beforehand rather
	// than deferring it.
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

const e2eNumscript = `send [USD/2 100] (
	source = @world
	destination = @dest
)
set_account_meta(@dest, "k", [USD/2 100])
`

// e2eSpecsMissingSchema deliberately has no $schema: that's the only thing
// --migrate fixes, and it should not affect whether the tests pass either
// way.
const e2eSpecsMissingSchema = `{
	"testCases": [
		{
			"it": "-",
			"expect.metadata": [
				{ "account": "dest", "key": "k", "value": "USD/2 100" }
			],
			"expect.postings": null
		}
	]
}
`

func writeE2ESpecsFixture(t *testing.T) (dir string, specsPath string) {
	dir = t.TempDir()
	numscriptPath := filepath.Join(dir, "main.num")
	specsPath = numscriptPath + ".specs.json"

	require.NoError(t, os.WriteFile(numscriptPath, []byte(e2eNumscript), 0644))
	require.NoError(t, os.WriteFile(specsPath, []byte(e2eSpecsMissingSchema), 0644))

	return dir, specsPath
}

func TestE2ETestPassesWithoutSchema(t *testing.T) {
	dir, _ := writeE2ESpecsFixture(t)

	cmd := exec.Command(e2eBinaryPath, "test", dir)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func TestE2ETestMigrateAddsSchema(t *testing.T) {
	dir, specsPath := writeE2ESpecsFixture(t)

	cmd := exec.Command(e2eBinaryPath, "test", "--migrate", dir)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	require.Contains(t, string(out), "migrated")

	migrated, err := os.ReadFile(specsPath)
	require.NoError(t, err)
	require.Contains(t, string(migrated), `"$schema"`)
	require.Contains(t, string(migrated), `"value": "USD/2 100"`)

	// running again against the now-migrated file passes cleanly, with no
	// further migration needed.
	rerun := exec.Command(e2eBinaryPath, "test", dir)
	rerunOut, err := rerun.CombinedOutput()
	require.NoError(t, err, string(rerunOut))
}
