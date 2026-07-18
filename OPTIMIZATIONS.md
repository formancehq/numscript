# `feat/exp/optimize-vm` — optimizations vs `feat/exp/vm`

Everything on this branch is a performance change over `feat/exp/vm`; behavior is
identical (validated by running the whole script-test corpus in both modes, see
`internal/compiler/scripts_test.go`).

Two layers, with different opt-in status:

- **Peepholes** are opt-in at compile time: `Compile` is unchanged and naive;
  `CompileWithOptimizations` runs them. Nothing else changes the bytecode.
- **Runtime & VM changes** (the allocation/queue work below) are always active —
  they speed up any compiled program, naive or optimized, since both execute on
  the same VM and `runtime.RunState`.

Most of these target the **warm** path (a reused `Vm`/`RunState` across many
runs); allocations they remove are per-run, so the win compounds with reuse.

## Peepholes (`internal/compiler/peephole*.go`)

Each is a pass `func([]vInstr) ([]vInstr, bool)`; `optimize` runs them to a
fixpoint. `defaultPeepholes()` = `monetaryFold`, `fundsBypass`,
`postFromUnbounded`, `deadCode`.

- **`monetaryFold`** — removes the `mk_monetary(A,M)` → `get_asset`/`get_amount`
  round-trip: consumers read the asset/amount registers directly. The dead
  `mk_monetary` is then dropped by `deadCode`.
- **`fundsBypass`** — skips the `runtime` funds queue for `send`s whose
  source→destination pairing is static, rewriting `pull_account` →
  `take_account` (debit, no queue) and `send_to_account` → `post_account`
  (posting, no debit). It segments the stream into send-statement regions (by
  `set_current_asset`, where the queue is provably empty) and classifies each by
  pull count P / send count S:
  - **single→single** (P=1, S=1 drain-all) — one take, one post.
  - **fan-out** (P=1, S>1) — one take, one post per destination. Restricted to
    **allotment** destinations (shares provably in `[0,got]` summing to `got`); an
    inorder dest is left to the queue because a negative `max` clause can push a
    send's cap above the available funds, which the queue clamps but a raw post
    would not.
  - **fan-in** (P>1, S=1 drain-all) — one take + one post per source. Guarded by
    statically-distinct source accounts (else the queue's `compactAt` would
    coalesce same-account funds into one posting) and no early-exit jump (an
    inorder cap-exhaustion jump would skip pulls, leaving their amounts stale) —
    so it fires for allotment sources and send-all inorder, not capped inorder.

  The true N×M case (both sides branch) keeps the queue. Impact is CPU-only
  (allocs are already ~1/op from the runtime work below): single→single ~9%,
  fan-out ~13%, fan-in ~negligible (the queue's O(1) front-pop + pooled amounts
  leave little to remove for a clean FIFO drain).
- **`postFromUnbounded`** — the aggressive follow-up to `fundsBypass` for a
  single→single send whose source is **unbounded** (`@world`, or `allowing
  unbounded overdraft`). It collapses the `take_account` +
  `check_enough_funds` + `post_account` triple into a single
  `post_from_unbounded` (`Op_PostFromUnbounded`), deleting two pieces of work
  that are pure overhead for an unbounded source:
  - the **enough-funds check** — an unbounded pull makes exactly `cap`
    available and can never be short (`got == cap == needed`), so the check
    always passes.
  - the **source-balance debit** — the `entryFor` map hit + `big.Int` subtract
    that `take_account` does only so a later `balance()` read of the source
    stays correct.

  Dropping the debit is sound only when the source balance is never observed
  afterward, enforced by two guards:
  - the program contains **no `balance()` read at all** (no `fetchBalance`) —
    the only way a script observes a running balance directly; if any exists
    the pass bails wholesale rather than reason about which account it reads.
  - a **non-`@world` source must not be pulled/taken elsewhere** in the
    program — a later *bounded* pull of the same account would fold the missing
    debit into its available-funds math and change the postings. `@world` is
    exempt: the VM always treats it as unbounded, so its debit is never folded
    into any later pull.

  The postings are unchanged (the destination is still credited by the emitted
  posting); only the unobservable source debit and the always-true check go
  away. Impact (warm, CPU-only): `send [USD/2 42] (@world → @dest)` goes
  **126.8 → 77.3 ns/op (−39%)** vs the plain `take`+`post` bypass, **−48%** vs
  naive; allocs unchanged at 1/op. Bounded sources are ineligible and
  unaffected.

  **Dead-credit elision (leaf destination).** The pass runs only when there are
  no `balance()` reads — under which the *destination credit* the emitted
  posting performs (an `addToBalance(dst)`, i.e. a balance-map `entryFor`
  lookup) is itself dead whenever `dst` is a **leaf**: never a later funding
  source (a bounded pull would fold its balance) and never `save`d. When it
  proves this (the dst account, a compile-time constant, is absent from every
  pull/take account set and every `save`), it emits `Op_PostFromUnboundedLeaf`
  instead — same posting, but `PostDirectNoCredit` skips the credit. That
  removes the `entryFor` map lookup, the single most expensive operation left in
  the fused path (~40 ns — ~half a world send, more than all the big.Int
  arithmetic combined). Impact: `@world → @dest` goes **77.3 → 44 ns/op (−43%)**,
  **−70% vs naive**; it lands *below* the crediting `DirectPost` floor (54.9)
  precisely because that floor still does the credit's lookup. Non-leaf / saved
  destinations keep the credit (plain `Op_PostFromUnbounded`).
