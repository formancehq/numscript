package interpreter

import (
	"math/big"
	"slices"
)

type Sender struct {
	Name   string
	Amount *big.Int
	Color  string
}

type fundsStack struct {
	senders []Sender
}

func newFundsStack(senders []Sender) fundsStack {
	senders = slices.Clone(senders)

	// TODO do not modify arg
	// TODO clone big ints so that we can manipulate them
	slices.Reverse(senders)
	return fundsStack{
		senders: senders,
	}
}

func (s *fundsStack) compactTop() {
	for len(s.senders) >= 2 {
		first := s.senders[len(s.senders)-1]
		second := s.senders[len(s.senders)-2]

		if second.Amount.Cmp(big.NewInt(0)) == 0 {
			s.senders = append(s.senders[0:len(s.senders)-2], first)
			continue
		}

		if first.Name != second.Name || first.Color != second.Color {
			return
		}

		s.senders = append(s.senders[0:len(s.senders)-2], Sender{
			Name:   first.Name,
			Color:  first.Color,
			Amount: new(big.Int).Add(first.Amount, second.Amount),
		})
	}
}

func (s *fundsStack) Pull(requiredAmount *big.Int) []Sender {
	// clone so that we can manipulate this arg
	requiredAmount = new(big.Int).Set(requiredAmount)

	// TODO preallocate for perfs
	var out []Sender

	for requiredAmount.Cmp(big.NewInt(0)) != 0 && len(s.senders) != 0 {
		s.compactTop()

		available := s.senders[len(s.senders)-1]
		s.senders = s.senders[:len(s.senders)-1]

		switch available.Amount.Cmp(requiredAmount) {
		case -1: // not enough:
			out = append(out, available)
			requiredAmount.Sub(requiredAmount, available.Amount)

		case 1: // more than enough
			s.senders = append(s.senders, Sender{
				Name:   available.Name,
				Color:  available.Color,
				Amount: new(big.Int).Sub(available.Amount, requiredAmount),
			})
			fallthrough

		case 0: // exactly the same
			out = append(out, Sender{
				Name:   available.Name,
				Color:  available.Color,
				Amount: new(big.Int).Set(requiredAmount),
			})
			return out
		}

	}

	return out
}
