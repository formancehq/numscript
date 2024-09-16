package interpreter

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ReconcileTestCase struct {
	Currency string

	Senders     []Sender
	Receivers   []Receiver
	Expected    []Posting
	ExpectedErr error
}

func runReconcileTestCase(t *testing.T, tc ReconcileTestCase) {
	if tc.Currency == "" {
		tc.Currency = "COIN"
	}

	got, err := Reconcile(tc.Currency, tc.Senders, tc.Receivers)

	require.Equal(t, tc.ExpectedErr, err)
	assert.Equal(t, tc.Expected, got)
}

func TestReconcileEmpty(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{})
}

func TestReconcileSingletonExactMatch(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency:  "COIN",
		Senders:   []Sender{{"src", big.NewInt(10)}},
		Receivers: []Receiver{{"dest", big.NewInt(10)}},
		Expected:  []Posting{{"src", "dest", big.NewInt(10), "COIN"}},
	})
}

func TestReconcileZero(t *testing.T) {
	// TODO double check
	runReconcileTestCase(t, ReconcileTestCase{
		Currency:  "COIN",
		Senders:   []Sender{{"src", big.NewInt(0)}},
		Receivers: []Receiver{{"dest", big.NewInt(0)}},
		Expected:  nil,
		// []Posting{
		// 	// {"src", "dest", big.NewInt(0), "COIN"}
		// },
	})
}

func TestNoReceiversLeft(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{{
			"src",
			big.NewInt(10),
		}},
	})
}

func TestReconcileSendersRemainder(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency: "EUR",
		Senders:  []Sender{{"src", big.NewInt(100)}},
		Receivers: []Receiver{
			{
				"d1",
				big.NewInt(70),
			},
			{
				"d2",
				big.NewInt(30),
			}},
		Expected: []Posting{
			{"src", "d1", big.NewInt(70), "EUR"},
			{"src", "d2", big.NewInt(30), "EUR"},
		},
	})
}

func TestReconcileWhenSendersAreSplit(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency: "EUR",
		Senders: []Sender{
			{"s1", big.NewInt(20)},
			{"s2", big.NewInt(30)},
		},
		Receivers: []Receiver{{"d", big.NewInt(50)}},
		Expected: []Posting{
			{"s1", "d", big.NewInt(20), "EUR"},
			{"s2", "d", big.NewInt(30), "EUR"},
		},
	})
}

func TestMany(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency: "EUR",
		Senders: []Sender{
			{"s1", big.NewInt(80 + 20)},
			{"s2", big.NewInt(1000)},
		},
		Receivers: []Receiver{
			{"d1", big.NewInt(80)},
			{"d2", big.NewInt(20 + 123)},
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
		Currency: "EUR",
		Senders: []Sender{
			{"s1", big.NewInt(80 + 20)},
			{"s2", big.NewInt(1000)},
		},
		Receivers: []Receiver{
			{"d1", big.NewInt(80)},
			{"d2", big.NewInt(20 + 123)},
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
		Currency: "EUR",
		Senders: []Sender{
			{"src1", big.NewInt(1)},
			{"src2", big.NewInt(10)},
			{"src2", big.NewInt(20)},
		},
		Receivers: []Receiver{{"d", big.NewInt(31)}},
		Expected: []Posting{
			{"src1", "d", big.NewInt(1), "EUR"},
			{"src2", "d", big.NewInt(30), "EUR"},
		},
	})
}

func TestReconcileKept(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency: "GEM",
		Senders: []Sender{
			{"src", big.NewInt(100)},
		},
		Receivers: []Receiver{
			{"dest", big.NewInt(50)},
			{"<kept>", big.NewInt(50)}},
		Expected: []Posting{
			{"src", "dest", big.NewInt(50), "GEM"},
		},
	})
}

func TestReconcileEmptyMonetaryForDest(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency: "GEM",
		Senders: []Sender{
			{"src", big.NewInt(100)},
		},
		Receivers: []Receiver{
			{"dest", nil},
		},
		Expected: []Posting{
			{"src", "dest", big.NewInt(100), "GEM"},
		},
	})
}

func TestReconcileSendAllMixed(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency: "GEM",
		Senders: []Sender{
			{"src", big.NewInt(100)},
		},
		Receivers: []Receiver{
			{"d1", big.NewInt(20)},
			{"d2", nil},
		},
		Expected: []Posting{
			{"src", "d1", big.NewInt(20), "GEM"},
			{"src", "d2", big.NewInt(80), "GEM"},
		},
	})
}

func TestReconcileSendMultiSrc(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency: "GEM",
		Senders: []Sender{
			{"src1", big.NewInt(10)},
			{"src2", big.NewInt(20)},
		},
		Receivers: []Receiver{
			{"dest", nil},
		},
		Expected: []Posting{
			{"src1", "dest", big.NewInt(10), "GEM"},
			{"src2", "dest", big.NewInt(20), "GEM"},
		},
	})
}
