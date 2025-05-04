package funds_stack

import (
	"math/big"
	"slices"
)

type Sender struct {
	Name   string
	Amount *big.Int
}

type FundsStack struct {
	senders []Sender
}

func NewFundsStack(senders []Sender) FundsStack {
	// TODO do not modify arg
	// TODO clone big ints so that we can manipulate them
	// TODO compact
	slices.Reverse(senders)
	return FundsStack{
		senders: senders,
	}
}

// TODO should return err as well? When not enough funds
func (s *FundsStack) Pull(requiredAmount *big.Int) []Sender {
	// clone so that we can manipulate this arg
	requiredAmount = new(big.Int).Set(requiredAmount)

	// TODO preallocate for perfs
	var out []Sender

	for requiredAmount.Cmp(big.NewInt(0)) != 0 && len(s.senders) != 0 {
		available := s.senders[len(s.senders)-1]

		switch available.Amount.Cmp(requiredAmount) {
		case -1: // not enough:
			out = append(out, available)
			s.senders = s.senders[:len(s.senders)-1]
			requiredAmount.Sub(requiredAmount, available.Amount)

		default: // enough:
			out = append(out, Sender{
				Name:   available.Name,
				Amount: requiredAmount,
			})
			available.Amount.Sub(available.Amount, requiredAmount)
			return out
		}

	}

	return out
}

// Get the total quantity of allocated funds
func (*FundsStack) Size() *big.Int {
	panic("TODO implement .Size()")
}
