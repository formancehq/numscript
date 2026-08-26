# Differential testing handoff notes (scratch, not for commit)

This is a working-notes dump to resume differential-testing work after a
context clear. It is **not version-controlled on purpose** — do not `git
add`/commit it. Delete it once the bug below is resolved and its learnings
are captured wherever makes sense (commit message, code comment, issue).

## Goal (original ask, 3 steps)

1. Vendor the legacy `ledger/internal/machine` interpreter into this repo as
   an "oracle" — as close to verbatim as possible, isolated from the public
   API.
2. Port `numscript_gen` (a Haskell/QuickCheck random-numscript-program
   generator) to Go, integrated with Go's native fuzzing.
3. Build a differential-testing harness: generate random programs, run them
   through both the oracle and this repo's interpreter, compare results.

All three are implemented. Right now we're mid-investigation of a real bug
the harness found (see "Open bug" below) — that's the priority to resume.

## Architecture

Three Go module roots in this one repo (deliberately — see "why nested
modules" below):

- **root module** (`github.com/formancehq/numscript`) — this repo as it
  already existed, plus:
  - `builder/` (public package) — extended with new combinators (see
    "What was built" below).
  - `internal/gen/` (new) — Go port of the Haskell generator. Only depends
    on `builder/` + stdlib, so it stays in the root module.
- **`internal/oracle/`** — its own nested Go module (own `go.mod`,
  `module github.com/formancehq/numscript/internal/oracle`). Contains the
  vendored legacy machine at `internal/oracle/machine/...`.
- **`internal/difftest/`** — its own nested Go module (own `go.mod`,
  `module github.com/formancehq/numscript/internal/difftest`), with
  `replace` directives pointing at `../..` (root) and `../oracle`. Contains
  the differential-testing harness and the `FuzzDiff` fuzz target.

**Why nested modules, not just `internal/` packages:** the oracle needs
three extra dependencies (`github.com/formancehq/go-libs/v5`,
`github.com/antlr/antlr4/runtime/Go/antlr`, `github.com/logrusorgru/aurora`)
that the user explicitly did not want added to the root module's
`go.mod`/`go.sum`. A directory with its own `go.mod` is automatically
excluded from the parent module's build graph (`go build ./...` from root
never sees it), so this fully isolates those dependencies. Same reasoning
extends to `internal/difftest` (it needs both the oracle's and the root
module's code, so it also can't live in the root module without pulling
those deps back in).

**Gotcha this implies:** Go's `internal/` import-visibility rule is a
path-prefix check independent of module boundaries — an importer can reach
`.../internal/oracle/...` only if the importer's own module path *also*
starts with `github.com/formancehq/numscript/`. That's why `internal/oracle`
and `internal/difftest` are named the way they are (nested under the
`numscript` path) rather than given arbitrary standalone module names.

**Running/testing each module:** you must `cd` into it first — commands
from the repo root do not see nested modules.

```bash
cd internal/oracle   && go build ./... && go vet ./... && go test ./...
cd internal/difftest && go build ./... && go vet ./...
cd internal/difftest && go test . -run FuzzDiff -v          # seed corpus + saved failures
cd internal/difftest && go test . -fuzz=FuzzDiff -fuzztime=60s   # actual fuzzing (needs `.` not `./...`)
```

Known trap: Go's fuzzer replays saved failures from
`internal/difftest/testdata/fuzz/FuzzDiff/` **before** starting new
exploration, and aborts on the first failure — so a run stops almost
instantly if a known failure is sitting there. To explore past a known
issue, either fix it, `rm` the corpus file, or (what I did) temporarily
copy the fuzz function under a different name so it starts with no cached
corpus.

## What was built (all currently UNCOMMITTED except where noted)

