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

func TestFormatMonetary(t *testing.T) {
	src := `send [COIN 100] (
  source = $src
  destination = $dest
)`

	AssertIsFormatted(t, src)
}

func TestFormatMaxSrc(t *testing.T) {
	src := `send $amt (
  source = max [COIN 10] from $src
  destination = $dest
)`

	AssertIsFormatted(t, src)
}

func TestFormatAddr(t *testing.T) {
	src := `send $amount (
  source = @src
  destination = @dest
)`

	AssertIsFormatted(t, src)
}

func TestInorder(t *testing.T) {
	src := `send $amount (
  source = {
    @s1
    @s2
  }
  destination = @dest
)`

	AssertIsFormatted(t, src)
}

func TestInorderSrcNested(t *testing.T) {
	src := `send $amount (
  source = {
    @s1
    {
      @s2
      @s3
    }
  }
  destination = @dest
)`

	AssertIsFormatted(t, src)
}

func TestInorderSrcNestedInMax(t *testing.T) {
	src := `send $amount (
  source = max $cap from {
    @s1
    {
      @s2
      @s3
    }
  }
  destination = @dest
)`

	AssertIsFormatted(t, src)
}

func TestInorderDest(t *testing.T) {
	src := `send $amount (
  source = @src
  destination = {
    max $cap to @d1
    max [COIN 10] kept
    remaining to @d3
  }
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
