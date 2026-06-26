package runtime_test

import (
	"reflect"
	"testing"

	"github.com/formancehq/numscript/internal/runtime"
)

// --- test helpers ---------------------------------------------------------

// mockStore is a Store that returns preset balances and counts how many times
// each (account, asset) pair is fetched, so tests can assert lazy/cached reads.
type mockStore struct {
	balances map[runtime.PairKey]int64
	calls    map[runtime.PairKey]int
}

func newMockStore(initial map[runtime.PairKey]int64) *mockStore {
	b := make(map[runtime.PairKey]int64, len(initial))
	for k, v := range initial {
		b[k] = v
	}
	return &mockStore{balances: b, calls: make(map[runtime.PairKey]int)}
}

func (m *mockStore) GetBalance(account, asset string) int64 {
	k := runtime.PairKey{account, asset}
	m.calls[k]++
	return m.balances[k] // 0 if absent
}

func (m *mockStore) callCount(account, asset string) int {
	return m.calls[runtime.PairKey{account, asset}]
}

const usd = "USD"

func newRS(initial map[runtime.PairKey]int64) (*runtime.RunState, *mockStore) {
	store := newMockStore(initial)
	rs := runtime.New(store)
	rs.SetCurrentAsset(usd)
	return rs, store
}

func strptr(s string) *string { return &s }

func wantBalance(t *testing.T, rs *runtime.RunState, account string, want int64) {
	t.Helper()
	if got := rs.GetAccountBalance(account, usd); got != want {
		t.Errorf("balance(%s) = %d, want %d", account, got, want)
	}
}

func wantReturn(t *testing.T, label string, got, want int64) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %d, want %d", label, got, want)
	}
}

func wantPostings(t *testing.T, rs *runtime.RunState, want []runtime.Posting) {
	t.Helper()
	got := rs.GetPostings()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("postings mismatch\n got: %+v\nwant: %+v", got, want)
	}
}

// --- GetAccountBalance / caching -----------------------------------------

func TestGetAccountBalance_FetchesFromStore(t *testing.T) {
	rs, store := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	wantBalance(t, rs, "A", 100)
	if store.callCount("A", usd) != 1 {
		t.Errorf("expected 1 store fetch, got %d", store.callCount("A", usd))
	}
}

func TestGetAccountBalance_EmptyAssetUsesCurrent(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 42})
	if got := rs.GetAccountBalance("A", ""); got != 42 {
		t.Errorf("got %d, want 42 (empty asset should resolve to currentAsset)", got)
	}
}

func TestGetAccountBalance_MissingIsZeroAndCached(t *testing.T) {
	rs, store := newRS(nil)
	if got := rs.GetAccountBalance("ghost", usd); got != 0 {
		t.Errorf("missing account = %d, want 0", got)
	}
	// second read must not re-hit the store even though value is 0
	_ = rs.GetAccountBalance("ghost", usd)
	if c := store.callCount("ghost", usd); c != 1 {
		t.Errorf("zero balance not cached: store called %d times, want 1", c)
	}
}

func TestCaching_FetchedOnlyOnce(t *testing.T) {
	rs, store := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	for i := 0; i < 5; i++ {
		rs.GetAccountBalance("A", usd)
	}
	if c := store.callCount("A", usd); c != 1 {
		t.Errorf("store called %d times, want 1", c)
	}
}

func TestCaching_WriteThroughCompounds(t *testing.T) {
	// Pull decreases the balance; the next read must see the decreased value
	// without consulting the store again.
	rs, store := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	rs.Pull("A", 30, runtime.BoundedOverdraft(0)) // A -> 70
	wantBalance(t, rs, "A", 70)
	rs.Pull("A", 20, runtime.BoundedOverdraft(0)) // A -> 50
	wantBalance(t, rs, "A", 50)
	if c := store.callCount("A", usd); c != 1 {
		t.Errorf("store consulted %d times across pulls, want 1", c)
	}
}

// --- Pull (bounded) -------------------------------------------------------

func TestPull_BoundedClampedByBalance(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	got := rs.Pull("A", 200, runtime.BoundedOverdraft(0)) // min(max(0,100+0),200)=100
	wantReturn(t, "Pull", got, 100)
	wantBalance(t, rs, "A", 0)
}

func TestPull_BoundedClampedByCap(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	got := rs.Pull("A", 30, runtime.BoundedOverdraft(0)) // min(100,30)=30
	wantReturn(t, "Pull", got, 30)
	wantBalance(t, rs, "A", 70)
}

func TestPull_BoundedWithOverdraftBound(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	// eff = max(0, 100+50) = 150 ; available = min(150, 200) = 150
	got := rs.Pull("A", 200, runtime.BoundedOverdraft(50))
	wantReturn(t, "Pull", got, 150)
	wantBalance(t, rs, "A", -50) // overdraft used
}

func TestPull_NegativeCapClampedToZero(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	got := rs.Pull("A", -5, runtime.BoundedOverdraft(0))
	wantReturn(t, "Pull", got, 0)
	wantBalance(t, rs, "A", 100)
}

