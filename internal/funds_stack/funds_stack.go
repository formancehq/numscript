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
	slices.Reverse(senders)
	return FundsStack{
		senders: senders,
	}
}

func (s *FundsStack) compactTop() {
	for len(s.senders) >= 2 {
		first := s.senders[len(s.senders)-1]
		second := s.senders[len(s.senders)-2]

		if second.Amount.Cmp(big.NewInt(0)) == 0 {
			s.senders = append(s.senders[0:len(s.senders)-2], first)
			continue
		}

		if first.Name != second.Name {
			return
		}

		s.senders = append(s.senders[0:len(s.senders)-2], Sender{
			Name:   first.Name,
			Amount: new(big.Int).Add(first.Amount, second.Amount),
		})
	}
}

func (s *FundsStack) Pull(requiredAmount *big.Int) []Sender {
	// clone so that we can manipulate this arg
	requiredAmount = new(big.Int).Set(requiredAmount)

	// TODO preallocate for perfs
	var out []Sender

	for requiredAmount.Cmp(big.NewInt(0)) != 0 && len(s.senders) != 0 {
		s.compactTop()

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
			if available.Amount.Cmp(big.NewInt(0)) == 0 {
				s.senders = s.senders[:len(s.senders)-1]
			}
			return out
		}

	}

	return out
}
