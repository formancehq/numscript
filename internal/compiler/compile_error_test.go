package compiler

// White-box tests asserting the concrete CompilerError produced for invalid
// programs. They call compileProgramToVirtual directly, since the public Compile
// stringifies the error and would lose the type.

import (
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/typecheck"
	"github.com/stretchr/testify/require"
)

func TestE2E_RejectsUnboundVariable(t *testing.T) {
	parsed := parser.Parse(`send [C 10] (source = $undeclared destination = @d)`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToVirtual(parsed.Value)
	require.IsType(t, TypeError{}, cErr)
	require.IsType(t, typecheck.UnboundVariable{}, cErr.(TypeError).Kind)
}

func TestE2E_RejectsTypeMismatch(t *testing.T) {
	parsed := parser.Parse(`vars { string $s } send [C 10] (source = $s destination = @d)`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToVirtual(parsed.Value)
	require.IsType(t, TypeError{}, cErr)
	require.IsType(t, typecheck.TypeMismatch{}, cErr.(TypeError).Kind)
}

func TestE2E_AllotmentDuplicateRemaining(t *testing.T) {
	parsed := parser.Parse(`
		send [USD/2 100] (
			source = @world
			destination = {
				remaining to @a
				remaining to @b
			}
		)
	`)
	require.Empty(t, parsed.Errors)
	_, cErr := compileProgramToVirtual(parsed.Value)
	require.IsType(t, DuplicateRemaining{}, cErr)
}
