package builder_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/builder"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestSrcAllowingUnboundedOverdraft(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(100)),
		),
		builder.SrcAllowingUnboundedOverdraft(
			builder.SrcColored(
				builder.ExprAccount("tmp:acc"),
				builder.ExprString("ABCDEF"),
			),
		),
		builder.DestAccount(
			builder.ExprAccount("dest"),
		),
	)

	_, _, script := builder.BuildProgram(stmt)
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`vars {
  account $account_0
  account $account_1
  string $string_0
  asset $asset_0
}

send [$asset_0 100] (
  source = $account_0 \ $string_0 allowing unbounded overdraft
  destination = $account_1
)`))
}
