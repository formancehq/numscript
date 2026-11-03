package compiler_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"slices"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/formancehq/numscript/internal/vm"

	"github.com/stretchr/testify/require"
)

// Compile does not call vm.Verify — the VM assumes its own compiler's output is
// well formed, and paying for a full static pass on every compile would make
// that assumption cost something. This test is what earns the assumption: every
// script in the corpus must produce bytecode the verifier accepts, plain and
// optimized alike, and that bytecode must survive its encoding.
func TestCompiledCorpusPassesVerify(t *testing.T) {
	rawSpecs, err := specs_format.ReadSpecsFiles([]string{scriptsFolder})
	require.NoError(t, err)

	for _, rawSpec := range rawSpecs {
		rel, err := filepath.Rel(scriptsFolder, rawSpec.NumscriptPath)
		require.NoError(t, err)

		t.Run(rel, func(t *testing.T) {
			if slices.Contains(scriptsBlacklist, rel) {
				t.Skip("blacklisted: not supported yet")
			}

			var specs specs_format.Specs
			require.NoError(t, json.Unmarshal(rawSpec.SpecsFileContent, &specs))

			featureFlags := make(map[string]struct{}, len(specs.FeatureFlags))
			for _, flag := range specs.FeatureFlags {
				featureFlags[flag] = struct{}{}
			}

			parsed := parser.Parse(rawSpec.NumscriptContent)
			require.Empty(t, parsed.Errors)

			for name, opts := range map[string]compiler.Options{"plain": {}, "optimized": {Optimize: true}} {
				t.Run(name, func(t *testing.T) {
					_, program, cErr := compiler.CompileWithOptions(parsed.Value, featureFlags, opts)
					require.Nil(t, cErr)

					require.NoError(t, vm.Verify(program))

					encoded := program.Encode()
					decoded, err := vm.DecodeProgram(encoded)
					require.NoError(t, err)
					require.True(t, bytes.Equal(encoded, decoded.Encode()), "Encode/Decode/Encode is not byte-stable")
				})
			}
		})
	}
}
