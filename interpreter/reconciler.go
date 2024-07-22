package interpreter

import (
	"math/big"
	"slices"
)

type Allotment[T interface{}] struct {
	Ratio big.Rat
	Value T
}

type Posting struct {
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Amount      *big.Int `json:"amount"`
	Asset       string   `json:"asset"`
}

type ReconcileError struct {
	error
}

type Sender struct {
	Name     string
	Monetary *big.Int
	Asset    string
}

type Receiver struct {
	Name     string
	Monetary *big.Int
	Asset    string
}

func Reconcile(senders []Sender, receivers []Receiver) ([]Posting, error) {
	var postings []Posting

	for {
		receiver, empty := popStack(&receivers)
		if empty {
			break
		}

		sender, empty := popStack(&senders)
		if empty {
			return nil, ReconcileError{}
		}

		var postingAmount big.Int
		snd := (*big.Int)(sender.Monetary)
		rcv := (*big.Int)(receiver.Monetary)

		switch snd.Cmp(rcv) {
		case 0: /* sender.Monetary == receiver.Monetary */
			postingAmount = *sender.Monetary
		case -1: /* sender.Monetary < receiver.Monetary */
			var monetary big.Int
			receivers = append(receivers, Receiver{
				Name:     receiver.Name,
				Monetary: monetary.Sub(receiver.Monetary, sender.Monetary),
				Asset:    sender.Asset,
			})
			postingAmount = *sender.Monetary
		case 1: /* sender.Monetary > receiver.Monetary */
			var monetary big.Int
			senders = append(senders, Sender{
				Name:     sender.Name,
				Monetary: monetary.Sub(sender.Monetary, receiver.Monetary),
				Asset:    sender.Asset,
			})
			postingAmount = *receiver.Monetary
		}

		if postingAmt := big.Int(postingAmount); postingAmt.BitLen() == 0 {
			continue
		}

		var postingToMerge *Posting
		if len(postings) != 0 {
			posting := &postings[len(postings)-1]
			if posting.Source == sender.Name && posting.Destination == receiver.Name {
				postingToMerge = posting
			}
		}

		if postingToMerge == nil {
			postings = append(postings, Posting{
				Source:      sender.Name,
				Destination: receiver.Name,
				Amount:      &postingAmount,
				Asset:       sender.Asset,
			})
		} else {
			// postingToMerge.Amount += postingAmount
			postingToMerge.Amount.Add(postingToMerge.Amount, &postingAmount)
		}
	}

	slices.Reverse(postings)
	return postings, nil
}

func popStack[T any](stack *[]T) (T, bool) {
	l := len(*stack)
	if l == 0 {
		var t T
		return t, true
	}

	popped := (*stack)[l-1]
	*stack = (*stack)[:l-1]
	return popped, false
}
