package interpreter

import (
	"fmt"
	"math/big"
	"slices"
)

type Sender struct {
	Name   string
	Amount *big.Int
	Color  string
}

type fundsQueue struct {
	asset   string
	senders []Sender
}

// Create a fundsQueue from a slice of senders.
func newFundsQueue(senders []Sender) fundsQueue {
	queue := fundsQueue{
		senders: []Sender{},
	}
	queue.Push(senders...)
	return queue
}

// Push senders to this fundsQueue
func (s *fundsQueue) Push(senders ...Sender) {
	for _, sender := range senders {
		s.PushOne(sender)
	}
}

// Push a single sender to this queue
func (s *fundsQueue) PushOne(sender Sender) {
	if sender.Amount.Cmp(big.NewInt(0)) == 0 {
		return
	}
	if len(s.senders) == 0 {
		s.senders = []Sender{sender}
		return
	}
	last := s.senders[len(s.senders)-1]
	if last.Name == sender.Name && last.Color == sender.Color {
		last.Amount.Add(last.Amount, sender.Amount)
	} else {
		s.senders = append(s.senders, sender)
	}
}

// Pull everything from this queue
func (s *fundsQueue) PullAll() []Sender {
	senders := s.senders
	s.senders = []Sender{}
	return senders
}

// Pull at most maxAmount from this queue, with any color
func (s *fundsQueue) PullAnything(maxAmount *big.Int) []Sender {
	return s.Pull(maxAmount, nil)
}

func (s *fundsQueue) PullColored(maxAmount *big.Int, color string) []Sender {
	return s.Pull(maxAmount, &color)
}
func (s *fundsQueue) PullUncolored(maxAmount *big.Int) []Sender {
	return s.PullColored(maxAmount, "")
}

// Pull at most maxAmount from this queue, with the given color
func (s *fundsQueue) Pull(maxAmount *big.Int, color *string) []Sender {
	// clone so that we can manipulate this arg
	maxAmount = new(big.Int).Set(maxAmount)

	// TODO preallocate for perfs
	out := newFundsQueue([]Sender{})
	offset := 0

	for maxAmount.Cmp(big.NewInt(0)) != 0 && len(s.senders) > offset {

		frontSender := s.senders[offset]

		if color != nil && frontSender.Color != *color {
			offset += 1
			continue
		}

		switch frontSender.Amount.Cmp(maxAmount) {
		case -1: // not enough
			maxAmount.Sub(maxAmount, frontSender.Amount)
			out.Push(frontSender)
			s.senders = slices.Delete(s.senders, offset, offset+1)
		case 1: // more than enough
			out.Push(Sender{
				Name:   frontSender.Name,
				Amount: maxAmount,
				Color:  frontSender.Color,
			})
			s.senders[offset].Amount.Sub(s.senders[offset].Amount, maxAmount)
			return out.senders
		case 0: // exactly enough
			out.Push(s.senders[offset])
			s.senders = slices.Delete(s.senders, offset, offset+1)
			return out.senders
		}
	}

	return out.senders
}

// Clone the queue so that you can safely mutate one without mutating the other
func (s fundsQueue) Clone() fundsQueue {
	return fundsQueue{
		senders: slices.Clone(s.senders),
		asset:   s.asset,
	}
}

func (s fundsQueue) String() string {
	out := ">"
	for i, sender := range s.senders {
		if sender.Color == "" {
			out += fmt.Sprintf("%v from %v", sender.Amount, sender.Name)
		} else {
			out += fmt.Sprintf("%v from %v\\%v", sender.Amount, sender.Name, sender.Color)
		}
		if i != len(s.senders)-1 {
			out += ", "
		}
	}
	out += ">"
	return out
}
