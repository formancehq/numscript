package interpreter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ReconcileTestCase struct {
	Senders     []Sender
	Receivers   []Receiver
	Expected    []Posting
	ExpectedErr error
}

func runReconcileTestCase(t *testing.T, tc ReconcileTestCase) {
	got, err := Reconcile(tc.Senders, tc.Receivers)

	require.Equal(t, tc.ExpectedErr, err)
	assert.Equal(t, tc.Expected, got)
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
	// TODO delete test?
	t.Skip()

	runReconcileTestCase(t, ReconcileTestCase{
		Receivers: []Receiver{{"dest", big.NewInt(10), "EUR"}},
		ExpectedErr: ReconcileError{
			Receiver:  Receiver{"dest", big.NewInt(10), "EUR"},
			Receivers: make([]Receiver, 0),
		},
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

func TestMany(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{
			{"s1", big.NewInt(80 + 20), "EUR"},
			{"s2", big.NewInt(1000), "EUR"},
		},
		Receivers: []Receiver{
			{"d1", big.NewInt(80), "EUR"},
			{"d2", big.NewInt(20 + 123), "EUR"},
		},
		Expected: []Posting{
			{"s1", "d1", big.NewInt(80), "EUR"},
			{"s1", "d2", big.NewInt(20), "EUR"},
			{"s2", "d2", big.NewInt(123), "EUR"},
		},
	})
}

func TestReconcileManySendersManyReceivers(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{
			{"s1", big.NewInt(80 + 20), "EUR"},
			{"s2", big.NewInt(1000), "EUR"},
		},
		Receivers: []Receiver{
			{"d1", big.NewInt(80), "EUR"},
			{"d2", big.NewInt(20 + 123), "EUR"},
		},
		Expected: []Posting{
			{"s1", "d1", big.NewInt(80), "EUR"},
			{"s1", "d2", big.NewInt(20), "EUR"},
			{"s2", "d2", big.NewInt(123), "EUR"},
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

func TestReconcileKept(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{
			{"src", big.NewInt(100), "GEM"},
		},
		Receivers: []Receiver{
			{"dest", big.NewInt(50), "EUR"},
			{"<kept>", big.NewInt(50), "EUR"}},
		Expected: []Posting{
			{"src", "dest", big.NewInt(50), "GEM"},
		},
	})
}

func TestReconcileEmptyMonetaryForDest(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{
			{"src", big.NewInt(100), "GEM"},
		},
		Receivers: []Receiver{
			{"dest", nil, "EUR"},
		},
		Expected: []Posting{
			{"src", "dest", big.NewInt(100), "GEM"},
		},
	})
}

func TestReconcileSendAllMixed(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{
			{"src", big.NewInt(100), "GEM"},
		},
		Receivers: []Receiver{
			{"d1", big.NewInt(20), "GEM"},
			{"d2", nil, "GEM"},
		},
		Expected: []Posting{
			{"src", "d1", big.NewInt(20), "GEM"},
			{"src", "d2", big.NewInt(80), "GEM"},
		},
	})
}

func TestReconcileSendMultiSrc(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{
			{"src1", big.NewInt(10), "GEM"},
			{"src2", big.NewInt(20), "GEM"},
		},
		Receivers: []Receiver{
			{"dest", nil, "GEM"},
		},
		Expected: []Posting{
			{"src1", "dest", big.NewInt(10), "GEM"},
			{"src2", "dest", big.NewInt(20), "GEM"},
		},
	})
}
