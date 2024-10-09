package format_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/format"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestFormatSimpleAddr(t *testing.T) {

	src := `send $amount (
  source = $src
  destination = $dest
)`

	AssertIsFormatted(t, src)
}

func AssertIsFormatted(t *testing.T, src string) {
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	formatted := format.Format(parsed.Value)
	require.Equal(t, src, formatted)
}

// TODO vars
// TODO allotment
// TODO inorder
// TODO max
