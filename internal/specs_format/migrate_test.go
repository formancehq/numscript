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

	// re-marshaling is still idempotent, even though nothing was reported changed
	var specs specs_format.Specs
	require.NoError(t, json.Unmarshal(out, &specs))
	require.Equal(t, specs_format.SchemaURL, specs.Schema)
}

func TestMigrateSpecsContentMissingSchema(t *testing.T) {
	raw := []byte(`{
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
	require.True(t, changed)

	var specs specs_format.Specs
	require.NoError(t, json.Unmarshal(out, &specs))
	require.Equal(t, specs_format.SchemaURL, specs.Schema)
	require.Equal(t, "USD/2 100", specs.TestCases[0].ExpectAccountsMeta[0].Value)
}

func TestMigrateSpecsContentStaleSchema(t *testing.T) {
	raw := []byte(`{
  "$schema": "https://raw.githubusercontent.com/formancehq/numscript/main/specs.schema.json",
  "testCases": [{ "it": "d1" }]
}
`)

	out, changed, err := specs_format.MigrateSpecsContent(raw)
	require.NoError(t, err)
	require.True(t, changed)

	var specs specs_format.Specs
	require.NoError(t, json.Unmarshal(out, &specs))
	require.Equal(t, specs_format.SchemaURL, specs.Schema)
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
