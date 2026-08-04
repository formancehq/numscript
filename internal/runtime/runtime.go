// Package runtime is the funds engine shared by the VM and the tree-walking
// interpreter: per-(account, scope, asset, color) balances, a FIFO queue of
// funding sources fed by Pull/PullUncapped, and the postings produced by
// Send/SendUncapped.
//
// Balances load lazily. An entry holds only the net delta applied this run until
// an operation needs the absolute value (a bounded pull, a send-all, a balance()
// read); the Store is consulted then and the starting balance folded in. An
// unbounded pull (nil overdraft) makes cap available regardless of balance, so it
// never triggers a fetch — that is what keeps send-from-@world free of Store
// round-trips. Nothing here knows the name "world": callers decide which accounts
// are unbounded and pass a nil overdraft.
//
// Color "" means uncolored. Pull tags queued funds with a color; Send drains only
// matching sources, preserving the position of the rest.
//
// A *RunState is NOT safe for concurrent use. Use one per execution.
//
// Amounts are *big.Int, a mutable reference type, so this package clones values
// it ingests from the Store and amounts it intends to mutate, only mutates
// big.Ints it privately owns (queued source amounts), and never hands out a live
// reference to internal state.
package runtime

import (
	"errors"
	"math/big"
)

// ErrNoOpenMark is the only error MarkEnd returns. Well-formed bytecode matches
// pushes to ends, so it only surfaces for hand-written IR or a hand-crafted .numb.
var ErrNoOpenMark = errors.New("runtime: no open mark to end")

// Store supplies the starting balance for an (account, asset, color) triple.
// Implementations should return 0 (or nil, treated as 0) for unknown triples, not
// an error. The returned *big.Int is cloned on ingest, so the Store may reuse it.
type Store interface {
	GetBalance(account, asset, color string) (*big.Int, error)
}

// Posting is aliased as the interpreter's public Posting type, so the json tags
// define the public ledger serialization contract — keep field names and order
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

// PairKey identifies a balance slot. Scope is a second dimension of the account
// (scoped accounts hold separate balances); the VM leaves it empty.
type PairKey struct {
	Account string
	Scope   string
	Asset   string
	Color   string
}

// source is a funding entry queued by Pull / PullUncapped. The amount is
// privately owned by the queue and may be mutated in place.
type source struct {
	account string
	scope   string
	amount  *big.Int
	color   string
}

// While baseLoaded is false, amount is the net delta applied this run and the
// Store has not been consulted; once true, the starting balance has been folded
// in and amount is the absolute running balance.
type balanceEntry struct {
	amount     big.Int
	baseLoaded bool
}

// mark is one open region: the source-queue depth and the posting-list length at
// the matching MarkPush. A rewinding MarkEnd rolls both back to it.
type mark struct {
	sources  int32
	postings int32
}

// The zero value is not usable; call New.
type RunState struct {
	store        Store
	balances     map[PairKey]*balanceEntry
	sources      []source // FIFO: front = index 0
	postings     []Posting
	currentAsset string

	// marks is the stack of open marks. LIFO, so the innermost open region is the
	// last element.
	marks []mark
}

// New creates an empty RunState backed by store.
func New(store Store) *RunState {
	return &RunState{
		store:    store,
		balances: make(map[PairKey]*balanceEntry),
	}
}

// SetCurrentAsset sets the asset used when an operation omits one.
//
// PRE: no mark is open (see HasOpenMark). A rewinding MarkEnd repays queued funds
// into the current asset's balance, so changing the asset mid-region would repay
// into the wrong one.
func (s *RunState) SetCurrentAsset(asset string) {
	s.currentAsset = asset
}

// Reset clears all per-execution state and rebinds the store, retaining
// map/slice capacity so a RunState can be reused across executions. Any mark left
// open by a run that failed mid-region is dropped.
//
// GetPostings returns copies, so a result obtained before Reset stays valid.
func (s *RunState) Reset(store Store) {
	s.store = store
	clear(s.balances)
	s.sources = s.sources[:0]
	s.postings = s.postings[:0]
	s.marks = s.marks[:0]
	s.currentAsset = ""
}

