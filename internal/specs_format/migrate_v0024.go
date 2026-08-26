package specs_format

import (
	"encoding/json"
	"math/big"
	"sort"

	"github.com/formancehq/numscript/internal/interpreter"
)

// The types below mirror the specs format as it shipped in v0.0.24 (commit
// 1b42b98), the last tagged release before balances and metadata became row
// arrays instead of nested maps (see interpreter.Balances / AccountsMetadata
// history). A real, released specs file can be in this shape, unlike the
// short-lived tagged-metadata-value format from fadd1f8, which was never
// released and isn't supported here.
//
// Known gap: v0.0.24 encoded an asset's color into its name (e.g. "USD_RED"),
// rather than as today's separate Color field. This conversion carries that
// name through as-is rather than trying to decode it, since the two can't be
// told apart from the string alone (a real asset might just contain an
// underscore). A file that relied on colored assets needs a manual fix after
// migrating.
type legacyV0024Balances map[string]map[string]*big.Int

type legacyV0024AccountsMetadata map[string]map[string]string

type legacyV0024Specs struct {
	Schema       string                      `json:"$schema,omitempty"`
	FeatureFlags []string                    `json:"featureFlags,omitempty"`
	Balances     legacyV0024Balances         `json:"balances,omitempty"`
	Vars         interpreter.VariablesMap    `json:"variables,omitempty"`
	Meta         legacyV0024AccountsMetadata `json:"metadata,omitempty"`
	TestCases    []legacyV0024TestCase       `json:"testCases,omitempty"`
}

type legacyV0024TestCase struct {
	It string `json:"it"`

	Balances legacyV0024Balances         `json:"balances,omitempty"`
	Vars     interpreter.VariablesMap    `json:"variables,omitempty"`
	Meta     legacyV0024AccountsMetadata `json:"metadata,omitempty"`

	Focus bool `json:"focus,omitempty"`
	Skip  bool `json:"skip,omitempty"`

	ExpectMissingFunds   bool `json:"expect.error.missingFunds,omitempty"`
	ExpectNegativeAmount bool `json:"expect.error.negativeAmount,omitempty"`

	// These three didn't change shape between v0.0.24 and today (new fields
	// were all added as omitempty), so they unmarshal directly into the
	// current types.
	ExpectPostings  []interpreter.Posting `json:"expect.postings,omitempty"`
	ExpectMovements Movements             `json:"expect.movements,omitempty"`

	ExpectTxMeta             map[string]string           `json:"expect.txMetadata,omitempty"`
	ExpectAccountsMeta       legacyV0024AccountsMetadata `json:"expect.metadata,omitempty"`
	ExpectEndBalances        legacyV0024Balances         `json:"expect.endBalances,omitempty"`
	ExpectEndBalancesInclude legacyV0024Balances         `json:"expect.endBalances.include,omitempty"`
}

// parseLegacyV0024Specs reports whether raw parses as a v0.0.24-shaped specs
// file. Only called after the current Specs shape has already failed to
// parse, so this is the fallback path, not the common case.
func parseLegacyV0024Specs(raw []byte) (legacyV0024Specs, bool) {
	var legacy legacyV0024Specs
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return legacyV0024Specs{}, false
	}
	return legacy, true
}

func upgradeLegacyV0024Specs(legacy legacyV0024Specs) Specs {
	testCases := make([]TestCase, 0, len(legacy.TestCases))
	for _, tc := range legacy.TestCases {
		testCases = append(testCases, TestCase{
			It:                       tc.It,
			Balances:                 upgradeLegacyBalances(tc.Balances),
			Vars:                     tc.Vars,
			Meta:                     upgradeLegacyAccountsMetadata(tc.Meta),
			Focus:                    tc.Focus,
			Skip:                     tc.Skip,
			ExpectMissingFunds:       tc.ExpectMissingFunds,
			ExpectNegativeAmount:     tc.ExpectNegativeAmount,
			ExpectPostings:           tc.ExpectPostings,
			ExpectTxMeta:             upgradeLegacyTxMeta(tc.ExpectTxMeta),
			ExpectAccountsMeta:       upgradeLegacySetAccountsMetadata(tc.ExpectAccountsMeta),
			ExpectEndBalances:        upgradeLegacyBalances(tc.ExpectEndBalances),
			ExpectEndBalancesInclude: upgradeLegacyBalances(tc.ExpectEndBalancesInclude),
			ExpectMovements:          tc.ExpectMovements,
		})
	}

	return Specs{
		Schema:       legacy.Schema,
		FeatureFlags: legacy.FeatureFlags,
		Balances:     upgradeLegacyBalances(legacy.Balances),
		Vars:         legacy.Vars,
		Meta:         upgradeLegacyAccountsMetadata(legacy.Meta),
		TestCases:    testCases,
	}
}

// upgradeLegacyBalances flattens {account: {asset: amount}} into sorted rows,
// so the migrated file's diff is deterministic across runs.
func upgradeLegacyBalances(legacy legacyV0024Balances) interpreter.Balances {
	if legacy == nil {
		return nil
	}
	rows := make(interpreter.Balances, 0, len(legacy))
	for account, byAsset := range legacy {
		for asset, amount := range byAsset {
			rows = append(rows, interpreter.BalanceRow{
				Account: account,
				Asset:   asset,
				Amount:  amount,
			})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Account != rows[j].Account {
			return rows[i].Account < rows[j].Account
		}
		return rows[i].Asset < rows[j].Asset
	})
	return rows
}

// upgradeLegacyAccountsMetadata flattens {account: {key: value}} into sorted
// rows, the same way upgradeLegacyBalances does.
func upgradeLegacyAccountsMetadata(legacy legacyV0024AccountsMetadata) interpreter.AccountsMetadata {
	if legacy == nil {
		return nil
	}
	rows := make(interpreter.AccountsMetadata, 0, len(legacy))
	for account, byKey := range legacy {
		for key, value := range byKey {
			rows = append(rows, interpreter.AccountMetadataRow{
				Account: account,
				Key:     key,
				Value:   value,
			})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Account != rows[j].Account {
			return rows[i].Account < rows[j].Account
		}
		return rows[i].Key < rows[j].Key
	})
	return rows
}

// upgradeLegacySetAccountsMetadata is upgradeLegacyAccountsMetadata's
// counterpart for expect.metadata, whose row type carries an extra Scope
// field the legacy shape never had (scopes didn't exist in v0.0.24).
func upgradeLegacySetAccountsMetadata(legacy legacyV0024AccountsMetadata) interpreter.SetAccountsMetadata {
	if legacy == nil {
		return nil
	}
	rows := make(interpreter.SetAccountsMetadata, 0, len(legacy))
	for account, byKey := range legacy {
		for key, value := range byKey {
			rows = append(rows, interpreter.SetAccountMetadataRow{
				Account: account,
				Key:     key,
				Value:   value,
			})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Account != rows[j].Account {
			return rows[i].Account < rows[j].Account
		}
		return rows[i].Key < rows[j].Key
	})
	return rows
}

// upgradeLegacyTxMeta flattens {key: value} into sorted rows.
func upgradeLegacyTxMeta(legacy map[string]string) ExpectedTxMeta {
	if legacy == nil {
		return nil
	}
	rows := make(ExpectedTxMeta, 0, len(legacy))
	for key, value := range legacy {
		rows = append(rows, TxMetadataRow{Key: key, Value: value})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Key < rows[j].Key
	})
	return rows
}
