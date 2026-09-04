package specs_format

import (
	"encoding/json"
	"strings"
)

// MigrateSpecsContent rewrites a specs file's raw JSON to the current format:
// v0.0.24's nested-map balances/metadata are flattened to today's row arrays
// (see migrate_v0024.go), and a missing or non-canonical $schema is pointed
// at SchemaURL. It reports whether anything changed, so a caller can gate
// writing the result behind a flag while still surfacing that a file is
// stale.
//
// $schema staleness is detected structurally, not by comparing byte-for-byte
// against a canonical re-marshal — otherwise a file that is already current,
// just formatted differently (e.g. one row per line), would be falsely
// flagged. Once a file IS stale, though, the rewrite re-marshals the whole
// struct, so it also canonicalizes formatting to match what the rest of the
// CLI generates (e.g. test-init) — this is not a minimal-diff patch.
func MigrateSpecsContent(raw []byte) (out []byte, changed bool, err error) {
	specs, structureChanged, err := parseSpecsForMigration(raw)
	if err != nil {
		return nil, false, err
	}

	schemaStale := !strings.HasSuffix(specs.Schema, "/v1.specs.schema.json")
	if !schemaStale && !structureChanged {
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

// parseSpecsForMigration parses raw as the current Specs shape, falling back
// to the v0.0.24 shape (see migrate_v0024.go) if that fails. It reports
// whether the fallback was used, i.e. whether the file's structure (not just
// its $schema) needed upgrading.
func parseSpecsForMigration(raw []byte) (specs Specs, structureChanged bool, err error) {
	if err := json.Unmarshal(raw, &specs); err == nil {
		return specs, false, nil
	}

	legacy, ok := parseLegacyV0024Specs(raw)
	if !ok {
		// Re-parse to surface the current shape's error: it applies to more
		// specs files (v0.0.24 was never the only released shape a file could
		// predate) and is what the rest of the CLI's error output expects.
		return Specs{}, false, json.Unmarshal(raw, &specs)
	}

	return upgradeLegacyV0024Specs(legacy), true, nil
}
