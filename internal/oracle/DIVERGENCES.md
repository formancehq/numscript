# Oracle vs ledger vs numscript

`internal/oracle/` is a copy of the ledger repo's `internal/machine`, vendored
so the differential harness has a second implementation to compare against. It
only works as a check while it stays faithful to ledger, so every edit to it is
written down here.

Baseline: `formancehq/ledger` `internal/machine`, main at `8880965c2`.
Everything below was measured by running all three engines, not recalled.

## The four differences

| # | what | ledger | oracle | numscript |
|---|---|---|---|---|
| 1 | negative send amount | `insufficient funds` | `insufficient funds` | rejects the amount |
| 2 | `kept` | does not consume the funding | consumes it | consumes it |
| 3 | `save` past the balance | balance goes negative | goes negative | floors at zero |
| 4 | source-side negative `max` | rejects the script | rejects the script | clause contributes zero |

The oracle matches ledger on 1, 3 and 4. It does not on 2, which is the one
place it cannot be trusted to check numscript.

`TestDifferentialSweep` is green: 0 divergences over 3000 seeds, and 161 of
those scripts tolerated under #1.

---

## 1. Negative send amount

```numscript
send [COIN 0] - [COIN 90] (
  source = @alice        // alice has 1000
  destination = @bob
)
```

| engine | result |
|---|---|
| ledger | `account(s) @alice had/have insufficient funds` |
| oracle | same as ledger |
| numscript | `Cannot send negative amount: -90` |

Alice has 1000 and is asked for -90. The amount is invalid at any balance, so
calling it a funding problem points at the wrong thing, and ledger's API maps it
to HTTP 400 `INSUFFICIENT_FUND`.

Ledger guards `OP_TAKE_MAX` but not `OP_TAKE`, so every bounded source falls
through to the insufficient-funds branch.

Upstream agrees it is a bug. `vm/machine_negative_amount_test.go` pins the
current behaviour with a `BUG:` comment. The fix, ledger PR #2060, was closed
unmerged and its branch deleted, so main still has no guard.

**The oracle copies ledger exactly, missing guard and all.** `Compare` absorbs the
difference instead, under the name `negative amount vs missing funds`. Both
engines reject the script and no money moves either way; only the error differs.

The tolerance runs in one direction: the interpreter naming a negative amount
while the oracle blames the funds. The reverse is #4, and stays a mismatch.

The sweep reports how often it fires -- currently 161 of 3000 scripts, 5.4%. If
that number grows, the tolerance has turned into a blind spot.

The oracle used to carry the missing guard itself. Tolerating in `Compare`
instead keeps `vm/machine.go` byte-identical to ledger apart from the vendored
types, and makes the cost countable on every run rather than invisible.

**Open question:** re-file the fix upstream.

## 2. `kept`

**Ledger does not consume a kept funding. It returns to the pool and funds
whatever comes next. numscript consumes it.**

```numscript
// acc0 = 1000, acc1 = 1000
send [COIN *] (
  source = { @acc1  @acc0 }
  destination = {
    max [COIN 400] kept
    remaining to @dst
  }
)
send [COIN *] (
  source = @acc1
  destination = @sink
)
```

| engine | who keeps the 400 | `@dst` | `@sink` |
|---|---|---|---|
| ledger | `@acc0`, and `@acc1` is drained | 1600 | 0 |
| oracle + numscript | `@acc1` | 1600 | **400** |

Both send 1600 to `@dst`. Within a single statement the difference is only which
account the money came from; totals per destination agree either way. Across
statements it is real, because `kept` decides which account keeps the money, and
the second statement reads that account.

Ledger's own `TestKeptComplex` asserts its behaviour deliberately. This is not a
bug upstream.

**The oracle was changed to match numscript**, in two places in
`script/compiler/destination.go`:

| site | ledger does | oracle does |
|---|---|---|
| `DestInOrderContext` | reassemble, reverse, take, reverse — re-attributing the kept amount to the bottom of the pool | keeps `subkept` as-is |
| `VisitAllocDestination` | `Bump(1)` before `OP_FUNDING_ASSEMBLE`, putting `subkept` at the front where the next portion spends it again | no bump, so it lands at the back and survives as leftover |

Both are needed. Changing only the first takes the sweep from 7 failing scripts
to 4.

**What it bought:** the sweep went from 7 failing and 85 tolerated scripts to
zero of each, and `Compare`'s `kept source attribution` tolerance was deleted.
That tolerance would have swallowed a genuinely wrong source in any script
mentioning `kept`, and it returned before metadata was compared.

**What it cost:** on `kept`, the oracle is no longer independent of numscript. A
`kept` regression in the interpreter will not be caught here.
`TestOracleKeptComplex` covers it instead: ledger's own `TestKeptComplex` script
with the expectations changed, keeping ledger's original numbers in the comment
so the difference stays visible.

**Open question:** whether ledger should adopt this. If it ever does, delete the
two edits rather than keep them.

## 3. `save` past the balance

```numscript
save [COIN 100] from @src        // src only has 50

send [COIN *] (
  source = @src allowing overdraft up to [COIN 100]
  destination = @dst
)
```

