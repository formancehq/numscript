package runtime_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/formancehq/numscript/internal/runtime"
)

// --- test helpers ---------------------------------------------------------

// mockStore is a Store that returns preset balances and counts how many times
// each (account, asset, color) triple is fetched, so tests can assert
// lazy/cached reads.
type mockStore struct {
	balances map[runtime.PairKey]*big.Int
	calls    map[runtime.PairKey]int
}

func newMockStore(initial map[runtime.PairKey]int64) *mockStore {
	b := make(map[runtime.PairKey]*big.Int, len(initial))
	for k, v := range initial {
		b[k] = big.NewInt(v)
	}
	return &mockStore{balances: b, calls: make(map[runtime.PairKey]int)}
}

func (m *mockStore) GetBalance(account, scope, asset, color string) (*big.Int, error) {
	k := runtime.PairKey{account, scope, asset, color}
	m.calls[k]++
	if v, ok := m.balances[k]; ok {
		return v, nil
	}
	return new(big.Int), nil // 0 if absent
}

func (m *mockStore) callCount(account, asset string) int {
	return m.calls[runtime.PairKey{account, "", asset, ""}]
}

const usd = "USD"

func newRS(initial map[runtime.PairKey]int64) (*runtime.RunState, *mockStore) {
	store := newMockStore(initial)
	rs := runtime.New(store)
	rs.SetCurrentAsset(usd)
	return rs, store
}

func strptr(s string) *string { return &s }

// accBal reads a balance; the mock store never errors, so a failure is fatal.
func accBal(rs *runtime.RunState, account, scope, asset, color string) *big.Int {
	b, err := rs.GetAccountBalance(account, scope, asset, color)
	if err != nil {
		panic(err)
	}
	return b
}

// pull adapts the out-param Pull to a value-returning form for test ergonomics.
func pull(rs *runtime.RunState, src string, cap, overdraft *big.Int, color string) *big.Int {
	out := new(big.Int)
	_ = rs.Pull(out, src, "", cap, overdraft, color)
	return out
}

// pullUncapped adapts the out-param PullUncapped to a value-returning form.
func pullUncapped(rs *runtime.RunState, src string, overdraftBound *big.Int, color string) *big.Int {
	out := new(big.Int)
	_ = rs.PullUncapped(out, src, "", overdraftBound, color)
	return out
}

func wantBalance(t *testing.T, rs *runtime.RunState, account string, want int64) {
	t.Helper()
	if got := accBal(rs, account, "", usd, ""); got.Cmp(big.NewInt(want)) != 0 {
		t.Errorf("balance(%s) = %s, want %d", account, got, want)
	}
}

func wantReturn(t *testing.T, label string, got *big.Int, want int64) {
	t.Helper()
	if got.Cmp(big.NewInt(want)) != 0 {
		t.Errorf("%s = %s, want %d", label, got, want)
	}
}

func wantPostings(t *testing.T, rs *runtime.RunState, want []runtime.Posting) {
	t.Helper()
	got := rs.GetPostings()
	mismatch := len(got) != len(want)
	for i := 0; !mismatch && i < len(got); i++ {
		g, w := got[i], want[i]
		if g.Source != w.Source || g.Destination != w.Destination ||
			g.Asset != w.Asset || g.Color != w.Color || g.Amount.Cmp(w.Amount) != 0 {
			mismatch = true
		}
	}
	if mismatch {
		t.Errorf("postings mismatch\n got: %s\nwant: %s", fmtPostings(got), fmtPostings(want))
	}
}

func fmtPostings(ps []runtime.Posting) string {
	out := "["
	for _, p := range ps {
		out += "{" + p.Source + "->" + p.Destination + " " + p.Amount.String() + " " + p.Asset
		if p.Color != "" {
			out += " " + p.Color
		}
		out += "}"
	}
	return out + "]"
}

// --- GetAccountBalance / caching -----------------------------------------

func TestGetAccountBalance_FetchesFromStore(t *testing.T) {
	rs, store := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	wantBalance(t, rs, "A", 100)
	if store.callCount("A", usd) != 1 {
		t.Errorf("expected 1 store fetch, got %d", store.callCount("A", usd))
	}
}

func TestGetAccountBalance_EmptyAssetUsesCurrent(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 42})
	if got := accBal(rs, "A", "", "", ""); got.Cmp(big.NewInt(42)) != 0 {
		t.Errorf("got %d, want 42 (empty asset should resolve to currentAsset)", got)
	}
}

func TestGetAccountBalance_MissingIsZeroAndCached(t *testing.T) {
	rs, store := newRS(nil)
	if got := accBal(rs, "ghost", "", usd, ""); got.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("missing account = %d, want 0", got)
	}
	// second read must not re-hit the store even though value is 0
	_ = accBal(rs, "ghost", "", usd, "")
	if c := store.callCount("ghost", usd); c != 1 {
		t.Errorf("zero balance not cached: store called %d times, want 1", c)
	}
}

func TestCaching_FetchedOnlyOnce(t *testing.T) {
	rs, store := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	for i := 0; i < 5; i++ {
		accBal(rs, "A", "", usd, "")
	}
	if c := store.callCount("A", usd); c != 1 {
		t.Errorf("store called %d times, want 1", c)
	}
}

func TestCaching_WriteThroughCompounds(t *testing.T) {
	// Pull decreases the balance; the next read must see the decreased value
	// without consulting the store again.
	rs, store := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	pull(rs, "A", big.NewInt(30), big.NewInt(0), "") // A -> 70
	wantBalance(t, rs, "A", 70)
	pull(rs, "A", big.NewInt(20), big.NewInt(0), "") // A -> 50
	wantBalance(t, rs, "A", 50)
	if c := store.callCount("A", usd); c != 1 {
		t.Errorf("store consulted %d times across pulls, want 1", c)
	}
}

