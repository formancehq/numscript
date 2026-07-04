// Package runtime is a Go port of the OCaml run_state module, extended with
// color (sub-asset fungibility) support. It is the shared funds engine driven by
// both the VM and the tree-walking interpreter.
//
// It tracks per-(account, asset, color) balances, an ordered FIFO queue of
// funding sources produced by Pull/PullUncapped, and the list of postings
// produced by Send/SendUncapped. It is the state layer the VM's PullAccount /
// SendToAccount / CheckEnoughFunds opcodes call into.
//
// Balances are tracked per (account, asset, color) triple as a balanceEntry that
// separates the movement applied this run from the account's starting balance.
// Until something needs the absolute balance, an entry holds only the net delta
// (debits/credits) and the Store is never consulted: an unbounded pull (from
// @world, or `allowing unbounded overdraft`) makes cap available regardless of
// balance, so it just records a debit. The starting balance is fetched from the
// Store lazily, the first time an operation actually needs the absolute value (a
// bounded pull, a send-all, or a balance() read), and folded into the delta;
// from then on the entry holds the absolute balance and further ops mutate it in
// place. This keeps the common send-from-@world path free of Store round-trips
// while a later balance(@world) still reports the correct running balance.
//
// Color is a plain string; the empty string "" means "uncolored". Pull tags the
// funds it queues with a color, and Send drains only the sources whose color
// matches the requested one, skipping (but preserving the position of)
// non-matching funds.
//
// Concurrency: a *RunState is mutable and NOT safe for concurrent use. Use one
// per execution.
//
// Numeric model: all amounts are *big.Int (arbitrary precision), matching the
// numscript interpreter. Because *big.Int is a mutable reference type, this
// package is careful about aliasing: it clones values it ingests from the Store
// and clones caller-supplied amounts it intends to mutate, it only mutates
// big.Ints it privately owns (queued source amounts), and it never hands out a
// live reference to its internal state (GetAccountBalance returns a copy;
// GetPostings returns a fresh slice whose amounts are valid until the next Reset).
package runtime

import "math/big"

// Store supplies the authoritative starting balance for an (account, asset,
// color) triple. A triple never seen by the ledger is fetched once, then cached.
// Implementations should return 0 (or nil, treated as 0) for unknown triples,
// not an error. The returned *big.Int is cloned on ingest, so the Store may
// safely reuse it.
type Store interface {
	GetBalance(account, asset, color string) (*big.Int, error)
}

// Posting is a recorded movement of Amount units of Asset (of the given Color)
// from Source to Destination. It is the single source of truth for the
// interpreter's public Posting type (aliased there), hence the json tags: field
// names and order define the public ledger serialization contract — keep them
// stable. The VM leaves the scope fields empty; the interpreter fills them.
type Posting struct {
	Source           string   `json:"source"`
	SourceScope      string   `json:"sourceScope,omitempty"`
	Destination      string   `json:"destination"`
	DestinationScope string   `json:"destinationScope,omitempty"`
	Amount           *big.Int `json:"amount"`
	Asset            string   `json:"asset"`
	Color            string   `json:"color,omitempty"`
}

type ExecutionResult struct {
	Postings         []Posting         `json:"postings"`
	Metadata         map[string]string `json:"txMeta"`
	AccountsMetadata AccountsMetadata  `json:"accountsMeta"`
}

// PairKey identifies a balance slot: an (account, scope, asset, color) tuple.
// Exported so a Store mock/adapter can build the same keys. Scope is a second
// dimension of the account (scoped accounts hold separate balances); the VM
// leaves it empty.
type PairKey struct {
	Account string
	Scope   string
	Asset   string
	Color   string
}

// source is an internal funding entry queued by Pull / PullUncapped. It carries
// the scope and color of the funds so Send can filter and so postings/refunds
// land on the right (scope, asset, color) balance. The amount is privately owned
// by the queue and may be mutated in place.
type source struct {
	account string
	scope   string
	amount  *big.Int
	color   string
}

// balanceEntry tracks one (account, asset, color) balance. While baseLoaded is
// false, amount is the net delta applied this run (starting from 0) and the
// Store has not been consulted; once baseLoaded is true, the Store's starting
// balance has been folded in and amount is the absolute running balance.
type balanceEntry struct {
	amount     big.Int
	baseLoaded bool
}

