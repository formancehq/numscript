package runtime_test

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/formancehq/numscript/internal/runtime"
)

func acct(i int) string { return fmt.Sprintf("acc%d", i) }

// Draining many queued sources front-to-back must stay correct (and FIFO) now
// that the front pop is head++ instead of a slice shift.
func TestSend_ManySourcesDrainInFIFOOrder(t *testing.T) {
	const n = 64
	initial := map[runtime.PairKey]int64{}
	for i := 0; i < n; i++ {
		initial[runtime.PairKey{acct(i), "", usd, ""}] = 1
	}
	rs, _ := newRS(initial)
	for i := 0; i < n; i++ {
		pull(rs, acct(i), big.NewInt(1), big.NewInt(0), "")
	}

	rs.SendUncapped(strptr("dest"), "", nil)

	got := rs.GetPostings()
	if len(got) != n {
		t.Fatalf("got %d postings, want %d", len(got), n)
	}
	for i := 0; i < n; i++ {
		if got[i].Source != acct(i) || got[i].Destination != "dest" ||
			got[i].Amount.Cmp(big.NewInt(1)) != 0 {
			t.Fatalf("posting %d = %s->%s %s, want %s->dest 1",
				i, got[i].Source, got[i].Destination, got[i].Amount, acct(i))
		}
	}
}

// After a send fully drains the queue, the next pull must rewind the backing
// array to the front (rewindIfEmpty) without corrupting anything across many
// pull/send cycles — this is the path that keeps the backing bounded.
func TestQueue_ReusedAcrossPullSendCycles(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{{"A", "", usd, ""}: 1000})
	const cycles = 20
	for k := 0; k < cycles; k++ {
		pull(rs, "A", big.NewInt(10), big.NewInt(0), "")
		rs.Send(strptr("dest"), "", big.NewInt(10), nil)
	}

	got := rs.GetPostings()
	if len(got) != cycles {
		t.Fatalf("got %d postings, want %d", len(got), cycles)
	}
	for i, p := range got {
		if p.Source != "A" || p.Destination != "dest" || p.Amount.Cmp(big.NewInt(10)) != 0 {
			t.Fatalf("posting %d = %s->%s %s, want A->dest 10", i, p.Source, p.Destination, p.Amount)
		}
	}
	wantBalance(t, rs, "A", 1000-cycles*10)
}

// A colored drain that consumes the front (head++), skips a non-matching source,
// then consumes a matching one behind the skip (a mid removal with head > 0) must
// remain correct, leaving the skipped source live for a later drain.
func TestSend_ColoredDrainWithDeadPrefix(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{
		{"A", "", usd, "red"}:  10,
		{"B", "", usd, "blue"}: 10,
		{"C", "", usd, "red"}:  10,
	})
	pull(rs, "A", big.NewInt(10), big.NewInt(0), "red")
	pull(rs, "B", big.NewInt(10), big.NewInt(0), "blue")
	pull(rs, "C", big.NewInt(10), big.NewInt(0), "red")

	// drain reds: A (front → head++), skip B, C (mid removal with head==1)
	rs.Send(strptr("dest"), "", big.NewInt(100), strptr("red"))
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "dest", Amount: big.NewInt(10), Asset: usd, Color: "red"},
		{Source: "C", Destination: "dest", Amount: big.NewInt(10), Asset: usd, Color: "red"},
	})

	// the skipped blue source is still queued; drain it now (front → head++)
	rs.SendUncapped(strptr("dest2"), "", nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "dest", Amount: big.NewInt(10), Asset: usd, Color: "red"},
		{Source: "C", Destination: "dest", Amount: big.NewInt(10), Asset: usd, Color: "red"},
		{Source: "B", Destination: "dest2", Amount: big.NewInt(10), Asset: usd, Color: "blue"},
	})
}

// Snapshot/Restore uses absolute indices; it must stay correct when a preceding
// send left a dead prefix (head > 0) so the mark is taken over a partially
// consumed queue.
func TestSnapshotRestore_OverDeadPrefix(t *testing.T) {
	rs, _ := newRS(map[runtime.PairKey]int64{
		{"A", "", usd, ""}: 50,
		{"B", "", usd, ""}: 50,
		{"C", "", usd, ""}: 100,
	})
	pull(rs, "A", big.NewInt(50), big.NewInt(0), "")
	pull(rs, "B", big.NewInt(50), big.NewInt(0), "")

	// consume A exactly → head advances to 1, B stays live (dead prefix of 1)
	rs.Send(strptr("X"), "", big.NewInt(50), nil)
	wantBalance(t, rs, "A", 0)

	// snapshot over the dead prefix, speculatively pull C, then backtrack
	mark := rs.Snapshot()
	pull(rs, "C", big.NewInt(100), big.NewInt(0), "")
	wantBalance(t, rs, "C", 0) // debited by the speculative pull
	rs.Restore(mark)
	wantBalance(t, rs, "C", 100) // repaid on restore

	// B must still be intact and drainable
	rs.SendUncapped(strptr("Y"), "", nil)
	wantPostings(t, rs, []runtime.Posting{
		{Source: "A", Destination: "X", Amount: big.NewInt(50), Asset: usd},
		{Source: "B", Destination: "Y", Amount: big.NewInt(50), Asset: usd},
	})
	wantBalance(t, rs, "B", 0)
}

// Drain a large queue front-to-back on a reused RunState. With the head index the
// per-pop cost is O(1), so the whole drain is O(n) rather than O(n^2).
func BenchmarkSend_ManySources(b *testing.B) {
	const n = 512
	initial := map[runtime.PairKey]int64{}
	for i := 0; i < n; i++ {
		initial[runtime.PairKey{acct(i), "", usd, ""}] = 1
	}
	store := newMockStore(initial)
	rs := runtime.New(store)
	dest := "dest"
	one := big.NewInt(1)
	zero := big.NewInt(0)
	out := new(big.Int)

	b.ReportAllocs()
	b.ResetTimer()
	for iter := 0; iter < b.N; iter++ {
		rs.Reset(store)
		rs.SetCurrentAsset(usd)
		for i := 0; i < n; i++ {
			rs.Pull(out, acct(i), "", one, zero, "")
		}
		rs.SendUncapped(&dest, "", nil)
	}
}
