package interpreter

import (
	"math/big"
)

type Sender struct {
	Name   string
	Amount *big.Int
	Color  string
}

type stack[T any] struct {
	Head T
	Tail *stack[T]
}

func fromSlice[T any](slice []T) *stack[T] {
	// TODO make it stack-safe
	if len(slice) == 0 {
		return nil
	}
	return &stack[T]{
		Head: slice[0],
		Tail: fromSlice(slice[1:]),
	}
}

type fundsStack struct {
	senders *stack[Sender]
}

func newFundsStack(senders []Sender) fundsStack {
	return fundsStack{
		senders: fromSlice(senders),
	}
}

func (s *fundsStack) compactTop() {
	for s.senders != nil && s.senders.Tail != nil {

		first := s.senders.Head
		second := s.senders.Tail.Head

		if second.Amount.Cmp(big.NewInt(0)) == 0 {
			s.senders = &stack[Sender]{Head: first, Tail: s.senders.Tail.Tail}
			continue
		}

		if first.Name != second.Name || first.Color != second.Color {
			return
		}

		s.senders = &stack[Sender]{
			Head: Sender{
				Name:   first.Name,
				Color:  first.Color,
				Amount: new(big.Int).Add(first.Amount, second.Amount),
			},
			Tail: s.senders.Tail.Tail,
		}
	}
}

func (s *fundsStack) PullAll() []Sender {
	var senders []Sender
	for s.senders != nil {
		senders = append(senders, s.senders.Head)
		s.senders = s.senders.Tail
	}
	return senders
}

func getLastCellOrNil(stack *stack[Sender]) *stack[Sender] {
	for {
		if stack.Tail == nil {
			return stack
		}
		stack = stack.Tail
	}
}

// TODO(perf) we can keep the reference of the last cell to have an O(1) push
func (s *fundsStack) Push(senders ...Sender) {
	newTail := fromSlice(senders)
	if s.senders == nil {
		s.senders = newTail
	} else {
		cell := getLastCellOrNil(s.senders)
		cell.Tail = newTail
	}
}

func (s *fundsStack) PullAnything(requiredAmount *big.Int) []Sender {
	return s.Pull(requiredAmount, nil)
}

func (s *fundsStack) PullColored(requiredAmount *big.Int, color string) []Sender {
	return s.Pull(requiredAmount, &color)
}
func (s *fundsStack) PullUncolored(requiredAmount *big.Int) []Sender {
	return s.PullColored(requiredAmount, "")
}

func (s *fundsStack) Pull(requiredAmount *big.Int, color *string) []Sender {
	// clone so that we can manipulate this arg
	requiredAmount = new(big.Int).Set(requiredAmount)

	// TODO preallocate for perfs
	var out []Sender

	for requiredAmount.Cmp(big.NewInt(0)) != 0 && s.senders != nil {
		s.compactTop()

		available := s.senders.Head
		s.senders = s.senders.Tail

		if color != nil && available.Color != *color {
			out1 := s.Pull(requiredAmount, color)
			s.senders = &stack[Sender]{
				Head: available,
				Tail: s.senders,
			}
			out = append(out, out1...)
			break
		}

		switch available.Amount.Cmp(requiredAmount) {
		case -1: // not enough:
			out = append(out, available)
			requiredAmount.Sub(requiredAmount, available.Amount)

		case 1: // more than enough
			s.senders = &stack[Sender]{
				Head: Sender{
					Name:   available.Name,
					Color:  available.Color,
					Amount: new(big.Int).Sub(available.Amount, requiredAmount),
				},
				Tail: s.senders,
			}
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

// Treat this stack as debts and filter out senders by "repaying" debts
func (s *fundsStack) RepayWith(credits *fundsStack, asset string) []Posting {
	var postings []Posting

	for s.senders != nil {
		// Peek head from debts and try to pull that much
		hd := s.senders.Head

		senders := credits.Pull(hd.Amount, &hd.Color)
		totalRepayed := big.NewInt(0)
		for _, sender := range senders {
			totalRepayed.Add(totalRepayed, sender.Amount)
			postings = append(postings, Posting{
				Source:      sender.Name,
				Destination: hd.Name,
				Amount:      sender.Amount,
				Asset:       coloredAsset(asset, &sender.Color),
			})
		}

		pulled := s.Pull(totalRepayed, &hd.Color)
		if len(pulled) == 0 {
			break
		}

		// careful: infinite loops possible with different colors
		// break
	}

	return postings
}