- **`internal/oracle/`** — vendored legacy machine. **This one IS
  committed** — commit `4ebeb49 "feat: copy paste ledger machine"` on
  branch `test/differential-testing`, made by the user outside of any
  Claude session, after reviewing the vendoring diff. Nothing to redo here.
  Explicit list of intentional deviations from the original ledger source is
  in that commit / was described when it was built (import-path rewrite,
  `ledger.*` types replaced by local `ResultPosting`/`ResultPostings`/
  `Account` in `internal/oracle/machine/vm/oracle_types.go`, since the
  original `ledger.Posting` name collided with an existing local `Posting`
  type already in `vm/machine.go`).
- **`builder/` extensions** (root module, uncommitted): `SrcCapped`,
  `SrcAllotment`, `DestInorder` (+ `DestInorderClause`), `DestAllotment`,
  `KeptOrDest`/`Kept()`/`To()`, `Portion`/`NewPortion`/`AllotmentClause[T]`
  (new file `builder/allotment.go`), and `UnsafeAccount` (raw `@literal`
  account reference bypassing the vars pool — used by the generator so it
  doesn't pollute every script with 16 account vars). Snapshot tests added
  to `builder/builder_test.go` for all of these, all passing.
- **`internal/gen/`** (root module, uncommitted): faithful Go port of
  `numscript_gen`'s `Gen.hs`/`Numscript.hs`/`Utils.hs` (send-only scope,
  matching the Haskell tool 1:1 — not the broader common grammar). Files:
  `ast.go` (intermediate AST), `gen.go` (generation logic + weights, ported
  verbatim from reading the actual Haskell source, not a secondhand
  summary), `cleanup.go` (port of `cleanupNumscript`), `convert.go` (AST →
  `builder.Statement`), `api.go` (`GenerateScript`, `RandFromBytes`).
  Exposed API: `gen.GenerateScript(rng *rand.Rand) (vars map[string]string,
  script string)`.
- **`internal/difftest/`** (own nested module, uncommitted): `run_new.go`
  (drives this repo's interpreter via the public `numscript` package),
  `run_oracle.go` (drives the oracle), `compare.go` (normalizes + diffs
  results), `difftest.go` (`RunOne` orchestration), `difftest_test.go`
  (`FuzzDiff`), `cmd/rundiff/main.go` (standalone batch runner).

## Bugs found and fixed while building this (all in the uncommitted work)

1. **`builder.BuildProgram` never inserted a separator between multiple
   statements** — consecutive `send` blocks concatenated into invalid
   syntax (`...)send [...] (...`). Only surfaced once real multi-statement
   programs were generated (every prior hand-written test built exactly one
   statement). Fixed in `builder/builder.go` (`BuildProgram` now writes
   `"\n\n"` between statements) + regression test `TestMultipleStatements`.
2. **`run_oracle.go` used `vm.EmptyStore`** instead of `vm.StaticStore{}`.
   `EmptyStore.GetBalances` never materializes zero-value rows (unlike
   `StaticStore{}`, an empty map, which does), and the oracle's own
   `credit()` only updates a balance if the map entry already exists — so
   with `EmptyStore`, in-script seed funding (`world -> accN` statements
   earlier in the same script) was silently dropped, causing bogus "missing
   balance" errors later. Fixed by switching to `vm.StaticStore{}`.
3. **`vm.Machine.SetVarsFromJSON` mutates (clears) the map it's given.**
   `difftest.RunOne` passes the same `vars` map to both `runNew` and
   `runOracle`; the oracle's call was destructively consuming it, so a
   failure report's printed `vars` reflected post-mutation empty state, not
   what was actually used. Fixed by giving each engine `maps.Clone(vars)` in
   both `run_new.go` and `run_oracle.go`.
4. **`compare.go` design fix:** "oracle rejects a script the new interpreter
   compiles" must be tolerated as an expected outcome, not a mismatch —
   `internal/gen`'s cleanup pass (like the original Haskell one) doesn't
   catch every case where unboundedness propagates up through nested
   inorder blocks. The real reference tool (`numscript_gen/app/Main.hs`)
   explicitly buckets this outcome the same way. The reverse direction (new
   interpreter rejects something the oracle compiled) is still flagged, on
   the reasoning that the generator only emits old-machine-compatible
   syntax, so the new interpreter rejecting it would be a real regression.

## Open bug — priority to resume

### The minimal repro (confirmed, reproduces every time)

```numscript
send [COIN 100] (
  source = @world
  destination = @acc0
)

send [COIN 100] (
  source = @acc0
  destination = {
    max [COIN 60] to @acc1
    max [COIN 40] kept
    max [COIN 50] kept
    remaining to @acc2
  }
)
```

- **This repo's interpreter**: succeeds. Postings: `world→acc0: 100`,
  `acc0→acc1: 60`. (60 to `@acc1`, 40 "kept" exactly exhausts the 100, the
  third/fourth clauses get nothing — internally consistent.)
- **Oracle (legacy machine)**: fails with `insufficient funds` on `@acc0`,
  even though `@acc0` has exactly 100 and the statement never asks for more
  than 100.

**Structural requirement, empirically confirmed via bisection:** it takes
*exactly two or more* consecutive `max ... kept` clauses positioned *after*
the funding pool is already fully exhausted by an earlier clause, followed
by a trailing non-`kept` `remaining` (or `max ... to <account>`) clause. With
only **one** `kept` clause before `remaining to @acc2`, both engines agree
and succeed — the bug needs the second one.

No source allotment, no overdraft, no vars, no nested destination blocks are
needed to reproduce it — those were all in the originally-fuzzed 21-statement
script but are irrelevant; this three-clause version (well, four:
`max`/`kept`/`kept`/`remaining`) is confirmed minimal by direct testing (see
"How to reproduce quickly" below to re-verify after a context clear).

### Working hypothesis (NOT fully confirmed — this is where to resume)

Traced with `m.Debug = true` on the oracle's `vm.Machine` (prints full VM
stack + balances + opcode before every `tick()`) against the original
larger repro. Right before the crash, the stack was:

