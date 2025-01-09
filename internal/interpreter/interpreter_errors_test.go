package interpreter_test

import (
	"context"
	"testing"

	"github.com/PagoPlus/numscript-wasm/internal/interpreter"
	"github.com/PagoPlus/numscript-wasm/internal/parser"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func matchErrWithSnapshots(t *testing.T, src string, vars map[string]string, runOpt interpreter.StaticStore) {
	parsed := parser.Parse(src)
	_, err := interpreter.RunProgram(context.Background(), parsed.Value, vars, runOpt, nil)
	require.NotNil(t, err)
	snaps.MatchSnapshot(t, err.GetRange().ShowOnSource(parsed.Source))
}

func TestShowUnboundVar(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = $unbound_var
  destination = @dest
)`, nil, interpreter.StaticStore{})

}

func TestShowMissingFundsSingleAccount(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = @a
  destination = @dest
)`, nil, interpreter.StaticStore{})

}

func TestShowMissingFundsInorder(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = {
    @a
    @b
	}
  destination = @dest
)`, nil, interpreter.StaticStore{})
}

func TestShowMissingFundsAllotment(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = {
    1/2 from @a
		remaining from @world
	}
  destination = @dest
)`, nil, interpreter.StaticStore{})
}

func TestShowMissingFundsMax(t *testing.T) {
	matchErrWithSnapshots(t, `send [COIN 10] (
  source = max [COIN 2] from {
    1/2 from @world
    remaining from @world
  }
  destination = @dest
)`, nil, interpreter.StaticStore{})
}

func TestShowMetadataNotFound(t *testing.T) {
	matchErrWithSnapshots(t, `vars {
  number $my_var = meta(@acc, "key")
}
`, nil, interpreter.StaticStore{})
}

func TestShowTypeError(t *testing.T) {
	matchErrWithSnapshots(t, `send 42 (
	source = @a
	destination = @b
)
`, nil, interpreter.StaticStore{})
}

func TestShowInvalidTypeErr(t *testing.T) {
	matchErrWithSnapshots(t, `vars {
  invalid_t $x
}
`, map[string]string{"x": "42"},
		interpreter.StaticStore{},
	)
}
