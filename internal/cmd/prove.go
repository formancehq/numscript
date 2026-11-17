package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/prove"
	"github.com/formancehq/numscript/internal/prove/encode"
	"github.com/formancehq/numscript/internal/prove/prop"
	"github.com/formancehq/numscript/internal/prove/smt"

	"github.com/spf13/cobra"
)

type ProveArgs struct {
	Props        []string
	PropsPath    string
	Solver       string
	Timeout      time.Duration
	EmitSMT      string
	OutFormatOpt string
}

// exitCodeError carries the exit code the prove command ends with.
type exitCodeError struct {
	code int
}

func (e exitCodeError) Error() string { return fmt.Sprintf("exit %d", e.code) }

func loadInstrs(scriptPath string) ([]ir.Instr, error) {
	if !strings.HasSuffix(scriptPath, ".ir") {
		return compileScriptToIR(scriptPath)
	}
	content, err := readIRSource(scriptPath)
	if err != nil {
		return nil, err
	}
	src := string(content)
	instrs, irErrs := ir.Parse(src)
	if len(irErrs) != 0 {
		for _, irErr := range irErrs {
			fmt.Fprintln(os.Stderr, irErr.Error())
			fmt.Fprintln(os.Stderr, irErr.Range.ShowOnSource(src))
		}
		return nil, fmt.Errorf("parsing failed")
	}
	return instrs, nil
}

func loadProps(scriptPath string, opts ProveArgs) ([]prop.Property, error) {
	var props []prop.Property
	for i, src := range opts.Props {
		p, errs := prop.ParseInline(src)
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "--prop[%d]: %s\n", i+1, e.Error())
				fmt.Fprintln(os.Stderr, e.Range.ShowOnSource(src))
			}
			return nil, fmt.Errorf("invalid property")
		}
		if p.Name == "" {
			p.Name = fmt.Sprintf("prop%d", i+1)
		}
		props = append(props, p)
	}

	path := opts.PropsPath
	if path == "" && len(opts.Props) == 0 {
		path = scriptPath + ".props"
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("no properties: pass --prop, or create '%s'", path)
		}
	}
	if path != "" {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		src := string(content)
		parsed, errs := prop.ParseFile(src)
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "%s:%s\n", path, e.Error())
				fmt.Fprintln(os.Stderr, e.Range.ShowOnSource(src))
			}
			return nil, fmt.Errorf("invalid properties file")
		}
		props = append(props, parsed...)
	}
	return props, nil
}

func proveScript(scriptPath string, opts ProveArgs) error {
	instrs, err := loadInstrs(scriptPath)
	if err != nil {
		return err
	}
	props, err := loadProps(scriptPath, opts)
	if err != nil {
		return err
	}

	solver := smt.Solver{Path: opts.Solver, Timeout: opts.Timeout}
	if err := solver.Available(); err != nil {
		return err
	}

	report, err := prove.Check(context.Background(), instrs, props, prove.Options{Solver: solver})
	if err != nil {
		var unsupported encode.UnsupportedError
		var propErr prove.PropertyError
		switch {
		case errors.As(err, &unsupported):
			return fmt.Errorf("cannot prove: %w", err)
		case errors.As(err, &propErr):
			for _, e := range propErr.Errors {
				fmt.Fprintf(os.Stderr, "property %q: %s\n", propErr.Property.Name, e.Error())
			}
			return fmt.Errorf("invalid property")
		}
		return err
	}

	if opts.EmitSMT != "" {
		for _, res := range report.Results {
			if opts.EmitSMT == stdioPath {
				fmt.Printf("; ---- %s\n%s\n", res.Property.Name, res.SMT)
				continue
			}
			if err := os.MkdirAll(opts.EmitSMT, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(opts.EmitSMT, res.Property.Name+".smt2"), []byte(res.SMT), 0o644); err != nil {
				return err
			}
		}
	}

	switch opts.OutFormatOpt {
	case OutputFormatJson:
		out, err := report.JSON(scriptPath)
		if err != nil {
			return err
		}
		if _, err := os.Stdout.Write(out); err != nil {
			return err
		}
	case OutputFormatPretty:
		if _, err := os.Stdout.WriteString(report.Pretty(scriptPath)); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid output format: %s", opts.OutFormatOpt)
	}

	if code := report.ExitCode(); code != 0 {
		return exitCodeError{code}
	}
	return nil
}

func getProveCmd() *cobra.Command {
	opts := ProveArgs{}

	cmd := cobra.Command{
		Use:   "prove <file.num|file.ir>",
		Short: "Prove properties of a script for every initial balance",
		Long: `Compile a numscript file (or parse a textual IR file) and check properties
written in the property language against it, for every initial balance. Each
property is either proved, refuted with a counterexample (initial balances the
real vm confirms), or left unknown when the solver times out.

Properties come from --prop (repeatable) and/or --props <file>; with neither,
'<script>.props' next to the script is used. See prove.md for the language.

Exit codes: 0 every property holds; 1 a property has a counterexample; 2 no
counterexample but a property is unknown; 3 the script is not supported, a
property is invalid, the solver is missing, or the prover disagrees with the vm.

Needs an SMT-LIB2 solver on PATH: z3 (default) or cvc5 (--solver cvc5).
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := proveScript(args[0], opts)
			if err != nil {
				cmd.SilenceUsage = true
				var exit exitCodeError
				if errors.As(err, &exit) {
					os.Exit(exit.code)
				}
				cmd.SilenceErrors = true
				fmt.Fprintln(os.Stderr, err)
				os.Exit(3)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&opts.Props, "prop", nil, "A property to check ('[name:] formula'); repeatable")
	cmd.Flags().StringVar(&opts.PropsPath, "props", "", "Path of a properties file (default: '<script>.props')")
	cmd.Flags().StringVar(&opts.Solver, "solver", "z3", "SMT-LIB2 solver executable: z3, cvc5, or a path")
	cmd.Flags().DurationVar(&opts.Timeout, "timeout", 10*time.Second, "Solver time limit per property")
	cmd.Flags().StringVar(&opts.EmitSMT, "emit-smt", "", "Write each property's SMT-LIB2 problem to this directory ('-' for stdout)")
	cmd.Flags().StringVarP(&opts.OutFormatOpt, "output-format", "o", OutputFormatPretty, "Set the output format. Available options: pretty, json.")

	return &cmd
}