// --- Pull (bounded) -------------------------------------------------------

func TestPull_BoundedClampedByBalance(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	got := pull(rs, "A", big.NewInt(200), big.NewInt(0), "") // min(max(0,100+0),200)=100
	wantReturn(t, "Pull", got, 100)
	wantBalance(t, rs, "A", 0)
}

func TestPull_BoundedClampedByCap(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	got := pull(rs, "A", big.NewInt(30), big.NewInt(0), "") // min(100,30)=30
	wantReturn(t, "Pull", got, 30)
	wantBalance(t, rs, "A", 70)
}

func TestPull_BoundedWithOverdraftBound(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	// eff = max(0, 100+50) = 150 ; available = min(150, 200) = 150
	got := pull(rs, "A", big.NewInt(200), big.NewInt(50), "")
	wantReturn(t, "Pull", got, 150)
	wantBalance(t, rs, "A", -50) // overdraft used
}

func TestPull_NegativeCapClampedToZero(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	got := pull(rs, "A", big.NewInt(-5), big.NewInt(0), "")
	wantReturn(t, "Pull", got, 0)
	wantBalance(t, rs, "A", 100)
}

func TestPull_NegativeOverdraftBoundClampedToZero(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	// bound clamped to 0 -> eff = 100 -> available = min(100, 200) = 100
	got := pull(rs, "A", big.NewInt(200), big.NewInt(-1000), "")
	wantReturn(t, "Pull", got, 100)
	wantBalance(t, rs, "A", 0)
}

func TestPull_NegativeStoreBalanceBounded(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: -20})
	// eff = max(0, -20+0) = 0 -> available = min(0, cap) = 0
	got := pull(rs, "A", big.NewInt(50), big.NewInt(0), "")
	wantReturn(t, "Pull", got, 0)
	wantBalance(t, rs, "A", -20)
}

func TestPull_WritesIntoOutAndDoesNotAliasQueue(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	out := new(big.Int)
	_ = rs.Pull(out, "A", "", big.NewInt(60), big.NewInt(0), "")
	if out.Cmp(big.NewInt(60)) != 0 {
		t.Fatalf("out written = %s, want 60", out)
	}
	// Mutating out afterwards must not corrupt the queued source (it's a copy).
	out.SetInt64(999)
	_ = rs.Send(strptr("X"), "", big.NewInt(60), nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(60)},
	})
}

func TestPull_OutCanBeReused(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100, {"B", "", usd, ""}: 100})
	out := new(big.Int)
	_ = rs.Pull(out, "A", "", big.NewInt(30), big.NewInt(0), "")
	if out.Cmp(big.NewInt(30)) != 0 {
		t.Fatalf("first = %s, want 30", out)
	}
	_ = rs.Pull(out, "B", "", big.NewInt(45), big.NewInt(0), "") // same buffer
	if out.Cmp(big.NewInt(45)) != 0 {
		t.Fatalf("second = %s, want 45", out)
	}
	// both pulls landed in the queue independently
	_ = rs.SendUncapped(strptr("X"), "", nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(30)},
		{Source: "B", Destination: "X", Asset: usd, Amount: big.NewInt(45)},
	})
}

func TestPull_DoesNotMutateCapOrOverdraft(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 10})
	cap := big.NewInt(200)
	ovd := big.NewInt(50)
	out := new(big.Int)
	_ = rs.Pull(out, "A", "", cap, ovd, "") // eff = 10+50 = 60 < 200 -> available 60
	if out.Cmp(big.NewInt(60)) != 0 {
		t.Errorf("available = %s, want 60", out)
	}
	if cap.Cmp(big.NewInt(200)) != 0 {
		t.Errorf("cap mutated: %s", cap)
	}
	if ovd.Cmp(big.NewInt(50)) != 0 {
		t.Errorf("overdraft mutated: %s", ovd)
	}
}

// --- Pull (unbounded) -----------------------------------------------------

func TestPull_UnboundedTakesFullCap(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 30})
	got := pull(rs, "A", big.NewInt(100), nil, "")
	wantReturn(t, "Pull", got, 100)
	wantBalance(t, rs, "A", -70) // balance can go negative
}

func TestPull_UnboundedNegativeCapClampedToZero(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 30})
	got := pull(rs, "A", big.NewInt(-10), nil, "")
	wantReturn(t, "Pull", got, 0)
	wantBalance(t, rs, "A", 30)
}

// --- PullUncapped ---------------------------------------------------------

func TestPullUncapped_Basic(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	got := pullUncapped(rs, "A", big.NewInt(0), "")
	wantReturn(t, "PullUncapped", got, 100)
	wantBalance(t, rs, "A", 0)
}

func TestPullUncapped_WithOverdraft(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	got := pullUncapped(rs, "A", big.NewInt(50), "")
	wantReturn(t, "PullUncapped", got, 150)
	wantBalance(t, rs, "A", -50)
}

func TestPullUncapped_WritesIntoOutAndDoesNotAliasQueue(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	out := new(big.Int)
	_ = rs.PullUncapped(out, "A", "", big.NewInt(0), "")
	if out.Cmp(big.NewInt(100)) != 0 {
		t.Fatalf("out = %s, want 100", out)
	}
	out.SetInt64(999) // mutate after: queued source must be an independent copy
	_ = rs.SendUncapped(strptr("X"), "", nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(100)},
	})
}

