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
	debits := vacc.getDebits(asset)

	postings, sender := repayWithSender(debits, asset, sender)

	credits := vacc.getCredits(asset)
	credits.Push(sender)

	return postings
}

// Treat this stack as debts and use the sender to repay debt.
// Return the sender updated with the left amt (and the emitted postings)
func repayWithSender(s *fundsStack, asset string, credit Sender) ([]Posting, Sender) {
	// clone the amount so that we can modify it
	credit.Amount = new(big.Int).Set(credit.Amount)

	var postings []Posting

	// Take away the debt that the credit allows for
	clearedDebt := s.PullColored(credit.Amount, credit.Color)
	for _, receiver := range clearedDebt {
		switch creditAccount := credit.Account.(type) {
		case VirtualAccount:
			pulled := creditAccount.Pull(asset, nil, receiver)
			postings = append(postings, pulled...)

		case AccountAddress:
			// TODO do we need this in the other case?
			credit.Amount.Sub(credit.Amount, receiver.Amount)

			switch receiverAccount := receiver.Account.(type) {
			case AccountAddress:
				postings = append(postings, Posting{
					Source:      string(creditAccount),
					Destination: string(receiverAccount),
					Amount:      receiver.Amount,
					Asset:       coloredAsset(asset, &credit.Color),
				})

			case VirtualAccount:
				panic("TODO repay vacc")
			}
		}

	}

	return postings, credit

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
				// TODO either include in coverage or simply this
				panic("UNRECHED")

				return receiverAccount.Receive(asset, Sender{
					vacc,
					pulledSender.Amount,
					receiver.Color,
				})
			}
		}
	}

	// TODO it looks like we aren't using overdraft now. How's that possible?
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
