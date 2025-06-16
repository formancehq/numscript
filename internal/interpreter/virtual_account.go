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
	return defaultMapGet(vacc.credits, asset, func() *fundsStack {
		fs := newFundsStack(nil)
		return &fs
	})
}

func (vacc *VirtualAccount) getDebits(asset string) *fundsStack {
	return defaultMapGet(vacc.debits, asset, func() *fundsStack {
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
	// TODO check debits first
	// debits := vacc.getDebits(asset)

	debits := vacc.getDebits(asset)

	postings, sender := debits.RepayWithSender(asset, sender)

	credits := vacc.getCredits(asset)
	credits.Push(sender)

	return postings
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
func (vacc *VirtualAccount) Pull(asset string, overdraft *big.Int, receiver Sender) []Posting {
	if overdraft == nil {
		overdraft = new(big.Int).Set(receiver.Amount)
	}

	credits := vacc.getCredits(asset)
	pulled := credits.PullColored(receiver.Amount, receiver.Color)

	remainingAmt := new(big.Int).Set(receiver.Amount)
	var postings []Posting
	for _, pulledSender := range pulled {
		switch pulledSenderAccount := pulledSender.Account.(type) {
		case VirtualAccount:
			recPostings := pulledSenderAccount.Pull(asset, overdraft, receiver)
			postings = append(postings, recPostings...)
			continue

		case AccountAddress:
			remainingAmt.Sub(remainingAmt, pulledSender.Amount)
			switch receiverAccount := receiver.Account.(type) {
			case AccountAddress:
				postings = append(postings, Posting{
					Source:      string(pulledSenderAccount),
					Destination: string(receiverAccount),
					Amount:      pulledSender.Amount,
					Asset:       coloredAsset(asset, &receiver.Color),
				})

			case VirtualAccount:
				// receiverAccount.Receive()
				panic("TODO handle virtual account in Pull()")
			}
		}

	}

	allowedDebt := utils.MinBigInt(remainingAmt, overdraft)
	if allowedDebt.Cmp(big.NewInt(0)) == 1 {
		// If we didn't pull enough and we're allowed to overdraft,
		// push the amount to debts WITHOUT emitting the corresponding postings (yet)
		debits := vacc.getDebits(asset)
		debits.Push(Sender{
			Account: receiver.Account,
			Color:   receiver.Color,
			Amount:  allowedDebt,
		})
	}

	return postings
}
