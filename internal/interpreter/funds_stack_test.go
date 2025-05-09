package interpreter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnoughBalance(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(100)},
	})

	out := stack.Pull(big.NewInt(2))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	}, out)

}

func TestSimple(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(10)},
	})

	out := stack.Pull(big.NewInt(5))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(3)},
	}, out)

	out = stack.Pull(big.NewInt(7))
	require.Equal(t, []Sender{
		{Name: "s2", Amount: big.NewInt(7)},
	}, out)
}

func TestPullZero(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(10)},
	})

	out := stack.Pull(big.NewInt(0))
	require.Equal(t, []Sender(nil), out)
}

func TestCompactFunds(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s1", Amount: big.NewInt(10)},
	})

	out := stack.Pull(big.NewInt(5))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(5)},
	}, out)
}

func TestCompactFunds3Times(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s1", Amount: big.NewInt(3)},
		{Name: "s1", Amount: big.NewInt(1)},
	})

	out := stack.Pull(big.NewInt(6))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(6)},
	}, out)
}

func TestCompactFundsWithEmptySender(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(0)},
		{Name: "s1", Amount: big.NewInt(10)},
	})

	out := stack.Pull(big.NewInt(5))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(5)},
	}, out)
}

func TestMissingFunds(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	})

	out := stack.Pull(big.NewInt(300))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	}, out)
}

func TestNoZeroLeftovers(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(10)},
		{Name: "s2", Amount: big.NewInt(15)},
	})

	stack.Pull(big.NewInt(10))

	out := stack.Pull(big.NewInt(15))
	require.Equal(t, []Sender{
		{Name: "s2", Amount: big.NewInt(15)},
	}, out)
}

func TestReconcileColoredAssetExactMatch(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "src", Amount: big.NewInt(10), Color: "X"},
		{Name: "s2", Amount: big.NewInt(15)},
	})

	out := stack.Pull(big.NewInt(10))
	require.Equal(t, []Sender{
		{Name: "src", Amount: big.NewInt(10), Color: "X"},
	}, out)

}

func TestReconcileColoredManyDestPerSender(t *testing.T) {

	stack := newFundsStack([]Sender{
		{"src", big.NewInt(10), "X"},
	})

	out := stack.Pull(big.NewInt(5))
	require.Equal(t, []Sender{
		{Name: "src", Amount: big.NewInt(5), Color: "X"},
	}, out)

	out = stack.Pull(big.NewInt(5))
	require.Equal(t, []Sender{
		{Name: "src", Amount: big.NewInt(5), Color: "X"},
	}, out)

}

func TestReconcileColoredManySenderColors(t *testing.T) {
	c1 := ("c1")
	c2 := ("c2")

	stack := newFundsStack([]Sender{
		{"src", big.NewInt(1), c1},
		{"src", big.NewInt(1), c2},
	})

	out := stack.Pull(big.NewInt(2))
	require.Equal(t, []Sender{
		{Name: "src", Amount: big.NewInt(1), Color: c1},
		{Name: "src", Amount: big.NewInt(1), Color: c2},
	}, out)

}
