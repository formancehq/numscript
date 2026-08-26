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

func TestSimpleSendWithFlag(t *testing.T) {
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

	_, _, script := builder.BuildProgramWithFeatureFlags([]string{
		"my-flag",
		"another-flag",
	}, stmt)
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`#![feature("my-flag", "another-flag")]
vars {
  account $account_0
  account $account_1
  asset $asset_0
}

send [$asset_0 42] (
  source = $account_0
  destination = $account_1
)`))
}

func TestSendAll(t *testing.T) {
	stmt := builder.StmtSendAll(
		builder.ExprAsset("COIN"),
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

send [$asset_0 *] (
  source = $account_0
  destination = $account_1
)`))
}

func TestSrcAllowingBoundedOverdraft(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(100)),
		),
		builder.SrcAccountOverdraft(
			builder.ExprAccount("tmp:acc"),
			builder.BoundedOverdraft(
				builder.ExprMonetary(
					builder.ExprAsset("USD/2"),
					builder.ExprNumberBigInt(big.NewInt(100)),
				),
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
  asset $asset_0
}

send [$asset_0 100] (
  source = $account_0 allowing overdraft up to [$asset_0 100]
  destination = $account_1
)`),
	)
}

func TestColoredAllowingUnboundedOverdraft(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(100)),
		),
		builder.SrcColoredOverdraft(
			builder.ExprAccount("tmp:acc"),
			builder.ExprString("ABCDEF"),
			builder.UnboundedOverdraft(),
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
				builder.SrcAccountOverdraft(
					builder.ExprAccount("src_nested2"),
					builder.UnboundedOverdraft(),
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
      $account_3 allowing unbounded overdraft
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

func TestSrcCapped(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(42)),
		),
		builder.SrcCapped(
			builder.ExprMonetary(
				builder.ExprAsset("USD/2"),
				builder.ExprNumberBigInt(big.NewInt(10)),
			),
			builder.SrcAccount(
				builder.ExprAccount("src"),
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
  asset $asset_0
}

send [$asset_0 42] (
  source = max [$asset_0 10] from $account_0
  destination = $account_1
)`))
}

func TestSrcAllotment(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(42)),
		),
		builder.SrcAllotment(
			builder.AllotmentClause[builder.Source]{
				Portion: builder.NewPortion(big.NewInt(1), big.NewInt(3)),
				Payload: builder.SrcAccount(builder.ExprAccount("a")),
			},
			builder.AllotmentClause[builder.Source]{
				Portion: builder.NewPortion(big.NewInt(2), big.NewInt(3)),
				Payload: builder.SrcAccount(builder.ExprAccount("b")),
			},
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
  asset $asset_0
}

send [$asset_0 42] (
  source = {
    1/3 from $account_0
    2/3 from $account_1
  }
  destination = $account_2
)`))
}

func TestDestInorder(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(42)),
		),
		builder.SrcAccount(
			builder.ExprAccount("src"),
		),
		builder.DestInorder(
			[]builder.DestInorderClause{
				{
					Max: builder.ExprMonetary(
						builder.ExprAsset("USD/2"),
						builder.ExprNumberBigInt(big.NewInt(10)),
					),
					Dest: builder.To(builder.DestAccount(builder.ExprAccount("a"))),
				},
			},
			builder.Kept(),
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
  destination = {
    max [$asset_0 10] to $account_1
    remaining kept
  }
)`))
}

func TestDestAllotment(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(42)),
		),
		builder.SrcAccount(
			builder.ExprAccount("src"),
		),
		builder.DestAllotment(
			builder.AllotmentClause[builder.KeptOrDest]{
				Portion: builder.NewPortion(big.NewInt(1), big.NewInt(2)),
				Payload: builder.To(builder.DestAccount(builder.ExprAccount("a"))),
			},
			builder.AllotmentClause[builder.KeptOrDest]{
				Portion: builder.NewPortion(big.NewInt(1), big.NewInt(2)),
				Payload: builder.Kept(),
			},
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
  destination = {
    1/2 to $account_1
    1/2 kept
  }
)`))
}

func TestUnsafeAccount(t *testing.T) {
	stmt := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(42)),
		),
		builder.SrcAccount(
			builder.UnsafeAccount("world"),
		),
		builder.DestAccount(
			builder.UnsafeAccount("acc0"),
		),
	)

	// UnsafeAccount bypasses the vars pool entirely: no account entries in
	// the returned bindings map (the asset is still pooled as usual).
	vars, _, script := builder.BuildProgram(stmt)
	for k := range vars {
		require.NotContains(t, k, "account")
	}
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`vars {
  asset $asset_0
}

send [$asset_0 42] (
  source = @world
  destination = @acc0
)`))
}

func TestMultipleStatements(t *testing.T) {
	stmt1 := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(1)),
		),
		builder.SrcAccount(builder.UnsafeAccount("a")),
		builder.DestAccount(builder.UnsafeAccount("b")),
	)
	stmt2 := builder.StmtSend(
		builder.ExprMonetary(
			builder.ExprAsset("USD/2"),
			builder.ExprNumberBigInt(big.NewInt(2)),
		),
		builder.SrcAccount(builder.UnsafeAccount("b")),
		builder.DestAccount(builder.UnsafeAccount("c")),
	)

	_, _, script := builder.BuildProgram(stmt1, stmt2)
	snaps.MatchInlineSnapshot(t, script, snaps.Inline(`vars {
  asset $asset_0
}

send [$asset_0 1] (
  source = @a
  destination = @b
)

send [$asset_0 2] (
  source = @b
  destination = @c
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
