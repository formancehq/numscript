# How the vendored oracle differs from ledger

`internal/oracle/` is a vendored copy of the ledger repo's `internal/machine`.
It exists to be an *independent* check on this repo's interpreter, so every
difference from upstream is a liability: the more the oracle is bent toward
numscript, the less it can catch.

This file lists every difference and why it is there.

Upstream baseline: `github.com/formancehq/ledger`, `internal/machine`, main at
`8880965c2` (2026-09-16).

## Summary

| § | kind | count | state |
|---|---|---|---|
| 1 | Structural — vendoring mechanics, no behaviour | 6 | fine, ignore |
| 2 | Oracle behaviour **not** in ledger main | 2 | ① ledger PR #2060 closed unmerged; ② `kept`, aligned to numscript on purpose — **costs checking power**, see below |
| 3 | Genuine numscript ↔ ledger semantic differences | 2 open, 2 resolved | need a decision |
| 4 | Local workarounds replaced by upstream's fixes | 0 | **closed** — ledger PR #2059 merged |

`TestDifferentialSweep` is **green**: 0 divergence classes and 0 tolerated
scripts over 3000 seeds. It went green by changing the oracle (§2 ②), not by
adding a tolerance — that trade is what §2 ② is about.

## Regenerating the list

```sh
L=~/Documents/dev/formance/ledger/internal/machine
O=internal/oracle/machine
norm() { sed -E 's|github\.com/formancehq/ledger/internal/machine|M|g; s|github\.com/formancehq/numscript/internal/oracle/machine|M|g' "$1"; }
for f in $(cd $O && find . -name '*.go' | sed 's|^\./||' | sort); do
  [ -f "$L/$f" ] && diff <(norm "$L/$f") <(norm "$O/$f") > /dev/null || echo "DIFFERS: $f"
done
```

Anything that shows up and is not listed below is unrecorded drift — either
document it here or remove it.

As of 2026-09-17 the list is exactly these nine, and every one is accounted for:

| file | why | § |
|---|---|---|
| `account.go`, `asset.go`, `internal/accounts/accounts.go`, `internal/assets/asset.go` | local `accounts`/`assets` packages | 1 |
| `vm/oracle_types.go`, `vm/run.go`, `vm/store.go` | local `Account`/`ResultPosting`/`Zero` | 1 |
| `vm/machine.go` | local types (§1) **and** the `OP_TAKE` guard | 1 + 2 ① |
| `script/compiler/destination.go` | `kept` consumes the funding | 2 ② |

`script/compiler/destination.go` entered this list on 2026-09-17 and is the only
entry that is a deliberate behavioural change *away* from ledger and *toward*
numscript. `script/compiler/compiler.go` left it when ledger PR #2063 merged.

---

## 1. Structural — vendoring mechanics, no behaviour change

These exist only because the code was lifted out of the ledger module. They
carry no semantics and need no decision.

| what | where |
|---|---|
| Import paths rewritten `…/ledger/internal/machine` → `…/numscript/internal/oracle/machine` | everywhere |
| `ledger/pkg/{accounts,assets}` replaced by local `machine/internal/{accounts,assets}` | `account.go`, `asset.go`, `internal/…` |
| `ledger "…/ledger/internal"` import dropped; local `Account`, `ResultPosting`, `Zero` used instead | `vm/oracle_types.go`, `vm/run.go`, `vm/store.go`, `vm/machine.go` |
| Upstream's 13 `*_test.go` and `examples/basic.go` not copied | — |
| `smoke_test.go` added — confirms the vendoring itself didn't break anything | `internal/oracle/` |
| `vm/machine.go` deliberately left un-gofmt'd, to stay diffable against upstream | repo gofmt check excludes `/oracle/` |

## 2. Oracle behaviour not in ledger main

Two places where the oracle does not match ledger main. Each carries a comment
at the code site pointing back here.

| # | what | where | kind | upstream |
|---|---|---|---|---|
| ① | negative-amount guard on `OP_TAKE` | `vm/machine.go` | oracle is **ahead** of ledger | ledger PR #2060 — **closed unmerged**; main pins the behaviour as a known bug |
| ② | `kept` consumes the funding | `script/compiler/destination.go` | oracle is **aligned to numscript**, against ledger | ledger `461050af9`, asserted by its `TestKeptComplex` — no upstream change proposed |

The two are not the same kind of exception, and ② is the worse one.

① is a bug that makes the oracle disagree with numscript for a reason numscript
is right about; keeping it would produce noise, not signal.

