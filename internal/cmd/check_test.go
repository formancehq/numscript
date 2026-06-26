package cmd

import (
	"testing"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

func diagnosticAt(line int, character int) analysis.Diagnostic {
	return analysis.Diagnostic{
		Range: parser.Range{
			Start: parser.Position{Line: line, Character: character},
		},
	}
}

func TestSortDiagnosticsInSourceOrder(t *testing.T) {
	t.Parallel()

	diagnostics := []analysis.Diagnostic{
		diagnosticAt(3, 0),
		diagnosticAt(1, 10),
		diagnosticAt(1, 2),
		diagnosticAt(0, 5),
	}

	sortDiagnostics(diagnostics)

	require.Equal(t, []analysis.Diagnostic{
		diagnosticAt(0, 5),
		diagnosticAt(1, 2),
		diagnosticAt(1, 10),
		diagnosticAt(3, 0),
	}, diagnostics)
}

func TestSortDiagnosticsWithEqualPositions(t *testing.T) {
	t.Parallel()

	// Equal positions must not panic or misbehave: the comparator has to be
	// strict (a valid less function), unlike the previous GtEq-based one.
	diagnostics := []analysis.Diagnostic{
		diagnosticAt(2, 4),
		diagnosticAt(2, 4),
		diagnosticAt(0, 0),
	}

	sortDiagnostics(diagnostics)

	require.Equal(t, []analysis.Diagnostic{
		diagnosticAt(0, 0),
		diagnosticAt(2, 4),
		diagnosticAt(2, 4),
	}, diagnostics)
}
