package compiler_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/formancehq/numscript/internal/compiler"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/specs_format"

	"github.com/stretchr/testify/require"
)

// Property 2: every program our compiler emits passes the VM's sanity checks.
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
			parsed := parser.Parse(string(rawSpec.NumscriptContent))
			require.Empty(t, parsed.Errors)
			_, program, cErr := compiler.Compile(parsed.Value)
			require.Nil(t, cErr)
			require.NoError(t, program.Verify())
		})
	}
}
