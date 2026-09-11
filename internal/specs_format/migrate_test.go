package specs_format_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/stretchr/testify/require"
)

func TestMigrateSpecsContentAlreadyCurrent(t *testing.T) {
	raw := []byte(`{
  "$schema": "` + specs_format.SchemaURL + `",
  "testCases": [
    {
      "it": "d1",
      "expect.metadata": [
        { "account": "acc", "key": "k", "value": "USD/2 100" }
      ]
    }
  ]
}
`)

	out, changed, err := specs_format.MigrateSpecsContent(raw)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, raw, out)
}

// $schema is an editor hint, not part of the format: a file whose structure is
// already current is left byte-for-byte alone, whatever its $schema says.
func TestMigrateSpecsContentLeavesSchemaAlone(t *testing.T) {
	for name, schemaLine := range map[string]string{
		"missing":      "",
		"stale":        `"$schema": "https://raw.githubusercontent.com/formancehq/numscript/main/specs.schema.json",`,
		"foreign host": `"$schema": "https://example.invalid/v1.specs.schema.json",`,
	} {
		t.Run(name, func(t *testing.T) {
			raw := []byte(`{
  ` + schemaLine + `
  "testCases": [
    {
      "it": "d1",
      "expect.metadata": [
        { "account": "acc", "key": "k", "value": "USD/2 100" }
      ]
    }
  ]
}
`)

			out, changed, err := specs_format.MigrateSpecsContent(raw)
			require.NoError(t, err)
			require.False(t, changed)
			require.Equal(t, raw, out)
		})
	}
}

func TestMigrateSpecsContentParseErr(t *testing.T) {
	_, _, err := specs_format.MigrateSpecsContent([]byte("not json"))
	require.Error(t, err)
}

// TestMigrateSpecsContentLegacyV0024Shape uses the exact specs shape v0.0.24
// (commit 1b42b98, the last tagged release) generated: balances and metadata
// as nested maps rather than today's row arrays.
func TestMigrateSpecsContentLegacyV0024Shape(t *testing.T) {
	raw := []byte(`{
  "testCases": [
    {
      "it": "-",
      "balances": {
        "sales:042": { "EUR/2": 2500 },
        "users:053": { "EUR/2": 500 }
      },
      "variables": { "sale": "sales:042" },
      "metadata": {
        "sales:042": { "seller": "users:053" },
        "users:053": { "commission": "12.5%" }
      },
      "expect.metadata": {
        "sales:042": { "paid": "true" }
      },
      "expect.txMetadata": { "note": "payout" },
      "expect.postings": [
        { "source": "sales:042", "destination": "users:053", "amount": 88, "asset": "EUR/2" }
      ]
    }
  ]
}
`)

	out, changed, err := specs_format.MigrateSpecsContent(raw)
	require.NoError(t, err)
	require.True(t, changed)

	var specs specs_format.Specs
	require.NoError(t, json.Unmarshal(out, &specs))

	// a structural migration does point $schema at the current schema: the
	// file's shape really did change.
	require.Equal(t, specs_format.SchemaURL, specs.Schema)

	tc := specs.TestCases[0]
	require.ElementsMatch(t, interpreter.Balances{
		{Account: "sales:042", Asset: "EUR/2", Amount: big.NewInt(2500)},
		{Account: "users:053", Asset: "EUR/2", Amount: big.NewInt(500)},
	}, tc.Balances)
	require.Equal(t, interpreter.VariablesMap{"sale": "sales:042"}, tc.Vars)
	require.ElementsMatch(t, interpreter.AccountsMetadata{
		{Account: "sales:042", Key: "seller", Value: "users:053"},
		{Account: "users:053", Key: "commission", Value: "12.5%"},
	}, tc.Meta)
	require.Equal(t, interpreter.SetAccountsMetadata{
		{Account: "sales:042", Key: "paid", Value: "true"},
	}, tc.ExpectAccountsMeta)
	require.Equal(t, specs_format.ExpectedTxMeta{
		{Key: "note", Value: "payout"},
	}, tc.ExpectTxMeta)
	require.Len(t, tc.ExpectPostings, 1)
}

