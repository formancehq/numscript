package numscript_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/formancehq/numscript"
	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/stretchr/testify/require"
)

// See compiler-architecture.md, "Frozen bytecode".
const (
	bytecodeCorpusDir   = "testdata/bytecode"
	bytecodeFixturesDir = "internal/interpreter/testdata/script-tests"

	corpusProgramExt   = ".numb"
	corpusVarsFile     = "vars.nvar"
	corpusBalancesFile = "balances.json"
	corpusMetadataFile = "metadata.json"
	corpusPostingsFile = "postings.json"
	corpusErrorFile    = "error.json"
)

type corpusError struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

type corpusStore struct {
	balances map[[4]string]*big.Int
	metadata map[[3]string]string
}

func newCorpusStore(balances interpreter.Balances, metadata interpreter.AccountsMetadata) corpusStore {
	s := corpusStore{balances: map[[4]string]*big.Int{}, metadata: map[[3]string]string{}}
	for _, row := range balances {
		amount := row.Amount
		if amount == nil {
			amount = new(big.Int)
		}
		s.balances[[4]string{row.Account, row.Scope, row.Asset, row.Color}] = amount
	}
	for _, row := range metadata {
		s.metadata[[3]string{row.Account, row.Scope, row.Key}] = row.Value
	}
	return s
}

func (s corpusStore) GetBalance(_ context.Context, account, scope, asset, color string) (*big.Int, error) {
	if v, ok := s.balances[[4]string{account, scope, asset, color}]; ok {
		return new(big.Int).Set(v), nil
	}
	return new(big.Int), nil
}

func (s corpusStore) GetMetadata(_ context.Context, account, scope, key string) (string, bool, error) {
	v, ok := s.metadata[[3]string{account, scope, key}]
	return v, ok, nil
}

func TestFreezeBytecodeCorpus(t *testing.T) {
	if os.Getenv("NUMSCRIPT_FREEZE_BYTECODE") == "" {
		t.Skip("set NUMSCRIPT_FREEZE_BYTECODE=1 to write a freeze (`just freeze-bytecode`)")
	}

	name := numscript.CurrentBytecodeVersion.String()
	if suffix := os.Getenv("NUMSCRIPT_FREEZE_SUFFIX"); suffix != "" {
		name += "-" + suffix
	}
	dir := filepath.Join(bytecodeCorpusDir, name)
	_, statErr := os.Stat(dir)
	require.ErrorIs(t, statErr, fs.ErrNotExist, "%s exists: a freeze is never overwritten, pick a suffix", dir)

	rawSpecs, err := specs_format.ReadSpecsFiles([]string{bytecodeFixturesDir})
	require.NoError(t, err)
	require.NotEmpty(t, rawSpecs)

	frozen := 0
	for _, rawSpec := range rawSpecs {
		rel, err := filepath.Rel(bytecodeFixturesDir, rawSpec.NumscriptPath)
		require.NoError(t, err)
		programName := strings.ReplaceAll(strings.TrimSuffix(filepath.ToSlash(rel), ".num"), "/", "-")

		var specs specs_format.Specs
		require.NoError(t, json.Unmarshal(rawSpec.SpecsFileContent, &specs), rel)
		require.NotEmpty(t, specs.TestCases, "%s: no test cases", rel)
		for _, tc := range specs.TestCases {
			require.False(t, tc.Focus || tc.Skip, "%s: a freeze covers every case, remove focus/skip from %q", rel, tc.It)
		}

		featureFlags := map[string]struct{}{}
		for _, flag := range specs.FeatureFlags {
			featureFlags[flag] = struct{}{}
		}

		enc, program, cErr := numscript.CompileWithFeatureFlags(rawSpec.NumscriptContent, featureFlags)
		if notImplemented, ok := errors.AsType[numscript.CompilerFeatureNotImplemented](cErr); ok {
			t.Logf("skip %s: compiler has no %s yet", rel, notImplemented.Feature)
			continue
		}
		require.NoError(t, cErr, rel)

		require.NoError(t, os.MkdirAll(filepath.Join(dir, programName), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, programName+corpusProgramExt), program.Encode(), 0o644))

		used := map[string]bool{}
		for _, tc := range specs.TestCases {
			caseVars := interpreter.VariablesMap{}
			maps.Copy(caseVars, specs.Vars)
			maps.Copy(caseVars, tc.Vars)
			vars, encErr := enc.Encode(caseVars)
			require.NoError(t, encErr, "%s %q", rel, tc.It)

			balances := specs_format.MergeBalances(specs.Balances, tc.Balances)
			metadata := mergeMetadata(specs.Meta, tc.Meta)

			res, execErr := numscript.ExecVm(context.Background(), numscript.NewVm(program), &vars, newCorpusStore(balances, metadata))

			caseDir := filepath.Join(dir, programName, caseFolderName(tc.It, used))
			require.NoError(t, os.MkdirAll(caseDir, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(caseDir, corpusVarsFile), vars.Encode(), 0o644))
			writeCorpusJSON(t, filepath.Join(caseDir, corpusBalancesFile), balances)
			if len(metadata) != 0 {
				writeCorpusJSON(t, filepath.Join(caseDir, corpusMetadataFile), metadata)
			}
			if execErr != nil {
				writeCorpusJSON(t, filepath.Join(caseDir, corpusErrorFile), corpusErrorOf(t, execErr))
			} else {
				writeCorpusJSON(t, filepath.Join(caseDir, corpusPostingsFile), corpusResult(res))
			}
			frozen++
		}
	}
	t.Logf("froze %d cases into %s", frozen, dir)
}

