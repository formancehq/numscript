package interpreter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnoughBalance(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{Name: "s1", Amount: big.NewInt(100)},
	})

	out := queue.PullAnything(big.NewInt(2))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	}, out)

}

func TestPush(t *testing.T) {
	queue := newFundsQueue(nil)
	queue.Push(Sender{Name: "acc", Amount: big.NewInt(100)})

	out := queue.PullUncolored(big.NewInt(20))
	require.Equal(t, []Sender{
		{Name: "acc", Amount: big.NewInt(20)},
	}, out)

}

func TestSimple(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(10)},
	})

	out := queue.PullAnything(big.NewInt(5))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(3)},
	}, out)

	out = queue.PullAnything(big.NewInt(7))
	require.Equal(t, []Sender{
		{Name: "s2", Amount: big.NewInt(7)},
	}, out)
}

func TestPullZero(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(10)},
	})

	out := queue.PullAnything(big.NewInt(0))
	require.Equal(t, []Sender{}, out)
}

func TestCompactFunds(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s1", Amount: big.NewInt(10)},
	})

	out := queue.PullAnything(big.NewInt(5))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(5)},
	}, out)
}

func TestCompactFunds3Times(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s1", Amount: big.NewInt(3)},
		{Name: "s1", Amount: big.NewInt(1)},
	})

	out := queue.PullAnything(big.NewInt(6))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(6)},
	}, out)
}

func TestCompactFundsWithEmptySender(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(0)},
		{Name: "s1", Amount: big.NewInt(10)},
	})

	out := queue.PullAnything(big.NewInt(5))

	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(5)},
	}, out)
}

func TestMissingFunds(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	})

	out := queue.PullAnything(big.NewInt(300))
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	}, out)
}

func TestNoZeroLeftovers(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{Name: "s1", Amount: big.NewInt(10)},
		{Name: "s2", Amount: big.NewInt(15)},
	})

	queue.PullAnything(big.NewInt(10))

	out := queue.PullAnything(big.NewInt(15))
	require.Equal(t, []Sender{
		{Name: "s2", Amount: big.NewInt(15)},
	}, out)
}

func TestReconcileColoredManyDestPerSender(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{"src", big.NewInt(10), "X"},
	})

	out := queue.PullColored(big.NewInt(5), "X")
	require.Equal(t, []Sender{
		{Name: "src", Amount: big.NewInt(5), Color: "X"},
	}, out)

	out = queue.PullColored(big.NewInt(5), "X")
	require.Equal(t, []Sender{
		{Name: "src", Amount: big.NewInt(5), Color: "X"},
	}, out)

}

func TestPullColored(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{Name: "s1", Amount: big.NewInt(5)},
		{Name: "s2", Amount: big.NewInt(1), Color: "red"},
		{Name: "s3", Amount: big.NewInt(10)},
		{Name: "s4", Amount: big.NewInt(2), Color: "red"},
		{Name: "s5", Amount: big.NewInt(5)},
	})

	out := queue.PullColored(big.NewInt(2), "red")
	require.Equal(t, []Sender{
		{Name: "s2", Amount: big.NewInt(1), Color: "red"},
		{Name: "s4", Amount: big.NewInt(1), Color: "red"},
	}, out)

	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(5)},
		{Name: "s3", Amount: big.NewInt(10)},
		{Name: "s4", Amount: big.NewInt(1), Color: "red"},
		{Name: "s5", Amount: big.NewInt(5)},
	}, queue.PullAll())
}

func TestPullColoredComplex(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{"s1", big.NewInt(1), "c1"},
		{"s2", big.NewInt(1), "c2"},
	})

	out := queue.PullColored(big.NewInt(1), "c2")
	require.Equal(t, []Sender{
		{Name: "s2", Amount: big.NewInt(1), Color: "c2"},
	}, out)
}

func TestClone(t *testing.T) {

	fs := newFundsQueue([]Sender{
		{"s1", big.NewInt(10), ""},
	})

	cloned := fs.Clone()

	fs.PullAll()

	require.Equal(t, []Sender{
		{"s1", big.NewInt(10), ""},
	}, cloned.PullAll())

}

func TestCompactFundsAndPush(t *testing.T) {
	noCol := ""

	queue := newFundsQueue([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s1", Amount: big.NewInt(10)},
	})

	queue.Pull(big.NewInt(1), &noCol)

	queue.Push(Sender{
		Name:   "pushed",
		Amount: big.NewInt(42),
	})

	out := queue.PullAll()
	require.Equal(t, []Sender{
		{Name: "s1", Amount: big.NewInt(11)},
		{Name: "pushed", Amount: big.NewInt(42)},
	}, out)
}
