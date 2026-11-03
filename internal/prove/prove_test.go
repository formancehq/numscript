package prove

import (
	"context"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/prove/encode"
	"github.com/formancehq/numscript/internal/prove/prop"
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

func loadScript(t *testing.T, path string) []ir.Instr {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	if strings.HasSuffix(path, ".ir") {
		instrs, errs := ir.Parse(string(content))
		require.Empty(t, errs)
		return instrs
	}
	parsed := parser.Parse(string(content))
	require.Empty(t, parsed.Errors)
	instrs, _, err := compiler.CompileToIR(parsed.Value, nil)
	require.NoError(t, err)
	return instrs
}

// crossCheck runs the script on random initial balances and requires the
// property to hold on each: the only guard against an unsound "valid".
func crossCheck(t *testing.T, instrs []ir.Instr, p *encode.Problem, pr prop.Property, seed int64) {
	t.Helper()
	rng := rand.New(rand.NewSource(seed))
	for trial := 0; trial < 64; trial++ {
		before := map[encode.Key]*big.Int{}
		for _, k := range p.Keys {
			var v int64
			if rng.Intn(4) == 0 {
				v = []int64{0, 50, 100, 150}[rng.Intn(4)]
			} else {
				v = int64(rng.Intn(601)) - 200
			}
			before[k] = big.NewInt(v)
		}
		replay, err := encode.Replay(context.Background(), instrs, before, p.Keys, p.MoveKeys)
		require.NoError(t, err)
		require.True(t, prop.Eval(pr.Formula, lookup(before, replay), replay.Success),
			"property %q is false on the vm for balances %v", pr.Name, before)
	}
}

func TestCases(t *testing.T) {
	requireZ3(t)
	propsFiles, err := filepath.Glob("testdata/cases/*.props")
	require.NoError(t, err)
	require.NotEmpty(t, propsFiles)

	for _, propsPath := range propsFiles {
		scriptPath := strings.TrimSuffix(propsPath, ".props")
		t.Run(filepath.Base(scriptPath), func(t *testing.T) {
			instrs := loadScript(t, scriptPath)

			propsSrc, err := os.ReadFile(propsPath)
			require.NoError(t, err)
			props, errs := prop.ParseFile(string(propsSrc))
			require.Empty(t, errs)

			expectedSrc, err := os.ReadFile(scriptPath + ".expected")
			require.NoError(t, err)
			expected := map[string]string{}
			var unsupported string
			for _, line := range strings.Split(strings.TrimSpace(string(expectedSrc)), "\n") {
				name, status, ok := strings.Cut(line, ": ")
				require.True(t, ok, "malformed expected line %q", line)
				if name == "unsupported" {
					unsupported = status
					continue
				}
				expected[name] = status
			}

			report, err := Check(context.Background(), instrs, props, Options{Solver: smt.Solver{Timeout: 10 * time.Second}})
			if unsupported != "" {
				var unsupportedErr encode.UnsupportedError
				require.ErrorAs(t, err, &unsupportedErr)
				require.Equal(t, unsupported, unsupportedErr.Reason)
				return
			}
			require.NoError(t, err)
			require.Len(t, report.Results, len(expected))

			for i, res := range report.Results {
				want, ok := expected[res.Property.Name]
				require.True(t, ok, "no expectation for property %q", res.Property.Name)
				require.Equal(t, want, res.Status.String(), "property %q: %v %v", res.Property.Name, res.Err, res.Diffs)
				if res.Status == Valid {
					crossCheck(t, instrs, report.Problem, res.Property, int64(i))
				}
			}
		})
	}
}

func TestPropertyErrors(t *testing.T) {
	requireZ3(t)
	instrs := loadScript(t, "testdata/cases/simple-send.num")
	pr, errs := prop.ParseInline("after(@nobody, USD/2) == 0")
	require.Empty(t, errs)
	_, err := Check(context.Background(), instrs, []prop.Property{pr}, Options{})
	var propErr PropertyError
	require.ErrorAs(t, err, &propErr)
	require.Len(t, propErr.Errors, 1)
	require.Equal(t, "account @nobody is not referenced by the script (known: @dest, @src)", propErr.Errors[0].Msg)
}

func TestReportOutput(t *testing.T) {
	requireZ3(t)
	instrs := loadScript(t, "testdata/cases/simple-send.num")
	src, err := os.ReadFile("testdata/cases/simple-send.num.props")
	require.NoError(t, err)
	props, errs := prop.ParseFile(string(src))
	require.Empty(t, errs)
	report, err := Check(context.Background(), instrs, props, Options{Solver: smt.Solver{Timeout: 10 * time.Second}})
	require.NoError(t, err)
	require.Equal(t, 1, report.ExitCode())
	require.True(t, *report.SuccessReachable)
	require.True(t, *report.FailureReachable)

	out := report.Pretty("simple-send.num")
	require.Contains(t, out, "success reachable: yes; failure reachable: yes")
	require.Contains(t, out, "src_never_negative: after(@src, USD/2) >= 0")
	require.Contains(t, out, "initial balances:")
	require.Contains(t, out, "missing funds")
	require.Contains(t, out, "6 valid, 1 counterexample")

	js, err := report.JSON("simple-send.num")
	require.NoError(t, err)
	require.Contains(t, string(js), `"status": "counterexample"`)
	require.Contains(t, string(js), `"outcome": "fails"`)
}
