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
// Numeric model: all amounts are *big.Int (arbitrary precision), matching the
// numscript interpreter. Because *big.Int is a mutable reference type, this
// package is careful about aliasing: it clones values it ingests from the Store
// and clones caller-supplied amounts it intends to mutate, it only mutates
// big.Ints it privately owns (queued source amounts), and it never hands out a
// live reference to its internal state (GetAccountBalance / GetPostings return
// copies).
package runtime

import "math/big"

// Store supplies the authoritative starting balance for an (account, asset,
// color) triple. A triple never seen by the ledger is fetched once, then cached.
// Implementations should return 0 (or nil, treated as 0) for unknown triples,
// not an error. The returned *big.Int is cloned on ingest, so the Store may
// safely reuse it.
type Store interface {
	GetBalance(account, asset, color string) *big.Int
}

// Posting is a recorded movement of Amount units of Asset (of the given Color)
// from Source to Destination. It is the single source of truth for the
// interpreter's public Posting type (aliased there), hence the json tags: field
// names and order define the public ledger serialization contract — keep them
// stable.
type Posting struct {
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Amount      *big.Int `json:"amount"`
	Asset       string   `json:"asset"`
	Color       string   `json:"color,omitempty"`
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
// right (asset, color) balance. The amount is privately owned by the queue and
// may be mutated in place.
type source struct {
	account string
	amount  *big.Int
	color   string
}

// RunState is the Go port of the OCaml run_state. The zero value is not usable;
// call New. All fields are unexported to preserve the .mli interface boundary.
type RunState struct {
	store        Store
	balances     map[PairKey]*big.Int // write-through cache over store
	sources      []source             // FIFO: front = index 0
	postings     []Posting
	currentAsset string
}

// New creates an empty RunState backed by store.
func New(store Store) *RunState {
	return &RunState{
		store:    store,
		balances: make(map[PairKey]*big.Int),
	}
}

// SetCurrentAsset sets the asset used when an operation omits one.
func (s *RunState) SetCurrentAsset(asset string) {
	s.currentAsset = asset
}

// Prewarm seeds the balance cache with balances fetched in bulk, so runtime's
// lazy per-key Store.GetBalance path is never hit for them. This lets a caller
// keep a single batched balance round-trip (e.g. the interpreter's pre-pass that
// collects every needed (account, asset, color) and fetches them in one query)
// instead of paying one Store call per triple.
//
// Call it once, before any Pull/Send/Save/ForcePosting. Amounts are cloned, so
// the caller may reuse them. A key that is already cached is left untouched (the
// live value wins), so a stray double-call can never clobber computed state.
func (s *RunState) Prewarm(balances map[PairKey]*big.Int) {
	for key, amount := range balances {
		if _, ok := s.balances[key]; ok {
			continue
		}
		cloned := new(big.Int)
		if amount != nil {
			cloned.Set(amount)
		}
		s.balances[key] = cloned
	}
}

// Has reports whether (account, asset, color) is already in the balance cache
// (prewarmed or touched). Lets a caller skip re-fetching balances it already
// holds, without triggering a Store load.
func (s *RunState) Has(account, asset, color string) bool {
	_, ok := s.balances[PairKey{account, asset, color}]
	return ok
}

// AccountBalance is a single cached (asset, color, amount) entry for an account.
type AccountBalance struct {
	Asset  string
	Color  string
	Amount *big.Int
}

// AccountBalances returns copies of every cached balance entry for account. It
// only reports entries already in the cache (it does not consult the Store), so
// an account that was never prewarmed/touched yields an empty slice. Used by
// asset scaling, which must enumerate an account's holdings across scales.
func (s *RunState) AccountBalances(account string) []AccountBalance {
	var out []AccountBalance
	for key, amount := range s.balances {
		if key.Account == account {
			out = append(out, AccountBalance{
				Asset:  key.Asset,
				Color:  key.Color,
				Amount: new(big.Int).Set(amount),
			})
		}
	}
	return out
}

// GetAccountBalance returns the balance for (account, asset, color). An empty
// asset means "use currentAsset" (the OCaml ?asset default). The value is
// fetched from the Store on first access and cached thereafter. The returned
// *big.Int is a fresh copy: callers may keep or mutate it freely without
// affecting runtime state.
//
// Note: "" is the unset sentinel for asset, consistent with currentAsset
// starting as "". A real asset must never be the empty string. For color, ""
// is a legitimate value meaning "uncolored".
func (s *RunState) GetAccountBalance(account, asset, color string) *big.Int {
	if asset == "" {
		asset = s.currentAsset
	}
	return new(big.Int).Set(s.cachedBalance(account, asset, color))
}

// Pull mirrors the OCaml `pull`. It debits up to cap from src's (currentAsset,
// color) balance (clamped to non-negative), honoring the overdraft policy,
// queues the pulled amount as a funding source tagged with color, and writes the
// amount made available into out. The overdraft bound is an optional *big.Int
// (the OCaml `int64 option`):
//
//	overdraft == nil -> unbounded: available = max(0, cap)
//	overdraft == b   -> available = min(max(0, balance + max(0,b)), max(0, cap))
//	                    (pass big.NewInt(0) for the "balance only" default)
//
// The result is written into the caller-provided out (overwritten), avoiding a
// return allocation; out may be any addressable *big.Int (e.g. a VM register).
// Inputs cap and overdraft are not mutated. The only allocation per call is the
// queued source's own copy of the amount (it must outlive out and is mutated in
// place by compactAt/Send); the balance is debited in place on the cached value.
func (s *RunState) Pull(out *big.Int, src string, cap *big.Int, overdraft *big.Int, color string) {
	currentBal := s.cachedBalance(src, s.currentAsset, color)

	if overdraft == nil {
		out.Set(cap) // unbounded; clamped to >= 0 below
	} else {
		// eff = max(0, currentBal + max(0, overdraft))
		out.Set(currentBal)
		if overdraft.Sign() > 0 {
			out.Add(out, overdraft)
		}
		if out.Sign() < 0 {
			out.SetInt64(0)
		}
		// available = min(eff, cap); a cap < eff (incl. negative) wins here and
		// is clamped to >= 0 below
		if cap.Cmp(out) < 0 {
			out.Set(cap)
		}
	}
	if out.Sign() < 0 {
		out.SetInt64(0)
	}

	// queue the pulled funds — an independent copy (out stays the caller's; the
	// queued amount is mutated in place by compactAt/Send)
	amt := new(big.Int).Set(out)
	s.sources = append(s.sources, source{src, amt, color})

	// debit the source balance in place; the cache keeps the same *big.Int
	currentBal.Sub(currentBal, out)
}

// PullUncapped mirrors the OCaml `pull_uncapped`: makes available
// max(0, balance + overdraftBound) of src's (currentAsset, color) balance,
// queuing it only when positive.
func (s *RunState) PullUncapped(src string, overdraftBound *big.Int, color string) *big.Int {
	currentBal := s.cachedBalance(src, s.currentAsset, color)
	available := clampNonNeg(new(big.Int).Add(currentBal, overdraftBound))
	if available.Sign() > 0 {
		s.balances[PairKey{src, s.currentAsset, color}] = new(big.Int).Sub(currentBal, available)
		s.sources = append(s.sources, source{src, available, color})
	}
	return new(big.Int).Set(available)
}

// Send mirrors the OCaml `send`, extended with a color filter. It drains queued
// funding sources in FIFO order until cap is satisfied or eligible sources run
// out, and each emitted posting carries the *consumed source's* own color.
//
// The color filter selects which sources are eligible:
//
//	color == nil   -> match anything (fundsQueue.PullAnything); a single drain
//	                  may consume and emit funds of several colors at once. This
//	                  is the mode the interpreter's destinations use.
//	color != nil   -> only sources whose color == *color are consumed; others
//	                  are skipped and left in place (fundsQueue.PullColored /
//	                  PullUncolored, with *color == "" meaning uncolored).
//
// dest == nil is the "keep/refund" path: the source is credited back and no
// posting is emitted. A partially consumed source's remainder stays at its
// position.
func (s *RunState) Send(dest *string, cap *big.Int, color *string) {
	cap = new(big.Int).Set(cap) // clone: we decrement it as sources are consumed
	asset := s.currentAsset
	i := 0
	for cap.Sign() > 0 && i < len(s.sources) {
		s.compactAt(i) // merge the run of adjacent same-(account,color) funds at i
		src := s.sources[i]
		if color != nil && src.color != *color {
			i++ // filtered out: skip, leave in place
			continue
		}
		if src.amount.Cmp(cap) >= 0 {
			s.credit(dest, src, asset, cap)
			if diff := new(big.Int).Sub(src.amount, cap); diff.Sign() > 0 {
				s.sources[i].amount = diff // remainder stays at this position
			} else {
				s.removeAt(i)
			}
			return // cap fully satisfied
		}
		s.credit(dest, src, asset, src.amount)
		cap.Sub(cap, src.amount)
		s.removeAt(i) // do not advance i; the next source shifts into position i
	}
}

// SendUncapped mirrors the OCaml `send_uncapped`, extended with the same color
// filter as Send: color == nil drains every queued source (each posting keeping
// its own color); color != nil drains only matching ones, leaving others in
// place.
func (s *RunState) SendUncapped(dest *string, color *string) {
	asset := s.currentAsset
	i := 0
	for i < len(s.sources) {
		s.compactAt(i) // merge the run of adjacent same-(account,color) funds at i
		src := s.sources[i]
		if color != nil && src.color != *color {
			i++ // filtered out: skip, leave in place
			continue
		}
		s.credit(dest, src, asset, src.amount)
		s.removeAt(i)
	}
}

// ForcePosting records a direct movement of amount (of asset/color) from src to
// dst, bypassing the funding queue: it debits src, credits dst, and appends the
// posting. It is for movements the queue does not model — e.g. asset-scaling
// conversions (interpreter.forcePushPostingUncolored). Unlike Send it uses the
// explicit asset argument, which may differ from the current asset (a scaled
// asset). A non-positive amount is a no-op. PRE: the caller has already checked
// invariants (e.g. amount sign); no balance sufficiency check is performed.
func (s *RunState) ForcePosting(src, dst, asset, color string, amount *big.Int) {
	if amount.Sign() <= 0 {
		return
	}
	s.addToBalance(src, asset, color, new(big.Int).Neg(amount))
	s.addPosting(src, dst, asset, color, amount) // appends the posting and credits dst
}

// Save mirrors the numscript `save` statement: it protects funds from being
// pulled later by reducing the (account, asset, color) balance, floored at zero.
//
//	amount != nil -> balance = max(0, balance - amount)   (PRE: amount >= 0)
//	amount == nil -> "save all": a positive balance becomes 0; a negative
//	                 balance is left unchanged (= min(balance, 0))
func (s *RunState) Save(account, asset, color string, amount *big.Int) {
	cur := s.cachedBalance(account, asset, color)
	var next *big.Int
	if amount == nil {
		if cur.Sign() <= 0 {
			return // negative/zero balance left unchanged
		}
		next = new(big.Int) // floor positive to zero
	} else {
		next = new(big.Int).Sub(cur, amount)
		if next.Sign() < 0 {
			next.SetInt64(0)
		}
	}
	s.balances[PairKey{account, asset, color}] = next
}

// Snapshot returns a cheap marker of the current source-queue depth, for
// backtracking a speculative source evaluation (e.g. a `oneof` branch). It is
// just the queue length: O(1), no allocation, no map cloning.
func (s *RunState) Snapshot() int {
	return len(s.sources)
}

// Restore undoes every Pull/PullUncapped performed since the matching Snapshot:
// it repays each source queued after the mark back to the (account, color)
// balance it was debited from, then truncates the queue to the mark. Balances
// are restored exactly without cloning maps — repaying the queued amounts is the
// exact inverse of the debits Pull made.
//
// PRECONDITION: nothing queued after the mark has been sent, and the current
// asset is unchanged since the Snapshot. Both hold during source evaluation,
// which is the only place backtracking happens — Send runs later, in the
// destination phase. (compactAt may have folded same-(account,color) funds, but
// the fold preserves both per the merge key, so the repay still lands correctly.)
func (s *RunState) Restore(mark int) {
	for i := mark; i < len(s.sources); i++ {
		src := s.sources[i]
		s.addToBalance(src.account, s.currentAsset, src.color, src.amount)
	}
	s.sources = s.sources[:mark]
}

// GetPostings returns a deep copy of the recorded postings, so callers cannot
// mutate internal state (matching the OCaml Dynarray.to_list, which copies).
func (s *RunState) GetPostings() []Posting {
	out := make([]Posting, len(s.postings))
	for i, p := range s.postings {
		cp := p
		cp.Amount = new(big.Int).Set(p.Amount)
		out[i] = cp
	}
	return out
}

// --- internal helpers ---

// credit routes a consumed source amount either into a posting (dest != nil) or
// back to the source as a refund (dest == nil). The funds keep their color, so
// both the posting and the destination/source balance land on (asset, color).
// amount is treated as read-only.
func (s *RunState) credit(dest *string, src source, asset string, amount *big.Int) {
	if dest != nil {
		s.addPosting(src.account, *dest, asset, src.color, amount)
	} else if amount.Sign() > 0 {
		// refund the source: consume funding, emit no posting
		s.addToBalance(src.account, asset, src.color, amount)
	}
}

// cachedBalance returns the cached balance for (account, asset, color), fetching
// from the Store and caching on first access. Presence in the map distinguishes
// "already fetched (possibly 0)" from "not yet fetched". The Store's value is
// cloned on ingest so runtime never mutates a pointer the Store owns. The
// returned pointer is the live cache entry — internal callers must not mutate it
// in place; they replace the map entry with a freshly allocated value instead.
func (s *RunState) cachedBalance(account, asset, color string) *big.Int {
	key := PairKey{account, asset, color}
	if v, ok := s.balances[key]; ok {
		return v
	}
	fromStore := s.store.GetBalance(account, asset, color)
	cached := new(big.Int)
	if fromStore != nil {
		cached.Set(fromStore)
	}
	s.balances[key] = cached
	return cached
}

// addToBalance applies delta to (account, asset, color), loading the base value
// through the cache first so an un-fetched account is not treated as 0. delta is
// read-only; the cache entry is replaced with a freshly allocated sum.
func (s *RunState) addToBalance(account, asset, color string, delta *big.Int) {
	cur := s.cachedBalance(account, asset, color)
	s.balances[PairKey{account, asset, color}] = new(big.Int).Add(cur, delta)
}

// addPosting appends a posting verbatim and credits the destination balance.
// Non-positive amounts are ignored. Postings are never merged here: same-source
// funds are instead coalesced upstream in the source queue by compactAt, so a
// posting can only ever fuse adjacent funds *within* one drain — never across
// separate sends. This mirrors the interpreter's fundsQueue, which merges in the
// queue (compactTop), not in the posting list. amount is cloned into the posting.
func (s *RunState) addPosting(src, dst, asset, color string, amount *big.Int) {
	if amount.Sign() <= 0 {
		return
	}
	s.postings = append(s.postings, Posting{
		Source:      src,
		Destination: dst,
		Asset:       asset,
		Color:       color,
		Amount:      new(big.Int).Set(amount),
	})
	s.addToBalance(dst, asset, color, amount)
}

// compactAt coalesces the maximal run of funds at index i that share i's
// (account, color), folding each into s.sources[i], and drops any zero-amount
// entries it passes. This is the slice analogue of fundsQueue.compactTop: it
// merges adjacent same-source funds in the queue before they are drained, so
// one drain over them yields a single posting. Because it operates on the queue
// (which each send fully consumes) and never on the posting list, it cannot fuse
// funds belonging to different sends. The fold mutates s.sources[i].amount in
// place, which is safe because queued amounts are privately owned.
func (s *RunState) compactAt(i int) {
	for i+1 < len(s.sources) {
		next := s.sources[i+1]
		if next.amount.Sign() == 0 {
			s.removeAt(i + 1)
			continue
		}
		if next.account != s.sources[i].account || next.color != s.sources[i].color {
			return
		}
		s.sources[i].amount.Add(s.sources[i].amount, next.amount)
		s.removeAt(i + 1)
	}
}

// removeAt deletes the source at index i, preserving the order of the rest.
func (s *RunState) removeAt(i int) {
	s.sources = append(s.sources[:i], s.sources[i+1:]...)
}

// clampNonNeg clamps x to >= 0 in place and returns it (for runtime-owned
// intermediates).
func clampNonNeg(x *big.Int) *big.Int {
	if x.Sign() < 0 {
		x.SetInt64(0)
	}
	return x
}
