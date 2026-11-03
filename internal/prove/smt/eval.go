package smt

import (
	"fmt"
	"math/big"
)

// Value is a concrete Int or Bool.
type Value struct {
	Sort Sort
	Int  *big.Int
	Bool bool
}

func IntValue(v *big.Int) Value { return Value{Sort: SortInt, Int: new(big.Int).Set(v)} }
func BoolValue(b bool) Value    { return Value{Sort: SortBool, Bool: b} }

func (v Value) String() string {
	if v.Sort == SortBool {
		if v.Bool {
			return "true"
		}
		return "false"
	}
	return v.Int.String()
}

// Evaluator evaluates terms under an assignment of the script's free symbols,
// resolving defined names through the script and memoizing them.
type Evaluator struct {
	script *Script
	inputs map[string]Value
	memo   map[string]Value
}

func NewEvaluator(script *Script, inputs map[string]Value) *Evaluator {
	return &Evaluator{script: script, inputs: inputs, memo: map[string]Value{}}
}

func (e *Evaluator) Eval(t Term) (Value, error) {
	switch t := t.(type) {
	case intLit:
		return IntValue(t.v), nil
	case boolLit:
		return BoolValue(t.v), nil
	case symbol:
		if v, ok := e.memo[t.name]; ok {
			return v, nil
		}
		if v, ok := e.inputs[t.name]; ok {
			return v, nil
		}
		def, ok := e.script.Definition(t.name)
		if !ok {
			return Value{}, fmt.Errorf("smt: symbol %s has no value", t.name)
		}
		v, err := e.Eval(def)
		if err != nil {
			return Value{}, err
		}
		e.memo[t.name] = v
		return v, nil
	case app:
		return e.evalApp(t)
	}
	return Value{}, fmt.Errorf("smt: unknown term")
}

func (e *Evaluator) ints(args []Term) ([]*big.Int, error) {
	out := make([]*big.Int, len(args))
	for i, a := range args {
		v, err := e.Eval(a)
		if err != nil {
			return nil, err
		}
		if v.Sort != SortInt {
			return nil, fmt.Errorf("smt: expected an Int")
		}
		out[i] = v.Int
	}
	return out, nil
}

func (e *Evaluator) evalApp(t app) (Value, error) {
	switch t.op {
	case "+", "-", "*", "div", "mod":
		xs, err := e.ints(t.args)
		if err != nil {
			return Value{}, err
		}
		r := new(big.Int)
		switch t.op {
		case "+":
			r.Add(xs[0], xs[1])
		case "-":
			if len(xs) == 1 {
				r.Neg(xs[0])
			} else {
				r.Sub(xs[0], xs[1])
			}
		case "*":
			r.Mul(xs[0], xs[1])
		case "div":
			r.Div(xs[0], xs[1])
		case "mod":
			r.Mod(xs[0], xs[1])
		}
		return IntValue(r), nil
	case "<", "<=":
		xs, err := e.ints(t.args)
		if err != nil {
			return Value{}, err
		}
		c := xs[0].Cmp(xs[1])
		if t.op == "<" {
			return BoolValue(c < 0), nil
		}
		return BoolValue(c <= 0), nil
	case "=":
		a, err := e.Eval(t.args[0])
		if err != nil {
			return Value{}, err
		}
		b, err := e.Eval(t.args[1])
		if err != nil {
			return Value{}, err
		}
		if a.Sort == SortBool {
			return BoolValue(a.Bool == b.Bool), nil
		}
		return BoolValue(a.Int.Cmp(b.Int) == 0), nil
	case "not":
		a, err := e.Eval(t.args[0])
		if err != nil {
			return Value{}, err
		}
		return BoolValue(!a.Bool), nil
	case "and", "or":
		for _, a := range t.args {
			v, err := e.Eval(a)
			if err != nil {
				return Value{}, err
			}
			if t.op == "and" && !v.Bool {
				return BoolValue(false), nil
			}
			if t.op == "or" && v.Bool {
				return BoolValue(true), nil
			}
		}
		return BoolValue(t.op == "and"), nil
	case "=>":
		a, err := e.Eval(t.args[0])
		if err != nil {
			return Value{}, err
		}
		if !a.Bool {
			return BoolValue(true), nil
		}
		return e.Eval(t.args[1])
	case "ite":
		c, err := e.Eval(t.args[0])
		if err != nil {
			return Value{}, err
		}
		if c.Bool {
			return e.Eval(t.args[1])
		}
		return e.Eval(t.args[2])
	}
	return Value{}, fmt.Errorf("smt: unknown operator %s", t.op)
}
