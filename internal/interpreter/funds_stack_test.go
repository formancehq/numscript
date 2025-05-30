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

	out := stack.PullAnything(big.NewInt(2))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	}, out)

}

func TestPush(t *testing.T) {
	stack := newFundsStack(nil)
	stack.Push(Sender{Name: "acc", Amount: big.NewInt(100)})

	out := stack.PullUncolored(big.NewInt(20))
	require.Equal(t, []Sender{
		{Name: "acc", Amount: big.NewInt(20)},
	}, out)

}

func TestSimple(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(10)},
	})

	out := stack.PullAnything(big.NewInt(5))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(3)},
	}, out)

	out = stack.PullAnything(big.NewInt(7))
	require.Equal(t, []Sender{
		{Name: "s2", Amount: big.NewInt(7)},
	}, out)
}

func TestPullZero(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(10)},
	})

	out := stack.PullAnything(big.NewInt(0))
	require.Equal(t, []Sender(nil), out)
}

func TestCompactFunds(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s1", Amount: big.NewInt(10)},
	})

	out := stack.PullAnything(big.NewInt(5))
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

	out := stack.PullAnything(big.NewInt(6))
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

	out := stack.PullAnything(big.NewInt(5))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(5)},
	}, out)
}

func TestMissingFunds(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	})

	out := stack.PullAnything(big.NewInt(300))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	}, out)
}

func TestNoZeroLeftovers(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(10)},
		{Name: "s2", Amount: big.NewInt(15)},
	})

	stack.PullAnything(big.NewInt(10))

	out := stack.PullAnything(big.NewInt(15))
	require.Equal(t, []Sender{
		{Name: "s2", Amount: big.NewInt(15)},
	}, out)
}

func TestReconcileColoredManyDestPerSender(t *testing.T) {
	stack := newFundsStack([]Sender{
		{"src", big.NewInt(10), "X"},
	})

	out := stack.PullColored(big.NewInt(5), "X")
	require.Equal(t, []Sender{
		{Name: "src", Amount: big.NewInt(5), Color: "X"},
	}, out)

	out = stack.PullColored(big.NewInt(5), "X")
	require.Equal(t, []Sender{
		{Name: "src", Amount: big.NewInt(5), Color: "X"},
	}, out)

}

func TestPullColored(t *testing.T) {
	stack := newFundsStack([]Sender{
		{Name: "s1", Amount: big.NewInt(5)},
		{Name: "s2", Amount: big.NewInt(1), Color: "red"},
		{Name: "s3", Amount: big.NewInt(10)},
		{Name: "s4", Amount: big.NewInt(2), Color: "red"},
		{Name: "s5", Amount: big.NewInt(5)},
	})

	out := stack.PullColored(big.NewInt(2), "red")
	require.Equal(t, []Sender{
		{Name: "s2", Amount: big.NewInt(1), Color: "red"},
		{Name: "s4", Amount: big.NewInt(1), Color: "red"},
	}, out)

	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(5)},
		{Name: "s3", Amount: big.NewInt(10)},
		{Name: "s4", Amount: big.NewInt(1), Color: "red"},
		{Name: "s5", Amount: big.NewInt(5)},
	}, stack.PullAll())
}

func TestPullColoredComplex(t *testing.T) {
	stack := newFundsStack([]Sender{
		{"s1", big.NewInt(1), "c1"},
		{"s2", big.NewInt(1), "c2"},
	})

	out := stack.PullColored(big.NewInt(1), "c2")
	require.Equal(t, []Sender{
		{Name: "s2", Amount: big.NewInt(1), Color: "c2"},
	}, out)
}

func TestClone(t *testing.T) {

	fs := newFundsStack([]Sender{
		{"s1", big.NewInt(10), ""},
	})

	cloned := fs.Clone()

	fs.PullAll()

	require.Equal(t, []Sender{
		{"s1", big.NewInt(10), ""},
	}, cloned.PullAll())

}
