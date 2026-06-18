package builder_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/builder"
	"github.com/gkampitakis/go-snaps/snaps"
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
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`send [$asset_0 42] (
  source = $account_0
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

	_, script := builder.BuildProgram(stmt)
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`send [$asset_0 42] (
  source = {
    $account_0
    $account_1
  }
)`),
	)
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
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`send [$asset_0 42] (
  source = {
    $account_0
    $account_1
    {
      $account_2
      $account_3
    }
    $account_4
  }
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
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`send [$asset_0 42] (
  source = {
    $account_0 \ $string_0
    $account_1
  }
)`),
	)
}
