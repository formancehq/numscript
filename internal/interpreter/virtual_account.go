package interpreter

import (
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/utils"
)

type VirtualAccount struct {
	Dbg     string
	credits map[string]*fundsStack
	debits  map[string]*fundsStack
}

func (v VirtualAccount) WithDbg(dbg string) VirtualAccount {
	v.Dbg = dbg
	return v
}
func (v VirtualAccount) String() string {
	var name string
	if v.Dbg != "" {
		name = v.Dbg
	} else {
		name = "anonymous"
	}

	return fmt.Sprintf("#<virtual:%s>", name)
}

func NewVirtualAccount() VirtualAccount {
	return VirtualAccount{
		credits: map[string]*fundsStack{},
		debits:  map[string]*fundsStack{},
	}
}

func (vacc *VirtualAccount) getCredits(asset string) *fundsStack {
	return utils.MapGetOrPutDefault(vacc.credits, asset, func() *fundsStack {
		fs := newFundsStack(nil)
		return &fs
	})
}

func (vacc *VirtualAccount) getDebits(asset string) *fundsStack {
	return utils.MapGetOrPutDefault(vacc.debits, asset, func() *fundsStack {
		fs := newFundsStack(nil)
		return &fs
	})
}

// Send funds to virtual account and add them to the account's credits.
// When pulled, the account will return those funds (with a  FIFO policy).
//
// If the account has debts (with the same asset), we'll repay the debt first.
// In this case, the operation will emit the corresponding postings (if any).
func (vacc *VirtualAccount) Receive(asset string, sender Sender) []Posting {
	// when receiving funds, we need to use them to clear debts first (if any)
	debits := vacc.getDebits(asset)

	postings, remainingAmount := repayWithSender(debits, asset, sender)

	credits := vacc.getCredits(asset)

	sender.Amount = remainingAmount
	credits.Push(sender)

	return postings
}

func send(
	source AccountValue,
	destination AccountValue,
	amount *big.Int,
	asset string,
	color string,
) []Posting {
	switch source := source.(type) {
	case AccountAddress:

		switch destination := destination.(type) {
		case AccountAddress:
			return []Posting{{
				Source:      string(source),
				Destination: string(destination),
				Amount:      amount,
				Asset:       coloredAsset(asset, &color),
			}}
		case VirtualAccount:
			panic("TODO2")
		}

	case VirtualAccount:

		switch dest := destination.(type) {
		case AccountAddress:
			return source.Pull(asset, Sender{
				Account: dest,
				Amount:  amount,
				Color:   color,
			})

		case VirtualAccount:
			panic("TODO4")
		}

	}

	panic("non exhaustive match")
}

// Treat this stack as debts and use the sender to repay debt.
// Return the emitted postings and the remaining amount
func repayWithSender(s *fundsStack, asset string, credit Sender) ([]Posting, *big.Int) {
	remainingAmt := new(big.Int).Set(credit.Amount)

	var postings []Posting

	// Take away the debt that the credit allows for
	pulled := s.PullColored(credit.Amount, credit.Color)
	for _, pulledSender := range pulled {
		newPostings := send(
			credit.Account,
			pulledSender.Account,
			pulledSender.Amount,
			asset,
			credit.Color,
		)
		postings = append(postings, newPostings...)
	}

	for _, p := range postings {
		remainingAmt.Sub(remainingAmt, p.Amount)
	}

	return postings, remainingAmt

}

// Pull all the *immediately* available credits
func (vacc *VirtualAccount) PullCredits(asset string) []Sender {
	return vacc.getCredits(asset).PullAll()
}

// Pull funds from the virtual account.
//
// If the overdraft is bounded (overdraft==0 is no overdraft), it may be possible that we don't pull enough.
// In that case, the operation will still succeed, but the sum of sent amount will be lower than the requested amount.
//
// If the overdraft is higher than 0 or unbounded, is possible that the pulled amount is higher than the virtual account's credits.
// In this case, we'll add the pulled amount to the virtual account's debts.
func (vacc *VirtualAccount) Pull(asset string, receiver Sender) []Posting {
	credits := vacc.getCredits(asset)
	pulled := credits.PullColored(receiver.Amount, receiver.Color)

	remainingAmt := new(big.Int).Set(receiver.Amount)

	var postings []Posting

	for _, pulledSender := range pulled {
		newPostings := send(
			pulledSender.Account,
			receiver.Account,
			pulledSender.Amount,
			asset,
			receiver.Color,
		)
		postings = append(postings, newPostings...)
	}

	// TODO it looks like we aren't using overdraft now. How's that possible?
	if remainingAmt.Cmp(big.NewInt(0)) == 1 {
		// If we didn't pull enough and we're allowed to overdraft,
		// push the amount to debts WITHOUT emitting the corresponding postings (yet)
		debits := vacc.getDebits(asset)
		debits.Push(Sender{
			Account: receiver.Account,
			Color:   receiver.Color,
			Amount:  remainingAmt,
		})
	}

	return postings
}
