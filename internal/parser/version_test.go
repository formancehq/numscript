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

	require.Equal(t, parser.VersionMachine{}, res.Value.GetVersion())
}

func TestParseInterpreterVersion(t *testing.T) {
	res := parser.Parse(`
		// @version interpreter 12.34.56
	`)

	require.Equal(t, parser.VersionInterpreter{
		Major: 12,
		Minor: 34,
		Patch: 56,
	}, res.Value.GetVersion())
}

func TestParseInterpreterVersionWhenNotFirstComment(t *testing.T) {
	res := parser.Parse(`
		// first comment 
		// @version interpreter 0.0.1
	`)

	require.Equal(t, parser.VersionInterpreter{
		Patch: 1,
	}, res.Value.GetVersion())
}

func TestParseInvalid(t *testing.T) {
	res := parser.Parse(`
		// @version not a valid version
	`)

	require.Equal(t, nil, res.Value.GetVersion())
}

func TestParseFlag(t *testing.T) {
	res := parser.Parse(`
		// @feature_flag experimental-something
		// @feature_flag experimental-my-feature
		// @feature_flag another-flag
	`)

	require.Equal(t, map[string]struct{}{
		"experimental-something":  struct{}{},
		"experimental-my-feature": struct{}{},
		"another-flag":            struct{}{},
	}, res.Value.GetFlags())
}

func TestGtEq(t *testing.T) {
	t.Run("same version", func(t *testing.T) {
		v1 := parser.VersionInterpreter{
			Major: 1,
			Minor: 2,
			Patch: 3,
		}

		require.True(t,
			v1.GtEq(parser.VersionInterpreter{
				Major: 1,
				Minor: 2,
				Patch: 3,
			}),
		)
	})

	t.Run("left has higher major version", func(t *testing.T) {
		v1 := parser.VersionInterpreter{
			Major: 10,
			Minor: 1,
			Patch: 1,
		}

		require.True(t,
			v1.GtEq(parser.VersionInterpreter{
				Major: 1,
				Minor: 100,
				Patch: 100,
			}),
		)
	})

	t.Run("left has lower major version", func(t *testing.T) {
		v1 := parser.VersionInterpreter{
			Major: 2,
			Minor: 10,
			Patch: 10,
		}

		require.False(t,
			v1.GtEq(parser.VersionInterpreter{
				Major: 3,
				Minor: 1,
				Patch: 1,
			}),
		)
	})

	t.Run("same major, higher minor", func(t *testing.T) {
		v1 := parser.VersionInterpreter{
			Major: 1,
			Minor: 10,
			Patch: 10,
		}

		require.True(t,
			v1.GtEq(parser.VersionInterpreter{
				Major: 1,
				Minor: 1,
				Patch: 10,
			}),
		)
	})

	t.Run("same major, lower minor", func(t *testing.T) {
		v1 := parser.VersionInterpreter{
			Major: 42,
			Minor: 2,
			Patch: 10,
		}

		require.False(t,
			v1.GtEq(parser.VersionInterpreter{
				Major: 42,
				Minor: 10,
				Patch: 10,
			}),
		)
	})

	t.Run("same major and minor, lower patch", func(t *testing.T) {
		v1 := parser.VersionInterpreter{
			Major: 2,
			Minor: 2,
			Patch: 1,
		}

		require.False(t,
			v1.GtEq(parser.VersionInterpreter{
				Major: 2,
				Minor: 2,
				Patch: 10,
			}),
		)
	})

	t.Run("same major and minor, higher patch", func(t *testing.T) {
		v1 := parser.VersionInterpreter{
			Major: 2,
			Minor: 2,
			Patch: 100,
		}

		require.True(t,
			v1.GtEq(parser.VersionInterpreter{
				Major: 2,
				Minor: 2,
				Patch: 2,
			}),
		)
	})
}
