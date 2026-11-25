package encode

import (
	"math/big"
	"sort"

	"github.com/formancehq/numscript/internal/prove/smt"
)

// strVal is a str register: a constant known at encode time, or not.
type strVal struct {
	known bool
	s     string
}

// linForm is a portion register as an exact rational linear form over Int
// terms: sum of coef*term plus konst. It is what keeps portions in QF_LIA.
type linTerm struct {
	key string
	t   smt.Term
	c   *big.Rat
}

type linForm struct {
	terms []linTerm
	konst *big.Rat
}

func linConst(r *big.Rat) *linForm {
	return &linForm{konst: new(big.Rat).Set(r)}
}

func linOfInt(t smt.Term) *linForm {
	if n, ok := smt.AsInt(t); ok {
		return linConst(new(big.Rat).SetInt(n))
	}
	return &linForm{
		terms: []linTerm{{key: smt.Print(t), t: t, c: big.NewRat(1, 1)}},
		konst: new(big.Rat),
	}
}

func (f *linForm) isConst() bool { return len(f.terms) == 0 }

func (f *linForm) clone() *linForm {
	out := &linForm{konst: new(big.Rat).Set(f.konst)}
	for _, t := range f.terms {
		out.terms = append(out.terms, linTerm{t.key, t.t, new(big.Rat).Set(t.c)})
	}
	return out
}

func (f *linForm) scale(r *big.Rat) *linForm {
	out := &linForm{konst: new(big.Rat).Mul(f.konst, r)}
	if r.Sign() == 0 {
		return out
	}
	for _, t := range f.terms {
		out.terms = append(out.terms, linTerm{t.key, t.t, new(big.Rat).Mul(t.c, r)})
	}
	return out
}

func (f *linForm) add(g *linForm) *linForm {
	out := f.clone()
	out.konst.Add(out.konst, g.konst)
	for _, gt := range g.terms {
		found := false
		for i := range out.terms {
			if out.terms[i].key == gt.key {
				out.terms[i].c.Add(out.terms[i].c, gt.c)
				found = true
				break
			}
		}
		if !found {
			out.terms = append(out.terms, linTerm{gt.key, gt.t, new(big.Rat).Set(gt.c)})
		}
	}
	// drop cancelled terms, keep a deterministic order
	kept := out.terms[:0]
	for _, t := range out.terms {
		if t.c.Sign() != 0 {
			kept = append(kept, t)
		}
	}
	out.terms = kept
	sort.Slice(out.terms, func(i, j int) bool { return out.terms[i].key < out.terms[j].key })
	return out
}

func (f *linForm) sub(g *linForm) *linForm {
	return f.add(g.scale(big.NewRat(-1, 1)))
}

// mul is defined when at least one side is constant.
func (f *linForm) mul(g *linForm) (*linForm, bool) {
	switch {
	case f.isConst():
		return g.scale(f.konst), true
	case g.isConst():
		return f.scale(g.konst), true
	}
	return nil, false
}

// numerator returns the integer-coefficient form D*f and D > 0, the common
// denominator of every coefficient.
func (f *linForm) numerator() (smt.Term, *big.Int) {
	d := new(big.Int).Set(f.konst.Denom())
	for _, t := range f.terms {
		d.Mul(d, new(big.Int).Div(t.c.Denom(), new(big.Int).GCD(nil, nil, d, t.c.Denom())))
	}
	scale := func(r *big.Rat) *big.Int {
		n := new(big.Int).Mul(r.Num(), d)
		return n.Div(n, r.Denom())
	}
	sum := smt.Int(scale(f.konst))
	for _, t := range f.terms {
		sum = smt.Add(sum, smt.MulConst(scale(t.c), t.t))
	}
	return sum, d
}

// toInt floors the form, like portion_to_int.
func (f *linForm) toInt() smt.Term {
	num, d := f.numerator()
	return smt.DivConst(num, d)
}

func (f *linForm) sign() smt.Term {
	num, _ := f.numerator()
	return num
}

func (f *linForm) lt(g *linForm) smt.Term {
	return smt.Lt(f.sub(g).sign(), smt.Int64(0))
}

func (f *linForm) eq(g *linForm) smt.Term {
	return smt.Eq(f.sub(g).sign(), smt.Int64(0))
}

func (f *linForm) String() string {
	return smt.Print(f.sign()) + "/" + f.denomString()
}

func (f *linForm) denomString() string {
	_, d := f.numerator()
	return d.String()
}

// regVal is the value of a logical register, in the representation of its
// bank.
type regVal struct {
	t smt.Term // int or bool
	s strVal
	p *linForm
}
