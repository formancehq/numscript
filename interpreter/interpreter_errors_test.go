package interpreter_test

import (
	"testing"

	"github.com/formancehq/numscript/interpreter"
	"github.com/formancehq/numscript/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func matchErrWithSnapshots(t *testing.T, src string) {
	parsed := parser.Parse(src)

	_, err := interpreter.RunProgram(parsed.Value, interpreter.RunProgramOptions{})
	require.NotNil(t, err)
	snaps.MatchSnapshot(t, err.GetRange().ShowOnSource(parsed.Source))
}

func TestShowUnboundVar(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = $unbound_var
  destination = @dest
)`)

}

func TestShowMissingFundsSingleAccount(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = @a
  destination = @dest
)`)

}

func TestShowMissingFundsInorder(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = {
    @a
    @b
	}
  destination = @dest
)`)
}

func TestShowMissingFundsAllotment(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = {
    1/2 from @a
		remaining from @world
	}
  destination = @dest
)`)
}

func TestShowMissingFundsMax(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = max [COIN 2] from {
    1/2 from @world
    remaining from @world
  }
  destination = @dest
)`)
}
