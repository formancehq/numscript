package funds_stack_test

import (
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/funds_stack"
	"github.com/stretchr/testify/require"
)

func TestEnoughBalance(t *testing.T) {
	stack := funds_stack.NewFundsStack([]funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(100)},
	})

	out := stack.Pull(big.NewInt(2))
	require.Equal(t, []funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	}, out)

}

func TestSimple(t *testing.T) {
	stack := funds_stack.NewFundsStack([]funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(10)},
	})

	out := stack.Pull(big.NewInt(5))
	require.Equal(t, []funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(3)},
	}, out)

	out = stack.Pull(big.NewInt(7))
	require.Equal(t, []funds_stack.Sender{
		{Name: "s2", Amount: big.NewInt(7)},
	}, out)
}

func TestPullZero(t *testing.T) {
	stack := funds_stack.NewFundsStack([]funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(10)},
	})

	out := stack.Pull(big.NewInt(0))
	require.Equal(t, []funds_stack.Sender(nil), out)
}

func TestCompactFunds(t *testing.T) {
	stack := funds_stack.NewFundsStack([]funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s1", Amount: big.NewInt(10)},
	})

	out := stack.Pull(big.NewInt(5))
	require.Equal(t, []funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(5)},
	}, out)
}

func TestCompactFunds3Times(t *testing.T) {
	stack := funds_stack.NewFundsStack([]funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s1", Amount: big.NewInt(3)},
		{Name: "s1", Amount: big.NewInt(1)},
	})

	out := stack.Pull(big.NewInt(6))
	require.Equal(t, []funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(6)},
	}, out)
}

func TestCompactFundsWithEmptySender(t *testing.T) {
	stack := funds_stack.NewFundsStack([]funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(2)},
		{Name: "s2", Amount: big.NewInt(0)},
		{Name: "s1", Amount: big.NewInt(10)},
	})

	out := stack.Pull(big.NewInt(5))
	require.Equal(t, []funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(5)},
	}, out)
}

func TestMissingFunds(t *testing.T) {
	stack := funds_stack.NewFundsStack([]funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	})

	out := stack.Pull(big.NewInt(300))
	require.Equal(t, []funds_stack.Sender{
		{Name: "s1", Amount: big.NewInt(2)},
	}, out)
}
