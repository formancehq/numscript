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
	Name   string
	Amount *big.Int
	Color  *string
}

type Receiver struct {
	Name     string
	Monetary *big.Int
}

func Reconcile(asset string, senders []Sender, receivers []Receiver) ([]Posting, InterpreterError) {

	// We reverse senders and receivers once so that we can
	// treat them as stack and push/pop in O(1)
	slices.Reverse(senders)
	slices.Reverse(receivers)
	var postings []Posting

	for {
		receiver, empty := popStack(&receivers)
		if empty {
			break
		}

		// Ugly workaround
		if receiver.Name == KEPT_ADDR {
			sender, empty := popStack(&senders)
			if !empty {
				var newMon big.Int
				newMon.Sub(sender.Amount, receiver.Monetary)
				senders = append(senders, Sender{
					Name:   sender.Name,
					Amount: &newMon,
				})
			}
			continue
		}

		sender, empty := popStack(&senders)
		if empty {
			isReceivedAmtZero := receiver.Monetary.Cmp(big.NewInt(0)) == 0
			if isReceivedAmtZero {
				return postings, nil
			}

			return postings, nil
		}

		snd := (*big.Int)(sender.Amount)

		var postingAmount big.Int
		switch snd.Cmp(receiver.Monetary) {
		case 0: /* sender.Monetary == receiver.Monetary */
			postingAmount = *sender.Amount
		case -1: /* sender.Monetary < receiver.Monetary */
			receivers = append(receivers, Receiver{
				Name:     receiver.Name,
				Monetary: new(big.Int).Sub(receiver.Monetary, sender.Amount),
			})
			postingAmount = *sender.Amount
		case 1: /* sender.Monetary > receiver.Monetary */
			senders = append(senders, Sender{
				Name:   sender.Name,
				Amount: new(big.Int).Sub(sender.Amount, receiver.Monetary),
				Color:  sender.Color,
			})
			postingAmount = *receiver.Monetary
		}

		var postingToMerge *Posting
		if len(postings) != 0 {
			posting := &postings[len(postings)-1]
			if posting.Source == sender.Name && posting.Destination == receiver.Name {
				postingToMerge = posting
			}
		}

		if postingToMerge == nil || postingToMerge.Asset != coloredAsset(asset, sender.Color) {
			postings = append(postings, Posting{
				Source:      sender.Name,
				Destination: receiver.Name,
				Amount:      &postingAmount,
				Asset:       coloredAsset(asset, sender.Color),
			})
		} else {
			// postingToMerge.Amount += postingAmount
			postingToMerge.Amount.Add(postingToMerge.Amount, &postingAmount)
		}
	}

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
