package pretty_test

import (
	. "numscript/pretty"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderStr(t *testing.T) {
	doc := Text("abc")

	out := PrintDefault(doc)
	require.Equal(t, "abc", out)
}

func TestConcatDocs(t *testing.T) {
	doc := Concat(
		Text("abc"),
		Text("def"),
	)

	out := PrintDefault(doc)
	require.Equal(t, "abcdef", out)
}

func TestNil(t *testing.T) {
	doc := Empty()

	out := PrintDefault(doc)
	require.Equal(t, "", out)
}

func TestLinesZero(t *testing.T) {
	doc := Concat(
		Text("abc"),
		Lines(0),
		Text("def"),
	)

	out := PrintDefault(doc)
	require.Equal(t, "abc\ndef", out)
}

func TestManyLines(t *testing.T) {
	doc := Concat(
		Text("abc"),
		Lines(2),
		Text("def"),
	)

	out := PrintDefault(doc)
	require.Equal(t, "abc\n\n\ndef", out)
}

func TestNestingNoBreak(t *testing.T) {
	doc := Concat(
		Text("ab"),
		Nest(
			Text("cd"),
		),
	)

	out := PrintDefault(doc)
	require.Equal(t, "abcd", out)
}