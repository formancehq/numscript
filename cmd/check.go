package cmd

import (
	"fmt"
	"numscript/analysis"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

func check(path string) {
	dat, err := os.ReadFile(path)
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		return
	}

	res := analysis.CheckSource(string(dat))
	sort.Slice(res.Diagnostics, func(i, j int) bool {
		p1 := res.Diagnostics[i].Range.Start
		p2 := res.Diagnostics[j].Range.Start

		return p2.GtEq(p1)
	})

	for i, d := range res.Diagnostics {
		if i != 0 {
			fmt.Print("\n\n")
		}
		errType := analysis.SeverityToAnsiString(d.Kind.Severity())
		fmt.Printf("%s:%d:%d - %s\n%s\n", path, d.Range.Start.Line, d.Range.Start.Character, errType, d.Kind.Message())
	}

	if len(res.Diagnostics) != 0 {
		fmt.Printf("\n\n")
	}

	errorsCount := res.GetErrorsCount()
	if errorsCount != 0 {

		var pluralizedError string
		if errorsCount == 1 {
			pluralizedError = "error"
		} else {
			pluralizedError = "errors"

		}

		fmt.Printf("\033[31mFound %d %s\033[0m\n", errorsCount, pluralizedError)
		os.Exit(1)
	}

	fmt.Printf("No errors found ✅\n")
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "check a numscript file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		check(path)
	},
}