// Prewarm seeds the balance cache from a bulk fetch, so the lazy per-key
// Store.GetBalance path is never hit for those keys.
//
// Call it once, before any Pull/Send/Save/ForcePosting. Amounts are cloned. A key
// whose base is already loaded is left untouched (the live value wins), so a
// stray double-call cannot clobber computed state; a key that only holds a delta
// has the base folded into it here.
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
// loaded. An entry that only holds a delta so far reports false, so a caller
// still fetches and folds in the base.
func (s *RunState) Has(account, scope, asset, color string) bool {
	e := s.balances[PairKey{account, scope, asset, color}]
	return e != nil && e.baseLoaded
}

type AccountBalance struct {
	Asset  string
	Color  string
	Amount *big.Int
}

// AccountBalances returns copies of every tracked balance entry for account, with
// starting balances folded in so the amounts are absolute. It only reports
// triples already touched this run — it does not enumerate the Store — so an
// account never prewarmed or touched yields an empty slice.
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

// GetAccountBalance returns a fresh copy of the balance for (account, asset,
// color), fetching from the Store on first access.
//
// "" is the unset sentinel for asset, meaning "use currentAsset"; a real asset is
// never the empty string. For color, "" is a legitimate value: uncolored.
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

// Pull debits up to cap from src's (currentAsset, color) balance, queues the
// pulled amount as a funding source tagged with color, and writes the amount made
// available into out:
//
//	overdraft == nil -> unbounded: available = max(0, cap)
//	overdraft == b   -> available = min(max(0, balance + max(0,b)), max(0, cap))
//	                    (pass big.NewInt(0) for the "balance only" default)
//
// out is overwritten and may be any addressable *big.Int (e.g. a VM register).
// cap and overdraft are not mutated. The only allocation per call is the queued
// source's own copy of the amount, which must outlive out.
func (s *RunState) Pull(out *big.Int, src string, scope string, cap *big.Int, overdraft *big.Int, color string) error {
	if overdraft == nil {
		// available = max(0, cap), independent of the balance, so no Store fetch:
		// the debit is recorded as a delta and folded in only if later needed.
		out.Set(cap)
		if out.Sign() < 0 {
			out.SetInt64(0)
		}
		amt := new(big.Int).Set(out)
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

	// an independent copy: out stays the caller's, while the queued amount is
	// mutated in place by compactAt/Send
	amt := new(big.Int).Set(out)
	s.sources = append(s.sources, source{src, scope, amt, color})

	currentBal.Sub(currentBal, out)
	return nil
}

// PullUncapped makes available max(0, balance + max(0, overdraftBound)) of src's
// (currentAsset, color) balance, queuing it only when positive, and writes the
// available amount into out. As in Pull a negative overdraftBound is clamped to 0,
// so it never eats into a positive balance; pass big.NewInt(0) for the "balance
// only" default. overdraftBound is not mutated.
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
		amt := new(big.Int).Set(out)
		s.sources = append(s.sources, source{src, scope, amt, color})
		currentBal.Sub(currentBal, out) // debit in place; cache keeps the pointer
	}
	return nil
}

// Send drains queued funding sources in FIFO order until cap is satisfied or
// eligible sources run out. Each emitted posting carries the *consumed source's*
// own color.
//
//	color == nil   -> match anything; one drain may consume and emit funds of
//	                  several colors at once (the interpreter's destination mode)
//	color != nil   -> only sources whose color == *color are consumed; others are
//	                  skipped and left in place (*color == "" meaning uncolored)
//
// dest == nil is the keep/refund path: the source is credited back and no posting
// is emitted. A partially consumed source's remainder stays at its position.
//
// PRE: no mark is open (see HasOpenMark). Draining consumes sources from the
// front, renumbering the queue and leaving an open mark at the wrong boundary.
func (s *RunState) Send(dest *string, destScope string, cap *big.Int, color *string) error {
	cap = new(big.Int).Set(cap) // clone: we decrement it as sources are consumed
	asset := s.currentAsset
	i := 0
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
			if diff := new(big.Int).Sub(src.amount, cap); diff.Sign() > 0 {
				s.sources[i].amount = diff // remainder stays at this position
			} else {
				s.removeAt(i)
			}
			return nil // cap fully satisfied
		}
		if err := s.credit(dest, destScope, src, asset, src.amount); err != nil {
			return err
		}
		cap.Sub(cap, src.amount)
		s.removeAt(i) // do not advance i; the next source shifts into position i
	}
	return nil
}

// SendUncapped applies the same color filter as Send: color == nil drains every
// queued source (each posting keeping its own color); color != nil drains only
// matching ones, leaving others in place.
//
// PRE: no mark is open, for the same reason as Send.
func (s *RunState) SendUncapped(dest *string, destScope string, color *string) error {
	asset := s.currentAsset
	i := 0
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
		s.removeAt(i)
	}
	return nil
}

