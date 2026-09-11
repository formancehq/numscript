package specs_format

import (
	"encoding/json"
)

// MigrateSpecsContent rewrites a specs file's raw JSON to the current format.
// Two things can be outdated, independently of each other:
//
//   - the file's structure: v0.0.24's nested-map balances/metadata are
//     flattened to today's row arrays (see migrate_v0024.go);
//   - its asset names: v0.0.24's ASSET_COLOR encoding is split back into the
//     separate asset/color fields (see decodeLegacyColors).
//
// The second can be stale on its own, since the encoding hides inside a string
// field whose surrounding structure never changed — so it is checked on every
// file, not only on ones that failed to parse as the current shape.
//
// It reports whether anything changed, so a caller can gate writing the result
// behind a flag while still surfacing that a file is stale.
//
// Staleness is decided by content alone: $schema is an editor hint, not part of
// the format, and nothing in the CLI reads it to parse, validate or dispatch. A
// migrated file does get its $schema pointed at SchemaURL, since the file
// really did change.
//
// Once a file IS stale, the rewrite re-marshals the whole struct, so it also
// canonicalizes formatting to match what the rest of the CLI generates (e.g.
// test-init) — this is not a minimal-diff patch.
func MigrateSpecsContent(raw []byte) (out []byte, changed bool, err error) {
	specs, structureChanged, err := parseSpecsForMigration(raw)
	if err != nil {
		return nil, false, err
	}

	colorsDecoded := decodeLegacyColors(&specs)

	if !structureChanged && !colorsDecoded {
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
// whether the fallback was used, i.e. whether the file's structure needed
// upgrading.
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