func TestPullUncapped_ZeroNotQueuedNorDebited(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 0})
	got := pullUncapped(rs, "A", big.NewInt(0), "")
	wantReturn(t, "PullUncapped", got, 0)
	wantBalance(t, rs, "A", 0)
	// nothing queued -> a subsequent drain produces no postings
	_ = rs.SendUncapped(strptr("X"), "", nil)
	wantPostings(t, rs, []runtime.Posting{})
}

func TestPullUncapped_NegativeOverdraftBoundClamped(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 10})
	// bound clamped to 0 -> effective = max(0, 10+0) = 10
	got := pullUncapped(rs, "A", big.NewInt(-50), "")
	wantReturn(t, "PullUncapped", got, 10)
	wantBalance(t, rs, "A", 0)
}

func TestPullUncapped_NegativeEffectiveNotQueued(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: -5})
	got := pullUncapped(rs, "A", big.NewInt(0), "") // max(0, -5+0) = 0
	wantReturn(t, "PullUncapped", got, 0)
	wantBalance(t, rs, "A", -5)
	_ = rs.SendUncapped(strptr("X"), "", nil)
	wantPostings(t, rs, []runtime.Posting{})
}

// --- Send: FIFO, partial requeue, posting creation -----------------------

func TestSend_PartialConsumeRequeuesFront(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100, {"B", "", usd, ""}: 50})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "") // source A:100
	pull(rs, "B", big.NewInt(50), big.NewInt(0), "")  // source B:50

	_ = rs.Send(strptr("X"), "", big.NewInt(30), nil) // takes 30 from A, requeues A:70 at front
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(30)}})

	_ = rs.Send(strptr("Y"), "", big.NewInt(200), nil) // A:70 then B:50, both fully
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(30)},
		{Source: "A", Destination: "Y", Asset: usd, Amount: big.NewInt(70)},
		{Source: "B", Destination: "Y", Asset: usd, Amount: big.NewInt(50)},
	})
}

func TestSend_FIFOOrder(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 10, {"B", "", usd, ""}: 10, {"C", "", usd, ""}: 10})
	pull(rs, "A", big.NewInt(10), big.NewInt(0), "")
	pull(rs, "B", big.NewInt(10), big.NewInt(0), "")
	pull(rs, "C", big.NewInt(10), big.NewInt(0), "")
	_ = rs.Send(strptr("X"), "", big.NewInt(30), nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(10)},
		{Source: "B", Destination: "X", Asset: usd, Amount: big.NewInt(10)},
		{Source: "C", Destination: "X", Asset: usd, Amount: big.NewInt(10)},
	})
}

func TestSend_ExactMatchNoRequeue(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 50})
	pull(rs, "A", big.NewInt(50), big.NewInt(0), "")
	_ = rs.Send(strptr("X"), "", big.NewInt(50), nil) // exact
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(50)}})
	// nothing left
	_ = rs.SendUncapped(strptr("Y"), "", nil)
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(50)}})
}

func TestSend_CapExceedsAvailableDrains(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "")
	_ = rs.Send(strptr("X"), "", big.NewInt(500), nil) // more than available; drains 100, no leftover
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(100)}})
}

func TestSend_ZeroCapIsNoOp(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "")
	_ = rs.Send(strptr("X"), "", big.NewInt(0), nil)
	wantPostings(t, rs, []runtime.Posting{})
	// source remains -> uncapped drain still sees it
	_ = rs.SendUncapped(strptr("X"), "", nil)
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(100)}})
}

func TestSend_NegativeCapIsNoOp(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "")
	_ = rs.Send(strptr("X"), "", big.NewInt(-5), nil)
	wantPostings(t, rs, []runtime.Posting{})
}

func TestSend_NoSourcesIsNoOp(t *testing.T) {
	rs, _ := newRS(nil)
	_ = rs.Send(strptr("X"), "", big.NewInt(100), nil)
	wantPostings(t, rs, []runtime.Posting{})
}

// --- Send: posting merge --------------------------------------------------

func TestSend_MergesWithinSingleDrain(t *testing.T) {
	// Two same-source funds drained by ONE Send to the same destination merge
	// into a single posting (mirrors fundsQueue.compactTop within one Pull).
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	pull(rs, "A", big.NewInt(60), big.NewInt(0), "")   // source A:60
	pull(rs, "A", big.NewInt(40), big.NewInt(0), "")   // source A:40
	_ = rs.Send(strptr("X"), "", big.NewInt(100), nil) // drains both A:60 then A:40 -> one posting
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(100)}})
}

func TestSend_DoesNotMergeAcrossSeparateSends(t *testing.T) {
	// Two separate Send calls, same src->dst->asset, are NOT merged. This
	// matches the interpreter (fundsQueue), which only merges adjacent funds
	// within a single Pull, never across send statements.
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "")
	_ = rs.Send(strptr("X"), "", big.NewInt(40), nil) // posting A->X 40, requeue A:60
	_ = rs.Send(strptr("X"), "", big.NewInt(40), nil) // separate send: NOT merged, requeue A:20
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(40)},
		{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(40)},
	})
}

func TestSend_DoesNotMergeDifferentDestination(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "")
	_ = rs.Send(strptr("X"), "", big.NewInt(40), nil)
	_ = rs.Send(strptr("Y"), "", big.NewInt(40), nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(40)},
		{Source: "A", Destination: "Y", Asset: usd, Amount: big.NewInt(40)},
	})
}

// --- Send: destination balance credit (the cache-bug fix) -----------------

