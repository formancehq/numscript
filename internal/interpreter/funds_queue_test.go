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

	out := queue.PullAnything(big.NewInt(20))
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
	require.Equal(t, []Sender(nil), out)
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

func TestClone(t *testing.T) {

	fq := newFundsQueue([]Sender{
		{"s1", big.NewInt(10), ""},
	})

	cloned := fq.Clone()

	fq.PullAll()

	require.Equal(t, []Sender{
		{"s1", big.NewInt(10), ""},
	}, cloned.PullAll())

}

func TestCompactFundsAndPush(t *testing.T) {
	queue := newFundsQueue([]Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s1", Amount: big.NewInt(10)},
	})

	queue.Pull(big.NewInt(1))

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