- **`deadCode`** — drops pure instructions (loads, arithmetic) whose result is
  never read. Cleans up after the others.

## VM opcodes added (`internal/vm/instruction.go`)

Compact / specialized forms to cut per-instruction work:

- **`Op_LoadIntImm`** — small unsigned int constants (≤ u16) encoded inline in the
  instruction word; no const-pool entry, no `big.Int` copy on load.
- **`Op_PullAccountCapZero`** — single-word pull for the common plain-account case
  (cap present, overdraft 0, no color; world stays unbounded), vs the 2-word
  general `Op_PullAccount`.
- **`Op_Take` / `Op_TakeCapZero` / `Op_Post`** — the `fundsBypass` targets. Take =
  Pull without the queue append; Post = a direct posting with no debit.
  `Op_TakeCapZero` is the single-word plain-account form.
- **`Op_PostFromUnbounded`** — the `postFromUnbounded` target. Single word
  (`A=src`, `B=dst`, `C=cap`, same layout as `Op_Post`): emits the posting
  `src → dst` of `cap` in the current asset and credits `dst`, with **no source
  debit and no funds check** — the fused unbounded-source fast path.
- **`Op_PostFromUnboundedLeaf`** — same as `Op_PostFromUnbounded` but for a
  **leaf** destination: also skips the `dst` credit (and its balance-map lookup)
  via `runtime.PostDirectNoCredit`. Emitted only when the credit is provably
  dead.
- **`Op_TakeCapZeroSlot` / `Op_PostSlot`** — two-word `Op_TakeCapZero` /
  `Op_Post` carrying a compile-assigned **balance slot** in the second word
  (`ext.A`). The slot indexes `runtime`'s `balanceSlots` array instead of hashing
  the `PairKey` balance map. Emitted by `assignBalanceSlots` for constant
  `(account, asset)` accesses.

## Balance slots (`internal/compiler/assign_balance_slots.go`, `internal/runtime/slots.go`)

The balance store keys entries by a 4-string `PairKey{account, scope, asset,
color}` map. That lookup (`entryFor`) is ~40 ns — about **half a warm send**, more
than all the big.Int arithmetic combined. `assignBalanceSlots` (a post-peephole
pass, optimized build only) turns the common case into an array index:

- It assigns each compile-time-constant `(account, asset)` pair a dense integer
  **slot**, tracking the current asset from the enclosing `set_current_asset`
  (also a constant). The two ops a single→single bypass lowers to — the
  bounded-zero `take_account` (source debit) and `post_account` (destination
  credit) — are annotated and assembled as `Op_TakeCapZeroSlot` / `Op_PostSlot`.
- `runtime.entryForSlot(slot, key)` indexes a `[]*balanceEntry` (`balanceSlots`)
  instead of the map. The slot table persists across `Reset` (entries are pooled,
  never freed; `freshen` re-stamps on first touch per generation).

**Soundness with runtime-resolved (variable) accounts.** Accounts may only be
known at runtime (from vars), so slots cannot *replace* the map. Instead a slot
**caches the same `*balanceEntry` the map holds** — it is populated by going
through `entryFor` the first time — so a slotted (constant) access and a dynamic
(var) access that resolve to the same `(account, asset)` share one entry, and
balances stay coherent. Dynamic accounts/assets and every un-slotted op (`pull`,
`send`, `save`, `balance`) keep hitting the map against those same entries. A
variable account or asset, or more than 254 distinct pairs, leaves the access
dynamic (map).

- Impact (warm): `send [USD/2 10] (@src → @dest)` goes **156 → 119 ns/op (−24%)**
  — the source read+debit and the destination credit each drop from a map lookup
  (~28 ns/key) to an array index (~7 ns/key). The slotted opt now runs *below* the
  map-based `RuntimeBaseline` floor (143 ns). Allocs unchanged at 1/op.
- **Scope / not-yet:** only the `take`/`post` bypass ops are slotted; the
  queue path (`pull`/`send`, used by capped inorder and true N×M sends) still
  uses the map, so those shapes are unchanged. Slotting them is the natural next
  step. Coherence is covered by `TestE2E_SlotCoherence` and the corpus in both
  modes.

## Store adapter reuse (`internal/vm/vm.go`) — the last allocation

