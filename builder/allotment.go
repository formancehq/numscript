package builder

import "math/big"

// A portion literal, e.g. `1/3`. Unlike accounts/assets/strings/numbers,
// portions are never routed through the vars pool: numscript only allows a
// portion literal, a variable, or `remaining` in allotment position, and
// this builder only ever emits the literal form.
type Portion struct {
	num, denom *big.Int
}

func NewPortion(num, denom *big.Int) Portion {
	return Portion{num: num, denom: denom}
}

func (p Portion) render(env *env) {
	env.builder.WriteString(p.num.String())
	env.builder.WriteByte('/')
	env.builder.WriteString(p.denom.String())
}

// payload is the underlying shape shared by Source and KeptOrDest: both are
// render closures. AllotmentClause is generic over it so the same type can
// back both `sourceAllotment` and `destinationAllotment` clauses.
type payload interface {
	~func(*env, int)
}

// One clause of an allotment block: a portion plus its payload (a Source in
// source position, a KeptOrDest in destination position).
type AllotmentClause[T payload] struct {
	Portion Portion
	Payload T
}
