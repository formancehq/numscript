package parser_test

import (
	"numscript/parser"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
)

func TestFaultToleranceVarName(t *testing.T) {
	p := parser.Parse(`vars { monetary 42  }`)
	snaps.MatchSnapshot(t, p.Value)
	assert.NotEmpty(t, p.Errors)
}

func TestFaultToleranceSend(t *testing.T) {
	p := parser.Parse(`send `)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceMonetary(t *testing.T) {
	p := parser.Parse(`send [COIN]`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceNoAddr(t *testing.T) {
	p := parser.Parse(`send  (
  source = {
	@
  }
  destination = @
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceInvalidDest(t *testing.T) {
	p := parser.Parse(`send [COIN 10] (
    source = @a
    destination =
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceInvalidSrcTk(t *testing.T) {
	p := parser.Parse(`send [COIN 10] (
    source = max
    destination = @d
)`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceTrailingComma(t *testing.T) {
	p := parser.Parse(`set_tx_meta(1, )`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceDestinationNoRemainingMispelledFrom(t *testing.T) {
	p := parser.Parse(`send [COIN 10] (
		source = @a
		destination = {
			max [COIN 10] from @x
		}
	)
	`)
	snaps.MatchSnapshot(t, p.Value)
}

func TestFaultToleranceIncompleteOrigin(t *testing.T) {
	p := parser.Parse(`
vars {
	asset $a = 
}
	`)
	snaps.MatchSnapshot(t, p.Value)
}
