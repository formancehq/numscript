package interpreter_test

import (
	"testing"

	"github.com/formancehq/numscript/interpreter"
	"github.com/formancehq/numscript/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func matchErrWithSnapshots(t *testing.T, src string, runOpt interpreter.RunProgramOptions) {
	parsed := parser.Parse(src)
	_, err := interpreter.RunProgram(parsed.Value, runOpt)
	require.NotNil(t, err)
	snaps.MatchSnapshot(t, err.GetRange().ShowOnSource(parsed.Source))
}

func TestShowUnboundVar(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = $unbound_var
  destination = @dest
)`, interpreter.RunProgramOptions{})

}

func TestShowMissingFundsSingleAccount(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = @a
  destination = @dest
)`, interpreter.RunProgramOptions{})

}

func TestShowMissingFundsInorder(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = {
    @a
    @b
	}
  destination = @dest
)`, interpreter.RunProgramOptions{})
}

func TestShowMissingFundsAllotment(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = {
    1/2 from @a
		remaining from @world
	}
  destination = @dest
)`, interpreter.RunProgramOptions{})
}

func TestShowMissingFundsMax(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = max [COIN 2] from {
    1/2 from @world
    remaining from @world
  }
  destination = @dest
)`, interpreter.RunProgramOptions{})
}

func TestShowMetadataNotFound(t *testing.T) {
	matchErrWithSnapshots(t, `vars {
  number $my_var = meta(@acc, "key")
}
`, interpreter.RunProgramOptions{})
}

func TestShowTypeError(t *testing.T) {
	matchErrWithSnapshots(t, `send 42 (
	source = @a
	destination = @b
)
`, interpreter.RunProgramOptions{})
}

func TestShowInvalidTypeErr(t *testing.T) {
	matchErrWithSnapshots(t, `vars {
  invalid_t $x
}
`, interpreter.RunProgramOptions{
		Vars: map[string]string{"x": "42"},
	})
}
