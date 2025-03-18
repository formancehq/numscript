package parser_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestParseMachine(t *testing.T) {
	res := parser.Parse(`
		// @version machine
	`)

	v := res.Value.GetVersion()
	require.Equal(t, v, parser.VersionMachine{})
}

func TestParseInterpreterVersion(t *testing.T) {
	res := parser.Parse(`
		// @version interpreter 12.34.56
	`)

	v := res.Value.GetVersion()
	require.Equal(t, v, parser.VersionInterpreter{
		Major: 12,
		Minor: 34,
		Patch: 56,
	})
}

func TestParseInvalid(t *testing.T) {
	res := parser.Parse(`
		// @version not a valid version
	`)

	v := res.Value.GetVersion()
	require.Equal(t, v, nil)
}
