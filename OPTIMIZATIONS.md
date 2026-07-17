# `feat/exp/optimize-vm` — optimizations vs `feat/exp/vm`

Everything on this branch is a performance change over `feat/exp/vm`; behavior is
identical (validated by running the whole script-test corpus in both modes, see
`internal/compiler/scripts_test.go`). Optimization is **opt-in**: `Compile` is
unchanged and naive; `CompileWithOptimizations` runs the peepholes.

## Peepholes (`internal/compiler/peephole*.go`)

Each is a pass `func([]vInstr) ([]vInstr, bool)`; `optimize` runs them to a
fixpoint. `defaultPeepholes()` = `monetaryFold`, `fundsBypass`, `deadCode`.

- **`monetaryFold`** — removes the `mk_monetary(A,M)` → `get_asset`/`get_amount`
  round-trip: consumers read the asset/amount registers directly. The dead
  `mk_monetary` is then dropped by `deadCode`.
- **`fundsBypass`** — fuses a 1-source/1-destination `send` so it skips the
  `runtime` funds queue. Detects the shape by abstract-interpreting the queue
  over the instruction list (a drain-all send over a depth-1 queue); rewrites the
  `pull_account` → `take_account` (debit, no queue) and the `send_to_account` →
  `post_account` (posting, no debit). Fan-out / fan-in / N×M are left as
  pull/send and still use the runtime queue. See `queue-bypass` notes.
- **`deadCode`** — drops pure instructions (loads, arithmetic) whose result is
  never read. Cleans up after the other two.

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

## Runtime (`internal/runtime/runtime.go`)

There is **one** allocation strategy, not a naive/optimized pair — a single
`*big.Int` free-list reused across runs, plus a couple of scratch/index tricks:

- **`free` pool (`takeBig`/`putBig`)** — recycles `big.Int`s across `Reset`:
  queued-source amounts (reclaimed when consumed/merged/dropped) and posting
  amounts (reclaimed at `Reset`). So a **reused** VM does ~no per-run heap
  allocation on the funds path. (Balance amounts are not pooled — they live inline
  in `balanceEntry`.)
- **`head` FIFO index** — the source queue consumes from the front by `head++`
  instead of shifting the slice, so a front-to-back drain is O(1) per pop (an
  O(n²) drain becomes O(n)); `rewindIfEmpty` reclaims the dead prefix.
- **`capScratch`** — a reusable `big.Int` for `Send`'s decrementing cap, so a
  capped send allocates nothing.
- **`PostingsRef()`** — returns the internal postings slice with no copy, for
  hot-loop callers that consume it immediately (vs `GetPostings`, which copies).

## Exposure

- `compiler.CompileWithOptimizations(program)` → `(VarsEncoder, vm.Program, error)`
  is the only public entry point for the optimized build. It returns assembled
  bytecode; the intermediate virtual instructions (and the peephole passes) are
  package-private, so seeing the fused vInstr stream requires an in-package test
  (e.g. `TestOptimizedSimpleProgram` in `compiler_test.go`).