func TestBytecodeCorpusHasCurrentVersion(t *testing.T) {
	current := numscript.CurrentBytecodeVersion
	for _, name := range readCorpusFreezes(t) {
		if freezeVersion(t, name) == current {
			return
		}
	}
	t.Fatalf("no frozen bytecode for version %s: run `just freeze-bytecode`", current)
}

func TestBytecodeCorpus(t *testing.T) {
	freezes := readCorpusFreezes(t)
	require.NotEmpty(t, freezes)

	for _, freeze := range freezes {
		t.Run(freeze, func(t *testing.T) {
			version := freezeVersion(t, freeze)
			dir := filepath.Join(bytecodeCorpusDir, freeze)
			entries, err := os.ReadDir(dir)
			require.NoError(t, err)

			programs := map[string]bool{}
			caseDirs := map[string]bool{}
			for _, e := range entries {
				switch {
				case e.IsDir():
					caseDirs[e.Name()] = true
				case strings.HasSuffix(e.Name(), corpusProgramExt):
					programs[strings.TrimSuffix(e.Name(), corpusProgramExt)] = true
				default:
					t.Errorf("unexpected file %s", e.Name())
				}
			}
			require.Equal(t, sortedKeys(programs), sortedKeys(caseDirs), "every program has a folder of cases and vice versa")

			for _, name := range sortedKeys(programs) {
				t.Run(name, func(t *testing.T) {
					checkFrozenProgram(t, dir, name, version)
				})
			}
		})
	}
}

func checkFrozenProgram(t *testing.T, dir, name string, version numscript.BytecodeVersion) {
	programBytes, err := os.ReadFile(filepath.Join(dir, name+corpusProgramExt))
	require.NoError(t, err)
	programVersion, err := numscript.PeekCompiledProgramVersion(programBytes)
	require.NoError(t, err)
	require.Equal(t, version, programVersion, "program version differs from the freeze directory name")

	cases, err := os.ReadDir(filepath.Join(dir, name))
	require.NoError(t, err)
	require.NotEmpty(t, cases, "no cases")

	readable := numscript.CurrentBytecodeVersion.CanRead(programVersion)
	for _, c := range cases {
		require.True(t, c.IsDir(), "%s: not a case folder", c.Name())
		t.Run(c.Name(), func(t *testing.T) {
			caseDir := filepath.Join(dir, name, c.Name())
			varsBytes, err := os.ReadFile(filepath.Join(caseDir, corpusVarsFile))
			require.NoError(t, err)
			varsVersion, err := numscript.PeekVarsVersion(varsBytes)
			require.NoError(t, err)
			require.Equal(t, programVersion, varsVersion)

			if !readable {
				var unsupported numscript.UnsupportedBytecodeVersionError
				_, err := numscript.DecodeCompiledProgram(programBytes)
				require.ErrorAs(t, err, &unsupported)
				require.Equal(t, programVersion, unsupported.Encoded)
				_, err = numscript.DecodeVars(varsBytes)
				require.ErrorAs(t, err, &unsupported)
				require.Equal(t, varsVersion, unsupported.Encoded)
				return
			}

			program, err := numscript.DecodeCompiledProgram(programBytes)
			require.NoError(t, err)
			vars, err := numscript.DecodeVars(varsBytes)
			require.NoError(t, err)
			if programVersion == numscript.CurrentBytecodeVersion {
				require.Equal(t, programBytes, program.Encode(), "re-encoding changed the program bytes")
				require.Equal(t, varsBytes, vars.Encode(), "re-encoding changed the vars bytes")
			}
			info, err := numscript.VerifyCompiledProgramWithVars(program, &vars)
			require.NoError(t, err)
			require.True(t, info.CheckVars(&vars))

			var balances interpreter.Balances
			readCorpusJSON(t, filepath.Join(caseDir, corpusBalancesFile), &balances)
			var metadata interpreter.AccountsMetadata
			if _, statErr := os.Stat(filepath.Join(caseDir, corpusMetadataFile)); statErr == nil {
				readCorpusJSON(t, filepath.Join(caseDir, corpusMetadataFile), &metadata)
			}

			res, execErr := numscript.ExecVm(context.Background(), numscript.NewVm(program), &vars, newCorpusStore(balances, metadata))

			wantPostings, postingsErr := os.ReadFile(filepath.Join(caseDir, corpusPostingsFile))
			wantError, errorErr := os.ReadFile(filepath.Join(caseDir, corpusErrorFile))
			require.NotEqual(t, postingsErr == nil, errorErr == nil, "a case has exactly one of %s and %s", corpusPostingsFile, corpusErrorFile)

			if errorErr == nil {
				require.Error(t, execErr, "the frozen run failed, this one succeeded")
				var want corpusError
				require.NoError(t, json.Unmarshal(wantError, &want))
				got := corpusErrorOf(t, execErr)
				require.Equal(t, want.Type, got.Type)
				requireSameJSON(t, want.Value, got.Value)
				return
			}
			require.NoError(t, execErr, "the frozen run succeeded, this one failed")
			got, err := json.Marshal(corpusResult(res))
			require.NoError(t, err)
			requireSameJSON(t, wantPostings, got)
		})
	}
}

