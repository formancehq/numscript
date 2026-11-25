package ir

import (
	"math/big"

	"github.com/formancehq/numscript/internal/funds"
)

// dropProvenAsserts removes a check that passes on every path reaching it. A
// check that may fail stays where it is, so the first error a run hits is
// unchanged.
func dropProvenAsserts(instrs []Instr) []Instr {
	drop := map[int]bool{}
	walkConsts(instrs, func(i int, env constEnv) {
		if env != nil && checkPasses(instrs[i], env) {
			drop[i] = true
		}
	})
	if len(drop) == 0 {
		return nil
	}
	out := make([]Instr, 0, len(instrs)-len(drop))
	for i, in := range instrs {
		if !drop[i] {
			out = append(out, in)
		}
	}
	return out
}

func checkPasses(in Instr, env constEnv) bool {
	intOf := func(r Reg) (*big.Int, bool) { v, ok := env[r].(*big.Int); return v, ok }
	strOf := func(r Reg) (string, bool) { v, ok := env[r].(string); return v, ok }
	portionOf := func(r Reg) (*big.Rat, bool) { v, ok := env[r].(*big.Rat); return v, ok }

	switch in := in.(type) {
	case AssertNonNegativeAmount:
		v, ok := intOf(in.Amount)
		return ok && v.Sign() >= 0
	case AssertNonNegativeBalance:
		v, ok := intOf(in.Balance)
		return ok && v.Sign() >= 0
	case AssertNonNegativePortion:
		v, ok := portionOf(in.Portion)
		return ok && v.Sign() >= 0
	case AssertLeftover:
		v, ok := portionOf(in.Portion)
		return ok && v.Sign() >= 0 && (!in.Exact || v.Sign() == 0)
	case AssertSameAsset:
		l, lok := strOf(in.Left)
		r, rok := strOf(in.Right)
		return lok && rok && l == r
	case AssertValidAccount:
		v, ok := strOf(in.Account)
		return ok && funds.ValidateAccount(v)
	case AssertValidColor:
		v, ok := strOf(in.Color)
		return ok && funds.ValidateColor(v)
	case AssertValidScope:
		v, ok := strOf(in.Scope)
		return ok && funds.ValidateScope(v)
	case AssertUnscoped:
		v, ok := strOf(in.Scope)
		return ok && v == ""
	case CheckEnoughFunds:
		g, gok := intOf(in.Got)
		n, nok := intOf(in.Needed)
		return gok && nok && g.Cmp(n) == 0
	}
	return false
}
