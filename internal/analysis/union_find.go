package analysis

import (
	"fmt"

	"github.com/formancehq/numscript/internal/utils"
)

type Type interface {
	Resolve() Type
}

var _ Type = (*TVar)(nil)
var _ Type = (*Asset)(nil)

// Impls

func (t *TVar) Resolve() Type {
	if t.resolution == nil {
		return t
	}

	resolved := t.resolution.Resolve()
	if resolved == t {
		return t
	}

	// This bit doesn't change the behaviour but
	t.resolution = resolved

	return resolved
}

type TVar struct {
	resolution Type
}

type Asset string

func (a *Asset) Resolve() Type {
	return a
}

func Unify(t1 Type, t2 Type) (ok bool) {
	t1 = t1.Resolve()
	t2 = t2.Resolve()

	switch t1 := t1.(type) {
	case *Asset:
		switch t2 := t2.(type) {
		case *Asset:
			return string(*t1) == string(*t2)

		case *TVar:
			return Unify(t2, t1)
		}

	case *TVar:
		// We must avoid cycles when unifying a var with itself
		if t1 == t2 {
			return true
		}

		// t1 is a tvar, so we can always unify it with t2
		t1.resolution = t2
		return true
	}

	return false
}

func TypeToString(r Type) string {
	r = r.Resolve()
	switch r := r.(type) {
	case *TVar:
		return fmt.Sprintf("'%p", r)

	case *Asset:
		return string(*r)
	}

	return utils.NonExhaustiveMatchPanic[string](r)
}
