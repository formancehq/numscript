// Package runtime is a Go port of the OCaml run_state module.
//
// It tracks per-(account, asset) balances, an ordered FIFO queue of funding
// sources produced by Pull/PullUncapped, and the list of postings produced by
// Send/SendUncapped. It is the state layer the VM's PullAccount /
// SendToAccount / CheckEnoughFunds opcodes call into.
//
// Balances are sourced lazily from a Store and then cached write-through: the
// first read of an (account, asset) pair fetches from the Store and caches the
// result; every subsequent read and every debit/credit operates on the cached
// value. So once @acc is fetched and decreased, later reads see the decreased
// balance without consulting the Store again.
//
// Concurrency: a *RunState is mutable and NOT safe for concurrent use. Use one
// per execution.
//
// NOTE on numeric width: this layer is int64-native, matching the OCaml
// run_state (balances, postings, sources are all int64). The VM's register
// layer uses math/big. If your VM-level Store returns *big.Int, wrap it in an
// adapter that narrows to int64 here, or unify the two on one representation.
package runtime

// Store supplies the authoritative starting balance for an (account, asset)
// pair. A pair never seen by the ledger is fetched once, then cached.
// Implementations should return 0 for unknown pairs (not an error).
type Store interface {
	GetBalance(account, asset string) int64
}

// Posting mirrors Common_intf.posting: a recorded movement of Amount units of
// Asset from Source to Destination.
type Posting struct {
	Source      string
	Destination string
	Asset       string
	Amount      int64
}

// PairKey identifies a balance slot. Exported so a Store mock/adapter can build
// the same keys.
type PairKey struct {
	Account string
	Asset   string
}

// source is an internal funding entry queued by Pull / PullUncapped.
type source struct {
	account string
	amount  int64
}

// Overdraft expresses the OCaml inner `int64 option` overdraft bound:
//
//	UnboundedOverdraft()  -> OCaml None    (take the full cap)
//	BoundedOverdraft(n)   -> OCaml Some n  (clamp by balance + n)
//
// The OCaml `pull` default is `Some 0L`; pass BoundedOverdraft(0) for that.
type Overdraft struct {
	bounded bool
	bound   int64
}

func UnboundedOverdraft() Overdraft      { return Overdraft{bounded: false} }
func BoundedOverdraft(n int64) Overdraft { return Overdraft{bounded: true, bound: n} }

// RunState is the Go port of the OCaml run_state. The zero value is not usable;
// call New. All fields are unexported to preserve the .mli interface boundary.
type RunState struct {
	store        Store
	balances     map[PairKey]int64 // write-through cache over store
	sources      []source          // FIFO: front = index 0
	postings     []Posting
	currentAsset string
}

// New creates an empty RunState backed by store.
func New(store Store) *RunState {
	return &RunState{
		store:    store,
		balances: make(map[PairKey]int64),
	}
}

// SetCurrentAsset sets the asset used when an operation omits one.
func (s *RunState) SetCurrentAsset(asset string) {
	s.currentAsset = asset
}

// GetAccountBalance returns the balance for (account, asset). An empty asset
// means "use currentAsset" (the OCaml ?asset default). The value is fetched
// from the Store on first access and cached thereafter.
//
// Note: "" is the unset sentinel, consistent with currentAsset starting as "".
// A real asset must never be the empty string.
func (s *RunState) GetAccountBalance(account, asset string) int64 {
	if asset == "" {
		asset = s.currentAsset
	}
	return s.cachedBalance(account, asset)
}

// Pull mirrors the OCaml `pull`. It debits up to cap from src (clamped to
// non-negative), honoring the overdraft policy, queues the pulled amount as a
// funding source, and returns the amount made available.
//
//	UnboundedOverdraft  -> available = cap
//	BoundedOverdraft(b) -> available = min(max(0, balance + max(0,b)), cap)
func (s *RunState) Pull(src string, cap int64, ovd Overdraft) int64 {
	cap = nonNeg(cap)
	currentBal := s.GetAccountBalance(src, "") // resolves to currentAsset, caches

	var available int64
	if !ovd.bounded {
		available = cap
	} else {
		bound := nonNeg(ovd.bound)
		eff := nonNeg(currentBal + bound)
		available = min64(eff, cap)
	}

	// key is already cached by GetAccountBalance above, so direct write is safe.
	s.balances[PairKey{src, s.currentAsset}] = currentBal - available
	s.sources = append(s.sources, source{src, available})
	return available
}

