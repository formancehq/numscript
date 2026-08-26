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
// way ($schema is an editor hint, not part of the format).
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

// e2eLegacyV0024Numscript and e2eLegacyV0024Specs are the exact script and
// specs file v0.0.24 (commit 1b42b98, the last tagged release) generated:
// balances/metadata as nested maps rather than today's row arrays.
const e2eLegacyV0024Numscript = `vars {
	account $sale
	account $seller = meta($sale, "seller")
	portion $commission = meta($seller, "commission")
}
send [EUR/2 100] (
	source = $sale
	destination = {
		remaining to $seller
		$commission to @platform
	}
)
`

const e2eLegacyV0024Specs = `{
	"testCases": [
		{
			"it": "-",
			"balances": {
				"sales:042": { "EUR/2": 2500 },
				"users:053": { "EUR/2": 500 }
			},
			"variables": { "sale": "sales:042" },
			"metadata": {
				"sales:042": { "seller": "users:053" },
				"users:053": { "commission": "12.5%" }
			},
			"expect.postings": [
				{ "source": "sales:042", "destination": "users:053", "amount": 88, "asset": "EUR/2" },
				{ "source": "sales:042", "destination": "platform", "amount": 12, "asset": "EUR/2" }
			]
		}
	]
}
`

func TestE2ETestRejectsLegacyV0024Shape(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.num"), []byte(e2eLegacyV0024Numscript), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.num.specs.json"), []byte(e2eLegacyV0024Specs), 0644))

	cmd := exec.Command(e2eBinaryPath, "test", dir)
	out, err := cmd.CombinedOutput()

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
	require.Contains(t, string(out), "cannot unmarshal object")
}

func TestE2ETestMigrateConvertsLegacyV0024Shape(t *testing.T) {
	dir := t.TempDir()
	specsPath := filepath.Join(dir, "main.num.specs.json")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "main.num"), []byte(e2eLegacyV0024Numscript), 0644))
	require.NoError(t, os.WriteFile(specsPath, []byte(e2eLegacyV0024Specs), 0644))

	cmd := exec.Command(e2eBinaryPath, "test", "--migrate", dir)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	require.Contains(t, string(out), "migrated")

	migrated, err := os.ReadFile(specsPath)
	require.NoError(t, err)
	require.Contains(t, string(migrated), `"account": "sales:042"`)
	require.Contains(t, string(migrated), `"asset": "EUR/2"`)

	rerun := exec.Command(e2eBinaryPath, "test", dir)
	rerunOut, err := rerun.CombinedOutput()
	require.NoError(t, err, string(rerunOut))
}
