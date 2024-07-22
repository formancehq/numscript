package interpreter

import "math/big"

type AccountWithBalances struct {
	Balances map[string]*big.Int
}

type StaticStore map[string]*AccountWithBalances
