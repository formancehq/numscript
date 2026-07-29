package runtime

import (
	"github.com/formancehq/numscript/internal/utils"
)

// MetaValueType is the numscript type a metadata value was stringified from.
// Metadata is stored as text, so without it a host can't tell the string "42"
// from the number 42 — the interpreter keeps those distinct, and the type is what
// lets the VM's output be mapped back onto the same typed contract.
type MetaValueType uint8

const (
	MetaValueStr MetaValueType = iota
	MetaValueAccount
	MetaValueAsset
	MetaValueInt
	MetaValuePortion
	MetaValueMonetary
)

func (t MetaValueType) String() string {
	switch t {
	case MetaValueStr:
		return "str"
	case MetaValueAccount:
		return "account"
	case MetaValueAsset:
		return "asset"
	case MetaValueInt:
		return "int"
	case MetaValuePortion:
		return "portion"
	case MetaValueMonetary:
		return "monetary"
	default:
		return "invalid"
	}
}

// MetaValue is one metadata entry written by a script: the stringified value plus
// the type it was stringified from. Str/Account/Asset hold the bare text (no
// quotes, no leading @); Int and Portion hold the big.Int/big.Rat form; Monetary
// holds "ASSET amount".
type MetaValue struct {
	Value string
	Typ   MetaValueType
}

type AccountMetadata = map[string]MetaValue
type AccountsMetadata map[string]AccountMetadata

func (m AccountsMetadata) fetchAccountMetadata(account string) AccountMetadata {
	return utils.MapGetOrPutDefault(m, account, func() AccountMetadata {
		return AccountMetadata{}
	})
}

func (m AccountsMetadata) DeepClone() AccountsMetadata {
	cloned := make(AccountsMetadata)
	for account, accountBalances := range m {
		for asset, metadataValue := range accountBalances {
			clonedAccountBalances := cloned.fetchAccountMetadata(account)
			utils.MapGetOrPutDefault(clonedAccountBalances, asset, func() MetaValue {
				return metadataValue
			})
		}
	}
	return cloned
}

func (m AccountsMetadata) Merge(update AccountsMetadata) {
	for acc, accBalances := range update {
		cachedAcc := utils.MapGetOrPutDefault(m, acc, func() AccountMetadata {
			return AccountMetadata{}
		})

		for curr, amt := range accBalances {
			cachedAcc[curr] = amt
		}
	}
}

func (m AccountsMetadata) PrettyPrint() string {
	header := []string{"Account", "Name", "Value", "Type"}

	var rows [][]string
	for account, accMetadata := range m {
		for name, value := range accMetadata {
			row := []string{account, name, value.Value, value.Typ.String()}
			rows = append(rows, row)
		}
	}

	return utils.CsvPretty(header, rows, true)
}