```
[[COIN @acc0 95] [COIN 190]]     <- about to OP_TAKE(amount=190) from funding {@acc0: 95}
```

i.e. the VM tried to `Take` 190 units from a funding pool that only had 95
— that's the literal insufficient-funds trigger. Backing up further in the
trace, **two separate `max ... kept` clauses in the loop each independently
saw the exact same, non-reduced `[COIN @acc0 95]` funding value** as their
input, rather than the second one seeing an already-exhausted (empty)
pool. That smells like the destination-inorder compiler's loop not properly
threading the "funding remaining after this clause" forward from one
iteration to the next for the *kept accumulator* bookkeeping specifically
(the actual per-clause `TakeMax` semantics seemed fine in isolation) —
plausibly causing the "kept accumulator" to double-count the same leftover
(97 counted twice ≈ 190), which then makes the final `remaining` clause's
computed "amount still owed" wrong, exceeding what's actually available.

**This was NOT re-verified against the new minimal 4-clause repro above** —
that trace was done against the original bigger script. First thing to do
on resume: re-run the debug trace against the minimal repro (much easier to
read) and confirm/refute this hypothesis precisely against the bytecode.

Relevant source to read when resuming, in order of relevance:

1. `internal/oracle/machine/script/compiler/destination.go` —
   `VisitDestinationRecursive`, the `*parser.DestInOrderContext` case
   (roughly lines 36–118). This is the compiler emitting bytecode for a
   destination-inorder block; the per-clause loop (lines ~55–92) and the
   "kept accumulator" + trailing `remaining` handling (lines ~93–118) are
   both suspect.
2. `internal/oracle/machine/funding.go` — `Funding.Take` and
   `Funding.TakeMax`. `Take` (used for the "remaining" clause) has a
   special-case for `amount.Eq(Zero)` (line ~60) worth double-checking.