func TestSend_CreditsDestinationOverExistingStoreBalance(t *testing.T) {
	// X already has 500 in the store. Crediting must fetch that first, not
	// treat X as 0.
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100, {"X", "", usd, ""}: 500})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "")
	_ = rs.Send(strptr("X"), "", big.NewInt(100), nil)
	wantBalance(t, rs, "X", 600)
}

// --- Send: refund path (dest == nil) -------------------------------------

func TestSend_RefundCreditsSourceNoPosting(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "") // A -> 0, source A:100
	_ = rs.Send(nil, "", big.NewInt(60), nil)         // refund 60 to A, requeue A:40
	wantBalance(t, rs, "A", 60)
	wantPostings(t, rs, []runtime.Posting{})
	// remaining 40 still queued
	_ = rs.Send(strptr("X"), "", big.NewInt(100), nil)
	wantPostings(t, rs, []runtime.Posting{{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(40)}})
}

// --- SendUncapped ---------------------------------------------------------

func TestSendUncapped_DrainsAllToDestination(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100, {"B", "", usd, ""}: 50})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "")
	pull(rs, "B", big.NewInt(50), big.NewInt(0), "")
	_ = rs.SendUncapped(strptr("X"), "", nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(100)},
		{Source: "B", Destination: "X", Asset: usd, Amount: big.NewInt(50)},
	})
}

func TestSendUncapped_RefundsAll(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100, {"B", "", usd, ""}: 50})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "") // A -> 0
	pull(rs, "B", big.NewInt(50), big.NewInt(0), "")  // B -> 0
	_ = rs.SendUncapped(nil, "", nil)                 // refund both
	wantBalance(t, rs, "A", 100)
	wantBalance(t, rs, "B", 50)
	wantPostings(t, rs, []runtime.Posting{})
}

func TestSendUncapped_NoSourcesIsNoOp(t *testing.T) {
	rs, _ := newRS(nil)
	_ = rs.SendUncapped(strptr("X"), "", nil)
	wantPostings(t, rs, []runtime.Posting{})
}

// --- GetPostings returns a defensive copy --------------------------------

func TestGetPostings_ReturnsCopy(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "")
	_ = rs.Send(strptr("X"), "", big.NewInt(100), nil)

	p := rs.GetPostings()
	if len(p) != 1 {
		t.Fatalf("expected 1 posting, got %d", len(p))
	}
	p[0].Amount = big.NewInt(999999) // mutate the returned slice

	p2 := rs.GetPostings()
	if p2[0].Amount.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("internal posting was mutated via returned slice: amount=%d", p2[0].Amount)
	}
}

// --- big.Int precision (beyond int64) ------------------------------------

func TestBigInt_AmountsBeyondInt64(t *testing.T) {
	// 10^30 is far beyond int64's ~9.2*10^18 ceiling; the whole pipeline
	// (store -> Pull -> Send -> posting + balances) must carry it losslessly.
	huge, _ := new(big.Int).SetString("1000000000000000000000000000000", 10) // 1e30
	store := newMockStore(nil)
	store.balances[runtime.PairKey{"A", "", usd, ""}] = new(big.Int).Set(huge)
	rs := runtime.New(store)
	rs.SetCurrentAsset(usd)

	got := pull(rs, "A", new(big.Int).Set(huge), big.NewInt(0), "")
	if got.Cmp(huge) != 0 {
		t.Fatalf("Pull returned %s, want %s", got, huge)
	}
	_ = rs.Send(strptr("X"), "", new(big.Int).Set(huge), nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: new(big.Int).Set(huge)},
	})
	if bal := accBal(rs, "X", "", usd, ""); bal.Cmp(huge) != 0 {
		t.Errorf("X balance = %s, want %s", bal, huge)
	}
	if bal := accBal(rs, "A", "", usd, ""); bal.Sign() != 0 {
		t.Errorf("A balance = %s, want 0", bal)
	}
}

func TestBigInt_GetAccountBalanceReturnsCopy(t *testing.T) {
	// Mutating the returned balance must not corrupt the cache.
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	b := accBal(rs, "A", "", usd, "")
	b.SetInt64(999999)
	wantBalance(t, rs, "A", 100)
}

// --- Prewarm (batched balance seeding) -----------------------------------

func TestPrewarm_SeedsCacheAndSkipsStore(t *testing.T) {
	rs, store := newRS(nil) // store has nothing
	rs.Prewarm(map[runtime.PairKey]*big.Int{
		{"A", "", usd, ""}:    big.NewInt(100),
		{"B", "", usd, "red"}: big.NewInt(40),
	})
	wantBalance(t, rs, "A", 100)
	if b := accBal(rs, "B", "", usd, "red"); b.Cmp(big.NewInt(40)) != 0 {
		t.Errorf("B red = %s, want 40", b)
	}
	// nothing was fetched lazily — the batch seed covered it
	if c := store.callCount("A", usd); c != 0 {
		t.Errorf("store consulted %d times for A, want 0", c)
	}
}

func TestPrewarm_ClonesValues(t *testing.T) {
	rs, _ := newRS(nil)
	seed := big.NewInt(100)
	rs.Prewarm(map[runtime.PairKey]*big.Int{{"A", "", usd, ""}: seed})
	seed.SetInt64(999) // mutate caller's value after seeding
	wantBalance(t, rs, "A", 100)
}

func TestPrewarm_DoesNotClobberLiveValue(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	pull(rs, "A", big.NewInt(30), big.NewInt(0), "")                              // A -> 70
	rs.Prewarm(map[runtime.PairKey]*big.Int{{"A", "", usd, ""}: big.NewInt(100)}) // must NOT reset to 100
	wantBalance(t, rs, "A", 70)
}

