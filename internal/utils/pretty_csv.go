package utils

import (
	"fmt"
	"slices"
	"strings"

	"github.com/formancehq/numscript/internal/ansi"
)

// Fails if the header is shorter than any of the rows
func CsvPretty(
	header []string,
	rows [][]string,
	sortRows bool,
) string {
	if sortRows {
		slices.SortStableFunc(rows, func(x, y []string) int {
			strX := strings.Join(x, "|")
			strY := strings.Join(y, "|")
			if strX == strY {
				return 0
			} else if strX < strY {
				return -1
			} else {
				return 1
			}
		})
	}

	// -- Find paddings
	var maxLengths = make([]int, len(header))
	for fieldIndex, fieldName := range header {
		maxLen := len(fieldName)

		for _, row := range rows {
			// panics if row[fieldIndex] is out of bounds
			// thus we must never have unlabeled cols
			maxLen = max(maxLen, len(row[fieldIndex]))
		}

		maxLengths[fieldIndex] = maxLen
	}

	var allRows []string

	// -- Print header
	{
		var partialRow []string
		for index, fieldName := range header {
			paddedHeader := fmt.Sprintf("%-*s",
				maxLengths[index],
				fieldName,
			)
			partialRow = append(partialRow, fmt.Sprintf("| %s ", ansi.ColorCyan(paddedHeader)))
		}
		partialRow = append(partialRow, "|")
		allRows = append(allRows, strings.Join(partialRow, ""))
	}

	// -- Print rows
	for _, row := range rows {
		var partialRow []string
		for index, fieldName := range row {
			partialRow = append(partialRow, fmt.Sprintf("| %-*s ",
				maxLengths[index],
				fieldName,
			))
		}
		partialRow = append(partialRow, "|")
		allRows = append(allRows, strings.Join(partialRow, ""))
	}

	return strings.Join(allRows, "\n")
}

// CsvPrettyOmitEmptyCols renders the table like CsvPretty, but omits any column
// whose data cells are all empty (its header is dropped along with it). This lets
// callers always pass the full set of columns and have the optional ones (e.g. a
// color or scope dimension that no row populates) disappear automatically.
//
// Fails if the header is shorter than any of the rows.
func CsvPrettyOmitEmptyCols(
	header []string,
	rows [][]string,
	sortRows bool,
) string {
	// with no rows there's nothing to judge column emptiness from, so keep every
	// column and still render the header
	if len(rows) == 0 {
		return CsvPretty(header, rows, sortRows)
	}

	keep := make([]bool, len(header))
	for col := range header {
		for _, row := range rows {
			if row[col] != "" {
				keep[col] = true
				break
			}
		}
	}

	filteredHeader := make([]string, 0, len(header))
	for col, name := range header {
		if keep[col] {
			filteredHeader = append(filteredHeader, name)
		}
	}

	filteredRows := make([][]string, len(rows))
	for i, row := range rows {
		filtered := make([]string, 0, len(filteredHeader))
		for col := range header {
			if keep[col] {
				filtered = append(filtered, row[col])
			}
		}
		filteredRows[i] = filtered
	}

	return CsvPretty(filteredHeader, filteredRows, sortRows)
}

func CsvPrettyMap(keyName string, valueName string, m map[string]string) string {
	var rows [][]string
	for k, v := range m {
		rows = append(rows, []string{k, v})
	}

	return CsvPretty([]string{keyName, valueName}, rows, true)
}
