#!/usr/bin/env bash
# Migrate .num.specs.json balance entries from the legacy nested shape
# ({"": N, "COLOR": M}) to the value-object shape introduced when `color`
# became a first-class Posting field.
#
# Output rules per (account, asset) entry:
#   • uncolored-only           → bare number          "USD/2": 100
#   • single colored entry     → value-object         "USD/2": { "color": "RED", "amount": 50 }
#   • mixed / multi-color      → array of value-objects, sorted by color
#
# Usage:
#   scripts/migrate-specs-to-color.sh path/to/file.num.specs.json [more files...]
#   scripts/migrate-specs-to-color.sh $(find . -name '*.num.specs.json')
set -euo pipefail

if [[ $# -eq 0 ]]; then
  echo "usage: $0 file.num.specs.json [...]" >&2
  exit 2
fi

read -r -d '' JQ_FILTER <<'JQ' || true
def to_value_object:
  if type != "object" then .
  elif (keys_unsorted | length) == 0 then .
  elif (keys_unsorted | length) == 1 and keys_unsorted[0] == "" then
    .[""]
  elif (keys_unsorted | length) == 1 then
    {color: keys_unsorted[0], amount: .[keys_unsorted[0]]}
  else
    [ (to_entries | sort_by(.key))[] |
      if .key == "" then {amount: .value}
      else {color: .key, amount: .value}
      end ]
  end;

def transform_balances:
  if type != "object" then .
  else
    with_entries(.value |= (
      if type != "object" then .
      else with_entries(.value |= to_value_object)
      end
    ))
  end;

(if has("balances") then .balances |= transform_balances else . end) |
(if has("testCases") then
  .testCases |= map(
    (if has("balances") then .balances |= transform_balances else . end) |
    (if has("expect.endBalances") then ."expect.endBalances" |= transform_balances else . end) |
    (if has("expect.endBalances.include") then ."expect.endBalances.include" |= transform_balances else . end)
  )
else . end)
JQ

for f in "$@"; do
  tmp="$(mktemp)"
  jq --indent 2 "$JQ_FILTER" "$f" > "$tmp"
  # Preserve trailing newline behaviour of jq output.
  mv "$tmp" "$f"
done
