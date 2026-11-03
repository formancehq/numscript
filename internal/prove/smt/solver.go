package smt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Status int

const (
	Sat Status = iota
	Unsat
	Unknown
)

func (s Status) String() string {
	switch s {
	case Sat:
		return "sat"
	case Unsat:
		return "unsat"
	}
	return "unknown"
}

type Result struct {
	Status Status
	// Values is the get-value answer, keyed by the unquoted symbol name. Empty
	// unless the status is Sat.
	Values map[string]Value
	// Raw is the solver's whole stdout.
	Raw string
}

// Solver runs an SMT-LIB2 solver as a subprocess, feeding the script on stdin.
// Path is the executable ("z3" when empty). The soft timeout is passed with
// the solver's own flag when the executable is z3 or cvc5, so the solver
// answers unknown instead of being killed.
type Solver struct {
	Path    string
	Timeout time.Duration
}

var ErrSolverNotFound = errors.New("smt: solver not found")

func (s Solver) path() string {
	if s.Path == "" {
		return "z3"
	}
	return s.Path
}

// Available reports whether the solver executable can be found.
func (s Solver) Available() error {
	if _, err := exec.LookPath(s.path()); err != nil {
		return fmt.Errorf("%w: %s", ErrSolverNotFound, s.path())
	}
	return nil
}

func (s Solver) args() []string {
	ms := s.Timeout.Milliseconds()
	switch strings.TrimSuffix(filepath.Base(s.path()), ".exe") {
	case "z3":
		args := []string{"-in", "-smt2"}
		if ms > 0 {
			args = append(args, fmt.Sprintf("-t:%d", ms))
		}
		return args
	case "cvc5":
		args := []string{"--lang", "smt2"}
		if ms > 0 {
			args = append(args, fmt.Sprintf("--tlimit=%d", ms))
		}
		return args
	}
	return nil
}

// CheckSat runs script followed by (check-sat) and, when getValue is not
// empty, (get-value (...)) for those names.
func (s Solver) CheckSat(ctx context.Context, script string, getValue []string) (Result, error) {
	if err := s.Available(); err != nil {
		return Result{}, err
	}

	var input strings.Builder
	input.WriteString(script)
	input.WriteString("(check-sat)\n")
	if len(getValue) > 0 {
		input.WriteString("(get-value (")
		for i, name := range getValue {
			if i > 0 {
				input.WriteString(" ")
			}
			input.WriteString(Quote(name))
		}
		input.WriteString("))\n")
	}

	if s.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.Timeout+2*time.Second)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, s.path(), s.args()...)
	cmd.Stdin = strings.NewReader(input.String())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()
	raw := stdout.String()

	res, err := parseResult(raw)
	if err != nil {
		if runErr != nil {
			return Result{Raw: raw}, fmt.Errorf("smt: solver failed: %w\n%s%s", runErr, raw, stderr.String())
		}
		return Result{Raw: raw}, err
	}
	return res, nil
}

func parseResult(raw string) (Result, error) {
	nodes, err := parseSexprs(raw)
	if err != nil {
		return Result{}, err
	}
	res := Result{Raw: raw, Values: map[string]Value{}}
	status := false
	for _, n := range nodes {
		switch {
		case !n.isList:
			switch n.atom {
			case "sat":
				res.Status, status = Sat, true
			case "unsat":
				res.Status, status = Unsat, true
			case "unknown":
				res.Status, status = Unknown, true
			default:
				return Result{}, fmt.Errorf("smt: unexpected solver output %q", n.atom)
			}
		case len(n.list) == 2 && !n.list[0].isList && n.list[0].atom == "error":
			// get-value after unsat/unknown has no model to report, which is
			// not an error of ours
			if status && res.Status != Sat {
				continue
			}
			return Result{}, fmt.Errorf("smt: solver error: %s", n.list[1].atom)
		default:
			for _, pair := range n.list {
				if !pair.isList || len(pair.list) != 2 || pair.list[0].isList {
					return Result{}, fmt.Errorf("smt: unexpected solver output %s", pair)
				}
				v, err := parseValue(pair.list[1])
				if err != nil {
					return Result{}, err
				}
				res.Values[pair.list[0].atom] = v
			}
		}
	}
	if !status {
		return Result{}, fmt.Errorf("smt: solver printed no status:\n%s", raw)
	}
	return res, nil
}
