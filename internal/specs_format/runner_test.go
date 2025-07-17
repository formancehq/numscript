package specs_format_test

import (
	"bytes"
	"testing"

	"github.com/formancehq/numscript/internal/specs_format"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestShowDiff(t *testing.T) {
	var buf bytes.Buffer
	specs_format.ShowDiff(
		&buf,
		map[string]any{
			"common": "ok",
			"x":      42,
		},
		map[string]any{
			"common": "ok",
			"x":      100,
		},
	)
	snaps.MatchSnapshot(t, buf.String())
}

func TestSingleTest(t *testing.T) {
	var out bytes.Buffer

	script := `
		send [USD/2 100] (
			source = @world
			destination = @dest
		)
	`

	specs := `
		{
			"testCases": [
				{
					"it": "tfailing",
					"expect.postings": [{
						"source": "wrong-source",
						"destination": "dest",
						"asset": "USD/2",
						"amount": 100
					}]
				},
				{
					"it": "tpassing",
					"expect.postings": [{
						"source": "world",
						"destination": "dest",
						"asset": "USD/2",
						"amount": 100
					}]
				}
			] 
		}
	`

	success := specs_format.RunSpecs(&out, &out, []specs_format.RawSpec{
		{
			NumscriptPath:    "example.num",
			SpecsPath:        "example.num.specs.json",
			NumscriptContent: script,
			SpecsFileContent: []byte(specs),
		},
	})

	require.False(t, success)

	snaps.MatchSnapshot(t, out.String())
}

func TestComplexAssertions(t *testing.T) {
	var out bytes.Buffer

	script := `
		send [USD/2 100] (
			source = @alice
			destination = @dest
		)
	`

	specs := `
		{
			"testCases": [
				{
					"it": "send when there are enough funds",
					"balances": {
						"alice": { "USD/2": 9999 }
					},
					"expect.volumes": {
							"alice": { "USD/2": -100 },
							"dest": { "USD/2": 1 }
					},
					"expect.movements": {
						"alice": {
							"dest": { "EUR": 100 }
						}
					},
					"expect.missingFunds": true
				},
				{
					"it": "tpassing",
					"balances": {
						"alice": { "USD/2": 0 }
					},
					"expect.missingFunds": true
				}
			] 
		}
	`

	success := specs_format.RunSpecs(&out, &out, []specs_format.RawSpec{
		{
			NumscriptPath:    "example.num",
			SpecsPath:        "example.num.specs.json",
			NumscriptContent: script,
			SpecsFileContent: []byte(specs),
		},
	})

	require.False(t, success)

	snaps.MatchSnapshot(t, out.String())
}

func TestNoFilesErr(t *testing.T) {
	var out bytes.Buffer
	success := specs_format.RunSpecs(&out, &out, []specs_format.RawSpec{})
	require.False(t, success)
	snaps.MatchSnapshot(t, out.String())
}

func TestParseErrSpecs(t *testing.T) {
	var out bytes.Buffer

	success := specs_format.RunSpecs(&out, &out, []specs_format.RawSpec{
		{
			NumscriptPath:    "example.num",
			SpecsPath:        "example.num.specs.json",
			NumscriptContent: "",
			SpecsFileContent: []byte(`
		not a json
	`),
		},
	})
	require.False(t, success)
	snaps.MatchSnapshot(t, out.String())
}

func TestSchemaErrSpecs(t *testing.T) {
	var out bytes.Buffer

	success := specs_format.RunSpecs(&out, &out, []specs_format.RawSpec{
		{
			NumscriptPath:    "example.num",
			SpecsPath:        "example.num.specs.json",
			NumscriptContent: "",
			SpecsFileContent: []byte(`
		{ "balances": 42 }
	`),
		},
	})
	require.False(t, success)
	snaps.MatchSnapshot(t, out.String())
}

func TestNumscriptParseErr(t *testing.T) {
	var out bytes.Buffer

	success := specs_format.RunSpecs(&out, &out, []specs_format.RawSpec{
		{
			NumscriptPath:    "example.num",
			SpecsPath:        "example.num.specs.json",
			NumscriptContent: "!err",
			SpecsFileContent: []byte(`
		{ }
	`),
		},
	})
	require.False(t, success)
	snaps.MatchSnapshot(t, out.String())
}

func TestRuntimeErr(t *testing.T) {
	var out bytes.Buffer

	specs := `
		{
			"testCases": [
				{
					"it": "runs",
					"expect.missingFunds": false
				}
			] 
		}
	`

	success := specs_format.RunSpecs(&out, &out, []specs_format.RawSpec{
		{
			NumscriptPath:    "example.num",
			SpecsPath:        "example.num.specs.json",
			NumscriptContent: `send [USD/2 100] ( source = "ops!" destination = @world)`,
			SpecsFileContent: []byte(specs),
		},
	})
	require.False(t, success)
	snaps.MatchSnapshot(t, out.String())
}
