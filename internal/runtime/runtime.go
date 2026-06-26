// Package runtime is a Go port of the OCaml run_state module, extended with
// color (sub-asset fungibility) support to match the interpreter's fundsQueue.
//
// It tracks per-(account, asset, color) balances, an ordered FIFO queue of
// funding sources produced by Pull/PullUncapped, and the list of postings
// produced by Send/SendUncapped. It is the state layer the VM's PullAccount /
// SendToAccount / CheckEnoughFunds opcodes call into.
//
// Balances are sourced lazily from a Store and then cached write-through: the
// first read of an (account, asset, color) triple fetches from the Store and
// caches the result; every subsequent read and every debit/credit operates on
// the cached value. So once @acc is fetched and decreased, later reads see the
// decreased balance without consulting the Store again.
//
// Color is a plain string; the empty string "" means "uncolored". Pull tags the
// funds it queues with a color, and Send drains only the sources whose color
// matches the requested one, skipping (but preserving the position of)
// non-matching funds — exactly like the interpreter's fundsQueue.
//
// Concurrency: a *RunState is mutable and NOT safe for concurrent use. Use one
// per execution.
//
// NOTE on numeric width: this layer is int64-native, matching the OCaml
// run_state (balances, postings, sources are all int64). The VM's register
// layer uses math/big. If your VM-level Store returns *big.Int, wrap it in an
// adapter that narrows to int64 here, or unify the two on one representation.
package runtime

// Store supplies the authoritative starting balance for an (account, asset,
// color) triple. A triple never seen by the ledger is fetched once, then cached.
// Implementations should return 0 for unknown triples (not an error).
type Store interface {
	GetBalance(account, asset, color string) int64
}

// Posting mirrors Common_intf.posting: a recorded movement of Amount units of
// Asset (of the given Color) from Source to Destination.
type Posting struct {
	Source      string
	Destination string
	Asset       string
	Color       string
	Amount      int64
}

// PairKey identifies a balance slot. Exported so a Store mock/adapter can build
// the same keys. Despite the name it is an (account, asset, color) triple.
type PairKey struct {
	Account string
	Asset   string
	Color   string
}