// PullUncapped mirrors the OCaml `pull_uncapped`: makes available
// max(0, balance + overdraftBound), queuing it only when positive.
func (s *RunState) PullUncapped(src string, overdraftBound int64) int64 {
	currentBal := s.GetAccountBalance(src, "")
	available := nonNeg(currentBal + overdraftBound)
	if available > 0 {
		s.balances[PairKey{src, s.currentAsset}] = currentBal - available
		s.sources = append(s.sources, source{src, available})
	}
	return available
}

// Send mirrors the OCaml `send`. It drains queued funding sources FIFO until
// cap is satisfied or sources run out. dest == nil is the "keep/refund" path:
// the source is credited back and no posting is emitted. A partially consumed
// source's remainder is requeued at the front.
func (s *RunState) Send(dest *string, cap int64) {
	for cap > 0 {
		src, ok := s.popFirst()
		if !ok {
			return
		}
		asset := s.currentAsset

		credit := func(amount int64) {
			if dest != nil {
				s.addPosting(src.account, *dest, asset, amount)
			} else if amount > 0 {
				// refund the source: consume funding, emit no posting
				s.addToBalance(src.account, asset, amount)
			}
		}

		if src.amount >= cap {
			credit(cap)
			if diff := src.amount - cap; diff > 0 {
				s.pushFront(source{src.account, diff})
			}
			return // cap fully satisfied
		}
		credit(src.amount)
		cap -= src.amount
	}
}

// SendUncapped mirrors the OCaml `send_uncapped`: drains every queued source.
func (s *RunState) SendUncapped(dest *string) {
	for {
		src, ok := s.popFirst()
		if !ok {
			return
		}
		if src.amount > 0 {
			if dest != nil {
				s.addPosting(src.account, *dest, s.currentAsset, src.amount)
			} else {
				s.addToBalance(src.account, s.currentAsset, src.amount)
			}
		}
	}
}

// GetPostings returns a copy of the recorded postings, so callers cannot mutate
// internal state (matching the OCaml Dynarray.to_list, which copies).
func (s *RunState) GetPostings() []Posting {
	out := make([]Posting, len(s.postings))
	copy(out, s.postings)
	return out
}

// --- internal helpers ---

// cachedBalance returns the cached balance for (account, asset), fetching from
// the Store and caching on first access. Presence in the map distinguishes
// "already fetched (possibly 0)" from "not yet fetched".
func (s *RunState) cachedBalance(account, asset string) int64 {
	key := PairKey{account, asset}
	if v, ok := s.balances[key]; ok {
		return v
	}
	v := s.store.GetBalance(account, asset)
	s.balances[key] = v
	return v
}

// addToBalance applies delta to (account, asset), loading the base value
// through the cache first so an un-fetched account is not treated as 0.
func (s *RunState) addToBalance(account, asset string, delta int64) {
	cur := s.cachedBalance(account, asset)
	s.balances[PairKey{account, asset}] = cur + delta
}

// addPosting appends a posting, merging with the previous one when source,
// destination, and asset all match, and credits the destination balance.
// Non-positive amounts are ignored.
func (s *RunState) addPosting(src, dst, asset string, amount int64) {
	if amount <= 0 {
		return
	}
	n := len(s.postings)
	merged := false
	if n > 0 {
		last := &s.postings[n-1]
		if last.Source == src && last.Destination == dst && last.Asset == asset {
			last.Amount += amount
			merged = true
		}
	}
	if !merged {
		s.postings = append(s.postings, Posting{
			Source:      src,
			Destination: dst,
			Asset:       asset,
			Amount:      amount,
		})
	}
	s.addToBalance(dst, asset, amount)
}

// popFirst removes and returns the front source (FIFO).
func (s *RunState) popFirst() (source, bool) {
	if len(s.sources) == 0 {
		return source{}, false
	}
	first := s.sources[0]
	s.sources = s.sources[1:]
	return first, true
}

// pushFront requeues a source at the front (O(n), matches OCaml add_left).
func (s *RunState) pushFront(src source) {
	s.sources = append([]source{src}, s.sources...)
}

func nonNeg(x int64) int64 {
	if x < 0 {
		return 0
	}
	return x
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