The warm path sat at 1 alloc/op long after the funds work was pooled: `Exec`
boxed a fresh `runtimeStoreAdapter` **value** into `RunState`'s `Store` interface
field every call (that field outlives the call — the runstate fetches balances
lazily through it, once per distinct account per run). The fix keeps one adapter
on the `Vm` and hands it to the runstate **by pointer**: boxing a
`*runtimeStoreAdapter` into an interface stores the pointer in the interface word
(no heap copy), so a warm `Exec` allocates nothing for the store plumbing. The
`ctx`/`store` fields are refreshed in place each call (safe for the
single-threaded, one-`Exec`-at-a-time VM); this also wires `ctx` through, which
was previously dropped. Impact: **1 → 0 alloc/op**, and removing even the 32 B
malloc is worth real time (`world → dest` 44 → 30 ns, −32%; simple send
119 → 103 ns, −13%).

## Runtime (`internal/runtime/`)

The bulk of the win is here: driving per-run heap allocation on the hot (warm,
reused-VM) path **to zero** (see the store-adapter fix above for the final
alloc). Grouped by what they attack.

### Funds queue (`runtime.go`)

- **`free` pool (`takeBig`/`putBig`)** — recycles `big.Int`s across `Reset`:
  queued-source amounts (reclaimed when consumed/merged/dropped) and posting
  amounts (reclaimed at `Reset`).
- **`head` FIFO index** — the source queue consumes from the front by `head++`
  instead of shifting the slice, so a front-to-back drain is O(1) per pop (an
  O(n²) drain becomes O(n)); `rewindIfEmpty` reclaims the dead prefix.
- **`capScratch`** — a reusable `big.Int` for `Send`'s decrementing cap, so a
  capped send allocates nothing.
- **`PostingsRef()`** — returns the internal postings slice with no copy, for
  hot-loop callers that consume it immediately (vs `GetPostings`, which copies).

### Balance cache — generation-stamped (`runtime.go`)

- Balances live in `map[PairKey]*balanceEntry`. `Reset` bumps a **generation
  counter** (`s.gen`) instead of `clear`-ing the map; each `balanceEntry` records
  the gen it was last touched in and is **lazily reset on first access** in a new
  run (`freshen`), reusing the struct *and* its embedded `big.Int` backing array.
  So a reused VM re-allocates ~no balance entries for recurring accounts — where
  the old code allocated one `balanceEntry` per account per run.
- Impact (warm): single→single **5 → 1 alloc/op** (~−27% ns); a 3-source send
  **10 → 1 alloc/op**.
- **Caveat / prototype status:** no eviction yet — the map grows with distinct
  accounts across runs. Fine for a hot/recurring working set; an all-unique-account
  workload regresses to ~baseline allocs *and* grows memory unbounded. A production
  version needs a size cap (the gen doubles as an LRU clock; pin the
  compiler-known static keys as the floor).

### Portions & allotments — integer, no `big.Rat` (`allotment.go`, `runtime.go`)

- **`runtime.Portion{Num, Den big.Int}`** (unreduced, `Den > 0`) replaces
  `big.Rat` in the VM's `portionsRegs`. The portion ops (`MkPortion`,
  `SubPortion`, `PortionCopy`) run integer-only and **in place** (reusing the
  register's `big.Int` backing; `SubPortion` uses a VM `portScratch [2]big.Int`).
  Reduction is never needed — `floor(amount * Num / Den)` is invariant under it.
- **`MakeAllotment`** is a method on `RunState` using reusable scratch
  (`allotTotal`/`allotTmp`) and integer `floor(amount*Num/Den)` — the old free
  function allocated a `big.Rat` per portion.
- Rare paths preserved: `PortionToString` reduces via `big.Rat` for canonical
  output; `MetaPortion` converts `ParsePortion`'s `*big.Rat` (which stays
  `*big.Rat` — it is shared with the interpreter and the vars encoder).
- Impact (warm): allotment fan-out `{1/2 to @a; 1/2 to @b}` went **31 allocs /
  ~1100 ns → 1 alloc / ~390 ns** (this change removed the 30 `big.Rat` allocs;
  `fundsBypass` then removed the queue for the remaining CPU).
- **Caveat:** unreduced denominators grow multiplicatively across a `sub_portion`
  chain (the `1 − p₁ − p₂ …` leftover). Bounded/tiny for realistic allotments;
  a production version could reduce periodically if it ever mattered.

## Not yet done (candidates)

- **Right-size the `Vm` register banks** — `NewVm` allocates fixed `[256]` banks
  (`[256]big.Rat` alone ≈16KB, ~50KB total); this dominates the *cold* path. The
  assembler already knows the exact per-bank count (`regPool.next`); threading it
  into `Program` would size them to fit. (Warm reuse already amortizes this.)
- **`balanceEntry` cache eviction** — the generation cache never evicts, so an
  all-unique-account workload grows the map unbounded. A production version needs
  a size cap (the generation doubles as an LRU clock; pin the compiler-known
  static keys as the floor).

## Exposure

- `compiler.CompileWithOptimizations(program)` → `(VarsEncoder, vm.Program, error)`
  is the only public entry point for the optimized build. It returns assembled
  bytecode; the intermediate virtual instructions (and the peephole passes) are
  package-private, so seeing the fused vInstr stream requires an in-package test
  (e.g. `TestOptimizedSimpleProgram` in `compiler_test.go`).
