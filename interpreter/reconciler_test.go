package interpreter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ReconcileTestCase struct {
	Senders     []Sender
	Receivers   []Receiver
	Expected    []Posting
	ExpectedErr error
}

func runReconcileTestCase(t *testing.T, tc ReconcileTestCase) {
	got, err := Reconcile(tc.Senders, tc.Receivers)
	assert.Equal(t, got, tc.Expected)
	assert.Equal(t, err, tc.ExpectedErr)
}

func TestReconcileEmpty(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{})
}

func TestReconcileSingletonExactMatch(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders:   []Sender{{"src", big.NewInt(10), "COIN"}},
		Receivers: []Receiver{{"dest", big.NewInt(10), "COIN"}},
		Expected:  []Posting{{"src", "dest", big.NewInt(10), "COIN"}},
	})
}

func TestNoReceiversLeft(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{{
			"src",
			big.NewInt(10),
			"",
		}},
	})
}

func TestNoSendersLeft(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Receivers:   []Receiver{{"dest", big.NewInt(10), "EUR"}},
		ExpectedErr: ReconcileError{},
	})
}

func TestReconcileSendersRemainder(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{{"src", big.NewInt(100), "EUR"}},
		Receivers: []Receiver{
			{
				"d1",
				big.NewInt(70),
				"EUR",
			},
			{
				"d2",
				big.NewInt(30),
				"EUR",
			}},
		Expected: []Posting{
			{"src", "d1", big.NewInt(70), "EUR"},
			{"src", "d2", big.NewInt(30), "EUR"},
		},
	})
}

func TestReconcileWhenSendersAreSplit(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{
			{"s1", big.NewInt(20), "EUR"},
			{"s2", big.NewInt(30), "EUR"},
		},
		Receivers: []Receiver{{"d", big.NewInt(50), "EUR"}},
		Expected: []Posting{
			{"s1", "d", big.NewInt(20), "EUR"},
			{"s2", "d", big.NewInt(30), "EUR"},
		},
	})
}

func TestReconcileOverlapping(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{
			{"src1", big.NewInt(1), "EUR"},
			{"src2", big.NewInt(10), "EUR"},
			{"src2", big.NewInt(20), "EUR"},
		},
		Receivers: []Receiver{{"d", big.NewInt(31), "EUR"}},
		Expected: []Posting{
			{"src1", "d", big.NewInt(1), "EUR"},
			{"src2", "d", big.NewInt(30), "EUR"},
		},
	})
}