func TestPull_NegativeOverdraftBoundClampedToZero(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	// bound clamped to 0 -> eff = 100 -> available = min(100, 200) = 100
	got := rs.Pull("A", 200, runtime.BoundedOverdraft(-1000))
	wantReturn(t, "Pull", got, 100)
	wantBalance(t, rs, "A", 0)
}

func TestPull_NegativeStoreBalanceBounded(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: -20})
	// eff = max(0, -20+0) = 0 -> available = min(0, cap) = 0
	got := rs.Pull("A", 50, runtime.BoundedOverdraft(0))
	wantReturn(t, "Pull", got, 0)
	wantBalance(t, rs, "A", -20)
}

// --- Pull (unbounded) -----------------------------------------------------

func TestPull_UnboundedTakesFullCap(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 30})
	got := rs.Pull("A", 100, runtime.UnboundedOverdraft())
	wantReturn(t, "Pull", got, 100)
	wantBalance(t, rs, "A", -70) // balance can go negative
}

func TestPull_UnboundedNegativeCapClampedToZero(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 30})
	got := rs.Pull("A", -10, runtime.UnboundedOverdraft())
	wantReturn(t, "Pull", got, 0)
	wantBalance(t, rs, "A", 30)
}

// --- PullUncapped ---------------------------------------------------------

func TestPullUncapped_Basic(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	got := rs.PullUncapped("A", 0)
	wantReturn(t, "PullUncapped", got, 100)
	wantBalance(t, rs, "A", 0)
}

func TestPullUncapped_WithOverdraft(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	got := rs.PullUncapped("A", 50)
	wantReturn(t, "PullUncapped", got, 150)
	wantBalance(t, rs, "A", -50)
}

func TestPullUncapped_ZeroNotQueuedNorDebited(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 0})
	got := rs.PullUncapped("A", 0)
	wantReturn(t, "PullUncapped", got, 0)
	wantBalance(t, rs, "A", 0)
	// nothing queued -> a subsequent drain produces no postings
	rs.SendUncapped(strptr("X"))
	wantPostings(t, rs, []runtime.Posting{})
}

func TestPullUncapped_NegativeEffectiveNotQueued(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 10})
	got := rs.PullUncapped("A", -50) // max(0, 10-50) = 0
	wantReturn(t, "PullUncapped", got, 0)
	wantBalance(t, rs, "A", 10)
	rs.SendUncapped(strptr("X"))
	wantPostings(t, rs, []runtime.Posting{})
}

// --- Send: FIFO, partial requeue, posting creation -----------------------

func TestSend_PartialConsumeRequeuesFront(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100, {"B", usd}: 50})
	rs.Pull("A", 100, runtime.BoundedOverdraft(0)) // source A:100
	rs.Pull("B", 50, runtime.BoundedOverdraft(0))  // source B:50

	rs.Send(strptr("X"), 30) // takes 30 from A, requeues A:70 at front
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: 30}})

	rs.Send(strptr("Y"), 200) // A:70 then B:50, both fully
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: 30},
		{Source: "A", Destination: "Y", Asset: usd, Amount: 70},
		{Source: "B", Destination: "Y", Asset: usd, Amount: 50},
	})
}

func TestSend_FIFOOrder(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 10, {"B", usd}: 10, {"C", usd}: 10})
	rs.Pull("A", 10, runtime.BoundedOverdraft(0))
	rs.Pull("B", 10, runtime.BoundedOverdraft(0))
	rs.Pull("C", 10, runtime.BoundedOverdraft(0))
	rs.Send(strptr("X"), 30)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: 10},
		{Source: "B", Destination: "X", Asset: usd, Amount: 10},
		{Source: "C", Destination: "X", Asset: usd, Amount: 10},
	})
}

func TestSend_ExactMatchNoRequeue(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 50})
	rs.Pull("A", 50, runtime.BoundedOverdraft(0))
	rs.Send(strptr("X"), 50) // exact
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: 50}})
	// nothing left
	rs.SendUncapped(strptr("Y"))
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: 50}})
}

func TestSend_CapExceedsAvailableDrains(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	rs.Pull("A", 100, runtime.BoundedOverdraft(0))
	rs.Send(strptr("X"), 500) // more than available; drains 100, no leftover
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: 100}})
}

func TestSend_ZeroCapIsNoOp(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	rs.Pull("A", 100, runtime.BoundedOverdraft(0))
	rs.Send(strptr("X"), 0)
	wantPostings(t, rs, []runtime.Posting{})
	// source remains -> uncapped drain still sees it
	rs.SendUncapped(strptr("X"))
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: 100}})
}

func TestSend_NegativeCapIsNoOp(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	rs.Pull("A", 100, runtime.BoundedOverdraft(0))
	rs.Send(strptr("X"), -5)
	wantPostings(t, rs, []runtime.Posting{})
}

