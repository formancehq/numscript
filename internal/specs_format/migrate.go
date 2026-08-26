package specs_format

import (
	"encoding/json"
	"strings"
)

// MigrateSpecsContent rewrites a specs file's raw JSON to the current format:
// a missing or non-canonical $schema is pointed at SchemaURL. It reports
// whether anything changed, so a caller can gate writing the result behind a
// flag while still surfacing that a file is stale.
//
// Staleness is detected structurally (just the $schema field), not by
// comparing byte-for-byte against a canonical re-marshal — otherwise a file
// that is already current, just formatted differently (e.g. one row per
// line), would be falsely flagged. Once a file IS stale, though, the rewrite
// re-marshals the whole struct, so it also canonicalizes formatting to match
// what the rest of the CLI generates (e.g. test-init) — this is not a
// minimal-diff patch.
func MigrateSpecsContent(raw []byte) (out []byte, changed bool, err error) {
	var specs Specs
	if err := json.Unmarshal(raw, &specs); err != nil {
		return nil, false, err
	}

	if strings.HasSuffix(specs.Schema, "/v1.specs.schema.json") {
		return raw, false, nil
	}

	specs.Schema = SchemaURL

	marshaled, err := json.MarshalIndent(specs, "", "  ")
	if err != nil {
		return nil, false, err
	}
	marshaled = append(marshaled, '\n')

	return marshaled, true, nil
}