② is the oracle being bent toward numscript on a point where ledger is not
wrong, just different. That is exactly the liability this file exists to track:
**the oracle can no longer catch a `kept` regression in the interpreter**, because
both sides now implement the same rule. It was taken knowingly, to get the sweep
green on a difference that had already been decided in numscript's favour — but
it is a decision to revisit if ledger ever adopts numscript's semantics, at which
point ② should be deleted rather than kept.

Two entries that used to live here are gone: `save` evaluating its monetary
expression (ledger PR **#2063**, merged `95dfad77e`, backported to
`release/v2.3` as #2067 and `release/v2.4` as #2066) and the negative-amount
guard on `OP_SAVE` (ledger PR **#2068**, merged `35bb7ffac`). The oracle's
copies of `script/compiler/compiler.go` and that arm of `OP_SAVE` are now
byte-identical to main; `script/compiler/compiler.go` has dropped out of the
`DIFFERS` list entirely. `TestOracleSaveMonetaryExpression` and
`TestOracleSaveNegativeAmountRejected` stay in `smoke_test.go` as vendoring
checks — upstream has its own coverage for both in `vm/machine_test.go`, which
is not copied here.

### ① Negative-amount guard on `OP_TAKE`

```go
case program.OP_TAKE:
    mon := pop[machine.Monetary](m)
    funding := pop[machine.Funding](m)
    if mon.Amount.Ltz() {
        return true, fmt.Errorf("cannot send a monetary with a negative amount: [%s %s]", ...)
    }
```

Ledger guards `OP_TAKE_MAX` only, and has since 2023 (`0cc2844e4`). `OP_TAKE`
is unguarded, so a negative send amount — `send [EUR/2 0] - [EUR/2 476] (…)`,
which is ordinary numscript — reaches `funding.Take()` and surfaces as
`insufficient funds` rather than as a rejected script.

numscript rejects it outright (`Cannot send negative amount: -476`), which is
the behaviour that makes sense: the amount is invalid regardless of what funds
happen to be available, and reporting it as a funding problem sends you
looking in the wrong place.

**This is reachable, not defensive.** Removing the guard makes 161 of 3000
swept scripts (5.4%) diverge, all on the same missing-funds-vs-invalid-amount
classification.

**Status: acknowledged upstream as a bug, not fixed.** PR #2060
(`fix/machine-negative-amount-op-take`) was closed unmerged on 2026-09-16 and
its branch deleted. What landed instead, via #2059, is
`vm/machine_negative_amount_test.go` — a characterization of the behaviour
rather than a fix:

- `TestNegativeSendGuardedPaths` pins the four source shapes that compile to
  `OP_TAKE_MAX` and do produce `cannot send a monetary with a negative amount`.
- `TestNegativeSendAlwaysReportsInsufficientFunds`, commented `BUG:`, pins that
  every bounded source instead reports `insufficient funds`, regardless of
  balance, overdraft allowance, or source/destination shape — and notes the API
  maps it to HTTP 400 `INSUFFICIENT_FUND`.

So ledger agrees this is wrong and has written the current behaviour down as a
known bug, but main still has no guard on `OP_TAKE`, and the oracle's guard
stays until one lands. This is a weaker position than ② and ③ were in — those
had merged PRs — so the entry needs a decision rather than just waiting: either
re-file the fix, or drop the guard and move this to §3, accepting 161 of 3000
swept scripts (5.4%) diverging on error classification.

No separate pin needed: the sweep covers this one, and removing the guard
lights up 161 scripts immediately.

### ② `kept` consumes the funding

The rule, in one line: **ledger does not consume a kept funding, the oracle
does.** On ledger a `kept` portion returns to the pool and funds whatever comes
next; here it is set aside and repaid.

Ledger implements the non-consuming behaviour in two places in
`script/compiler/destination.go`, and the oracle drops both:

| site | ledger | oracle |
|---|---|---|
| `DestInOrderContext` | reassembles `remaining ++ subkept`, reverses, takes `kept_amt`, reverses back — re-attributing the kept amount to the **bottom** of the pool | `subkept` is kept as-is, so the **front** keeps |
| `VisitAllocDestination` | `Bump(1)` before `OP_FUNDING_ASSEMBLE`, putting `subkept` at the **front** of the pool, where the next portion spends it again | no bump, so `subkept` lands at the **back** and survives as leftover |

Within one statement the difference is only attribution — totals per
destination agree either way. Across statements it is not, because `kept`
decides *which account keeps the money*:

```
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
| numscript + oracle | `@acc1` (front) | 1600 | **400** |
| ledger | `@acc0` (bottom), `@acc1` drained | 1600 | **0** |

**What this cost.** Before the change the sweep carried a `kept source
attribution` tolerance in `Compare` that suppressed the single-statement
symptom — 85 of 3000 scripts — while 7 got past it and failed on totals. Both
numbers are now zero, and the tolerance has been deleted along with
`usesKept`/`aggregateByDestination`/`destTotalsDiffer`. That is a real gain:
the tolerance would have swallowed a genuinely wrong *source* in any script
mentioning `kept`, and it short-circuited before metadata was compared.

**What it cost instead** is stated above and is worth repeating: on `kept`, the
oracle is no longer independent of the interpreter. A `kept` regression in
numscript will not be caught here. `TestOracleKeptComplex` in `smoke_test.go`
is the compensating control — it is ledger's own `TestKeptComplex` script with
the expectations changed, and its comment carries ledger's original numbers, so
the delta stays visible and reviewable rather than silently absorbed.

## 3. Genuine numscript ↔ ledger semantic differences

Not oracle defects — the two engines really disagree. Each needs a product
decision about which is correct, and the answer may be a change to numscript.

The oracle is faithful to ledger on 3.2, 3.3 and 3.4. It is **not** on 3.1 any
more; that entry is kept here for the history and now points at §2 ②.

The open ones are pinned in `internal/difftest/regression_test.go`
(`TestKnownOpenDivergences`, `TestSourceSideNegativeMaxClauseTolerated`), which
assert that the divergence is *still there*. If one starts agreeing, the test
fails and this file needs updating. 3.1 moved to `TestKnownBugRepros`, which
asserts the opposite — that the two now agree.

### 3.1 `kept` source attribution — **resolved, by changing the oracle**

The two engines still disagree; the oracle no longer sits on ledger's side of
it. Ledger does not consume a kept funding, numscript does, and the oracle was
changed to consume it too. Full description, the two compiler sites, and what
the choice costs: **§2 ②**.

This is the one entry in §3 that was closed without a decision about which
semantics numscript should have. That question is still open — it was just
un-blocked from the sweep.

### 3.2 `save` beyond the account's balance

numscript's `runSaveStatement` floors the saved amount at the balance; ledger
subtracts unfloored and goes negative, and the next send then fails with
insufficient funds. Verified at ledger main `35bb7ffac`: `save [COIN 100] from
@src` with a balance of 50 gives `balances=map[@src:map[COIN:-50]]`,
`err=insufficient funds`.

The oracle used to floor at zero too — its comment said "matching the new
interpreter's behavior", i.e. the oracle bent toward the engine it exists to
check. That has been reverted, so the divergence is now visible.

It does not show up in the sweep (the generator rarely produces save-overdraw
followed by a bounded-overdraft draw on the same account), so it is pinned
directly in `TestKnownOpenDivergences`.

**Status: open, now visible.**

### 3.3 Source-side negative `max` clause

The oracle errors; numscript treats the clause as contributing nothing, which
can let a different source in the same list cover the shortfall. Known and
deliberately unfixed — closing it means changing the interpreter's ground-truth
behaviour, not just catching up to it.

`Compare` tolerates it by comparing missing-funds *classification* rather than
error text.

**Status: known, accepted, pinned both ways.**

### 3.4 Destination-side negative `max` clause

**Resolved.** PR #190 makes numscript error, matching ledger.

## 4. Local workarounds replaced by upstream's fixes

**Empty — closed by ledger PR #2059, merged as `8880965c2`.**

Two balance-resolution bugs the oracle used to work around locally, then
carried upstream's proposed form of while #2059 was in review:

| was | now |
|---|---|
| `UnresolvedResourceBalances map[string][]int` — tolerated address collisions | upstream's `map[int]string`, which cannot collide |
| `StaticStore.GetBalances` per-account `make` hoisted, written locally | upstream's form |

Both are now plain upstream code in main, so there is nothing left to track
here. `vm/store.go` has dropped out of the `DIFFERS` list apart from §1's type
substitutions.

The merge is not identical to the branch this section used to cite
(`081349d51`): review added a dedup of `balancesQuery`, which the oracle now
carries too.

```go
// several resources can alias the same account/asset pair, only query it once
for address, assets := range balancesQuery {
    slices.Sort(assets)
    balancesQuery[address] = slices.Compact(assets)
}
```

That is the one thing re-vendoring against the merge commit had to pick up
rather than simply delete — worth remembering when §2 eventually closes the
same way, since a merged PR is not always what its branch was.

The three collision reproducers are in `smoke_test.go`
(`TestOracleBalanceVarsOnAliasedAccountResources`, `…OnAliasedAssetResources`,
`…OnSameAccountDifferentAssets`) and still pass. Upstream's own versions live in
`vm/machine_test.go`, which is not copied here.
