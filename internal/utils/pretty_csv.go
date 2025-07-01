package utils

import (
	"fmt"
	"strings"

	"github.com/formancehq/numscript/internal/ansi"
)

// Fails if the header is shorter than any of the rows
func CsvPretty(
	header []string,
	rows [][]string,
) string {
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