// ForcePosting moves amount from src to dst bypassing the funding queue, for
// movements the queue does not model (e.g. asset-scaling conversions). Unlike Send
// it uses the explicit asset argument, which may differ from the current asset. A
// non-positive amount is a no-op; no balance sufficiency check is performed.
//
// Safe inside a region, unlike Send: it touches no queue entry, so a rewinding
// MarkEnd undoes it completely by reversing the posting.
func (s *RunState) ForcePosting(src, srcScope, dst, dstScope, asset, color string, amount *big.Int) error {
	if amount.Sign() <= 0 {
		return nil
	}
	if err := s.addToBalance(src, srcScope, asset, color, new(big.Int).Neg(amount)); err != nil {
		return err
	}
	return s.addPosting(src, srcScope, dst, dstScope, asset, color, amount) // appends the posting and credits dst
}

// Save implements the numscript `save` statement: it protects funds from being
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

// --- marks ---
//
// A region lets a caller try a source evaluation and undo it if it did not work
// out — the `oneof` shape: pull from a branch, and if the branch did not cover the
// cap, repay what it pulled and try the next one.
//
// The mark is a source-queue depth that the caller never holds: RunState owns a
// LIFO of them, so a caller cannot name a depth that was never marked, restore one
// out of order, or restore one twice. The queue can therefore never be truncated
// to a bogus depth.
//
// A depth only stays meaningful while nothing consumes the queue from the front
// and the asset a repay lands on is fixed, which Send, SendUncapped and
// SetCurrentAsset document as a precondition. That half is not enforced here; the
// VM checks it at Op_SendToAccount and Op_SetCurrentAsset.
//
// A rewind undoes pulls and postings, so a ForcePosting inside a region is rolled
// back. It does NOT undo queue *consumption*, which is why Send is refused inside
// a region; see the INVARIANT on MarkEnd.
//
// There is no "rewind but keep the mark": a retry closes with rewind and pushes
// again. Pushes and ends therefore match strictly, so "a region left open after a
// rewind" is not a state to detect but one that cannot be encoded. Neither op
// carries an operand, so mark depth is a function of the instruction stream alone
// and an IR verifier could prove balance statically. No such pass exists yet.

func (s *RunState) HasOpenMark() bool {
	return len(s.marks) > 0
}

// MarkPush opens a region at the current source-queue depth and posting-list
// length. Allocation-free once the stack has reached its high-water mark, since
// the backing array is retained across Reset.
func (s *RunState) MarkPush() {
	s.marks = append(s.marks, mark{
		sources:  int32(len(s.sources)),
		postings: int32(len(s.postings)),
	})
}

// MarkEnd closes the innermost region, always popping the mark.
//
// rewind == false commits: whatever was pulled stays queued and whatever was
// posted stays posted.
//
// rewind == true rolls the region back in two parts. Postings emitted since the
// MarkPush are reversed and dropped; each posting records its own asset, so this
// part does not depend on currentAsset. Sources still queued above the mark are
// repaid to the balance they were debited from, then the queue is truncated.
// compactAt may have folded funds, but the fold preserves (account, scope, color),
// so the repay still lands correctly. The two parts are both additive balance
// adjustments, so their order is conventional rather than required.
//
// INVARIANT: no Send may have run inside the region. Reversing a posting credits
// its source back, which is only correct because the queue entry that funded it was
// pulled *inside* the region and is therefore not also repaid by the second loop. A
// Send can consume entries queued *below* the mark, and for those the credit and
// the repay would both apply — inventing money. Hence Op_SendToAccount is refused
// while a mark is open; relaxing that requires undoing queue consumption too.
func (s *RunState) MarkEnd(rewind bool) error {
	if len(s.marks) == 0 {
		return ErrNoOpenMark
	}
	m := s.marks[len(s.marks)-1]
	s.marks = s.marks[:len(s.marks)-1]

	if !rewind {
		return nil
	}

	for i := len(s.postings) - 1; i >= int(m.postings); i-- {
		p := s.postings[i]
		s.subFromBalance(p.Destination, p.DestinationScope, p.Asset, p.Color, p.Amount)
		_ = s.addToBalance(p.Source, p.SourceScope, p.Asset, p.Color, p.Amount)
	}
	s.postings = s.postings[:m.postings]

	for i := int(m.sources); i < len(s.sources); i++ {
		src := s.sources[i]
		_ = s.addToBalance(src.account, src.scope, s.currentAsset, src.color, src.amount)
	}
	s.sources = s.sources[:m.sources]
	return nil
}

