package interpreter

import (
	"github.com/formancehq/numscript/internal/utils"
)

func (m AccountsMetadata) fetchAccountMetadata(account string) AccountMetadata {
	return defaultMapGet(m, account, func() AccountMetadata {
		return AccountMetadata{}
	})
}

func (m AccountsMetadata) DeepClone() AccountsMetadata {
	cloned := make(AccountsMetadata)
	for account, accountBalances := range m {
		for asset, metadataValue := range accountBalances {
			clonedAccountBalances := cloned.fetchAccountMetadata(account)
			defaultMapGet(clonedAccountBalances, asset, func() string {
				return metadataValue
			})
		}
	}
	return cloned
}

func (m AccountsMetadata) Merge(update AccountsMetadata) {
	for acc, accBalances := range update {
		cachedAcc := defaultMapGet(m, acc, func() AccountMetadata {
			return AccountMetadata{}
		})

		for curr, amt := range accBalances {
			cachedAcc[curr] = amt
		}
	}
}

func (m AccountsMetadata) PrettyPrint() string {
	header := []string{"Account", "Name", "Value"}

	var rows [][]string
	for account, accMetadata := range m {
		for name, value := range accMetadata {
			row := []string{account, name, value}
			rows = append(rows, row)
		}
	}

	return utils.CsvPretty(header, rows, true)
}
