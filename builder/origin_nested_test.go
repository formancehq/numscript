package builder_test

import (
	"strings"
	"testing"

	"github.com/formancehq/numscript"
	"github.com/formancehq/numscript/builder"
)

// An origin's account position takes any account expression, including
// another origin var. Rendering the outer origin then declares the inner one,
// growing env.originVars midway through the render loop.
func TestNestedOriginVars(t *testing.T) {
	inner := builder.NewAccountVarFromMeta(builder.ExprAccount("src"), "delegate")
	outer := builder.NewMonetaryVarFromBalance(inner, builder.ExprAsset("COIN"))

	_, _, script := builder.BuildProgram(
		builder.StmtSend(outer,
			builder.SrcAccount(builder.ExprAccount("world")),
			builder.DestAccount(builder.ExprAccount("dst"))),
	)

	if errs := numscript.Parse(script).GetParsingErrors(); len(errs) != 0 {
		t.Fatalf("generated script does not parse: %v\n%s", errs, script)
	}

	// The inner var must be declared before the var whose RHS references it:
	// both engines resolve origins in declaration order and dereference a
	// referenced resource without checking it resolved.
	innerDecl := strings.Index(script, "account $originvar_1 =")
	outerDecl := strings.Index(script, "monetary $originvar_0 =")
	if innerDecl < 0 || outerDecl < 0 {
		t.Fatalf("expected both origin vars to be declared, got:\n%s", script)
	}
	if innerDecl > outerDecl {
		t.Errorf("origin var declared after the var referencing it:\n%s", script)
	}
}

// Two origins sharing one nested origin var: the shared var is declared once,
// before both of its dependents.
func TestNestedOriginVarsShared(t *testing.T) {
	shared := builder.NewAccountVarFromMeta(builder.ExprAccount("src"), "delegate")
	a := builder.NewMonetaryVarFromBalance(shared, builder.ExprAsset("COIN"))
	b := builder.NewMonetaryVarFromBalance(shared, builder.ExprAsset("USD/2"))

	_, _, script := builder.BuildProgram(
		builder.StmtSend(a,
			builder.SrcAccount(builder.ExprAccount("world")),
			builder.DestAccount(builder.ExprAccount("dst"))),
		builder.StmtSend(b,
			builder.SrcAccount(builder.ExprAccount("world")),
			builder.DestAccount(builder.ExprAccount("dst"))),
	)

	if errs := numscript.Parse(script).GetParsingErrors(); len(errs) != 0 {
		t.Fatalf("generated script does not parse: %v\n%s", errs, script)
	}
	if n := strings.Count(script, "account $originvar_2 ="); n != 1 {
		t.Fatalf("expected the shared origin var declared exactly once, got %d:\n%s", n, script)
	}
	shrDecl := strings.Index(script, "account $originvar_2 =")
	for _, dep := range []string{"monetary $originvar_0 =", "monetary $originvar_1 ="} {
		if i := strings.Index(script, dep); i < 0 || shrDecl > i {
			t.Errorf("shared origin var not declared before %q:\n%s", dep, script)
		}
	}
}
