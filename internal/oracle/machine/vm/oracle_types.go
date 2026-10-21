package vm

// This file is NOT part of the verbatim vendored code from
// github.com/formancehq/ledger/internal/machine. It replaces the small
// set of symbols internal/machine originally imported from
// github.com/formancehq/ledger/internal (package `ledger`) — that package
// is Go-internal to a different module and cannot be imported here.
//
// Zero and Posting/Postings are copied verbatim in substance (renamed to
// ResultPosting/ResultPostings, since this package already declares its
// own unrelated local `Posting` type for internal execution bookkeeping
// in machine.go — the original code avoided this collision via the
// `ledger.` qualifier, which isn't available once vendored in-package)
// from:
//   - github.com/formancehq/ledger/internal/bigint.go
//   - github.com/formancehq/ledger/internal/posting.go
//
// Account is NOT verbatim: the original (internal/account.go) embeds
// bun.BaseModel and ORM tags for Postgres persistence. Only .Address and
// .Metadata are ever read from it anywhere in internal/machine, so this
// is trimmed to just those two fields.

import (
	"errors"
	"math/big"

	"github.com/formancehq/go-libs/v5/pkg/types/metadata"
	"github.com/formancehq/numscript/internal/oracle/machine/internal/accounts"
	"github.com/formancehq/numscript/internal/oracle/machine/internal/assets"
)

var Zero = big.NewInt(0)

type ResultPosting struct {
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Amount      *big.Int `json:"amount"`
	Asset       string   `json:"asset"`
}

func (p ResultPosting) GetSource() string {
	return p.Source
}

func (p ResultPosting) GetDestination() string {
	return p.Destination
}

func NewResultPosting(source string, destination string, asset string, amount *big.Int) ResultPosting {
	return ResultPosting{
		Source:      source,
		Destination: destination,
		Amount:      amount,
		Asset:       asset,
	}
}

type ResultPostings []ResultPosting

func (p ResultPostings) Reverse() ResultPostings {
	postings := make(ResultPostings, len(p))
	copy(postings, p)

	for i := range p {
		postings[i].Source, postings[i].Destination = postings[i].Destination, postings[i].Source
	}

	for i := 0; i < len(p)/2; i++ {
		postings[i], postings[len(postings)-i-1] = postings[len(postings)-i-1], postings[i]
	}

	return postings
}

func (p ResultPostings) Validate() (int, error) {
	for i, p := range p {
		if p.Amount == nil {
			return i, errors.New("no amount defined")
		}
		if p.Amount.Cmp(Zero) < 0 {
			return i, errors.New("negative amount")
		}
		if !accounts.ValidateAddress(p.Source) {
			return i, errors.New("invalid source address")
		}
		if !accounts.ValidateAddress(p.Destination) {
			return i, errors.New("invalid destination address")
		}
		if !assets.IsValid(p.Asset) {
			return i, errors.New("invalid asset")
		}
	}

	return 0, nil
}

type Account struct {
	Address  string            `json:"address"`
	Metadata metadata.Metadata `json:"metadata"`
}
