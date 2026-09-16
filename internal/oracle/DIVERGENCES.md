# How the vendored oracle differs from ledger

`internal/oracle/` is a vendored copy of the ledger repo's `internal/machine`.
It exists to be an *independent* check on this repo's interpreter, so every
difference from upstream is a liability: the more the oracle is bent toward
numscript, the less it can catch.

This file lists every difference and why it is there.

Upstream baseline: `github.com/formancehq/ledger`, `internal/machine`, main.

## Summary

| § | kind | count | state |
|---|---|---|---|
| 1 | Structural — vendoring mechanics, no behaviour | 6 | fine, ignore |
| 2 | Oracle fixes **not** in ledger main | 3 | filed: ledger PRs #2060, #2063, #2068 |
| 3 | Genuine numscript ↔ ledger semantic differences | 3 open, 1 resolved | need a decision |
| 4 | Local workarounds replaced by upstream's proposed fixes | 2 | waiting on ledger PR #2059 |

`TestDifferentialSweep` is **currently red on purpose** — see §3.1.

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

---

## 1. Structural — vendoring mechanics, no behaviour change

These exist only because the code was lifted out of the ledger module. They
carry no semantics and need no decision.

| what | where |
|---|---|
| Import paths rewritten `…/ledger/internal/machine` → `…/numscript/internal/oracle/machine` | everywhere |
| `ledger/pkg/{accounts,assets}` replaced by local `machine/internal/{accounts,assets}` | `account.go`, `asset.go`, `internal/…` |
| `ledger "…/ledger/internal"` import dropped; local `Account`, `ResultPosting`, `Zero` used instead | `vm/oracle_types.go`, `vm/run.go`, `vm/store.go`, `vm/machine.go` |
| Upstream's 9 `*_test.go` and `examples/basic.go` not copied | — |
| `smoke_test.go` added — confirms the vendoring itself didn't break anything | `internal/oracle/` |
| `vm/machine.go` deliberately left un-gofmt'd, to stay diffable against upstream | repo gofmt check excludes `/oracle/` |

## 2. Oracle fixes not yet integrated in ledger main

Three ledger bugs the oracle fixes locally. All are filed upstream and all
should be dropped from the oracle once their PR merges, at which point this
section becomes empty again.

These are the only places the oracle is knowingly *ahead* of ledger rather than
faithful to it, which is a deliberate exception: a bug that makes the oracle
disagree with numscript for a reason numscript is right about produces noise,
not signal. Each carries a comment at the code site pointing back here.

| # | what | where | upstream |
|---|---|---|---|
| ① | negative-amount guard on `OP_TAKE` | `vm/machine.go` | ledger PR #2060 |
| ② | `save` evaluates its monetary expression | `script/compiler/compiler.go` | ledger PR #2063 (merged, unreleased) |
| ③ | negative-amount guard on `OP_SAVE` | `vm/machine.go` | ledger PR #2068 |

② and ③ are one story and must be read together: ② is what makes ③ reachable.
Do not take ② without ③.

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

**Status: reported upstream — ledger PR #2060**
(`fix/machine-negative-amount-op-take`), which adds the same guard, with the
same message, to `OP_TAKE`. When that merges, drop this guard from the oracle
and re-vendor.

No separate pin needed: the sweep covers this one, and removing the guard
lights up 161 scripts immediately.

### ② `save` evaluates its monetary expression

```go
} else if mon := c.GetMon(); mon != nil {
    typ, _, compErr := p.VisitExpr(mon, true)   // upstream: VisitExpr(mon, false), then PushAddress
```

`VisitExpr` has two modes and only the push path emits arithmetic:

```go
if push {
    ... OP_MONETARY_SUB
}
return machine.TypeMonetary, lhsAddr, nil   // the LEFT operand's address
```

Upstream's `VisitSaveFromAccount` used the address path, so no operator was
emitted and the address pushed was the left operand's. `save [COIN 50] -
[COIN 40]` compiled to bytecode identical to `save [COIN 50]`.

With `@alice` at 100:

| script | ledger reserved | correct |
|---|---|---|
| `save [COIN 50] - [COIN 40]` | 50 | 10 |
| `save [COIN 90] + [COIN 5]` | 90 | 95 |

No error — just the wrong amount reserved, in either direction depending on the
operator. numscript's `runSaveStatement` evaluates the whole expression and was
never affected; `send` was never affected either, since `VisitMonetary` takes
the address only to derive the asset and then evaluates properly with a push.

**Status: reported upstream — ledger PR #2063**
(`fix/machine-save-drops-arithmetic`), backported to `release/v2.3` (#2067) and
`release/v2.4` (#2066). When it merges, drop this from the oracle and
re-vendor.

Pinned by `TestOracleSaveMonetaryExpression` in `smoke_test.go`, because the
sweep does **not** cover it: `internal/gen` never emits a binary monetary
expression in `save` position, which is why the harness did not find this
itself. Closing that generator gap would make the pin redundant.

### ③ Negative-amount guard on `OP_SAVE`

```go
case machine.Monetary:
    if v.Amount.Ltz() {
        return true, machine.NewErrNegativeAmount(
            "tried to save a negative amount: [%s %s]", string(v.Asset), v.Amount)
    }
