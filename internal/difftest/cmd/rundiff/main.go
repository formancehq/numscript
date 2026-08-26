// Command rundiff runs a batch of differential tests between this repo's
// interpreter and the vendored legacy ledger "machine" (internal/oracle),
// outside of `go test -fuzz`. Useful for a quick, large ad-hoc sweep, or for
// reproducing/inspecting a specific seed.
//
// Usage:
//
//	go run ./cmd/rundiff -n 100000
//	go run ./cmd/rundiff -seed 12345 -n 1 -verbose
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/formancehq/numscript/internal/difftest"
)

func main() {
	n := flag.Int("n", 1000, "number of random scripts to run")
	seed := flag.Int64("seed", 1, "base RNG seed; each iteration uses seed+i")
	verbose := flag.Bool("verbose", false, "print every generated script, not just mismatches")
	stopOnFirst := flag.Bool("stop-on-first", true, "stop at the first mismatch found")
	flag.Parse()

	ctx := context.Background()
	mismatches := 0

	for i := range *n {
		rngSeed := *seed + int64(i)
		rng := rand.New(rand.NewSource(rngSeed))

		c := difftest.RunOne(ctx, rng)

		if *verbose {
			fmt.Printf("--- seed %d ---\n%s\n\n", rngSeed, c.Script)
		}

		if c.Verdict.Mismatch {
			mismatches++
			fmt.Printf("MISMATCH at seed %d: %s\n\nvars: %v\n\nscript:\n%s\n\n", rngSeed, c.Verdict.Reason, c.Vars, c.Script)
			if *stopOnFirst {
				os.Exit(1)
			}
		}
	}

	fmt.Printf("ran %d scripts, %d mismatches\n", *n, mismatches)
	if mismatches > 0 {
		os.Exit(1)
	}
}
