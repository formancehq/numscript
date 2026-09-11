package specs_format

import (
	"encoding/json"
	"math/big"
	"regexp"
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
// v0.0.24 also encoded an asset's color into its name (e.g. "USD_COL/4" for
// asset USD/4 colored COL) rather than as today's separate Color field; see
// decodeLegacyColors below for how that is undone.
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

// legacyColoredAssetRegexp matches v0.0.24's ASSET_COLOR encoding, splitting it
// into the asset's base name, the color, and the optional precision suffix.
//
// The match is unambiguous: an asset name can't contain an underscore (the
// lexer's ASSET token doesn't allow one, and interpreter.checkAssetName
// rejects it), so an underscore here can only be the encoding's separator. The
// two halves are constrained to the same shapes the interpreter accepts today
// for an asset and a color respectively.
var legacyColoredAssetRegexp = regexp.MustCompile(`^([A-Z][A-Z0-9]{0,16})_([A-Z]{1,16})(\/\d{1,6})?$`)

// splitLegacyColoredAsset undoes the ASSET_COLOR encoding, reporting whether
// src was encoded at all. v0.0.24 inserted the color before the precision
// suffix, so "USD/4" colored "COL" became "USD_COL/4".
func splitLegacyColoredAsset(src string) (asset string, color string, ok bool) {
	m := legacyColoredAssetRegexp.FindStringSubmatch(src)
	if m == nil {
		return src, "", false
	}
	return m[1] + m[3], m[2], true
}

// decodeLegacyColors rewrites every ASSET_COLOR-encoded asset in specs into the
// separate asset/color fields colors became in cbc11e8, reporting whether
// anything changed.
//
// Unlike the rest of this file, this runs on every specs file rather than only
// on ones that failed to parse as the current shape. The encoding lives inside
// a string field whose surrounding structure never changed, so a v0.0.24 file
// that used colors but no nested balance/metadata maps parses cleanly as the
// current shape and would otherwise never be offered for decoding.
//
// A row that already carries a Color is left alone: it can't be a legacy row,
// and overwriting it would lose information.
func decodeLegacyColors(specs *Specs) bool {
	changed := decodeBalanceColors(specs.Balances)
	for i := range specs.TestCases {
		tc := &specs.TestCases[i]
		changed = decodeBalanceColors(tc.Balances) || changed
		changed = decodeBalanceColors(tc.ExpectEndBalances) || changed
		changed = decodeBalanceColors(tc.ExpectEndBalancesInclude) || changed
		changed = decodePostingColors(tc.ExpectPostings) || changed
		changed = decodeMovementColors(tc.ExpectMovements) || changed
	}
	return changed
}

func decodeBalanceColors(rows interpreter.Balances) bool {
	changed := false
	for i := range rows {
		if rows[i].Color != "" {
			continue
		}
		asset, color, ok := splitLegacyColoredAsset(rows[i].Asset)
		if !ok {
			continue
		}
		rows[i].Asset, rows[i].Color = asset, color
		changed = true
	}
	return changed
}

func decodePostingColors(postings []interpreter.Posting) bool {
	changed := false
	for i := range postings {
		if postings[i].Color != "" {
			continue
		}
		asset, color, ok := splitLegacyColoredAsset(postings[i].Asset)
		if !ok {
			continue
		}
		postings[i].Asset, postings[i].Color = asset, color
		changed = true
	}
	return changed
}

func decodeMovementColors(movements Movements) bool {
	changed := false
	for i := range movements {
		if movements[i].Color != "" {
			continue
		}
		asset, color, ok := splitLegacyColoredAsset(movements[i].Asset)
		if !ok {
			continue
		}
		movements[i].Asset, movements[i].Color = asset, color
		changed = true
	}
	return changed
}
