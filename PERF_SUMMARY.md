# Performance summary — `feat/exp/optimize-vm`

Wrap-up of the compiler/VM/runtime performance work on this branch, measured
against the base VM branch `feat/exp/vm`. For the *mechanism* behind each number
see [`OPTIMIZATIONS.md`](./OPTIMIZATIONS.md); this doc is just the scoreboard.

## Setup

- **Machine:** Apple M3 Pro (`darwin/arm64`), `cpu: Apple M3 Pro`.
- **Command:** `go test ./internal/compiler/ -bench=… -benchmem -run=^$ -count=6`
- **Numbers are medians** of 6 runs. All benchmarks measure the **warm** path: a
  reused `Vm` / `RunState` across iterations (parse/compile/assemble happen once,
  up front), so what's timed is a single `Exec`.
- Source (`internal/compiler/bench_test.go`, `bench_world_test.go`):
  - **simple** — `send [USD/2 10] (source = @src, destination = @dest)`
  - **capped** — inorder `{ @a ; max [USD/2 5] from @b ; @c } → @dest`
  - **world** — `send [USD/2 42] (source = @world, destination = @dest)`

Three reference points per script:
- **TreeWalker** — the tree-walking interpreter on a pre-parsed AST.
- **CompiledVM / …Opt** — compiled bytecode on the register VM (naive vs
  `CompileWithOptimizations`).
- **RuntimeBaseline** — the **floor**: `runtime.RunState` driven directly, doing
  exactly the funds ops the script lowers to, with no AST walk and no bytecode
  dispatch. The gap VM→floor is dispatch overhead; the gap interpreter→floor is
  front-end overhead.

## Headline: base VM branch vs this branch (warm)

| Script (warm) | `feat/exp/vm` | `feat/exp/optimize-vm` | Δ ns | allocs |
|---|---:|---:|---:|---:|
| simple `send` (VM) | 302.5 ns | **103 ns** (opt+slots) | **−66%** | 10 → **0** |
| capped inorder (VM) | 853.3 ns | **510 ns** (opt) | **−40%** | 23 → **0** |
| world → dest (VM) | ~queue, 10 allocs | **30 ns** (opt, leaf) | — | 10 → **0** |

The dominant win is **allocations → 0/op** on the warm path (runtime rewrite +
the reused store adapter), compounded by the peepholes (opt-in), the unbounded
fast path, and balance slots.

**The last allocation.** The warm path sat at 1 alloc/op for a while: `Exec`
boxed a fresh `runtimeStoreAdapter` value into `RunState`'s `Store` interface
field every call (that field outlives the call — the runstate fetches balances
lazily). Reusing one adapter on the `Vm` and handing it over **by pointer**
(boxing a pointer into an interface stores it in the interface word, no heap
copy) drops it to **0 alloc/op** — and removing even a 32 B malloc is worth
real time: `world → dest` 44 → 30 ns (−32%), simple send 119 → 103 ns (−13%).

## VM vs interpreter vs floor — `feat/exp/optimize-vm`

Current warm-path state (all figures post the balance-slots and store-adapter
work). The `CompiledVM*` rows run through `Exec` and now allocate nothing; the
hand-written `*Baseline` floors still box their own store adapter (1 alloc).

**simple `send`**

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| TreeWalker (interpreter) | 1603 | 2705 | 47 |
| RuntimeBaseline (map-based floor) | 143 | 16 | 1 |
| CompiledVM (naive) | 166 | 0 | 0 |
| **CompiledVMOpt (peephole + slots)** | **103** | **0** | **0** |

The slotted opt (`103 ns`) runs *below* the `RuntimeBaseline` "floor" (`143 ns`):
that floor still keys the balance map by a 4-string `PairKey` and still boxes its
own store adapter (1 alloc), while the optimized VM indexes a small array and
reuses its adapter (0 alloc). Once you replace the map, the old floor stops being
a floor.

**capped inorder**

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| TreeWalkerCapped | 3812 | 6019 | 97 |
| RuntimeBaselineCapped (floor) | 452.9 | 16 | 1 |
| CompiledVMCapped (naive) | 534 | 0 | 0 |
| **CompiledVMOptCapped (peephole)** | **510** | **0** | **0** |