// --- ForcePosting (direct src->dst, bypassing the queue) -----------------

func TestForcePosting_DebitsSourceCreditsDestAndRecords(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100, {"B", "", usd, ""}: 10})
	_ = rs.ForcePosting("A", "", "B", "", usd, "", big.NewInt(30))
	wantBalance(t, rs, "A", 70)
	wantBalance(t, rs, "B", 40)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "B", Asset: usd, Amount: big.NewInt(30)},
	})
}

func TestForcePosting_UsesExplicitAssetNotCurrent(t *testing.T) {
	// asset-scaling emits postings on a scaled asset, distinct from currentAsset.
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", "USD/2", ""}: 500})
	rs.SetCurrentAsset(usd) // current asset is USD, but we post on USD/2
	_ = rs.ForcePosting("A", "", "B", "", "USD/2", "", big.NewInt(500))
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "B", Asset: "USD/2", Amount: big.NewInt(500)},
	})
	if b := accBal(rs, "A", "", "USD/2", ""); b.Sign() != 0 {
		t.Errorf("A USD/2 = %s, want 0", b)
	}
}

func TestForcePosting_ZeroIsNoOp(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	_ = rs.ForcePosting("A", "", "B", "", usd, "", big.NewInt(0))
	wantBalance(t, rs, "A", 100)
	wantPostings(t, rs, []runtime.Posting{})
}

// --- Save (numscript `save` statement) -----------------------------------

func TestSave_ReducesByAmount(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	_ = rs.Save("A", "", usd, "", big.NewInt(30))
	wantBalance(t, rs, "A", 70)
}

func TestSave_FlooredAtZero(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 20})
	_ = rs.Save("A", "", usd, "", big.NewInt(50)) // would be -30, floored to 0
	wantBalance(t, rs, "A", 0)
}

func TestSave_AllZeroesPositiveBalance(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 80})
	_ = rs.Save("A", "", usd, "", nil) // save all
	wantBalance(t, rs, "A", 0)
}

func TestSave_AllLeavesNegativeUntouched(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: -40})
	_ = rs.Save("A", "", usd, "", nil)
	wantBalance(t, rs, "A", -40)
}

func TestSave_ThenPullSeesProtectedBalance(t *testing.T) {
	// after saving, a bounded Pull can only take what's left
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	_ = rs.Save("A", "", usd, "", big.NewInt(70)) // A -> 30 available
	got := pull(rs, "A", big.NewInt(100), big.NewInt(0), "")
	wantReturn(t, "Pull", got, 30)
	wantBalance(t, rs, "A", 0)
}

// --- marks (cheap oneof backtracking) ------------------------------------

func TestMark_RewindUndoesPullsAndBalances(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100, {"B", "", usd, ""}: 80})

	rs.MarkPush() // at depth 0
	pull(rs, "A", big.NewInt(60), big.NewInt(0), "")
	pull(rs, "B", big.NewInt(50), big.NewInt(0), "")
	// balances debited, two sources queued
	wantBalance(t, rs, "A", 40)
	wantBalance(t, rs, "B", 30)

	require.NoError(t, rs.MarkEnd(true))
	// balances repaid, queue emptied, and the region is closed by the same call
	wantBalance(t, rs, "A", 100)
	wantBalance(t, rs, "B", 80)
	require.False(t, rs.HasOpenMark())
	_ = rs.SendUncapped(strptr("X"), "", nil)
	wantPostings(t, rs, []runtime.Posting{}) // nothing left to send
}

func TestMark_OneofFailedBranchThenRealBranch(t *testing.T) {
	// Models `oneof` exactly as the interpreter and the compiled bytecode emit it:
	// branch 1 falls short, so its region is closed with a rewind and a fresh one is
	// opened for branch 2, which covers the amount and commits.
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 30, {"B", "", usd, ""}: 100})

	rs.MarkPush()

	// branch 1: @A can only provide 30 of the needed 100 -> abandon
	got := pull(rs, "A", big.NewInt(100), big.NewInt(0), "")
	wantReturn(t, "branch1 pull", got, 30) // short
	require.NoError(t, rs.MarkEnd(true))
	wantBalance(t, rs, "A", 30) // A untouched after backtrack

	// close-and-reopen: the fresh mark is identical to the one just rewound
	rs.MarkPush()

	// branch 2: @B covers it
	got = pull(rs, "B", big.NewInt(100), big.NewInt(0), "")
	wantReturn(t, "branch2 pull", got, 100)

	require.NoError(t, rs.MarkEnd(false))
	_ = rs.Send(strptr("X"), "", big.NewInt(100), nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "B", Destination: "X", Asset: usd, Amount: big.NewInt(100)},
	})
	wantBalance(t, rs, "A", 30)
	wantBalance(t, rs, "B", 0)
}

func TestMark_RewindKeepsSourcesQueuedBeforeThePush(t *testing.T) {
	// A mark opened mid-stream must only undo what came after it.
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100, {"B", "", usd, ""}: 100})
	pull(rs, "A", big.NewInt(40), big.NewInt(0), "") // kept

	rs.MarkPush()
	pull(rs, "B", big.NewInt(70), big.NewInt(0), "") // undone
	require.NoError(t, rs.MarkEnd(true))

	wantBalance(t, rs, "A", 60)  // still debited
	wantBalance(t, rs, "B", 100) // repaid
	_ = rs.SendUncapped(strptr("X"), "", nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(40)},
	})
}