// GetPostings returns a fresh slice, so callers cannot alter the internal
// length/order. Posting amounts are write-once — addPosting clones on append and
// never mutates an existing posting — so the *big.Int values are shared rather
// than deep-cloned.
func (s *RunState) GetPostings() []Posting {
	out := make([]Posting, len(s.postings))
	copy(out, s.postings)
	return out
}

// --- internal helpers ---

// credit routes a consumed source amount either into a posting (dest != nil) or
// back to the source as a refund (dest == nil). amount is read-only.
func (s *RunState) credit(dest *string, destScope string, src source, asset string, amount *big.Int) error {
	if dest != nil {
		return s.addPosting(src.account, src.scope, *dest, destScope, asset, src.color, amount)
	} else if amount.Sign() > 0 {
		// refund the source: consume funding, emit no posting
		return s.addToBalance(src.account, src.scope, asset, src.color, amount)
	}
	return nil
}

// entryFor creates a fresh zero-delta entry (base not yet loaded) if absent.
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
// Store is scope-agnostic: scoped balances are seeded via Prewarm, and the VM (the
// only path hitting the Store) never uses scopes.
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

// absoluteBalance returns a live pointer into the entry, loading the starting
// balance from the Store on first access. Internal callers may mutate it in place
// to debit/credit; it is never aliased externally.
func (s *RunState) absoluteBalance(account, scope, asset, color string) (*big.Int, error) {
	key := PairKey{account, scope, asset, color}
	e := s.entryFor(key)
	if err := s.loadBase(key, e); err != nil {
		return nil, err
	}
	return &e.amount, nil
}

// addToBalance never consults the Store: a delta on a not-yet-loaded entry just
// accumulates (folded against the base later, if ever needed), and a delta on a
// loaded entry mutates the absolute balance in place. delta is read-only. The
// error return is kept for call-site symmetry; it is always nil.
func (s *RunState) addToBalance(account, scope, asset, color string, delta *big.Int) error {
	e := s.entryFor(PairKey{account, scope, asset, color})
	e.amount.Add(&e.amount, delta)
	return nil
}

// subFromBalance is addToBalance with the sign flipped, so a caller undoing a
// movement does not have to allocate a negated copy of the amount. delta is
// read-only.
func (s *RunState) subFromBalance(account, scope, asset, color string, delta *big.Int) {
	e := s.entryFor(PairKey{account, scope, asset, color})
	e.amount.Sub(&e.amount, delta)
}

// addPosting appends a posting and credits the destination balance. Non-positive
// amounts are ignored. Postings are never merged here: same-source funds are
// coalesced upstream in the queue by compactAt, so a posting can only fuse
// adjacent funds *within* one drain, never across separate sends. amount is cloned
// into the posting.
func (s *RunState) addPosting(src, srcScope, dst, dstScope, asset, color string, amount *big.Int) error {
	if amount.Sign() <= 0 {
		return nil
	}
	s.postings = append(s.postings, Posting{
		Source:           src,
		SourceScope:      srcScope,
		Destination:      dst,
		DestinationScope: dstScope,
		Asset:            asset,
		Color:            color,
		Amount:           new(big.Int).Set(amount),
	})
	return s.addToBalance(dst, dstScope, asset, color, amount)
}

// compactAt coalesces the maximal run of funds at index i sharing i's (account,
// scope, color) into s.sources[i], dropping any zero-amount entries it passes, so
// one drain over them yields a single posting. It operates on the queue and never
// on the posting list, so it cannot fuse funds belonging to different sends. The
// fold mutates s.sources[i].amount in place, safe because queued amounts are
// privately owned.
func (s *RunState) compactAt(i int) {
	for i+1 < len(s.sources) {
		next := s.sources[i+1]
		if next.amount.Sign() == 0 {
			s.removeAt(i + 1)
			continue
		}
		if next.account != s.sources[i].account || next.scope != s.sources[i].scope || next.color != s.sources[i].color {
			return
		}
		s.sources[i].amount.Add(s.sources[i].amount, next.amount)
		s.removeAt(i + 1)
	}
}

// removeAt preserves the order of the remaining sources.
func (s *RunState) removeAt(i int) {
	s.sources = append(s.sources[:i], s.sources[i+1:]...)
}
