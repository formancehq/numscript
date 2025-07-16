package cmd

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestShowDiff(t *testing.T) {
	var buf bytes.Buffer
	showDiff(
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

	success := runRawSpecs(&out, &out, []rawSpec{
		{
			NumscriptPath:    "exmaple.num",
			SpecsPath:        "example.num.specs.json",
			NumscriptContent: script,
			SpecsFileContent: []byte(specs),
		},
	})

	require.False(t, success)

	snaps.MatchSnapshot(t, out.String())
}
