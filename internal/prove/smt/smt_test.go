package smt

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func requireZ3(t *testing.T) {
	t.Helper()
	if err := (Solver{}).Available(); err != nil {
		if os.Getenv("NUMSCRIPT_REQUIRE_Z3") != "" {
			t.Fatal(err)
		}
		t.Skip(err)
	}
}

func TestFolding(t *testing.T) {
	x := Symbol("x", SortInt)
	require.Equal(t, "7", Print(Add(Int64(3), Int64(4))))
	require.Equal(t, "x", Print(Add(x, Int64(0))))
	require.Equal(t, "0", Print(Sub(x, x)))
	require.Equal(t, "(- 3)", Print(Neg(Int64(3))))
	require.Equal(t, "(- x)", Print(Neg(x)))
	require.Equal(t, "x", Print(MulConst(big.NewInt(1), x)))
	require.Equal(t, "(* 2 x)", Print(MulConst(big.NewInt(2), x)))
	require.Equal(t, "(- 3)", Print(DivConst(Int64(-5), big.NewInt(2))))
	require.Equal(t, "1", Print(ModConst(Int64(-5), big.NewInt(2))))
	require.Equal(t, "true", Print(Lt(Int64(0), Int64(10))))
	require.Equal(t, "false", Print(Lt(x, x)))
	require.Equal(t, "true", Print(And(Bool(true), Bool(true))))
	require.Equal(t, "false", Print(And(Symbol("a", SortBool), Bool(false))))
	require.Equal(t, "(and a b)", Print(And(Symbol("a", SortBool), Bool(true), Symbol("b", SortBool))))
	require.Equal(t, "(or a b c)", Print(Or(Or(Symbol("a", SortBool), Symbol("b", SortBool)), Symbol("c", SortBool))))
	require.Equal(t, "a", Print(Not(Not(Symbol("a", SortBool)))))
	require.Equal(t, "x", Print(Ite(Bool(true), x, Int64(1))))
	require.Equal(t, "x", Print(Ite(Symbol("c", SortBool), x, x)))
	require.Equal(t, "c", Print(Ite(Symbol("c", SortBool), Bool(true), Bool(false))))
	require.Equal(t, "(not c)", Print(Ite(Symbol("c", SortBool), Bool(false), Bool(true))))
	require.Equal(t, "(ite (< x 0) 0 x)", Print(Max0(x)))
	require.Equal(t, "(ite (< x 0) 0 x)", Print(Max0(Max0(x))))
	require.Equal(t, "(ite (< x 10) x 10)", Print(Min(x, Int64(10))))
	require.Equal(t, "3", Print(Min(Int64(3), Int64(10))))
	require.Equal(t, "b", Print(Implies(Bool(true), Symbol("b", SortBool))))
	require.Equal(t, "true", Print(Implies(Symbol("a", SortBool), Bool(true))))
	require.Equal(t, "(not a)", Print(Iff(Symbol("a", SortBool), Bool(false))))
	// a min of two non-negatives needs no outer clamp
	require.Equal(t, "(ite (< (ite (< x 0) 0 x) 10) (ite (< x 0) 0 x) 10)", Print(Max0(Min(Max0(x), Int64(10)))))
}

func TestPrintQuoting(t *testing.T) {
	require.Equal(t, "$r9@18", Print(Symbol("$r9@18", SortInt)))
	require.Equal(t, "|before(@src,USD/2)|", Print(Symbol("before(@src,USD/2)", SortInt)))
	require.Equal(t, "|1x|", Print(Symbol("1x", SortInt)))
	require.Equal(t, "(= |a b| (- 5))", Print(Eq(Symbol("a b", SortInt), Int64(-5))))
}

func TestScriptDefineAliasesLeaves(t *testing.T) {
	s := NewScript()
	x := s.DeclareConst("x", SortInt)
	require.Equal(t, x, s.Define("y", x))
	y := s.Define("y", Add(x, Int64(1)))
	require.Equal(t, "y", Print(y))
	def, ok := s.Definition("y")
	require.True(t, ok)
	require.Equal(t, "(+ x 1)", Print(def))
	require.Equal(t, "(set-logic QF_LIA)\n(set-option :produce-models true)\n(declare-const x Int)\n(define-fun y () Int (+ x 1))\n", s.Text())
	require.Panics(t, func() { s.Define("y", Add(x, Int64(2))) })
}

func TestParseResult(t *testing.T) {
	res, err := parseResult("sat\n((|before(@src,USD/2)| (- 5))\n (fails true)\n (n 42))\n")
	require.NoError(t, err)
	require.Equal(t, Sat, res.Status)
	require.Equal(t, "-5", res.Values["before(@src,USD/2)"].String())
	require.Equal(t, "true", res.Values["fails"].String())
	require.Equal(t, "42", res.Values["n"].String())

	res, err = parseResult("unsat\n(error \"line 12 column 10: model is not available\")\n")
	require.NoError(t, err)
	require.Equal(t, Unsat, res.Status)

	res, err = parseResult("unknown\n")
	require.NoError(t, err)
	require.Equal(t, Unknown, res.Status)

	_, err = parseResult("(error \"line 3 column 4: unknown constant foo\")\nsat\n")
	require.Error(t, err)

	_, err = parseResult("")
	require.Error(t, err)
}

func TestEvaluator(t *testing.T) {
	s := NewScript()
	x := s.DeclareConst("x", SortInt)
	y := s.Define("y", Add(DivConst(x, big.NewInt(2)), ModConst(x, big.NewInt(2))))
	z := s.Define("z", Ite(Lt(y, Int64(0)), Neg(y), Max0(Sub(y, Int64(10)))))
	ev := NewEvaluator(s, map[string]Value{"x": IntValue(big.NewInt(-5))})
	v, err := ev.Eval(z)
	require.NoError(t, err)
	// -5 div 2 = -3, -5 mod 2 = 1, y = -2, z = 2
	require.Equal(t, "2", v.String())
	v, err = ev.Eval(And(Lt(x, Int64(0)), Implies(Bool(false), Bool(false))))
	require.NoError(t, err)
	require.True(t, v.Bool)
	_, err = ev.Eval(Symbol("w", SortInt))
	require.Error(t, err)
}

func TestZ3RoundTrip(t *testing.T) {
	requireZ3(t)
	s := NewScript()
	x := s.DeclareConst("before(@src,USD/2)", SortInt)
	s.Assert(Lt(x, Int64(-3)))
	res, err := Solver{Timeout: 5 * time.Second}.CheckSat(context.Background(), s.Text(), []string{"before(@src,USD/2)"})
	require.NoError(t, err)
	require.Equal(t, Sat, res.Status)
	require.True(t, res.Values["before(@src,USD/2)"].Int.Cmp(big.NewInt(-3)) < 0)

	s.Assert(Gt(x, Int64(0)))
	res, err = Solver{Timeout: 5 * time.Second}.CheckSat(context.Background(), s.Text(), []string{"before(@src,USD/2)"})
	require.NoError(t, err)
	require.Equal(t, Unsat, res.Status)

	_, err = Solver{Path: "no-such-solver"}.CheckSat(context.Background(), s.Text(), nil)
	require.ErrorIs(t, err, ErrSolverNotFound)
}
