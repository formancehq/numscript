package prove

import (
	"math/big"
	"time"

	"github.com/formancehq/numscript/internal/prove/encode"
	"github.com/formancehq/numscript/internal/prove/prop"
)

type Status int

const (
	// Valid: the property holds for every initial balance.
	Valid Status = iota
	// Counterexample: the solver found initial balances on which the property
	// is false, and the VM confirmed them.
	Counterexample
	// Unknown: the solver gave up (timeout).
	Unknown
	// ToolBug: the solver's model does not reproduce on the VM, or the
	// property does not evaluate to false on it. A bug in the prover.
	ToolBug
	// SolverError: the solver could not be run or rejected the problem.
	SolverError
)

func (s Status) String() string {
	switch s {
	case Valid:
		return "valid"
	case Counterexample:
		return "counterexample"
	case Unknown:
		return "unknown"
	case ToolBug:
		return "tool-bug"
	case SolverError:
		return "solver-error"
	}
	return "?"
}

func (s Status) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

type Result struct {
	Property prop.Property
	Status   Status
	Duration time.Duration
	// SMT is the complete problem the solver was given.
	SMT string
	Err error

	// Set on Counterexample and ToolBug.
	Before map[encode.Key]*big.Int
	Model  encode.Outcome
	Replay encode.Outcome
	// Diffs lists model/VM disagreements (ToolBug).
	Diffs []string
	// Replayed is the property's value on the VM run: false on a
	// counterexample, true means ToolBug.
	Replayed bool
}

type Report struct {
	Problem *encode.Problem
	// SuccessReachable and FailureReachable say whether some initial balance
	// makes the run succeed / fail; nil when the solver could not tell.
	SuccessReachable *bool
	FailureReachable *bool
	Results          []Result
}

// ExitCode maps a report to the CLI's contract: 0 all valid, 1 a
// counterexample, 2 only unknowns otherwise, 3 a tool or solver problem.
func (r *Report) ExitCode() int {
	code := 0
	for _, res := range r.Results {
		switch res.Status {
		case Counterexample:
			if code < 1 {
				code = 1
			}
		case Unknown:
			if code < 2 {
				code = 2
			}
		case ToolBug, SolverError:
			code = 3
		}
	}
	return code
}
