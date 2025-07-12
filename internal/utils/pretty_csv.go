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
	var maxLengths []int = make([]int, len(header))
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
			partialRow = append(partialRow, fmt.Sprintf("| %-*s ",
				maxLengths[index],
				ansi.ColorCyan(fieldName),
			))
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

func CsvPrettyMap(keyName string, valueName string, m map[string]string) string {
	var rows [][]string
	for k, v := range m {
		rows = append(rows, []string{k, v})
	}

	return CsvPretty([]string{keyName, valueName}, rows, true)
}
