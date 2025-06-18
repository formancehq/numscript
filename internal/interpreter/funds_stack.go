package interpreter

import (
	"math/big"
)

type Sender struct {
	Account AccountValue
	Amount  *big.Int
	Color   string
}

type stack[T any] struct {
	Head T
	Tail *stack[T]

	// Instead of keeping a single ref of the lastCell and updating the invariant on every push/pop operation,
	// we keep a cache of the last cell on every cell.
	// This makes code much easier and we don't risk breaking the invariant and producing wrong results and other subtle issues
	//
	// While, unlike keeping a single reference (like golang's queue `container/list` package does), this is not always O(1),
	// the amortized time should still be O(1) (the number of steps of traversal while searching the last elem is not higher than the number of .Push() calls)
	lastCell *stack[T]
}

func (s *stack[T]) getLastCell() *stack[T] {
	// check if this is the last cell without reading cache first
	if s.Tail == nil {
		return s
	}

	// if not, check if cache is present
	if s.lastCell != nil {
		// even if it is, it may be a stale value (as more values could have been pushed), so we check the value recursively
		lastCell := s.lastCell.getLastCell()
		// we do path compression so that next time we get the path immediately
		s.lastCell = lastCell
		return lastCell
	}

	// if no last value is cached, we traverse recursively to find it
	s.lastCell = s.Tail.getLastCell()
	return s.lastCell
}

func fromSlice[T any](slice []T) *stack[T] {
	var ret *stack[T]
	// TODO use https://pkg.go.dev/slices#Backward in golang 1.23
	for i := len(slice) - 1; i >= 0; i-- {
		ret = &stack[T]{
			Head: slice[i],
			Tail: ret,
		}
	}
	return ret
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

		if first.Account != second.Account || first.Color != second.Color {
			return
		}

		s.senders = &stack[Sender]{
			Head: Sender{
				Account: first.Account,
				Color:   first.Color,
				Amount:  new(big.Int).Add(first.Amount, second.Amount),
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

func (s *fundsStack) Push(senders ...Sender) {
	newTail := fromSlice(senders)
	if s.senders == nil {
		s.senders = newTail
	} else {
		cell := s.senders.getLastCell()
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
					Account: available.Account,
					Color:   available.Color,
					Amount:  new(big.Int).Sub(available.Amount, requiredAmount),
				},
				Tail: s.senders,
			}
			fallthrough

		case 0: // exactly the same
			out = append(out, Sender{
				Account: available.Account,
				Color:   available.Color,
				Amount:  new(big.Int).Set(requiredAmount),
			})
			return out
		}

	}

	return out
}

// Clone the stack so that you can safely mutate one without mutating the other
func (s fundsStack) Clone() fundsStack {
	fs := newFundsStack(nil)

	senders := s.senders
	for senders != nil {
		fs.Push(senders.Head)
		senders = senders.Tail
	}

	return fs
}
