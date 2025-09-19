package analysis_test

import (
	"testing"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/stretchr/testify/require"
)

func TestResolveConcrete(t *testing.T) {
	t1 := analysis.TAsset("USD")
	out := t1.Resolve()
	require.Equal(t, &t1, out)
}

func TestUnifyConcreteWhenNotSame(t *testing.T) {
	t1 := analysis.TAsset("USD")
	t2 := analysis.TAsset("EUR")
	ok := analysis.Unify(&t1, &t2)
	require.False(t, ok)
}

func TestUnifyConcreteWhenSame(t *testing.T) {
	t1 := analysis.TAsset("USD")
	t2 := analysis.TAsset("USD")
	ok := analysis.Unify(&t1, &t2)
	require.True(t, ok)
}

func TestUnifyItselfIsNoop(t *testing.T) {
	t1 := &analysis.TVar{}
	ok := analysis.Unify(t1, t1)
	require.True(t, ok)

	require.Same(t, t1.Resolve(), t1)
}

func TestResolveUnbound(t *testing.T) {
	t1 := &analysis.TVar{}
	require.Same(t, t1.Resolve(), t1)
}

func TestUnifyVarWithConcrete(t *testing.T) {
	t1 := &analysis.TVar{}
	t2 := analysis.TAsset("USD")

	ok := analysis.Unify(t1, &t2)
	require.True(t, ok)

	require.Same(t, t1.Resolve(), &t2)
}

func TestUnifyTransitive(t *testing.T) {
	t1 := &analysis.TVar{}
	t2 := &analysis.TVar{}
	t3 := &analysis.TVar{}

	// t1->t2->t3

	ok := analysis.Unify(t1, t2)
	require.True(t, ok)

	ok = analysis.Unify(t1, t3)
	require.True(t, ok)

	t4 := analysis.TAsset("USD")
	ok = analysis.Unify(t1, &t4)
	require.True(t, ok)

	require.Same(t, t1.Resolve(), &t4)
	require.Same(t, t2.Resolve(), &t4)
	require.Same(t, t3.Resolve(), &t4)
}

func TestUnifyTransitiveInverse(t *testing.T) {
	t1 := &analysis.TVar{}
	t2 := &analysis.TVar{}
	t3 := &analysis.TVar{}

	// t1->t2->t3

	ok := analysis.Unify(t1, t2)
	require.True(t, ok)

	ok = analysis.Unify(t1, t3)
	require.True(t, ok)

	t4 := analysis.TAsset("USD")
	ok = analysis.Unify(t3, &t4)
	require.True(t, ok)

	require.Same(t, t1.Resolve(), &t4)
	require.Same(t, t2.Resolve(), &t4)
	require.Same(t, t3.Resolve(), &t4)
}