// RunState is the Go port of the OCaml run_state. The zero value is not usable;
// call New. All fields are unexported to preserve the .mli interface boundary.
type RunState struct {
	store        Store
	balances     map[PairKey]*balanceEntry
	sources      []source // FIFO; the live window is sources[head:]
	head         int      // index of the front: consuming it is head++, no shift
	postings     []Posting
	currentAsset string

	// free recycles *big.Int across runs to avoid per-run allocation. It holds
	// runtime-OWNED big.Ints that never escape past the next Reset: queued source
	// amounts (reclaimed when a source is consumed, merged, or dropped, and any
	// leftovers at Reset) and posting amounts (reclaimed on Reset — see the
	// GetPostings/PostingsRef lifetime contract). Balance amounts are NOT pooled:
	// they live inline in balanceEntry (a value big.Int), not as separate pointers.
	// takeBig() returns a (possibly dirty) one to overwrite; putBig() returns a dead
	// one.
	free []*big.Int

	// capScratch is a reusable big.Int for Send's decrementing cap, so a capped
	// send allocates nothing. Safe to reuse: Send is not reentrant.
	capScratch big.Int
}

func (s *RunState) takeBig() *big.Int {
	n := len(s.free)
	if n == 0 {
		return new(big.Int)
	}
	v := s.free[n-1]
	s.free = s.free[:n-1]
	return v
}

func (s *RunState) putBig(x *big.Int) {
	s.free = append(s.free, x)
}

// New creates an empty RunState backed by store.
func New(store Store) *RunState {
	return &RunState{
		store:    store,
		balances: make(map[PairKey]*balanceEntry),
	}
}

// SetCurrentAsset sets the asset used when an operation omits one.
func (s *RunState) SetCurrentAsset(asset string) {
	s.currentAsset = asset
}

// Reset clears all per-execution state — the balance cache, the source queue,
// the postings, and the current asset — and rebinds the store, while retaining
// the underlying map/slice capacity. This lets a single RunState be reused
// across executions without reallocating its containers (the balances map and
// the sources/postings slices keep their backing storage).
//
// Note: Reset recycles the posting amounts into the pool, so a GetPostings /
// PostingsRef result from before this Reset must not be used afterward (see their
// lifetime contract). A RunState that is never Reset keeps such results valid.
func (s *RunState) Reset(store Store) {
	s.store = store
	clear(s.balances)
	// reclaim any leftover source amounts: they are runtime-owned and now dead, so
	// recycle them for the next run. Only the live window (sources[head:]) still
	// holds amounts that haven't been recycled; the dead prefix left by
	// front-consumption (head > 0) was already recycled at consume time.
	for i := s.head; i < len(s.sources); i++ {
		s.free = append(s.free, s.sources[i].amount)
	}
	s.sources = s.sources[:0]
	s.head = 0
	for i := range s.postings {
		s.free = append(s.free, s.postings[i].Amount)
	}
	s.postings = s.postings[:0]
	s.currentAsset = ""
}

// Prewarm seeds the balance cache with balances fetched in bulk, so runtime's
// lazy per-key Store.GetBalance path is never hit for them. This lets a caller
// keep a single batched balance round-trip (e.g. the interpreter's pre-pass that
// collects every needed (account, asset, color) and fetches them in one query)
// instead of paying one Store call per triple.
//
// Call it once, before any Pull/Send/Save/ForcePosting. Amounts are cloned, so
// the caller may reuse them. A key whose base is already loaded is left untouched
// (the live value wins), so a stray double-call can never clobber computed state;
// a key that only holds a delta so far has the base folded into it here.
func (s *RunState) Prewarm(balances map[PairKey]*big.Int) {
	for key, amount := range balances {
		e := s.balances[key]
		if e == nil {
			e = &balanceEntry{}
			s.balances[key] = e
		} else if e.baseLoaded {
			continue
		}
		if amount != nil {
			e.amount.Add(&e.amount, amount) // fold base into any accumulated delta
		}
		e.baseLoaded = true
	}
}

// Has reports whether (account, asset, color) already has its starting balance
// loaded (prewarmed or fetched). An entry that only holds a delta so far reports
// false, so a caller (e.g. fetchAndPrewarm) still fetches and folds in the base.
func (s *RunState) Has(account, scope, asset, color string) bool {
	e := s.balances[PairKey{account, scope, asset, color}]
	return e != nil && e.baseLoaded
}

