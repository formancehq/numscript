package interpreter_test

import (
	"testing"

	"github.com/formancehq/numscript/interpreter"
	"github.com/formancehq/numscript/parser"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestShowUnboundVar(t *testing.T) {
	parsed := parser.Parse(`send [COIN 10] (
  source = $unbound_var
  destination = @dest
)`)

	_, err := interpreter.RunProgram(parsed.Value, interpreter.RunProgramOptions{})
	snaps.MatchSnapshot(t, err.GetRange().ShowOnSource(parsed.Source))
}