func TestMigrateSpecsContentUnsupportedShapeStillErrors(t *testing.T) {
	// The fadd1f8 tagged-metadata-value format: never released, and distinct
	// from v0.0.24's shape (whose leaf values are plain strings), so it
	// should still fail rather than silently "succeed" as a legacy parse.
	raw := []byte(`{
  "testCases": [
    {
      "it": "d1",
      "expect.metadata": [
        { "account": "acc", "key": "k", "value": {"type": "monetary", "asset": "USD/2", "amount": "100"} }
      ]
    }
  ]
}
`)

	_, _, err := specs_format.MigrateSpecsContent(raw)
	require.Error(t, err)
}

// A v0.0.24 file that used colors but no nested balance/metadata maps parses
// cleanly as the current shape, so the encoded assets are the only thing left
// to migrate. This is the case the structural check alone can't catch.
func TestMigrateSpecsContentDecodesColorsInCurrentShape(t *testing.T) {
	raw := []byte(`{
  "featureFlags": ["experimental-asset-colors"],
  "balances": [
    { "account": "acc", "asset": "COIN_RED", "amount": 1 }
  ],
  "testCases": [
    {
      "it": "-",
      "expect.postings": [
        { "source": "src", "destination": "dest", "amount": 10, "asset": "USD_COL/4" }
      ],
      "expect.movements": [
        { "source": "src", "destination": "dest", "amount": 10, "asset": "USD_COL/4" }
      ],
      "expect.endBalances": [
        { "account": "dest", "asset": "USD_COL/4", "amount": 10 }
      ]
    }
  ]
}
`)

	out, changed, err := specs_format.MigrateSpecsContent(raw)
	require.NoError(t, err)
	require.True(t, changed)

	var specs specs_format.Specs
	require.NoError(t, json.Unmarshal(out, &specs))

	require.Equal(t, interpreter.Balances{
		{Account: "acc", Asset: "COIN", Color: "RED", Amount: big.NewInt(1)},
	}, specs.Balances)

	tc := specs.TestCases[0]
	require.Equal(t, "USD/4", tc.ExpectPostings[0].Asset)
	require.Equal(t, "COL", tc.ExpectPostings[0].Color)
	require.Equal(t, "USD/4", tc.ExpectMovements[0].Asset)
	require.Equal(t, "COL", tc.ExpectMovements[0].Color)
	require.Equal(t, "USD/4", tc.ExpectEndBalances[0].Asset)
	require.Equal(t, "COL", tc.ExpectEndBalances[0].Color)
}

// Both migrations at once: the nested-map structure AND encoded assets inside it.
func TestMigrateSpecsContentDecodesColorsInLegacyShape(t *testing.T) {
	raw := []byte(`{
  "testCases": [
    {
      "it": "-",
      "balances": { "acc": { "COIN_RED": 1, "COIN": 100 } }
    }
  ]
}
`)

	out, changed, err := specs_format.MigrateSpecsContent(raw)
	require.NoError(t, err)
	require.True(t, changed)

	var specs specs_format.Specs
	require.NoError(t, json.Unmarshal(out, &specs))

	require.ElementsMatch(t, interpreter.Balances{
		{Account: "acc", Asset: "COIN", Amount: big.NewInt(100)},
		{Account: "acc", Asset: "COIN", Color: "RED", Amount: big.NewInt(1)},
	}, specs.TestCases[0].Balances)
}

// A row that already carries a color can't be a legacy row, so it is left
// exactly as-is — including its asset, however odd it looks.
func TestMigrateSpecsContentLeavesExplicitColorAlone(t *testing.T) {
	raw := []byte(`{
  "balances": [
    { "account": "acc", "asset": "COIN_RED", "color": "BLUE", "amount": 1 }
  ]
}
`)

	out, changed, err := specs_format.MigrateSpecsContent(raw)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, raw, out)
}

// Assets that merely resemble the encoding must not be split: the separator is
// a single underscore between an asset name and a [A-Z]{1,16} color.
func TestMigrateSpecsContentLeavesNonEncodedAssetsAlone(t *testing.T) {
	for _, asset := range []string{
		"USD/2",     // no underscore at all
		"COIN",      // ditto
		"USD_",      // empty color
		"USD_red/2", // colors are upper-case
		"USD_A_B",   // two separators
		"_RED",      // empty asset name
	} {
		t.Run(asset, func(t *testing.T) {
			raw := []byte(`{
  "balances": [
    { "account": "acc", "asset": "` + asset + `", "amount": 1 }
  ]
}
`)

			out, changed, err := specs_format.MigrateSpecsContent(raw)
			require.NoError(t, err)
			require.False(t, changed)
			require.Equal(t, raw, out)
		})
	}
}