// source is an internal funding entry queued by Pull / PullUncapped. It carries
// the color of the funds so Send can filter and so postings/refunds land on the
// right (asset, color) balance.
type source struct {
	account string
	amount  int64
	color   string
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

// GetAccountBalance returns the balance for (account, asset, color). An empty
// asset means "use currentAsset" (the OCaml ?asset default). The value is
// fetched from the Store on first access and cached thereafter.
//
// Note: "" is the unset sentinel for asset, consistent with currentAsset
// starting as "". A real asset must never be the empty string. For color, ""
// is a legitimate value meaning "uncolored".
func (s *RunState) GetAccountBalance(account, asset, color string) int64 {
	if asset == "" {
		asset = s.currentAsset
	}
	return s.cachedBalance(account, asset, color)
}

// Pull mirrors the OCaml `pull`. It debits up to cap from src's (currentAsset,
// color) balance (clamped to non-negative), honoring the overdraft policy,
// queues the pulled amount as a funding source tagged with color, and returns
// the amount made available.
//
//	UnboundedOverdraft  -> available = cap
//	BoundedOverdraft(b) -> available = min(max(0, balance + max(0,b)), cap)
func (s *RunState) Pull(src string, cap int64, ovd Overdraft, color string) int64 {
	cap = nonNeg(cap)
	currentBal := s.GetAccountBalance(src, "", color) // resolves to currentAsset, caches

	var available int64
	if !ovd.bounded {
		available = cap
	} else {
		bound := nonNeg(ovd.bound)
		eff := nonNeg(currentBal + bound)
		available = min64(eff, cap)
	}

	// key is already cached by GetAccountBalance above, so direct write is safe.
	s.balances[PairKey{src, s.currentAsset, color}] = currentBal - available
	s.sources = append(s.sources, source{src, available, color})
	return available
}

// PullUncapped mirrors the OCaml `pull_uncapped`: makes available
// max(0, balance + overdraftBound) of src's (currentAsset, color) balance,
// queuing it only when positive.
func (s *RunState) PullUncapped(src string, overdraftBound int64, color string) int64 {
	currentBal := s.GetAccountBalance(src, "", color)
	available := nonNeg(currentBal + overdraftBound)
	if available > 0 {
		s.balances[PairKey{src, s.currentAsset, color}] = currentBal - available
		s.sources = append(s.sources, source{src, available, color})
	}
	return available
}

// Send mirrors the OCaml `send`, extended with a color filter. It drains queued
// funding sources whose color matches `color` in FIFO order until cap is
// satisfied or matching sources run out. Sources of a different color are
// skipped and left in place (matching fundsQueue.Pull's color-skip). dest ==
// nil is the "keep/refund" path: the source is credited back and no posting is
// emitted. A partially consumed source's remainder stays at its position.
func (s *RunState) Send(dest *string, cap int64, color string) {
	asset := s.currentAsset
	i := 0
	for cap > 0 && i < len(s.sources) {
		s.compactAt(i) // merge the run of adjacent same-(account,color) funds at i
		src := s.sources[i]
		if src.color != color {
			i++ // different color: skip, leave in place
			continue
		}
		if src.amount >= cap {
			s.credit(dest, src, asset, cap)
			if diff := src.amount - cap; diff > 0 {
				s.sources[i].amount = diff // remainder stays at this position
			} else {
				s.removeAt(i)
			}
			return // cap fully satisfied
		}
		s.credit(dest, src, asset, src.amount)
		cap -= src.amount
		s.removeAt(i) // do not advance i; the next source shifts into position i
	}
}

// SendUncapped mirrors the OCaml `send_uncapped`, extended with a color filter:
// drains every queued source whose color matches `color`, leaving others in
// place.
func (s *RunState) SendUncapped(dest *string, color string) {
	asset := s.currentAsset
	i := 0
	for i < len(s.sources) {
		s.compactAt(i) // merge the run of adjacent same-(account,color) funds at i
		src := s.sources[i]
		if src.color != color {
			i++ // different color: skip, leave in place
			continue
		}
		s.credit(dest, src, asset, src.amount)
		s.removeAt(i)
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

// credit routes a consumed source amount either into a posting (dest != nil) or
// back to the source as a refund (dest == nil). The funds keep their color, so
// both the posting and the destination/source balance land on (asset, color).
func (s *RunState) credit(dest *string, src source, asset string, amount int64) {
	if dest != nil {
		s.addPosting(src.account, *dest, asset, src.color, amount)
	} else if amount > 0 {
		// refund the source: consume funding, emit no posting
		s.addToBalance(src.account, asset, src.color, amount)
	}
}

// cachedBalance returns the cached balance for (account, asset, color), fetching
// from the Store and caching on first access. Presence in the map distinguishes
// "already fetched (possibly 0)" from "not yet fetched".
func (s *RunState) cachedBalance(account, asset, color string) int64 {
	key := PairKey{account, asset, color}
	if v, ok := s.balances[key]; ok {
		return v
	}
	v := s.store.GetBalance(account, asset, color)
	s.balances[key] = v
	return v
}

// addToBalance applies delta to (account, asset, color), loading the base value
// through the cache first so an un-fetched account is not treated as 0.
func (s *RunState) addToBalance(account, asset, color string, delta int64) {
	cur := s.cachedBalance(account, asset, color)
	s.balances[PairKey{account, asset, color}] = cur + delta
}

// addPosting appends a posting verbatim and credits the destination balance.
// Non-positive amounts are ignored. Postings are never merged here: same-source
// funds are instead coalesced upstream in the source queue by compactAt, so a
// posting can only ever fuse adjacent funds *within* one drain — never across
// separate sends. This mirrors the interpreter's fundsQueue, which merges in the
// queue (compactTop), not in the posting list.
func (s *RunState) addPosting(src, dst, asset, color string, amount int64) {
	if amount <= 0 {
		return
	}
	s.postings = append(s.postings, Posting{
		Source:      src,
		Destination: dst,
		Asset:       asset,
		Color:       color,
		Amount:      amount,
	})
	s.addToBalance(dst, asset, color, amount)
}

// compactAt coalesces the maximal run of funds at index i that share i's
// (account, color), folding each into s.sources[i], and drops any zero-amount
// entries it passes. This is the slice analogue of fundsQueue.compactTop: it
// merges adjacent same-source funds in the queue before they are drained, so
// one drain over them yields a single posting. Because it operates on the queue
// (which each send fully consumes) and never on the posting list, it cannot fuse
// funds belonging to different sends.
func (s *RunState) compactAt(i int) {
	for i+1 < len(s.sources) {
		next := s.sources[i+1]
		if next.amount == 0 {
			s.removeAt(i + 1)
			continue
		}
		if next.account != s.sources[i].account || next.color != s.sources[i].color {
			return
		}
		s.sources[i].amount += next.amount
		s.removeAt(i + 1)
	}
}

// removeAt deletes the source at index i, preserving the order of the rest.
func (s *RunState) removeAt(i int) {
	s.sources = append(s.sources[:i], s.sources[i+1:]...)
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
