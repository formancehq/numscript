package typecheck_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/typecheck"

	"github.com/stretchr/testify/require"
)

func check(t *testing.T, src string) typecheck.Result {
	t.Helper()
	parsed := parser.Parse(src)
	require.Empty(t, parsed.Errors)
	return typecheck.Check(parsed.Value)
}

func kinds(res typecheck.Result) []typecheck.ErrorKind {
	out := make([]typecheck.ErrorKind, len(res.Errors))
	for i, e := range res.Errors {
		out[i] = e.Kind
	}
	return out
}

func TestValidProgram(t *testing.T) {
	res := check(t, `
		vars { account $acc = @src }
		send [USD/2 10] (source = $acc destination = @dest)
	`)
	require.Empty(t, res.Errors)
	require.Equal(t, typecheck.TypeAccount, res.VarTypes["acc"])
}

func TestInvalidType(t *testing.T) {
	res := check(t, `vars { invalid $x }`)
	require.Equal(t, []typecheck.ErrorKind{typecheck.InvalidType{Name: "invalid"}}, kinds(res))
}

func TestDuplicateVariable(t *testing.T) {
	res := check(t, `vars { account $x account $x }`)
	require.Equal(t, []typecheck.ErrorKind{typecheck.DuplicateVariable{Name: "x"}}, kinds(res))
}

func TestUnboundVariable(t *testing.T) {
	res := check(t, `send [C 10] (source = $nope destination = @d)`)
	require.Equal(t, []typecheck.ErrorKind{typecheck.UnboundVariable{Name: "nope", Type: typecheck.TypeAccount}}, kinds(res))
}

func TestTypeMismatch(t *testing.T) {
	// a string var used where an account is expected
	res := check(t, `vars { string $s } send [C 10] (source = $s destination = @d)`)
	require.Equal(t, []typecheck.ErrorKind{
		typecheck.TypeMismatch{Expected: typecheck.TypeAccount, Got: typecheck.TypeString},
	}, kinds(res))
}

func TestUnknownFunction(t *testing.T) {
	res := check(t, `vars { number $n = nope() }`)
	require.Equal(t, []typecheck.ErrorKind{typecheck.UnknownFunction{Name: "nope"}}, kinds(res))
}

func TestBadArity(t *testing.T) {
	res := check(t, `vars { monetary $m = balance(@a) }`)
	require.Equal(t, []typecheck.ErrorKind{typecheck.BadArity{Expected: 2, Actual: 1}}, kinds(res))
}

func TestExprTypes(t *testing.T) {
	res := check(t, `send [USD/2 10] (source = @a destination = @b)`)
	// the monetary literal is typed
	send := res // just assert no errors + monetary present via a scan
	require.Empty(t, send.Errors)
	found := false
	for _, ty := range res.ExprTypes {
		if ty == typecheck.TypeMonetary {
			found = true
		}
	}
	require.True(t, found, "expected a monetary-typed expr")
}
