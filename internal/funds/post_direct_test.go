package funds_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/formancehq/numscript/internal/funds"
)

func TestPostDirect_CreditsDestinationOnly(t *testing.T) {
	rs, store := newRS(map[funds.PairKey]int64{{Account: "src", Asset: usd}: 100})

	require.NoError(t, rs.PostDirect("src", "", "dst", "", "", big.NewInt(30)))

	require.Equal(t, []funds.Posting{
		{Source: "src", Destination: "dst", Asset: usd, Amount: big.NewInt(30)},
	}, rs.GetPostings())
	require.False(t, rs.Has("dst", "", usd, ""), "the credit is a delta: no fetch")
	require.Equal(t, 0, store.callCount("src", usd))

	require.Equal(t, int64(30), accBal(rs, "dst", "", usd, "").Int64())
	require.Equal(t, int64(100), accBal(rs, "src", "", usd, "").Int64())
}

func TestPostDirect_NonPositiveIsNoOp(t *testing.T) {
	rs, _ := newRS(nil)

	require.NoError(t, rs.PostDirect("src", "", "dst", "", "", big.NewInt(0)))
	require.NoError(t, rs.PostDirect("src", "", "dst", "", "", big.NewInt(-1)))

	require.Empty(t, rs.GetPostings())
	require.Equal(t, int64(0), accBal(rs, "dst", "", usd, "").Int64())
}

func TestPostDirect_CarriesScopesAndColor(t *testing.T) {
	rs, _ := newRS(nil)

	require.NoError(t, rs.PostDirect("src", "s1", "dst", "s2", "RED", big.NewInt(5)))

	require.Equal(t, []funds.Posting{{
		Source:           "src",
		SourceScope:      "s1",
		Destination:      "dst",
		DestinationScope: "s2",
		Asset:            usd,
		Color:            "RED",
		Amount:           big.NewInt(5),
	}}, rs.GetPostings())
	require.Equal(t, int64(5), accBal(rs, "dst", "s2", usd, "RED").Int64())
	require.Equal(t, int64(0), accBal(rs, "dst", "", usd, "").Int64())
}

func TestPostDirect_ClonesAmount(t *testing.T) {
	rs, _ := newRS(nil)
	amount := big.NewInt(5)

	require.NoError(t, rs.PostDirect("src", "", "dst", "", "", amount))
	amount.SetInt64(99)

	require.Equal(t, int64(5), rs.GetPostings()[0].Amount.Int64())
}