func TestMark_CommitKeepsWhatTheRegionPulled(t *testing.T) {
	// The success path: a branch covered the amount, so the region is popped
	// without a rewind and its funds stay queued.
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})

	rs.MarkPush()
	pull(rs, "A", big.NewInt(70), big.NewInt(0), "")
	require.NoError(t, rs.MarkEnd(false))

	require.False(t, rs.HasOpenMark())
	wantBalance(t, rs, "A", 30)
	_ = rs.SendUncapped(strptr("X"), "", nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Amount: big.NewInt(70)},
	})
}

func TestMark_NestedRegionsRewindIndependently(t *testing.T) {
	// The inner rewind must undo only the inner region; the outer one's funds
	// survive it and are only undone by the outer rewind.
	rs, _ := newRS(map[runtime.PairKey]int64{
		{"A", "", usd, ""}: 100, {"B", "", usd, ""}: 100, {"C", "", usd, ""}: 100,
	})

	rs.MarkPush()
	pull(rs, "A", big.NewInt(10), big.NewInt(0), "")

	rs.MarkPush()
	pull(rs, "B", big.NewInt(20), big.NewInt(0), "")
	require.NoError(t, rs.MarkEnd(true))

	wantBalance(t, rs, "A", 90)  // outer pull survives the inner rewind
	wantBalance(t, rs, "B", 100) // inner pull undone
	require.True(t, rs.HasOpenMark(), "outer mark must still be open")

	pull(rs, "C", big.NewInt(30), big.NewInt(0), "")
	require.NoError(t, rs.MarkEnd(true)) // outer rewind: undoes A and C, and closes

	require.False(t, rs.HasOpenMark())
	wantBalance(t, rs, "A", 100)
	wantBalance(t, rs, "C", 100)
	_ = rs.SendUncapped(strptr("X"), "", nil)
	wantPostings(t, rs, []runtime.Posting{})
}

// A caller cannot name a queue depth, so the only way to misuse a mark is to
// end one that was never pushed. Both flags must report it rather than
// panicking on an out-of-range truncation, which is what the old index-valued
// Restore did.
func TestMark_EndWithNoOpenMark(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})

	require.ErrorIs(t, rs.MarkEnd(true), runtime.ErrNoOpenMark)
	require.ErrorIs(t, rs.MarkEnd(false), runtime.ErrNoOpenMark)

	// and after a balanced region has closed
	rs.MarkPush()
	require.NoError(t, rs.MarkEnd(false))
	require.ErrorIs(t, rs.MarkEnd(true), runtime.ErrNoOpenMark)
	require.ErrorIs(t, rs.MarkEnd(false), runtime.ErrNoOpenMark)
}

// HasOpenMark is what lets a caller enforce the precondition Send and
// SetCurrentAsset document: the VM checks it at the two opcodes instead of
// discovering the damage afterwards.
func TestMark_HasOpenMarkReportsAnOpenRegion(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	require.False(t, rs.HasOpenMark())

	rs.MarkPush()
	require.True(t, rs.HasOpenMark())
	rs.MarkPush()
	require.True(t, rs.HasOpenMark())

	require.NoError(t, rs.MarkEnd(false))
	require.True(t, rs.HasOpenMark(), "the outer region is still open")
	require.NoError(t, rs.MarkEnd(false))
	require.False(t, rs.HasOpenMark())
}

// ForcePosting inside a region must be undone by a rewind: it is the one
// posting-emitting operation that is legal inside one (it consumes no queue entry),
// and it is what a scaled source does inside a `oneof` branch.
func TestMark_RewindReversesForcePostings(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})

	rs.MarkPush()
	require.NoError(t, rs.ForcePosting("A", "", "swap", "", usd, "", big.NewInt(30)))
	wantBalance(t, rs, "A", 70)
	wantBalance(t, rs, "swap", 30)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "swap", Asset: usd, Amount: big.NewInt(30)},
	})

	require.NoError(t, rs.MarkEnd(true))

	// posting dropped, and both sides of it credited back
	wantPostings(t, rs, []runtime.Posting{})
	wantBalance(t, rs, "A", 100)
	wantBalance(t, rs, "swap", 0)
	require.False(t, rs.HasOpenMark())
}

// The realistic scaling shape: a swap out and back in a different asset, plus a
// pull, all inside one region. A rewind must leave every asset untouched.
func TestMark_RewindReversesAScaledSwapAndItsPull(t *testing.T) {
	const eur3 = "EUR/3"
	rs, _ := newRS(map[runtime.PairKey]int64{
		{"acc", "", usd, ""}:  1,
		{"acc", "", eur3, ""}: 10,
	})

	rs.MarkPush()
	// acc converts 10 EUR/3 through @swap into 1 more of the current asset
	require.NoError(t, rs.ForcePosting("acc", "", "swap", "", eur3, "", big.NewInt(10)))
	require.NoError(t, rs.ForcePosting("swap", "", "acc", "", usd, "", big.NewInt(1)))
	// then pulls what it now holds
	pull(rs, "acc", big.NewInt(2), big.NewInt(0), "")
	require.Len(t, rs.GetPostings(), 2)

	require.NoError(t, rs.MarkEnd(true))

	wantPostings(t, rs, []runtime.Posting{})
	wantBalance(t, rs, "acc", 1) // usd back to its starting balance
	if got := accBal(rs, "acc", "", eur3, ""); got.Cmp(big.NewInt(10)) != 0 {
		t.Errorf("balance(acc, %s) = %s, want 10", eur3, got)
	}
	if got := accBal(rs, "swap", "", eur3, ""); got.Sign() != 0 {
		t.Errorf("balance(swap, %s) = %s, want 0", eur3, got)
	}
	wantBalance(t, rs, "swap", 0)
}

