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

		if receiver.Monetary == nil {
			slices.Reverse(senders)
			for _, sender := range senders {
				// empty all the senders
				postings = append(postings, Posting{
					Source:      sender.Name,
					Destination: receiver.Name,
					Amount:      sender.Monetary,
					Asset:       sender.Asset,
				})
			}
			break
		}

		// Ugly workaround
		if receiver.Name == "<kept>" {
			// TODO test kept + send*

			// reduce sender by amt

			sender, empty := popStack(&senders)
			if !empty {
				var newMon big.Int
				newMon.Sub(sender.Monetary, receiver.Monetary)
				if newMon.Sign() == -1 {
					panic("NEG" + newMon.String())
				}

				senders = append(senders, Sender{
					Name:     sender.Name,
					Asset:    sender.Asset,
					Monetary: &newMon,
				})
			}
			continue
		}

		rcv := (*big.Int)(receiver.Monetary)

		sender, empty := popStack(&senders)
		if empty {
			isReceivedAmtZero := rcv.Cmp(big.NewInt(0)) == 0
			if isReceivedAmtZero {
				return postings, nil
			}

			return postings, nil
			// return nil, ReconcileError{
			// 	Receiver:  receiver,
			// 	Receivers: receivers,
			// }
		}

		snd := (*big.Int)(sender.Monetary)

		var postingAmount big.Int
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
