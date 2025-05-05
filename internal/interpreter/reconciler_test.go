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
	t.Parallel()
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
		Senders:   []Sender{{"src", big.NewInt(10), nil}},
		Receivers: []Receiver{{"dest", big.NewInt(10)}},
		Expected:  []Posting{{"src", "dest", big.NewInt(10), "COIN"}},
	})
}

func TestReconcileZero(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency:  "COIN",
		Senders:   []Sender{{"src", big.NewInt(0), nil}},
		Receivers: []Receiver{{"dest", big.NewInt(0)}},
		Expected: []Posting{
			{"src", "dest", big.NewInt(0), "COIN"},
		},
		ExpectedErr: nil,
	})
}

func TestNoReceiversLeft(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Senders: []Sender{{
			"src",
			big.NewInt(10),
			nil,
		}},
	})
}

func TestReconcileSendersRemainder(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency: "EUR",
		Senders:  []Sender{{"src", big.NewInt(100), nil}},
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
			{"s1", big.NewInt(20), nil},
			{"s2", big.NewInt(30), nil},
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
			{"s1", big.NewInt(80 + 20), nil},
			{"s2", big.NewInt(1000), nil},
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
			{"s1", big.NewInt(80 + 20), nil},
			{"s2", big.NewInt(1000), nil},
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
			{"src1", big.NewInt(1), nil},
			{"src2", big.NewInt(10), nil},
			{"src2", big.NewInt(20), nil},
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
			{"src", big.NewInt(100), nil},
		},
		Receivers: []Receiver{
			{"dest", big.NewInt(50)},
			{"<kept>", big.NewInt(50)}},
		Expected: []Posting{
			{"src", "dest", big.NewInt(50), "GEM"},
		},
	})
}

func TestReconcileColoredAssetExactMatch(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency: "COIN",
		Senders: []Sender{
			{"src", big.NewInt(10), pointer("x")},
		},
		Receivers: []Receiver{{"dest", big.NewInt(10)}},
		Expected:  []Posting{{"src", "dest", big.NewInt(10), "COIN_x"}},
	})
}

func TestReconcileColoredManyDestPerSender(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency: "COIN",
		Senders: []Sender{
			{"src", big.NewInt(10), pointer("x")},
		},
		Receivers: []Receiver{
			{"d1", big.NewInt(5)},
			{"d2", big.NewInt(5)},
		},
		Expected: []Posting{
			{"src", "d1", big.NewInt(5), "COIN_x"},
			{"src", "d2", big.NewInt(5), "COIN_x"},
		},
	})
}

func TestReconcileColoredManySenderColors(t *testing.T) {
	runReconcileTestCase(t, ReconcileTestCase{
		Currency: "COIN",
		Senders: []Sender{
			{"src", big.NewInt(1), pointer("c1")},
			{"src", big.NewInt(1), pointer("c2")},
		},
		Receivers: []Receiver{
			{"dest", big.NewInt(2)},
		},
		Expected: []Posting{
			{"src", "dest", big.NewInt(1), "COIN_c1"},
			{"src", "dest", big.NewInt(1), "COIN_c2"},
		},
	})
}

func pointer[T any](x T) *T {
	return &x
}
