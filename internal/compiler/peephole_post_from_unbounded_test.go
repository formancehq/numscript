package compiler

import (
	"strings"
	"testing"

	"github.com/formancehq/numscript/internal/parser"
)

// countOp reports how many post_from_unbounded instrs the optimized stream for
// src contains (via the vInstr dump).
func optDump(t *testing.T, src string) string {
	t.Helper()
	parsed := parser.Parse(src)
	if len(parsed.Errors) != 0 {
		t.Fatalf("parse: %v", parsed.Errors)
	}
	cp, err := compileProgramToVirtual(parsed.Value)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	return dump(optimize(cp.instructions, defaultPeepholes()))
}

func TestPostFromUnbounded_World(t *testing.T) {
	dump := optDump(t, `send [USD/2 42] (source = @world destination = @dest)`)
	if !strings.Contains(dump, "post_from_unbounded") {
		t.Fatalf("expected fused op for @world source, got:\n%s", dump)
	}
	if strings.Contains(dump, "take_account") || strings.Contains(dump, "check_enough_funds") {
		t.Fatalf("take/check should be gone, got:\n%s", dump)
	}
}

func TestPostFromUnbounded_UnboundedOverdraft(t *testing.T) {
	dump := optDump(t, `send [USD/2 42] (source = @a allowing unbounded overdraft destination = @dest)`)
	if !strings.Contains(dump, "post_from_unbounded") {
		t.Fatalf("expected fused op for unbounded-overdraft source, got:\n%s", dump)
	}
}

func TestPostFromUnbounded_LeafDstSkipsCredit(t *testing.T) {
	// @dest is a leaf (never a source, never saved): the credit is dead.
	dump := optDump(t, `send [USD/2 42] (source = @world destination = @dest)`)
	if !strings.Contains(dump, "post_from_unbounded_leaf") {
		t.Fatalf("expected leaf (no-credit) variant, got:\n%s", dump)
	}
}

func TestPostFromUnbounded_NonLeafDstKeepsCredit(t *testing.T) {
	// @mid is credited then used as a source: its balance is observed, so the
	// credit must stay (plain post_from_unbounded, not the leaf variant).
	src := `send [USD/2 42] (source = @world destination = @mid)
send [USD/2 10] (source = @mid destination = @dest)`
	dump := optDump(t, src)
	if !strings.Contains(dump, "post_from_unbounded(") {
		t.Fatalf("expected crediting variant for non-leaf dst, got:\n%s", dump)
	}
	if strings.Contains(dump, "post_from_unbounded_leaf") {
		t.Fatalf("non-leaf dst must keep its credit, got:\n%s", dump)
	}
}

func TestPostFromUnbounded_SavedDstKeepsCredit(t *testing.T) {
	// @dest is saved, so its balance is read: keep the credit.
	src := `save [USD/2 1] from @dest
send [USD/2 42] (source = @world destination = @dest)`
	dump := optDump(t, src)
	if strings.Contains(dump, "post_from_unbounded_leaf") {
		t.Fatalf("saved dst must keep its credit, got:\n%s", dump)
	}
}

func TestPostFromUnbounded_BoundedSourceIneligible(t *testing.T) {
	// @src is a plain (bounded-zero) account, not unbounded.
	dump := optDump(t, `send [USD/2 42] (source = @src destination = @dest)`)
	if strings.Contains(dump, "post_from_unbounded") {
		t.Fatalf("bounded source must NOT use the fused op, got:\n%s", dump)
	}
}

func TestPostFromUnbounded_BalanceReadDisables(t *testing.T) {
	// A balance() read anywhere disables the pass (the source debit becomes
	// observable). Uses @world twice: one send + one metadata balance read.
	src := `set_tx_meta("b", balance(@world, USD/2))
send [USD/2 42] (source = @world destination = @dest)`
	dump := optDump(t, src)
	if strings.Contains(dump, "post_from_unbounded") {
		t.Fatalf("balance() present must disable the pass, got:\n%s", dump)
	}
}

func TestPostFromUnbounded_NonWorldReusedIneligible(t *testing.T) {
	// @a unbounded in two sends: the second (bounded-capable) reuse means the
	// dropped debit could be observed, so a non-world reused source is skipped.
	src := `send [USD/2 42] (source = @a allowing unbounded overdraft destination = @x)
send [USD/2 42] (source = @a allowing unbounded overdraft destination = @y)`
	dump := optDump(t, src)
	if strings.Contains(dump, "post_from_unbounded") {
		t.Fatalf("reused non-world source must be skipped, got:\n%s", dump)
	}
}