3. `internal/oracle/machine/vm/machine.go` — the `tick()` switch, cases
   `OP_TAKE`, `OP_TAKE_MAX`, `OP_FUNDING_SUM`, `OP_FUNDING_ASSEMBLE`,
   `OP_FUNDING_REVERSE`, `OP_MONETARY_ADD`. **Not yet read in this
   investigation** — this is the actual runtime semantics of the opcodes the
   compiler above emits, and is probably necessary to fully confirm the
   hypothesis (the compiler reading alone wasn't conclusive).

### Which side is "buggy"?

Not yet determined — that was explicitly left for the user to judge once
they see the minimal repro (numscript semantics for `kept` interacting with
multiple trailing clauses might be genuinely ambiguous/undefined, not
necessarily a clear-cut bug in either engine). Whoever resumes this should
present the minimal repro's two behaviors and get a ruling before assuming
the oracle is wrong just because it errors.

### How to reproduce quickly after a context clear

Since `runNew`/`runOracle` are unexported, a repro test needs to be an
*internal* test file in `internal/difftest` (`package difftest`, not
`difftest_test`). Example (delete after use, don't commit):

```go
package difftest

import (
	"context"
	"testing"
)

func TestRepro(t *testing.T) {
	script := `send [COIN 100] (
  source = @world
  destination = @acc0
)

send [COIN 100] (
  source = @acc0
  destination = {
    max [COIN 60] to @acc1
    max [COIN 40] kept
    max [COIN 50] kept
    remaining to @acc2
  }
)`
	ctx := context.Background()
	newRes := runNew(ctx, script, map[string]string{})
	oracleRes := runOracle(ctx, script, map[string]string{})
	t.Logf("new:    %+v", newRes)
	t.Logf("oracle: %+v", oracleRes)
}
```

Run with `cd internal/difftest && go test . -run TestRepro -v`.

For the VM trace: add `m.Debug = true` right after `vm.NewMachine(*p)` in a
throwaway copy of `runOracle`'s body (or inline in the repro test, calling
`compiler.Compile`/`vm.NewMachine`/`SetVarsFromJSON`/`ResolveResources`/
`ResolveBalances`/`Execute` directly instead of going through `runOracle`).
It prints one block per VM instruction via `fmt.Println` — pipe to a file
and grep around `OP_TAKE`/`insufficient` to find the crash point fast.

### Secondary, not-yet-reduced finding (lower priority)

A fresh 10s fuzz burst (after filtering out the above finding) turned up a
**second, distinct** divergence: same general shape (source allotment +
destination with a `kept` clause), but here *both* engines succeed and the
per-destination-account totals match — they just disagree on which upstream
source account gets debited for one clause (new interpreter draws it all
from `world`; oracle splits it between a partially-used source account and
`world`). Postings count differs (22 vs 23) as a result. Not reduced to a
minimal repro yet. Likely related to the same general area (fund
repayment/attribution when `kept` interacts with a multi-source allotment)
but not confirmed to share a root cause with the primary bug above. The
corpus/RNG seed that found it was not saved (it was found via a throwaway
scratch fuzz target that got deleted) — would need to re-fuzz to rediscover
it, or investigate directly once the primary bug is understood.

## Git state at time of writing

- Current branch: `test/differential-testing`.
- Only commit beyond `main`: `4ebeb49 "feat: copy paste ledger machine"`
  (the `internal/oracle` vendoring — made by the user, not any Claude
  session).
- Everything else described above (`builder/*`, `internal/gen/`,
  `internal/difftest/`) is **uncommitted**, sitting in the working tree.
  Root `go.mod`/`go.sum` are untouched (confirmed via `git diff go.mod
  go.sum` being empty) — the isolation goal held throughout.
- `internal/difftest/testdata/fuzz/FuzzDiff/` currently has one saved
  corpus entry (`2fde656d41007dc7`) reproducing the *original* 21-statement
  form of the primary bug (pre-reduction). The minimal 4-clause repro above
  is not saved as a corpus file anywhere — it only exists in this document.
