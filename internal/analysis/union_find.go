package analysis

import (
	"fmt"

	"github.com/formancehq/numscript/internal/utils"
)

type Type interface {
	Resolve() Type
}

var _ Type = (*TVar)(nil)
var _ Type = (*TAsset)(nil)

// Impls

func (t *TVar) Resolve() Type {
	if t.resolution == nil {
		return t
	}

	resolved := t.resolution.Resolve()
	if resolved == t {
		return t
	}

	// This bit doesn't change the behaviour but allows to return the path right away
	// the next time we call Resolve()
	t.resolution = resolved

	return resolved
}

type TVar struct {
	resolution Type
}

type TAsset string

func (a *TAsset) Resolve() Type {
	return a
}

func Unify(t1 Type, t2 Type) (ok bool) {
	t1 = t1.Resolve()
	t2 = t2.Resolve()

	switch t1 := t1.(type) {
	case *TAsset:
		switch t2 := t2.(type) {
		case *TAsset:
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

	case *TAsset:
		return string(*r)
	}

	return utils.NonExhaustiveMatchPanic[string](r)
}

type TypePrinter struct {
	nextId uint16
	store  map[*TVar]uint16
}

func NewTypePrinter() TypePrinter {
	return TypePrinter{
		nextId: 0,
		store:  make(map[*TVar]uint16),
	}
}

func (p *TypePrinter) getVarId(v *TVar) uint16 {
	prev, ok := p.store[v]
	if ok {
		return prev
	}

	id := p.nextId
	p.nextId += 1
	p.store[v] = id
	return id
}

func (p *TypePrinter) Print(r Type) string {
	r = r.Resolve()
	switch r := r.(type) {
	case *TVar:
		return fmt.Sprintf("asset_%d", p.getVarId(r))

	case *TAsset:
		return string(*r)

	}

	return utils.NonExhaustiveMatchPanic[string](r)

}
