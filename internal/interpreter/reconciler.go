package interpreter

import (
	"fmt"
	"math/big"
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
}

type Receiver struct {
	Name   string
	Amount *big.Int
}

func Reconcile(asset string, senders []Sender, receivers []Receiver) []Posting {
	fundsStack := newFundsStack(senders)

	var postings []Posting

	for _, receiver := range receivers {
		senders := fundsStack.Pull(receiver.Amount)

		if receiver.Name == KEPT_ADDR {
			continue
		}

		for _, sender := range senders {
			postings = append(postings, Posting{
				Source:      sender.Name,
				Destination: receiver.Name,
				Amount:      sender.Amount,
				Asset:       asset,
			})
		}
	}

	return postings
}