// AccountBalance is a single cached (asset, color, amount) entry for an account.
type AccountBalance struct {
	Asset  string
	Color  string
	Amount *big.Int
}

// AccountBalances returns copies of every tracked balance entry for account,
// with each entry's starting balance folded in so the amounts are absolute. It
// only reports triples already touched this run (it does not enumerate the
// Store), so an account that was never prewarmed/touched yields an empty slice.
// Used by asset scaling, which must enumerate an account's holdings across scales.
func (s *RunState) AccountBalances(account, scope string) ([]AccountBalance, error) {
	var out []AccountBalance
	for key, e := range s.balances {
		if key.Account == account && key.Scope == scope {
			if err := s.loadBase(key, e); err != nil {
				return nil, err
			}
			out = append(out, AccountBalance{
				Asset:  key.Asset,
				Color:  key.Color,
				Amount: new(big.Int).Set(&e.amount),
			})
		}
	}
	return out, nil
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
func (s *RunState) GetAccountBalance(account, scope, asset, color string) (*big.Int, error) {
	if asset == "" {
		asset = s.currentAsset
	}
	bal, err := s.absoluteBalance(account, scope, asset, color)
	if err != nil {
		return nil, err
	}
	return new(big.Int).Set(bal), nil
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
func (s *RunState) Pull(out *big.Int, src string, scope string, cap *big.Int, overdraft *big.Int, color string) error {
	if overdraft == nil {
		// unbounded: available = max(0, cap), independent of the balance — so no
		// Store fetch. The debit is recorded as a delta and folded against the
		// starting balance only if the balance is later needed.
		out.Set(cap)
		if out.Sign() < 0 {
			out.SetInt64(0)
		}
		amt := s.takeBig()
		amt.Set(out)
		s.sources = append(s.sources, source{src, scope, amt, color})
		e := s.entryFor(PairKey{src, scope, s.currentAsset, color})
		e.amount.Sub(&e.amount, out)
		return nil
	}

	currentBal, err := s.absoluteBalance(src, scope, s.currentAsset, color)
	if err != nil {
		return err
	}

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
	if out.Sign() < 0 {
		out.SetInt64(0)
	}

	// queue the pulled funds — an independent (recycled) copy (out stays the
	// caller's; the queued amount is mutated in place by compactAt/Send)
	amt := s.takeBig()
	amt.Set(out)
	s.pushSource(src, scope, amt, color)

	// debit the source balance in place; the cache keeps the same *big.Int
	currentBal.Sub(currentBal, out)
	return nil
}

// PullUncapped mirrors the OCaml `pull_uncapped`: makes available
// max(0, balance + max(0, overdraftBound)) of src's (currentAsset, color)
// balance, queuing it only when positive, and writes the available amount into
// out. As in Pull, a negative overdraftBound is clamped to 0 (a nonsensical
// negative bound never eats into the positive balance); pass big.NewInt(0) for
// the "balance only" default.
//
// Like Pull, the result is written into the caller-provided out (no return
// allocation; out may be any addressable *big.Int). overdraftBound is not
// mutated. When the available amount is positive it costs one allocation (the
// queued source's own copy) and debits the balance in place; when it is zero
// nothing is queued, nothing is debited, and no allocation occurs.
func (s *RunState) PullUncapped(out *big.Int, src string, scope string, overdraftBound *big.Int, color string) error {
	currentBal, err := s.absoluteBalance(src, scope, s.currentAsset, color)
	if err != nil {
		return err
	}

	// available = max(0, currentBal + max(0, overdraftBound))
	out.Set(currentBal)
	if overdraftBound.Sign() > 0 {
		out.Add(out, overdraftBound)
	}
	if out.Sign() < 0 {
		out.SetInt64(0)
	}

	if out.Sign() > 0 {
		amt := s.takeBig()
		amt.Set(out)
		s.pushSource(src, scope, amt, color)
		currentBal.Sub(currentBal, out) // debit in place; cache keeps the pointer
	}
	return nil
}

// Send mirrors the OCaml `send`, extended with a color filter. It drains queued
// funding sources in FIFO order until cap is satisfied or eligible sources run
// out, and each emitted posting carries the *consumed source's* own color.
//
// The color filter selects which sources are eligible:
//
//	color == nil   -> match anything; a single drain may consume and emit funds
//	                  of several colors at once. This is the mode the interpreter's
//	                  destinations use.
//	color != nil   -> only sources whose color == *color are consumed; others are
//	                  skipped and left in place (*color == "" meaning uncolored).
//
// dest == nil is the "keep/refund" path: the source is credited back and no
// posting is emitted. A partially consumed source's remainder stays at its
// position.
func (s *RunState) Send(dest *string, destScope string, cap *big.Int, color *string) error {
	// copy cap into a reusable scratch we decrement as sources are consumed; the
	// caller's cap is left untouched and no allocation is made.
	s.capScratch.Set(cap)
	cap = &s.capScratch
	asset := s.currentAsset
	i := s.head
	for cap.Sign() > 0 && i < len(s.sources) {
		s.compactAt(i) // merge the run of adjacent same-(account,scope,color) funds at i
		src := s.sources[i]
		if color != nil && src.color != *color {
			i++ // filtered out: skip, leave in place
			continue
		}
		if src.amount.Cmp(cap) >= 0 {
			if err := s.credit(dest, destScope, src, asset, cap); err != nil {
				return err
			}
			src.amount.Sub(src.amount, cap) // remainder stays in place (no alloc)
			if src.amount.Sign() == 0 {
				s.putBig(src.amount)
				s.removeAt(i)
			}
			return nil // cap fully satisfied
		}
		if err := s.credit(dest, destScope, src, asset, src.amount); err != nil {
			return err
		}
		cap.Sub(cap, src.amount)
		s.putBig(src.amount) // source fully consumed; recycle its amount
		s.removeAt(i)
		if i < s.head {
			i = s.head // consumed the front (head advanced); resume at the new front
		}
		// otherwise a mid source was removed and the tail shifted into i; re-read i
	}
	return nil
}

// SendUncapped mirrors the OCaml `send_uncapped`, extended with the same color
// filter as Send: color == nil drains every queued source (each posting keeping
// its own color); color != nil drains only matching ones, leaving others in
// place.
func (s *RunState) SendUncapped(dest *string, destScope string, color *string) error {
	asset := s.currentAsset
	i := s.head
	for i < len(s.sources) {
		s.compactAt(i) // merge the run of adjacent same-(account,scope,color) funds at i
		src := s.sources[i]
		if color != nil && src.color != *color {
			i++ // filtered out: skip, leave in place
			continue
		}
		if err := s.credit(dest, destScope, src, asset, src.amount); err != nil {
			return err
		}
		s.putBig(src.amount) // source fully consumed; recycle its amount
		s.removeAt(i)
		if i < s.head {
			i = s.head // consumed the front (head advanced); resume at the new front
		}
	}
	return nil
}

// ForcePosting records a direct movement of amount (of asset/color) from src to
// dst, bypassing the funding queue: it debits src, credits dst, and appends the
// posting. It is for movements the queue does not model — e.g. asset-scaling
// conversions (interpreter.forcePushPostingUncolored). Unlike Send it uses the
// explicit asset argument, which may differ from the current asset (a scaled
// asset). A non-positive amount is a no-op. PRE: the caller has already checked
// invariants (e.g. amount sign); no balance sufficiency check is performed.
func (s *RunState) ForcePosting(src, srcScope, dst, dstScope, asset, color string, amount *big.Int) error {
	if amount.Sign() <= 0 {
		return nil
	}
	if err := s.addToBalance(src, srcScope, asset, color, new(big.Int).Neg(amount)); err != nil {
		return err
	}
	return s.addPosting(src, srcScope, dst, dstScope, asset, color, amount) // appends the posting and credits dst
}

// Save mirrors the numscript `save` statement: it protects funds from being
// pulled later by reducing the (account, asset, color) balance, floored at zero.
//
//	amount != nil -> balance = max(0, balance - amount)   (PRE: amount >= 0)
//	amount == nil -> "save all": a positive balance becomes 0; a negative
//	                 balance is left unchanged (= min(balance, 0))
func (s *RunState) Save(account, scope, asset, color string, amount *big.Int) error {
	cur, err := s.absoluteBalance(account, scope, asset, color)
	if err != nil {
		return err
	}
	// mutate the cached balance in place (it is runtime-owned and never aliased
	// externally — GetAccountBalance hands out copies — like addToBalance).
	if amount == nil {
		if cur.Sign() <= 0 {
			return nil // negative/zero balance left unchanged
		}
		cur.SetInt64(0) // floor positive to zero
		return nil
	}
	cur.Sub(cur, amount)
	if cur.Sign() < 0 {
		cur.SetInt64(0)
	}
	return nil
}

// Snapshot returns a cheap marker of the current source-queue depth, for
// backtracking a speculative source evaluation (e.g. a `oneof` branch). It is
// just the queue length: O(1), no allocation, no map cloning.
func (s *RunState) Snapshot() int {
	// Normalize a fully-drained queue back to the front so the mark (an absolute
	// index) stays valid even though front-consumption may have left head > 0 with
	// a dead prefix. Without this a stale head==len would make Restore truncate to
	// an out-of-range mark after the next Pull rewinds the backing.
	s.rewindIfEmpty()
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
func (s *RunState) Restore(mark int) error {
	for i := mark; i < len(s.sources); i++ {
		src := s.sources[i]
		if err := s.addToBalance(src.account, src.scope, s.currentAsset, src.color, src.amount); err != nil {
			return err
		}
	}
	s.sources = s.sources[:mark]
	return nil
}

// GetPostings returns the recorded postings in a fresh slice, so callers cannot
// alter the internal queue's length/order. The posting Amounts are shared, not
// deep-cloned (avoiding an allocation per posting).
//
// LIFETIME: the shared Amounts are pooled and recycled by the next Reset of this
// RunState, so the result is valid only until then. A RunState that is never
// Reset (e.g. the tree-walker's fresh-per-run state) keeps them valid for good.
// A caller that reuses a RunState across runs (the VM) and needs to retain a
// result past the next run must deep-copy the Amounts itself.
func (s *RunState) GetPostings() []Posting {
	out := make([]Posting, len(s.postings))
	copy(out, s.postings)
	return out
}

// PostingsRef returns the internal postings slice directly, with no copy — for
// hot-loop callers that consume the result immediately. It is valid only until
// the next Reset or the next posting append (which may reallocate the slice), and
// its Amounts are pooled (recycled by the next Reset). Do not retain it. Use
// GetPostings when you need an independent slice.
func (s *RunState) PostingsRef() []Posting {
	return s.postings
}

// --- internal helpers ---

// credit routes a consumed source amount either into a posting (dest != nil) or
// back to the source as a refund (dest == nil). The funds keep their color, so
// both the posting and the destination/source balance land on (asset, color).
// amount is treated as read-only.
func (s *RunState) credit(dest *string, destScope string, src source, asset string, amount *big.Int) error {
	if dest != nil {
		return s.addPosting(src.account, src.scope, *dest, destScope, asset, src.color, amount)
	} else if amount.Sign() > 0 {
		// refund the source: consume funding, emit no posting
		return s.addToBalance(src.account, src.scope, asset, src.color, amount)
	}
	return nil
}

// entryFor returns the entry for key, creating a fresh zero-delta one (base not
// yet loaded) if absent.
func (s *RunState) entryFor(key PairKey) *balanceEntry {
	e := s.balances[key]
	if e == nil {
		e = &balanceEntry{}
		s.balances[key] = e
	}
	return e
}

// loadBase folds e's starting balance in from the Store on first need, turning a
// delta-only entry into an absolute one. Idempotent once baseLoaded is set. The
// Store is scope-agnostic; scoped balances are seeded via Prewarm, and the VM
// (the only path hitting the Store) never uses scopes.
func (s *RunState) loadBase(key PairKey, e *balanceEntry) error {
	if e.baseLoaded {
		return nil
	}
	fromStore, err := s.store.GetBalance(key.Account, key.Asset, key.Color)
	if err != nil {
		return err
	}
	if fromStore != nil {
		e.amount.Add(&e.amount, fromStore) // absolute = base + accumulated delta
	}
	e.baseLoaded = true
	return nil
}

// absoluteBalance returns a live pointer to the (account, asset, color) absolute
// balance, loading the starting balance from the Store on first access. The
// returned pointer is the live entry value — internal callers may mutate it in
// place to debit/credit; it is never aliased externally (GetAccountBalance hands
// out copies).
func (s *RunState) absoluteBalance(account, scope, asset, color string) (*big.Int, error) {
	key := PairKey{account, scope, asset, color}
	e := s.entryFor(key)
	if err := s.loadBase(key, e); err != nil {
		return nil, err
	}
	return &e.amount, nil
}

// addToBalance applies delta to (account, asset, color). It never consults the
// Store: a delta added to a not-yet-loaded entry just accumulates against the
// running delta (folded against the base later, if ever needed), and a delta
// added to a loaded entry mutates the absolute balance in place. delta is
// read-only. The error return is kept for call-site symmetry; it is always nil.
func (s *RunState) addToBalance(account, scope, asset, color string, delta *big.Int) error {
	e := s.entryFor(PairKey{account, scope, asset, color})
	e.amount.Add(&e.amount, delta)
	return nil
}

// addPosting appends a posting verbatim and credits the destination balance.
// Non-positive amounts are ignored. Postings are never merged here: same-source
// funds are instead coalesced upstream in the source queue by compactAt, so a
// posting can only ever fuse adjacent funds *within* one drain — never across
// separate sends. amount is cloned into the posting.
func (s *RunState) addPosting(src, srcScope, dst, dstScope, asset, color string, amount *big.Int) error {
	if amount.Sign() <= 0 {
		return nil
	}
	// the posting amount is recycled from the pool and reclaimed at Reset; it is a
	// distinct big.Int from amount (takeBig never returns a live one), so the
	// following in-place balance credit does not alias it.
	amt := s.takeBig()
	amt.Set(amount)
	s.postings = append(s.postings, Posting{
		Source:           src,
		SourceScope:      srcScope,
		Destination:      dst,
		DestinationScope: dstScope,
		Asset:            asset,
		Color:            color,
		Amount:           amt,
	})
	return s.addToBalance(dst, dstScope, asset, color, amount)
}

// compactAt coalesces the maximal run of funds at index i that share i's
// (account, color), folding each into s.sources[i], and drops any zero-amount
// entries it passes. It merges adjacent same-source funds in the queue before
// they are drained, so
// one drain over them yields a single posting. Because it operates on the queue
// (which each send fully consumes) and never on the posting list, it cannot fuse
// funds belonging to different sends. The fold mutates s.sources[i].amount in
// place, which is safe because queued amounts are privately owned.
func (s *RunState) compactAt(i int) {
	for i+1 < len(s.sources) {
		next := s.sources[i+1]
		if next.amount.Sign() == 0 {
			s.putBig(next.amount) // dropped; recycle
			s.removeAt(i + 1)
			continue
		}
		if next.account != s.sources[i].account || next.scope != s.sources[i].scope || next.color != s.sources[i].color {
			return
		}
		s.sources[i].amount.Add(s.sources[i].amount, next.amount)
		s.putBig(next.amount) // merged away; recycle
		s.removeAt(i + 1)
	}
}

// removeAt deletes the live source at index i, preserving the order of the rest.
// Removing the front (i == head) is O(1): just advance head, leaving behind a
// dead entry whose amount the caller has already recycled. A mid removal — only
// the rare color-skip and compaction cases — shifts the suffix down and
// truncates, as before. This makes the common front-to-back drain O(1) per pop
// instead of O(n) (an O(n^2) drain becomes O(n)).
func (s *RunState) removeAt(i int) {
	if i == s.head {
		s.head++
		return
	}
	s.sources = append(s.sources[:i], s.sources[i+1:]...)
}

// pushSource appends a funding source at the tail of the queue, first rewinding
// the backing array to the front if the queue has been fully drained — so
// front-consumption's dead prefix doesn't make the slice grow across the
// pull/send cycles of a single run.
func (s *RunState) pushSource(account, scope string, amount *big.Int, color string) {
	s.rewindIfEmpty()
	s.sources = append(s.sources, source{account, scope, amount, color})
}

// rewindIfEmpty resets the queue to index 0 when it holds no live sources
// (head == len). The dead prefix left by front-consumption holds only
// already-recycled amounts, so discarding it loses nothing and keeps the backing
// array bounded by the max concurrently-queued sources rather than the total
// pulled over the run.
func (s *RunState) rewindIfEmpty() {
	if s.head == len(s.sources) {
		s.head = 0
		s.sources = s.sources[:0]
	}
}
