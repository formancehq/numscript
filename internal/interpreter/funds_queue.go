package interpreter

import (
	"math/big"
)

type Sender struct {
	Name   string
	Amount *big.Int
	Color  string
}

type queue[T any] struct {
	Head T
	Tail *queue[T]

	// Instead of keeping a single ref of the lastCell and updating the invariant on every push/pop operation,
	// we keep a cache of the last cell on every cell.
	// This makes code much easier and we don't risk breaking the invariant and producing wrong results and other subtle issues
	//
	// While, unlike keeping a single reference (like golang's queue `container/list` package does), this is not always O(1),
	// the amortized time should still be O(1) (the number of steps of traversal while searching the last elem is not higher than the number of .Push() calls)
	lastCell *queue[T]
}

func (s *queue[T]) getLastCell() *queue[T] {
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

func fromSlice[T any](slice []T) *queue[T] {
	var ret *queue[T]
	// TODO use https://pkg.go.dev/slices#Backward in golang 1.23
	for i := len(slice) - 1; i >= 0; i-- {
		ret = &queue[T]{
			Head: slice[i],
			Tail: ret,
		}
	}
	return ret
}

type fundsQueue struct {
	senders *queue[Sender]
}

func newFundsQueue(senders []Sender) fundsQueue {
	return fundsQueue{
		senders: fromSlice(senders),
	}
}

func (s *fundsQueue) compactTop() {
	for s.senders != nil && s.senders.Tail != nil {

		first := s.senders.Head
		second := s.senders.Tail.Head

		if second.Amount.Cmp(big.NewInt(0)) == 0 {
			s.senders = &queue[Sender]{Head: first, Tail: s.senders.Tail.Tail}
			continue
		}

		if first.Name != second.Name || first.Color != second.Color {
			return
		}

		s.senders = &queue[Sender]{
			Head: Sender{
				Name:   first.Name,
				Color:  first.Color,
				Amount: new(big.Int).Add(first.Amount, second.Amount),
			},
			Tail: s.senders.Tail.Tail,
		}
	}
}

func (s *fundsQueue) PullAll() []Sender {
	var senders []Sender
	for s.senders != nil {
		senders = append(senders, s.senders.Head)
		s.senders = s.senders.Tail
	}
	return senders
}

func (s *fundsQueue) Push(senders ...Sender) {
	newTail := fromSlice(senders)
	if s.senders == nil {
		s.senders = newTail
	} else {
		cell := s.senders.getLastCell()
		cell.Tail = newTail
	}
}

// PullAnything is the entry point used by pushReceiver to drain senders
// for a given amount, regardless of color. Each pulled Sender keeps its own
// Color so the receiver-side posting can carry it untouched.
func (s *fundsQueue) PullAnything(requiredAmount *big.Int) []Sender {
	return s.Pull(requiredAmount)
}

// Pull drains up to requiredAmount from the head of the queue. It does NOT
// verify that the queue holds at least requiredAmount — it simply returns
// whatever is there. The caller is responsible for checking completeness:
// tryTakingExact raises MissingFundsErr when the total sent is below the
// requested amount.
//
// The queue itself is bounded upstream: every pushSender call has already
// been capped by CalculateSafeWithdraw against the source's (asset, color)
// balance, so the queue never holds more than the source can legitimately
// commit. Color is carried by each Sender — Pull preserves it on the way out
// without ever inspecting it.
func (s *fundsQueue) Pull(requiredAmount *big.Int) []Sender {
	// clone so that we can manipulate this arg
	requiredAmount = new(big.Int).Set(requiredAmount)

	// TODO preallocate for perfs
	var out []Sender

	for requiredAmount.Cmp(big.NewInt(0)) != 0 && s.senders != nil {
		s.compactTop()

		available := s.senders.Head
		s.senders = s.senders.Tail

		switch available.Amount.Cmp(requiredAmount) {
		case -1: // not enough:
			out = append(out, available)
			requiredAmount.Sub(requiredAmount, available.Amount)

		case 1: // more than enough
			s.senders = &queue[Sender]{
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

// Clone the queue so that you can safely mutate one without mutating the other
func (s fundsQueue) Clone() fundsQueue {
	fq := newFundsQueue(nil)

	senders := s.senders
	for senders != nil {
		fq.Push(senders.Head)
		senders = senders.Tail
	}

	return fq
}
