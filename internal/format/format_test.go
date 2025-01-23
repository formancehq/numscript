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
)
`

	AssertIsFormatted(t, src)
}

func TestFormatManyStatements(t *testing.T) {

	src := `send $amount (
  source = $src
  destination = $dest
)

send $amount (
  source = $src
  destination = $dest
)
`

	AssertIsFormatted(t, src)
}

func TestSendAll(t *testing.T) {
	src := `send [COIN *] (
  source = @src
  destination = @dest
)
`

	AssertIsFormatted(t, src)
}

func TestOverdraftUnbounded(t *testing.T) {
	src := `send $ass (
  source = @src allowing unbounded overdraft
  destination = @dest
)
`

	AssertIsFormatted(t, src)
}

func TestOverdraftBounded(t *testing.T) {
	src := `send $ass (
  source = @src allowing overdraft up to $x
  destination = @dest
)
`

	AssertIsFormatted(t, src)
}

func TestFormatMonetary(t *testing.T) {
	src := `send [COIN 100] (
  source = $src
  destination = $dest
)
`

	AssertIsFormatted(t, src)
}

func TestFormatMaxSrc(t *testing.T) {
	src := `send $amt (
  source = max [COIN 10] from $src
  destination = $dest
)
`

	AssertIsFormatted(t, src)
}

func TestFormatAddr(t *testing.T) {
	src := `send $amount (
  source = @src
  destination = @dest
)
`

	AssertIsFormatted(t, src)
}

func TestInorder(t *testing.T) {
	src := `send $amount (
  source = {
    @s1
    @s2
  }
  destination = @dest
)
`

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
)
`

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
)
`

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
)
`

	AssertIsFormatted(t, src)
}

func TestAllotmentSrc(t *testing.T) {
	src := `send $amount (
  source = {
    1/3 from @a
    4% from @a1
    $var from @b
    remaining from @r
  }
  destination = @dest
)
`

	AssertIsFormatted(t, src)
}

func TestAllotmentDest(t *testing.T) {
	src := `send $amount (
  source = @src
  destination = {
    1/3 to @a
    4% kept
    $var to @b
    remaining to @r
  }
)
`

	AssertIsFormatted(t, src)
}

func TestFnCall(t *testing.T) {
	src := `set_tx_meta("k", 42)

example_fn(@acc, EUR, 1/2)
`

	AssertIsFormatted(t, src)
}

func TestVars(t *testing.T) {
	src := `vars {
  monetary $x
  account $a
  number $a = meta(@acc, "k")
  portion $a = something(ASSET, EUR/2, 42)
}

`

	AssertIsFormatted(t, src)
}

func TestMixAll(t *testing.T) {
	src := `vars {
  monetary $amt
}

send $amt (
  source = @a
  destination = @b
)

set_tx_meta("k", 42)
`

	AssertIsFormatted(t, src)
}

func TestPreserveStatementsComments(t *testing.T) {
	src := `// c1

set_tx_meta("k", 42)
`

	AssertIsFormatted(t, src)
}

func TestPreserveComments(t *testing.T) {
	t.Skip()

	src := `vars {
  // c1
  monetary $amt
}

// c2
send $amt (
  // c3
  source = @a
  destination = @b
)


// c4
set_tx_meta("k", 42)
`

	AssertIsFormatted(t, src)
}

func AssertIsFormatted(t *testing.T, src string) {
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)

	formatted := format.Format(parsed.Value)
	require.Equal(t, src, formatted)
}