// A rewind must reverse only the region's own postings; ones emitted before the
// push survive it.
func TestMark_RewindKeepsPostingsEmittedBeforeThePush(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})

	require.NoError(t, rs.ForcePosting("A", "", "keep", "", usd, "", big.NewInt(10)))
	rs.MarkPush()
	require.NoError(t, rs.ForcePosting("A", "", "drop", "", usd, "", big.NewInt(20)))
	require.NoError(t, rs.MarkEnd(true))

	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "keep", Asset: usd, Amount: big.NewInt(10)},
	})
	wantBalance(t, rs, "A", 90)
	wantBalance(t, rs, "keep", 10)
	wantBalance(t, rs, "drop", 0)
}

// Nested: the inner rewind drops only the inner posting; the outer rewind then drops
// the outer one too.
func TestMark_NestedRegionsReversePostingsIndependently(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})

	rs.MarkPush()
	require.NoError(t, rs.ForcePosting("A", "", "outer", "", usd, "", big.NewInt(10)))

	rs.MarkPush()
	require.NoError(t, rs.ForcePosting("A", "", "inner", "", usd, "", big.NewInt(20)))
	require.NoError(t, rs.MarkEnd(true))

	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "outer", Asset: usd, Amount: big.NewInt(10)},
	})
	wantBalance(t, rs, "inner", 0)
	wantBalance(t, rs, "A", 90) // outer posting survives the inner rewind

	require.NoError(t, rs.MarkEnd(true))

	wantPostings(t, rs, []runtime.Posting{})
	wantBalance(t, rs, "outer", 0)
	wantBalance(t, rs, "A", 100)
}

// Committing keeps them: a region that succeeded emits its postings for real.
func TestMark_CommitKeepsForcePostings(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})

	rs.MarkPush()
	require.NoError(t, rs.ForcePosting("A", "", "swap", "", usd, "", big.NewInt(30)))
	require.NoError(t, rs.MarkEnd(false))

	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "swap", Asset: usd, Amount: big.NewInt(30)},
	})
	wantBalance(t, rs, "A", 70)
	wantBalance(t, rs, "swap", 30)
}

// A rewind reverses a posting using the asset recorded *on the posting*, not
// currentAsset — so a posting in another asset is still undone correctly.
func TestMark_RewindUsesThePostingsOwnAsset(t *testing.T) {
	const other = "EUR/2"
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", other, ""}: 50})

	rs.MarkPush() // currentAsset is usd, the posting is in EUR/2
	require.NoError(t, rs.ForcePosting("A", "", "B", "", other, "", big.NewInt(20)))
	require.NoError(t, rs.MarkEnd(true))

	if got := accBal(rs, "A", "", other, ""); got.Cmp(big.NewInt(50)) != 0 {
		t.Errorf("balance(A, %s) = %s, want 50", other, got)
	}
	if got := accBal(rs, "B", "", other, ""); got.Sign() != 0 {
		t.Errorf("balance(B, %s) = %s, want 0", other, got)
	}
	// currentAsset untouched by the reversal
	wantBalance(t, rs, "A", 0)
}

func TestMark_ResetDropsAnOpenMark(t *testing.T) {
	// A run that fails mid-region must not leak its open marks into the
	// next run on a reused RunState.
	rs, store := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 100})
	rs.MarkPush()
	pull(rs, "A", big.NewInt(40), big.NewInt(0), "")
	require.True(t, rs.HasOpenMark())

	rs.Reset(store)
	require.False(t, rs.HasOpenMark())
	require.ErrorIs(t, rs.MarkEnd(false), runtime.ErrNoOpenMark)
}

// --- color ----------------------------------------------------------------

func TestColor_BalancesTrackedSeparatelyPerColor(t *testing.T) {
	// Same account+asset, two colors: each (account, asset, color) is its own
	// balance slot, fetched from the store independently.
	rs, store := newRS(map[runtime.PairKey]int64{
		{"A", "", usd, "red"}:  100,
		{"A", "", usd, "blue"}: 40,
	})
	if got := accBal(rs, "A", "", usd, "red"); got.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("red balance = %d, want 100", got)
	}
	if got := accBal(rs, "A", "", usd, "blue"); got.Cmp(big.NewInt(40)) != 0 {
		t.Errorf("blue balance = %d, want 40", got)
	}
	// uncolored slot is independent and absent -> 0
	if got := accBal(rs, "A", "", usd, ""); got.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("uncolored balance = %d, want 0", got)
	}
	if c := store.calls[runtime.PairKey{"A", "", usd, "red"}]; c != 1 {
		t.Errorf("red fetched %d times, want 1", c)
	}
}

func TestColor_PullTagsSourceAndPostingCarriesColor(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, "red"}: 100})
	pull(rs, "A", big.NewInt(60), big.NewInt(0), "red")
	_ = rs.Send(strptr("X"), "", big.NewInt(60), strptr("red"))
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Color: "red", Amount: big.NewInt(60)},
	})
	// destination credited on the colored slot, source debited on it
	if got := accBal(rs, "X", "", usd, "red"); got.Cmp(big.NewInt(60)) != 0 {
		t.Errorf("X red = %d, want 60", got)
	}
	if got := accBal(rs, "A", "", usd, "red"); got.Cmp(big.NewInt(40)) != 0 {
		t.Errorf("A red = %d, want 40", got)
	}
}