```

`OP_SAVE` subtracts the saved amount from the tracked balance without checking
its sign, and subtracting a negative **inflates** it. With `@alice` holding a
real 100 USD:

```
save [USD 10] - [USD 20] from @alice

send [USD *] (
  source = @alice
  destination = @bob
)
```

| version | sends |
|---|---|
| ledger before ② | 90 — the right operand was dropped, so this was `save [USD 10]` |
| ledger with ② only | **110 of a real 100** |
| with ③ | `tried to save a negative amount: [USD -10]` |
| numscript | `Cannot send negative amount: -10` |

So ② is what makes this reachable: before it, a negative could never arrive at
`OP_SAVE`, because the expression was never evaluated. **② must never ship
without ③** — on its own it turns a wrong-amount bug into a money-creation one.
At the time of writing ② is merged upstream but in no tag, so no released
ledger is affected.

**Status: reported upstream — ledger PR #2068.** Pinned by
`TestOracleSaveNegativeAmountRejected`; like ②, the sweep does not reach it.

Note the error style differs from ①: this uses the typed
`machine.NewErrNegativeAmount`, while the `OP_TAKE`/`OP_TAKE_MAX` guards use a
bare `fmt.Errorf`. Upstream's inconsistency, mirrored here on purpose.

## 3. Genuine numscript ↔ ledger semantic differences

Not oracle defects — the two engines really disagree. The oracle is faithful to
ledger on all of these. Each needs a product decision about which is correct,
and the answer may be a change to numscript.

All are pinned in `internal/difftest/regression_test.go`
(`TestKnownOpenDivergences`, `TestSourceSideNegativeMaxClauseTolerated`), which
assert that the divergence is *still there*. If one starts agreeing, the test
fails and this file needs updating.

### 3.1 `kept` source attribution — and its knock-on effect on totals

Ledger `461050af9` (in main since 2026-08-31) and numscript fix the same
in-order destination miscompilation with opposite semantics:

```
send [EUR/2 *] (
  source = { @a @b }                 // a=60, b=200
  destination = {
    max [EUR/2 50] kept
    max [EUR/2 100] to @out
    remaining kept
  }
)
```

| engine | `@out` funded from |
|---|---|
| numscript | `a:10, b:90` — `kept` consumes the **front** of the pool |
| ledger | `a:60, b:40` — `kept` is taken from the **bottom** |

Ledger's own `TestKeptComplex` asserts bottom-taking deliberately, so this is
intended upstream behaviour, not a bug. That test is ported into
`smoke_test.go` and passes, which is what pins that `461050af9` is genuinely
integrated here rather than approximated.

**Within one statement this is only attribution: totals per destination
agree.** That is what `Compare`'s `kept source attribution` tolerance allows —
85 of 3000 swept scripts.

**Across statements it is not.** `kept` decides *which account keeps the
money*, so it changes balances, and any later statement reading one diverges
for real:

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
| numscript | `@acc1` (front) | 1600 | **400** |
| ledger | `@acc0` (bottom), `@acc1` drained | 1600 | **0** |

So the tolerance in `Compare` is a real blind spot, not a normalization: it
suppresses the single-statement symptom of a difference that is observable in
totals as soon as a balance is read again. 7 of 3000 swept scripts get past it
and fail.

**Status: open, and the sweep is red because of it.** This was a deliberate
choice over quarantining `kept` in the generator or skip-listing the seeds:
the disagreement is real, and hiding it would make the harness look healthier
than it is. Deciding which semantics numscript should have is what closes it.

### 3.2 `save` beyond the account's balance

numscript's `runSaveStatement` floors the saved amount at the balance; ledger
subtracts unfloored and goes negative, and the next send then fails with
insufficient funds. Verified at ledger HEAD: `save [COIN 100] from @src` with a
balance of 50 gives `balances=map[@src:map[COIN:-50]]`, `err=insufficient funds`.

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

## 4. Local workarounds replaced by upstream's proposed fixes

Both were local reimplementations of bugs upstream has since fixed its own way,
on `fix/machine-balance-resource-collisions` (ledger PR **#2059**). The oracle
now carries upstream's version of each rather than its own.

| was | now |
|---|---|
| `UnresolvedResourceBalances map[string][]int` — tolerated address collisions | upstream's `map[int]string` (ledger `081349d51`), which cannot collide |
| `StaticStore.GetBalances` per-account `make` hoisted, written locally | upstream's form (ledger `52da79031`) |

**#2059 is still open, so these are not in ledger main either** — against the
baseline at the top of this file, the oracle is ahead here too, exactly like
§2. They are listed separately only because the code is upstream's own, not
something written here: when #2059 merges, the oracle matches main with no
further work, whereas each §2 entry has to be deleted by hand.

If #2059 is ever closed unmerged, these move into §2 and need filing like the
rest.

The three collision reproducers from `52da79031` are in `smoke_test.go`.