func readCorpusFreezes(t *testing.T) []string {
	entries, err := os.ReadDir(bytecodeCorpusDir)
	require.NoError(t, err)
	var names []string
	for _, e := range entries {
		require.True(t, e.IsDir(), "%s: not a freeze directory", e.Name())
		names = append(names, e.Name())
	}
	return names
}

func freezeVersion(t *testing.T, name string) numscript.BytecodeVersion {
	base, _, _ := strings.Cut(name, "-")
	var v numscript.BytecodeVersion
	_, err := fmt.Sscanf(base, "%d.%d", &v.Major, &v.Minor)
	require.NoError(t, err, "freeze directory %q is not named <major>.<minor>[-suffix]", name)
	return v
}

func corpusResult(res numscript.ExecutionResult) numscript.ExecutionResult {
	if res.Postings == nil {
		res.Postings = []numscript.Posting{}
	}
	return res
}

// Marshalled through a pointer: big.Int and big.Rat fields only have a
// pointer-receiver MarshalText, and a value field would encode as {}.
func corpusErrorOf(t *testing.T, err error) corpusError {
	v := reflect.ValueOf(err)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	ptr := reflect.New(v.Type())
	ptr.Elem().Set(v)
	value, mErr := json.Marshal(ptr.Interface())
	require.NoError(t, mErr)
	return corpusError{Type: v.Type().Name(), Value: value}
}

func mergeMetadata(outer, inner interpreter.AccountsMetadata) interpreter.AccountsMetadata {
	merged := interpreter.AccountsMetadata{}
	index := map[[3]string]int{}
	for _, rows := range []interpreter.AccountsMetadata{outer, inner} {
		for _, row := range rows {
			key := [3]string{row.Account, row.Scope, row.Key}
			if i, ok := index[key]; ok {
				merged[i] = row
			} else {
				index[key] = len(merged)
				merged = append(merged, row)
			}
		}
	}
	return merged
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func caseFolderName(it string, used map[string]bool) string {
	base := strings.Trim(nonSlug.ReplaceAllString(strings.ToLower(it), "-"), "-")
	if base == "" {
		base = "case"
	}
	name := base
	for i := 2; used[name]; i++ {
		name = fmt.Sprintf("%s-%d", base, i)
	}
	used[name] = true
	return name
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

func writeCorpusJSON(t *testing.T, path string, v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, append(data, '\n'), 0o644))
}

func readCorpusJSON(t *testing.T, path string, v any) {
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, v))
}

func requireSameJSON(t *testing.T, want, got []byte) {
	require.Equal(t, decodeJSONExact(t, want), decodeJSONExact(t, got))
}

func decodeJSONExact(t *testing.T, data []byte) any {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var v any
	require.NoError(t, dec.Decode(&v))
	return v
}