Capped still runs through the funds queue (`pull`/`send`), so it is not slotted
yet — its win is the alloc removal and the peepholes, not slots.

**world → dest**

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| WorldNaive (VM, queue) | 149.0 | 0 | 0 |
| **WorldOpt (fused leaf op + slots + 0 alloc)** | **30** | **0** | **0** |
| WorldBaselineTakePost (floor, debit kept) | 95.2 | 16 | 1 |
| WorldBaselineDirectPost (floor, debit skipped) | 54.9 | 16 | 1 |

`30 ns` is far below even the `DirectPost` floor (`54.9`): the leaf op elides the
destination credit's balance-map lookup, and the reused store adapter drops the
last allocation — so only the posting append and dispatch remain.

For reference, the same three on the base branch `feat/exp/vm`:

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| TreeWalker | 1635 | 2609 | 48 |
| RuntimeBaseline | 253.8 | 320 | 10 |
| CompiledVM | 302.5 | 336 | 10 |
| TreeWalkerCapped | 3997 | 6003 | 102 |
| RuntimeBaselineCapped | 744.9 | 848 | 23 |
| CompiledVMCapped | 853.3 | 864 | 23 |

## Front-end overhead: interpreter vs VM

The interesting quantity is *front-end overhead* — the cost each engine adds on
top of the shared funds-work floor. Subtract the floor from each engine:

**Front-end cost = engine − floor** (simple `send`, floor ≈ 143 ns):

| Engine | total | − floor | allocs over floor |
|---|---:|---:|---:|
| TreeWalker (interpreter) | 1603 | **~1460 ns** | +46 |
| CompiledVM (naive) | 166 | **~20 ns** | +0 |

The interpreter's AST walk adds **~1460 ns and ~46 allocations** per send; the
VM's dispatch + registers add only **tens of ns and zero allocations** — roughly
**two orders of magnitude** lighter per instruction. That gap is the whole reason
to run bytecode instead of walking the tree.

The optimized VM goes further still: with balance slots it now runs **below** the
map-based floor (`103 < 143 ns`), because the floor itself still pays the balance
map lookup *and* a store-adapter alloc that the optimized VM has eliminated. So
`engine − floor` only characterizes the naive VM and the interpreter — once slots
replace the map, the floor stops being a floor and the subtraction goes negative.

Cross-branch, the per-instruction VM front-end barely moved (it was always cheap);
what changed is the **floor** — the shared funds work — which the runtime rewrite
roughly halved (simple `253.8 → 143`, capped `744.9 → 452.9`) and, with the reused
store adapter, drove to **0 allocations** (`10 → 0` / `23 → 0`).

## The unbounded-source fast path (`postFromUnbounded`)

The single biggest per-shape win on this branch. For a single→single send from an
unbounded source (`@world` or `allowing unbounded overdraft`), the fused
`Op_PostFromUnbounded` drops the always-true enough-funds check *and* the
(unobservable) source-balance debit, emitting the posting directly:

| world → dest (warm VM) | ns/op | vs naive |
|---|---:|---:|
| WorldNaive (queue) | 149.0 | — |
| WorldOpt, prior `take`+`post` bypass | 126.8 | −15% |
| WorldOpt, fused `Op_PostFromUnbounded` (credits dst) | 77.3 | −48% |
| WorldOpt, `Op_PostFromUnboundedLeaf` (leaf dst) | 44 | −70% |
| **WorldOpt + reused store adapter (0 alloc)** | **30** | **−80%** |

Successive wins on the same shape:

- **`Op_PostFromUnbounded`** drops the always-true enough-funds check and the
  (unobservable) source-balance debit: `126.8 → 77.3` (−39%).
- **`Op_PostFromUnboundedLeaf`** additionally drops the *destination credit* when
  dst is a **leaf** (never a later funding source, never saved, no `balance()`
  read anywhere) — which removes the `entryFor` balance-map lookup, the single
  most expensive operation left in the path (~40 ns): `77.3 → 44` (−43%).
