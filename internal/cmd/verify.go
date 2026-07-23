package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/formancehq/numscript/internal/verify"

	"github.com/spf13/cobra"
)

func getVerifyCmd() *cobra.Command {
	var timeoutMs int
	var z3Path string
	var showModel bool
	var showSMT bool

	cmd := &cobra.Command{
		Use:   "verify <path> <query>",
		Short: "Formally verify a property of a numscript file using an SMT solver (z3)",
		Long: `Verify encodes a numscript script's execution semantics as SMT-LIB2 and
uses a local z3 binary to prove or refute a property, or to find a witness.

The query is a small boolean DSL, optionally prefixed with "prove:" or "find:"
(default: prove). Predicates: start_balance(acct, asset), end_balance(...),
sent(...), received(...), volumes(...), and fail. Example:

  numscript verify script.num 'prove: !fail => received("dest", "USD/2") == 10'
  numscript verify script.num 'find: fail'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, query := args[0], args[1]
			dat, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}

			res, err := verify.Verify(context.Background(), string(dat), query, verify.Options{
				TimeoutMs: timeoutMs,
				Z3Path:    z3Path,
			})
			if err != nil {
				return err
			}

			if showSMT {
				fmt.Printf("--- SMT sent to z3 ---\n%s----------------------\n\n", res.SMT)
			}

			fmt.Printf("[%s] %s\n", res.Mode, res.Verdict)
			if res.Model != "" && showModel {
				fmt.Printf("\nModel:\n%s\n", res.Model)
			}
			if res.Verdict == verify.Counterexample || res.Verdict == verify.Impossible || res.Verdict == verify.Unknown {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&timeoutMs, "timeout", 0, "z3 timeout in milliseconds (0 = default)")
	cmd.Flags().StringVar(&z3Path, "z3", "", "path to the z3 binary (default: look up 'z3' on PATH)")
	cmd.Flags().BoolVar(&showModel, "model", true, "print the counterexample/witness model")
	cmd.Flags().BoolVar(&showSMT, "smt", false, "print the full SMT-LIB2 script sent to z3 (encoding + compiled query)")
	return cmd
}