func TestSend_NoSourcesIsNoOp(t *testing.T) {
	rs, _ := newRS(nil)
	rs.Send(strptr("X"), 100)
	wantPostings(t, rs, []runtime.Posting{})
}

// --- Send: posting merge --------------------------------------------------

func TestSend_MergesConsecutiveSameTriple(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	rs.Pull("A", 100, runtime.BoundedOverdraft(0))
	rs.Send(strptr("X"), 40) // posting A->X 40, requeue A:60
	rs.Send(strptr("X"), 40) // merges into A->X 80, requeue A:20
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: 80}})
}

func TestSend_DoesNotMergeDifferentDestination(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	rs.Pull("A", 100, runtime.BoundedOverdraft(0))
	rs.Send(strptr("X"), 40)
	rs.Send(strptr("Y"), 40)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: 40},
		{Source: "A", Destination: "Y", Asset: usd, Amount: 40},
	})
}

// --- Send: destination balance credit (the cache-bug fix) -----------------

func TestSend_CreditsDestinationOverExistingStoreBalance(t *testing.T) {
	// X already has 500 in the store. Crediting must fetch that first, not
	// treat X as 0.
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100, {"X", usd}: 500})
	rs.Pull("A", 100, runtime.BoundedOverdraft(0))
	rs.Send(strptr("X"), 100)
	wantBalance(t, rs, "X", 600)
}

// --- Send: refund path (dest == nil) -------------------------------------

func TestSend_RefundCreditsSourceNoPosting(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	rs.Pull("A", 100, runtime.BoundedOverdraft(0)) // A -> 0, source A:100
	rs.Send(nil, 60)                               // refund 60 to A, requeue A:40
	wantBalance(t, rs, "A", 60)
	wantPostings(t, rs, []runtime.Posting{})
	// remaining 40 still queued
	rs.Send(strptr("X"), 100)
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: 40}})
}

// --- SendUncapped ---------------------------------------------------------

func TestSendUncapped_DrainsAllToDestination(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100, {"B", usd}: 50})
	rs.Pull("A", 100, runtime.BoundedOverdraft(0))
	rs.Pull("B", 50, runtime.BoundedOverdraft(0))
	rs.SendUncapped(strptr("X"))
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: 100},
		{Source: "B", Destination: "X", Asset: usd, Amount: 50},
	})
}

func TestSendUncapped_RefundsAll(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100, {"B", usd}: 50})
	rs.Pull("A", 100, runtime.BoundedOverdraft(0)) // A -> 0
	rs.Pull("B", 50, runtime.BoundedOverdraft(0))  // B -> 0
	rs.SendUncapped(nil)                           // refund both
	wantBalance(t, rs, "A", 100)
	wantBalance(t, rs, "B", 50)
	wantPostings(t, rs, []runtime.Posting{})
}

func TestSendUncapped_NoSourcesIsNoOp(t *testing.T) {
	rs, _ := newRS(nil)
	rs.SendUncapped(strptr("X"))
	wantPostings(t, rs, []runtime.Posting{})
}

// --- GetPostings returns a defensive copy --------------------------------

func TestGetPostings_ReturnsCopy(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", usd}: 100})
	rs.Pull("A", 100, runtime.BoundedOverdraft(0))
	rs.Send(strptr("X"), 100)

	p := rs.GetPostings()
	if len(p) != 1 {
		t.Fatalf("expected 1 posting, got %d", len(p))
	}
	p[0].Amount = 999999 // mutate the returned slice

	p2 := rs.GetPostings()
	if p2[0].Amount != 100 {
		t.Errorf("internal posting was mutated via returned slice: amount=%d", p2[0].Amount)
	}
}

// --- end-to-end flow ------------------------------------------------------

func TestEndToEnd_TwoSourcesSplitAcrossDestinations(t *testing.T) {
	rs, store := newRS(map[runtime.PairKey]int64{
		{"alice", usd}: 100,
		{"bob", usd}:   100,
		{"carol", usd}: 0,
		{"dave", usd}:  0,
	})
	rs.Pull("alice", 100, runtime.BoundedOverdraft(0))
	rs.Pull("bob", 100, runtime.BoundedOverdraft(0))

	rs.Send(strptr("carol"), 150) // alice:100 fully, bob:50 partial (requeue bob:50)
	rs.Send(strptr("dave"), 50)   // bob:50 fully

	wantPostings(t, rs, []runtime.Posting{
		{Source: "alice", Destination: "carol", Asset: usd, Amount: 100},
		{Source: "bob", Destination: "carol", Asset: usd, Amount: 50},
		{Source: "bob", Destination: "dave", Asset: usd, Amount: 50},
	})
	wantBalance(t, rs, "alice", 0)
	wantBalance(t, rs, "bob", 0)
	wantBalance(t, rs, "carol", 150)
	wantBalance(t, rs, "dave", 50)

	// each account fetched from store exactly once
	for _, acct := range []string{"alice", "bob", "carol", "dave"} {
		if c := store.callCount(acct, usd); c != 1 {
			t.Errorf("%s fetched %d times, want 1", acct, c)
		}
	}
}
