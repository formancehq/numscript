package specs_format_test

import (
	"encoding/json"
	"testing"

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
