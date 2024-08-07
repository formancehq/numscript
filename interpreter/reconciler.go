package interpreter

import (
	"fmt"
	"math/big"
	"slices"
)

type Posting struct {
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Amount      *big.Int `json:"amount"`
	Asset       string   `json:"asset"`
}

type ReconcileError struct {
	Receiver  Receiver
	Receivers []Receiver
}

func (e ReconcileError) Error() string {
	return fmt.Sprintf("Error reconciling senders and getters (receiver = %#v ; receivers = %v)", e.Receiver, e.Receivers)
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

func (r *Receiver) String() string {
	return fmt.Sprintf("<[%s %s] from  %s>", r.Asset, r.Monetary.String(), r.Name)
}

func Reconcile(senders []Sender, receivers []Receiver) ([]Posting, error) {
	var postings []Posting

	for {
		receiver, empty := popStack(&receivers)
		if empty {
			break
		}

		// Ugly workaround
		if receiver.Name == "<kept>" {
			continue
		}

		sender, empty := popStack(&senders)
		if empty {
			return nil, ReconcileError{
				Receiver:  receiver,
				Receivers: receivers,
			}
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
