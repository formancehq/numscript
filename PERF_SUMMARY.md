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
| simple `send` (VM) | 302.5 ns | **156.2 ns** (opt) | **−48%** | 10 → **1** |
| capped inorder (VM) | 853.3 ns | **531 ns** (opt) | **−38%** | 23 → **1** |
| world → dest (VM) | ~queue, 10 allocs | **44 ns** (opt, leaf) | — | 10 → **1** |

The dominant win is **allocations → 1/op** on the warm path (runtime rewrite,
always-on), compounded by the peepholes (opt-in) and the new unbounded-source
fast path.

## VM vs interpreter vs floor — `feat/exp/optimize-vm`

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| **simple `send`** | | | |
| TreeWalker (interpreter) | 1603 | 2705 | 47 |
| RuntimeBaseline (floor) | 137.5 | 16 | 1 |
| CompiledVM (naive) | 181.0 | 32 | 1 |
| CompiledVMOpt (peephole) | 156.2 | 32 | 1 |
| **capped inorder** | | | |
| TreeWalkerCapped | 3812 | 6019 | 97 |
| RuntimeBaselineCapped (floor) | 452.9 | 16 | 1 |
| CompiledVMCapped (naive) | 554.9 | 32 | 1 |
| CompiledVMOptCapped (peephole) | 531 | 32 | 1 |
| **world → dest** | | | |
| WorldNaive (VM, queue) | 149.0 | 32 | 1 |
| **WorldOpt (VM, fused leaf op)** | **44** | 32 | 1 |
| WorldBaselineTakePost (floor, debit kept) | 95.2 | 16 | 1 |
| WorldBaselineDirectPost (floor, debit skipped) | 54.9 | 16 | 1 |

The leaf op (`44 ns`) is below even the `DirectPost` floor (`54.9`, which still
credits dst): eliding the destination credit removes the `entryFor` balance-map
lookup entirely, so only the posting append and dispatch remain.

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

**Front-end cost = engine − floor** (simple `send`, this branch):

| Engine | total | − floor | front-end | allocs over floor |
|---|---:|---:|---:|---:|
| TreeWalker (interpreter) | 1603 | 137.5 | **1465.5 ns** | +46 |
| CompiledVM (naive) | 181.0 | 137.5 | **43.5 ns** | +0 |
| CompiledVMOpt | 156.2 | 137.5 | **18.7 ns** | +0 |

So the VM's front-end (dispatch + registers) is **~34× lighter** than the
interpreter's AST walk (naive), or **~78× lighter** optimized — and it does **0
extra allocations** where the interpreter's front-end does **~46**.

**VM-vs-interpreter speedup, floor removed** — `(interp − floor) / (vm − floor)`,
i.e. how many times smaller the VM's own overhead is:

| Script | interp − floor | vm − floor | **speedup** |
|---|---:|---:|---:|
| simple (base VM) | 1465.5 | 43.5 | **33.7×** |
| simple (Opt VM) | 1465.5 | 18.7 | **78.4×** |
| capped (base VM) | 3359.1 | 102.0 | **32.9×** |
| capped (Opt VM) | 3359.1 | 78.1 | **43.0×** |

Cross-branch, the front-end ratio itself barely moved (the VM was always cheap
per instruction); what changed is the **floor** — the shared funds work — which
the runtime rewrite roughly halved (simple `253.8 → 137.5`, capped
`744.9 → 452.9`) while collapsing allocs `10 → 1` / `23 → 1`.

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
| **WorldOpt, `Op_PostFromUnboundedLeaf` (leaf dst)** | **44** | **−70%** |

Two successive peephole wins on the same shape:

- **`Op_PostFromUnbounded`** drops the always-true enough-funds check and the
  (unobservable) source-balance debit: `126.8 → 77.3` (−39%).
- **`Op_PostFromUnboundedLeaf`** additionally drops the *destination credit* when
  dst is a **leaf** (never a later funding source, never saved, no `balance()`
  read anywhere) — which removes the `entryFor` balance-map lookup, the single
  most expensive operation left in the path (~40 ns): `77.3 → 44` (−43%).

`44 ns` lands *below* the crediting `DirectPost` floor (`54.9`) exactly because
that floor still does the credit's map lookup. Bounded sources and non-leaf /
saved / balance-read destinations are ineligible (guarded) and unchanged. This
matters because funding `@world → external leaf` is the most common numscript
shape.

## Takeaways

1. **Allocations are the story.** Warm path `10 → 1` (simple) and `23 → 1`
   (capped); this is ~half the ns win by itself and is *always on* (naive or
   optimized bytecode).
2. **The VM front-end was never the bottleneck** — it's 33–78× lighter than the
   interpreter's. The runtime funds work (the floor) is where the time goes, and
   that's what the rewrite attacked.
3. **Peepholes add a focused CPU layer on top:** ~9–13% general, up to **−70% for
   `@world → leaf` sends** via the fused direct-posting op + dead-credit elision.
4. **The balance map is the next floor.** The `entryFor` map lookup is ~40 ns —
   ~half a world send, more than all the big.Int arithmetic combined. Dead-credit
   elision removes it *when provably dead*; removing it in general (integer
   account slots) is the largest remaining structural lever.
5. **Behavior is identical** — the whole script-test corpus passes in both naive
   and optimized modes (`internal/compiler/scripts_test.go`).

### Caveats

- The balance cache has **no eviction** yet (`RuntimeBaselineUnique`: an
  all-distinct-account workload regresses to ~5 allocs/op and grows the map
  unbounded). Fine for a hot/recurring working set; a production version needs a
  size cap. See `OPTIMIZATIONS.md` § "Not yet done".
- Cold-start (fresh `Vm` per run) is not the target here and still allocates the
  fixed register banks; warm reuse amortizes it.
