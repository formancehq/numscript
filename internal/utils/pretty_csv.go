package utils

import (
	"fmt"
	"slices"
	"strings"

	"github.com/formancehq/numscript/internal/ansi"
)

// Rows shorter than the header are padded with empty cells;
// cells beyond the header length are ignored for padding purposes
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
			// rows shorter than the header are treated as having empty cells
			if fieldIndex < len(row) {
				maxLen = max(maxLen, len(row[fieldIndex]))
			}
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
		for index := range header {
			// missing cells in ragged rows are rendered as empty strings
			var fieldValue string
			if index < len(row) {
				fieldValue = row[index]
			}
			partialRow = append(partialRow, fmt.Sprintf("| %-*s ",
				maxLengths[index],
				fieldValue,
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