func TestColor_SendSkipsNonMatchingColorLeavingItQueued(t *testing.T) {
	// Queue order: red, blue, red. A red Send must drain the two red sources
	// (skipping blue, leaving it queued), exactly like fundsQueue.Pull's
	// color-skip.
	rs, _ := newRS(map[runtime.PairKey]int64{
		{"A", "", usd, "red"}:  50,
		{"B", "", usd, "blue"}: 30,
		{"C", "", usd, "red"}:  40,
	})
	pull(rs, "A", big.NewInt(50), big.NewInt(0), "red")
	pull(rs, "B", big.NewInt(30), big.NewInt(0), "blue")
	pull(rs, "C", big.NewInt(40), big.NewInt(0), "red")

	_ = rs.Send(strptr("X"), "", big.NewInt(100), strptr("red")) // only 90 red available; blue stays put
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Color: "red", Amount: big.NewInt(50)},
		{Source: "C", Destination: "X", Asset: usd, Color: "red", Amount: big.NewInt(40)},
	})

	// the skipped blue source is still queued and drains on a blue send
	_ = rs.Send(strptr("Y"), "", big.NewInt(100), strptr("blue"))
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Color: "red", Amount: big.NewInt(50)},
		{Source: "C", Destination: "X", Asset: usd, Color: "red", Amount: big.NewInt(40)},
		{Source: "B", Destination: "Y", Asset: usd, Color: "blue", Amount: big.NewInt(30)},
	})
}

func TestColor_SendDoesNotMergeAcrossColors(t *testing.T) {
	// Same src->dst->asset but different colors are distinct postings even
	// within consecutive drains.
	rs, _ := newRS(map[runtime.PairKey]int64{
		{"A", "", usd, "red"}:  40,
		{"A", "", usd, "blue"}: 40,
	})
	pull(rs, "A", big.NewInt(40), big.NewInt(0), "red")
	pull(rs, "A", big.NewInt(40), big.NewInt(0), "blue")
	_ = rs.SendUncapped(strptr("X"), "", strptr("red"))
	_ = rs.SendUncapped(strptr("X"), "", strptr("blue"))
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Color: "red", Amount: big.NewInt(40)},
		{Source: "A", Destination: "X", Asset: usd, Color: "blue", Amount: big.NewInt(40)},
	})
}

func TestColor_MatchAnyDrainsMixedColorsPreservingEach(t *testing.T) {
	// This is the mode the interpreter's destinations use (fundsQueue.PullAnything):
	// one drain (color == nil) consumes funds of several colors at once, and each
	// posting keeps its source fund's own color.
	rs, _ := newRS(map[runtime.PairKey]int64{
		{"A", "", usd, "red"}:  50,
		{"B", "", usd, "blue"}: 30,
		{"C", "", usd, ""}:     20,
	})
	pull(rs, "A", big.NewInt(50), big.NewInt(0), "red")
	pull(rs, "B", big.NewInt(30), big.NewInt(0), "blue")
	pull(rs, "C", big.NewInt(20), big.NewInt(0), "")

	_ = rs.Send(strptr("X"), "", big.NewInt(100), nil) // nil = match anything
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Asset: usd, Color: "red", Amount: big.NewInt(50)},
		{Source: "B", Destination: "X", Asset: usd, Color: "blue", Amount: big.NewInt(30)},
		{Source: "C", Destination: "X", Asset: usd, Color: "", Amount: big.NewInt(20)},
	})
	// destination credited on each respective color slot
	if b := accBal(rs, "X", "", usd, "red"); b.Cmp(big.NewInt(50)) != 0 {
		t.Errorf("X red = %s, want 50", b)
	}
	if b := accBal(rs, "X", "", usd, "blue"); b.Cmp(big.NewInt(30)) != 0 {
		t.Errorf("X blue = %s, want 30", b)
	}
}

func TestColor_RefundUsesSourceColor(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, "red"}: 100})
	pull(rs, "A", big.NewInt(100), big.NewInt(0), "red") // A red -> 0
	_ = rs.Send(nil, "", big.NewInt(60), strptr("red"))  // refund 60 to A's red slot
	if got := accBal(rs, "A", "", usd, "red"); got.Cmp(big.NewInt(60)) != 0 {
		t.Errorf("A red after refund = %d, want 60", got)
	}
	wantPostings(t, rs, []runtime.Posting{})
}

// --- end-to-end flow ------------------------------------------------------

func TestEndToEnd_TwoSourcesSplitAcrossDestinations(t *testing.T) {
	rs, store := newRS(map[runtime.PairKey]int64{
		{"alice", "", usd, ""}: 100,
		{"bob", "", usd, ""}:   100,
		{"carol", "", usd, ""}: 0,
		{"dave", "", usd, ""}:  0,
	})
	pull(rs, "alice", big.NewInt(100), big.NewInt(0), "")
	pull(rs, "bob", big.NewInt(100), big.NewInt(0), "")

	_ = rs.Send(strptr("carol"), "", big.NewInt(150), nil) // alice:100 fully, bob:50 partial (requeue bob:50)
	_ = rs.Send(strptr("dave"), "", big.NewInt(50), nil)   // bob:50 fully

	wantPostings(t, rs, []runtime.Posting{
		{Source: "alice", Destination: "carol", Asset: usd, Amount: big.NewInt(100)},
		{Source: "bob", Destination: "carol", Asset: usd, Amount: big.NewInt(50)},
		{Source: "bob", Destination: "dave", Asset: usd, Amount: big.NewInt(50)},
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