- **Reused store adapter** removes the last allocation: `44 → 30` (−32%).

`30 ns` is far *below* the crediting `DirectPost` floor (`54.9`): the leaf op
skips the credit's map lookup and the VM no longer boxes a store adapter, so only
the posting append and instruction dispatch remain. Bounded sources and non-leaf /
saved / balance-read destinations are ineligible (guarded) and unchanged. This
matters because funding `@world → external leaf` is the most common numscript
shape.

## Balance slots (`assignBalanceSlots` + `Op_*Slot`)

The structural attack on the balance map. The compiler assigns each constant
`(account, asset)` a dense integer **slot**; the two ops a single→single bypass
lowers to — the bounded-zero `take_account` (source debit) and `post_account`
(destination credit) — get two-word `Op_TakeCapZeroSlot` / `Op_PostSlot` forms
carrying the slot. The runtime indexes a small `[]*balanceEntry` instead of
hashing a 4-string `PairKey`.

| simple `send` (warm VM) | ns/op |
|---|---:|
| CompiledVMOpt (map) | 156 |
| CompiledVMOpt + slots | 119 |
| **+ reused store adapter (0 alloc)** | **103** |

Slots: `−24%` — the two map lookups (source read+debit, destination credit) drop
to array indexes (~40 ns → ~6 ns, matching the isolated micro-benchmark: map
~28 ns/key vs slot ~7 ns/key). The reused store adapter then removes the last
allocation (`119 → 103`). Capped (queue-based `pull`/`send`, not `take`/`post`)
is not yet slotted; `@world → leaf` has no balance access and is unaffected by
slots.

**Coherence with variables (the subtle part).** Accounts can be resolved at
runtime (from vars), so slots cannot simply *replace* the map. Instead each slot
**caches the same `*balanceEntry` the map holds** (populated by going through the
map the first time), so a slotted (constant-account) access and a dynamic (var)
access that resolve to the same account share one entry — balances stay coherent.
Dynamic accounts, dynamic assets, and un-slotted ops (pulls, sends, saves) all
keep using the map against those same entries. Verified by `TestE2E_SlotCoherence`
(a slotted read seeing a map-path credit, across a reused warm VM) and the whole
corpus in both modes.

## Takeaways

1. **Allocations are the story.** Warm path `10 → 0` (simple) and `23 → 0`
   (capped): the pooling drove it to 1/op, and reusing the store adapter (passed
   by pointer) removed the last one. This is ~half the ns win by itself and is
   *always on* (naive or optimized bytecode).
2. **The VM front-end was never the bottleneck** — it's ~two orders of magnitude
   lighter than the interpreter's AST walk (which also does ~46 allocs/send). The
   runtime funds work (the floor) is where the time goes, and that's what the
   rewrite attacked.
3. **Peepholes add a focused CPU layer on top:** ~9–13% general, up to **−70% for
   `@world → leaf` sends** (fused direct-posting + dead-credit elision) and
   **−24% for a plain `@src → @dest` send** (balance slots).
4. **The balance map WAS the next floor — and slots break through it.** The
   `entryFor` map lookup was ~40 ns (~half a world send, more than all big.Int
   arithmetic combined). Balance slots turn it into an array index, so the slotted
   opt now runs *below* the old map-based `RuntimeBaseline` floor. Slotting the
   queue path (`pull`/`send`) is the natural next step.
5. **Behavior is identical** — the whole script-test corpus passes in both naive
   and optimized modes (`internal/compiler/scripts_test.go`), plus targeted
   coherence tests for the slot cache.

### Caveats

- The balance cache has **no eviction** yet (`RuntimeBaselineUnique`: an
  all-distinct-account workload regresses to ~5 allocs/op and grows the map
  unbounded). Fine for a hot/recurring working set; a production version needs a
  size cap. See `OPTIMIZATIONS.md` § "Not yet done".
- Cold-start (fresh `Vm` per run) is not the target here and still allocates the
  fixed register banks; warm reuse amortizes it.
