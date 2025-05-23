package lsp_test

import (
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/analysis"
	lsp "github.com/formancehq/numscript/internal/lsp"
	"github.com/formancehq/numscript/internal/lsp/lsp_types"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

func performAction(t *testing.T,
	initial string,
	expected string,
	toEdit func(kind analysis.DiagnosticKind, program parser.Program) lsp_types.TextEdit,
) {
	res := analysis.CheckSource(initial)
	require.Len(t, res.Diagnostics, 1)

	first := res.Diagnostics[0]

	finalStr := performEdit(initial, toEdit(first.Kind, res.Program))

	require.Equal(t, expected, finalStr)
}

func TestCreateVarWhenNoVarsBlock(t *testing.T) {
	initial := `send [USD/2 100] (
  source = max $example from @a
  destination = @b
)
`

	final := `vars {
  monetary $example
}

send [USD/2 100] (
  source = max $example from @a
  destination = @b
)
`

	performAction(t, initial, final, func(kind analysis.DiagnosticKind, program parser.Program) lsp_types.TextEdit {
		return lsp.CreateVar(kind.(analysis.UnboundVariable), program)
	})

}

func TestCreateVarWhenAlreadyExistingVars(t *testing.T) {
	initial := `vars {
  monetary $example
}

send [USD/2 100] (
  source = max $example from $account
  destination = @b
)
`

	final := `vars {
  monetary $example
  account $account
}

send [USD/2 100] (
  source = max $example from $account
  destination = @b
)
`

	performAction(t, initial, final, func(kind analysis.DiagnosticKind, program parser.Program) lsp_types.TextEdit {
		return lsp.CreateVar(kind.(analysis.UnboundVariable), program)
	})

}

func TestCreateVarWhenAlreadyExistingVarsSameLine(t *testing.T) {
	initial := `vars { account $account }

send [USD/2 100] (
  source = max $example from $account
  destination = @b
)
`

	final := `vars { account $account
  monetary $example
}

send [USD/2 100] (
  source = max $example from $account
  destination = @b
)
`

	performAction(t, initial, final, func(kind analysis.DiagnosticKind, program parser.Program) lsp_types.TextEdit {
		return lsp.CreateVar(kind.(analysis.UnboundVariable), program)
	})

}

func TestCreateVarWhenEmptyVarsBlock(t *testing.T) {
	initial := `vars {
}

send [USD/2 100] (
  source = max [USD/2 100] from $account
  destination = @b
)
`

	final := `vars {
  account $account
}

send [USD/2 100] (
  source = max [USD/2 100] from $account
  destination = @b
)
`

	performAction(t, initial, final, func(kind analysis.DiagnosticKind, program parser.Program) lsp_types.TextEdit {
		return lsp.CreateVar(kind.(analysis.UnboundVariable), program)
	})

}

func TestPositionToOffset(t *testing.T) {
	str := `abc
def
ghi`

	require.Equal(t, positionToOffset(strings.Split(str, "\n"), lsp_types.Position{
		Line:      1,
		Character: 1,
	}), 5)

}

func TestPerformEdit(t *testing.T) {
	initial := `a
ins<>here
c
`

	require.Equal(t, `a
ins___here
c
`, performEdit(initial, lsp_types.TextEdit{
		Range: lsp_types.Range{
			Start: lsp_types.Position{Line: 1, Character: 3},
			End:   lsp_types.Position{Line: 1, Character: 5},
		},
		NewText: "___",
	}))

}

func TestPerformEdit2(t *testing.T) {
	initial := `abc`

	require.Equal(t, `LINE1
LINE2

abc`, performEdit(initial, lsp_types.TextEdit{
		// Empty range
		NewText: `LINE1
LINE2

`,
	}))

}

func positionToOffset(lines []string, position lsp_types.Position) int {
	// TODO: check indexes are 0-based

	offset := 0
	for _, line := range lines[0:position.Line] {
		// +1 for the newline which was trimmed in lines
		offset += len(line) + 1
	}

	offset += int(position.Character)

	return offset
}

func performEdit(initial string, textEdit lsp_types.TextEdit) string {
	lines := strings.Split(initial, "\n")

	startOffset := positionToOffset(lines, textEdit.Range.Start)
	endOffset := positionToOffset(lines, textEdit.Range.End)

	before := initial[0:startOffset]
	after := initial[endOffset:]

	return before + textEdit.NewText + after
}
