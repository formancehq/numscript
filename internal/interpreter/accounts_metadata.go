package interpreter

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
