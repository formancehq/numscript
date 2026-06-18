package builder_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/builder"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestSimpleSend(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(42)),
		),
		builder.SrcAccount(
			builder.ExprAccount("src"),
		),
		builder.DestAccount(
			builder.ExprAccount("dest"),
		),
	)

	_, script := builder.BuildProgram(stmt)
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`vars {
  account $account_0
  account $account_1
  asset $asset_0
}

send [$asset_0 42] (
  source = $account_0
  destination = $account_1
)`))
}

func TestInorder(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(42)),
		),
		builder.SrcInorder(
			builder.SrcAccount(
				builder.ExprAccount("src1"),
			),
			builder.SrcAccount(
				builder.ExprAccount("src2"),
			),
		),
		builder.DestAccount(
			builder.ExprAccount("dest"),
		),
	)

	vars, script := builder.BuildProgram(stmt)
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`vars {
  account $account_0
  account $account_1
  account $account_2
  asset $asset_0
}

send [$asset_0 42] (
  source = {
    $account_0
    $account_1
  }
  destination = $account_2
)`))

	require.Equal(t, map[string]string{
		"account_0": "src1",
		"account_1": "src2",
		"account_2": "dest",
		"asset_0":   "USD/2",
	}, vars)
}

func TestInorderNested(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(42)),
		),
		builder.SrcInorder(
			builder.SrcAccount(
				builder.ExprAccount("src1"),
			),
			builder.SrcAccount(
				builder.ExprAccount("src2"),
			),
			builder.SrcInorder(
				builder.SrcAccount(
					builder.ExprAccount("src_nested1"),
				),
				builder.SrcAccount(
					builder.ExprAccount("src_nested2"),
				),
			),
			builder.SrcAccount(
				builder.ExprAccount("src_upper"),
			),
		),
		builder.DestAccount(
			builder.ExprAccount("dest"),
		),
	)

	_, script := builder.BuildProgram(stmt)
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`vars {
  account $account_0
  account $account_1
  account $account_2
  account $account_3
  account $account_4
  account $account_5
  asset $asset_0
}

send [$asset_0 42] (
  source = {
    $account_0
    $account_1
    {
      $account_2
      $account_3
    }
    $account_4
  }
  destination = $account_5
)`))
}

func TestInorderWithColors(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(42)),
		),
		builder.SrcInorder(
			builder.SrcColored(
				builder.ExprAccount("acc"),
				builder.ExprString("col"),
			),
			builder.SrcAccount(
				builder.ExprAccount("src2"),
			),
		),
		builder.DestAccount(
			builder.ExprAccount("dest"),
		),
	)

	_, script := builder.BuildProgram(stmt)
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`vars {
  account $account_0
  account $account_1
  account $account_2
  string $string_0
  asset $asset_0
}

send [$asset_0 42] (
  source = {
    $account_0 \ $string_0
    $account_1
  }
  destination = $account_2
)`))
}