| engine | postings |
|---|---|
| ledger | `src->dst 50` |
| oracle | `src->dst 50` |
| numscript | `src->dst 100` |

numscript floors the save at the real balance of 50, leaving 0 saved, so the
whole 100 overdraft is available. Ledger subtracts unfloored to -50, so the
overdraft only brings it back to 50. Same script, twice the money moved.

Verified at ledger main `35bb7ffac`: `save [COIN 100] from @src` against a
balance of 50 gives `balances=map[@src:map[COIN:-50]]`.

The oracle used to floor at zero as well, with a comment saying it matched the
interpreter. That was the oracle bending toward the engine it exists to check,
and it has been reverted.

The generator rarely produces a save-overdraw followed by a bounded-overdraft
draw on the same account, so the sweep does not reach this.

**Open question:** whether numscript should stop flooring.

## 4. Source-side negative `max`

```numscript
send [COIN 100] (
  source = {
    max [COIN 0] - [COIN 10] from @a    // a and b both have 1000
    @b
  }
  destination = @dst
)
```

| engine | result |
|---|---|
| ledger | `cannot send a monetary with a negative amount: [COIN -10]` |
| oracle | same |
| numscript | `b->dst 100` — the clause contributes nothing and `@b` covers it |

Deliberately unfixed on the numscript side: closing it means changing the
interpreter's ground-truth behaviour, not just catching up to ledger. `Compare`
tolerates it by comparing missing-funds classification rather than error text.

**A ledger-side fix was built and rejected.** Ledger PR #2079
(https://github.com/formancehq/ledger/pull/2079) made a negative `max` clamp to
zero, matching numscript. It works, and it is green, but it needs a new opcode.

`OP_TAKE_MAX` receives two different things: a `max` cap, and the send amount of
a source with an unbounded fallback (`@world`, `allowing unbounded overdraft`).
Clamping inside the opcode fixes this case and at the same time turns
`send [COIN -90] (source = @world)` into a committed zero-amount transaction,
which moves ledger away from numscript on #1. Telling the two apart means a new
opcode the compiler emits only at the two `max` sites.

**Decision: leave ledger alone.** Adding an opcode to the machine is too
aggressive for what this buys, and the machine is the part where a mistake is
most expensive. Keeping ledger's current behaviour is the safer default.

So this stays a numscript-side question. Do not rebuild #2079.

The destination-side version of this was a real numscript bug and is fixed:
numscript now errors there, matching ledger.

**Open question:** whether numscript should reject.

---

## Where each one is pinned

| # | test |
|---|---|
| 1 | the sweep, as a named tolerance. `TestMissingFundsClassificationMismatchStillCaught` pins that the opposite direction is not tolerated |
| 2 | `TestKnownBugRepros` (the engines agree now) and `TestOracleKeptComplex` |
| 3 | `TestKnownOpenDivergences` |
| 4 | `TestSourceSideNegativeMaxClauseTolerated` and `TestMissingFundsClassificationMismatchStillCaught` |

`TestKnownOpenDivergences` asserts the divergence is still there. If a case
starts agreeing, it fails; update this file and move the case to
`TestKnownBugRepros` rather than deleting it.

## Vendoring mechanics

No semantics, no decisions needed.

| what | where |
|---|---|
| import paths rewritten to `…/numscript/internal/oracle/machine` | everywhere |
| `ledger/pkg/{accounts,assets}` replaced by local copies | `account.go`, `asset.go`, `internal/…` |
| `ledger` import dropped, local `Account`/`ResultPosting`/`Zero` used | `vm/oracle_types.go`, `vm/run.go`, `vm/store.go`, `vm/machine.go` |
| upstream's 13 `*_test.go` and `examples/basic.go` not copied | — |
| `smoke_test.go` added, to check the vendoring itself | `internal/oracle/` |
| `vm/machine.go` left un-gofmt'd, to stay diffable against upstream | the repo gofmt check excludes `/oracle/` |

The legacy machine also emits zero-amount postings where numscript emits none.
`run_oracle.go` drops them before comparing. That is a harness normalization,
not a difference between the engines.

## Checking for drift

```sh
L=~/Documents/dev/formance/ledger/internal/machine
O=internal/oracle/machine
norm() { sed -E 's|github\.com/formancehq/ledger/internal/machine|M|g; s|github\.com/formancehq/numscript/internal/oracle/machine|M|g' "$1"; }
for f in $(cd $O && find . -name '*.go' | sed 's|^\./||' | sort); do
  [ -f "$L/$f" ] && diff <(norm "$L/$f") <(norm "$O/$f") > /dev/null || echo "DIFFERS: $f"
done
```

Nine files differ. Only one is a behaviour change:

| file | why |
|---|---|
| `account.go`, `asset.go`, `internal/accounts/accounts.go`, `internal/assets/asset.go` | vendoring |
| `vm/oracle_types.go`, `vm/run.go`, `vm/store.go` | vendoring |
| `vm/machine.go` | vendoring |
| `script/compiler/destination.go` | `kept` (#2) |

Anything else that shows up is undocumented drift. Either record it here or
remove it.
