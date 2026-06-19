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

	_, _, script := builder.BuildProgram(stmt)
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

	vars, _, script := builder.BuildProgram(stmt)
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

	_, _, script := builder.BuildProgram(stmt)
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

	_, _, script := builder.BuildProgram(stmt)
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

func TestWithExternVar(t *testing.T) {
	// The builder module exposes a type-safe API to create scripts

	// We can create (typed) vars this way:
	accVar := builder.NewAccountVar()
	amtVar := builder.NewNumberVar()

	// sources, destinations, expressions, and statements are typed, so you can never
	// use a source node instead of a expression node, and so on.
	// In addition to that, expressions are typed, so you can't use a string expression
	// instead of a number expression
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprVar(&amtVar), // <- you can reference vars this way (identified by address)
		),
		builder.SrcAccount(
			builder.ExprVar(&accVar),
		),
		builder.DestAccount(
			builder.ExprAccount("dest"),
		),
	)

	// When you build the program, it'll create 3 values:
	vars, varsEnv, script := builder.BuildProgram(stmt)

	// 1: (vars) the map[string]string of KNOWN variables. This is generated via the
	// strings literals you pass (in the example above, USD/2 and @dest).
	// This way, you'll pass this map to the tx and numscript will handle interpolation instead of
	// handling that in this lib
	require.Equal(t, map[string]string{
		"asset_0":   "USD/2",
		"account_1": "dest",
	}, vars)

	// 2: (varsEnv) The env of the NAMES of each variable that is referenced within the script.
	// You'll reference them by ptr address. The "Fill*()" methods are typed, and return you the name of the var,
	// and the "stringified" value of the var content (in the case of account/asset/string, the string itself)
	// Behaviour of Fill*() of vars that are never referenced in the script (thus, whose name is never allocated) is undefined
	// (it may panic in the future)
	//
	// user code would likely be something like:
	//
	// varsCp := maps.Clone(vars)
	// k, v := varsEnv.FillAccount(..)
	// varsCp[k] = v
	// (etc)
	k, v := varsEnv.FillAccount(&accVar, "my_src")
	require.Equal(t, "account_0", k)
	require.Equal(t, "my_src", v)

	k, v = varsEnv.FillNumber(&amtVar, big.NewInt(42))
	require.Equal(t, "number_0", k)
	require.Equal(t, "42", v)

	// 3: (script) the script itself. It's generated while trying to be as "stable" as possible,
	// e.g. vars are declared with an hardcoded order of types, and with an increasing order of names
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`vars {
  account $account_0
  account $account_1
  asset $asset_0
  number $number_0
}

send [$asset_0 $number_0] (
  source = $account_0
  destination = $account_1
)`))
}
