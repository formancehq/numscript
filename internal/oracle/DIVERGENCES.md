# Oracle vs ledger vs numscript

`internal/oracle/` is a copy of the ledger repo's `internal/machine`, vendored
so the differential harness has a second implementation to compare against. It
only works as a check while it stays faithful to ledger, so every edit to it is
written down here.

Baseline: `formancehq/ledger` `internal/machine`, main at `8880965c2`.
Everything below was measured by running all three engines, not recalled.

## The six differences

| # | what | ledger | oracle | numscript |
|---|---|---|---|---|
| 1 | negative send amount | `insufficient funds` | `insufficient funds` | rejects the amount |
| 2 | `kept` | does not consume the funding | consumes it | consumes it |
| 3 | `save` past the balance | balance goes negative | floors at zero | floors at zero |
| 4 | source-side negative `max` | rejects the script | rejects the script | clause contributes zero |
| 5 | negative bounded overdraft cap | applies it as-is | applies it as-is | clamps it to zero |
| 6 | allotment portions above 100% with `remaining` | rejects the script | rejects the script | rejects the script (fixed 2026-09-25) |

The oracle matches ledger on 1, 4, 5 and 6. It does not on 2 and 3, the two
places it cannot be trusted to check numscript.

`TestDifferentialSweep` is green: 0 divergences over 3000 seeds, 112 of those
scripts tolerated under #1. With the generator's scenario blocks and an oracle
still faithful to ledger on `save`, the same seeds gave 9 divergences, all #3;
the numbers are under #3. The sweep also prints what it reached and what it
could compare -- see "What the sweep compares".

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

The sweep reports how often it fires -- currently 112 of 3000 scripts, 3.7%. If
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

Re-read 2026-09-18 against the interpreter's `pushReceiver` and
`fundsQueue.Pull`: with the two edits, both engines consume a kept funding from
the front of the pool, in inorder, allotment and nested destinations alike. The
one difference left is when the kept amount is credited back -- immediately on
the interpreter, at `OP_REPAY` on the oracle -- and no statement can read a
balance in between.

**Open question:** whether ledger should adopt this. If it ever does, delete the
two edits rather than keep them.

## 3. `save` past the balance

**Ledger lets `save` push a balance negative. numscript floors it at zero, and
so does the oracle since 2026-09-18.**

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
| oracle | `src->dst 100` |
| numscript | `src->dst 100` |

numscript floors the save at the real balance of 50, leaving 0, so the whole
100 overdraft is available. Ledger subtracts unfloored to -50, so the overdraft
only brings it back to 50. Same script, twice the money moved.

The floor also works from the other side of zero. On a balance of -50,
`save [COIN 0]` raises it to 0 on numscript and leaves it at -50 on ledger, so
the same bounded overdraft of 100 then moves 100 against 50: a save of nothing
changes the outcome.

Verified at ledger main `35bb7ffac`: `save [COIN 100] from @src` against a
balance of 50 gives `balances=map[@src:map[COIN:-50]]`.

**The oracle was changed to match numscript**, in `vm/machine.go`, `OP_SAVE`,
`case machine.Monetary`: the result of `balance.Sub(amount)` is floored at
zero. The `*` branch already agreed -- both engines zero a positive balance and
leave a negative one alone.

This is the second time the oracle floors. It originally did, with a comment
saying it matched the interpreter; that was reverted as the oracle bending
toward the engine it exists to check, and the divergence was left open. It is
reintroduced here as a recorded decision: numscript's behaviour is the one
wanted, and ledger is to follow it (see the open question), so the oracle
follows first.

**What it bought.** A sweep that reaches the shape and is green on it. The
generator now builds scenario blocks: a few statements sharing one (account,
asset), setups that move its balance and observers that read it, with amounts
drawn around the balance and the amounts already mentioned (`internal/gen/
scenario.go`). Half the seeds still take the old uniform path unchanged. Over
the sweep's 3000 seeds, counted on the AST by `gen.Shape`:

| | uniform generator | with scenario blocks |
|---|---|---|
| scripts containing a save | 891 | 1097 |
| containing a bounded overdraft | 2284 | 2303 |
| both, on the same account | 272 | 521 |
| both, on the same (account, asset) | 109 | 387 |
| ...in that order | 60 | 320 |
| ...where the save also overdraws | 4 | 26 |
| ...that diverged, oracle faithful to ledger | 0 | 9 |
| scripts both engines ran to completion | 932 | 1138 |

(The earlier grep-based count here -- 686 / 2284 / 130 / 7 / 2 / 0 -- did not
see var-valued and `*` saves; the bounded-overdraft row is the same.)

The nine were seeds 214, 510, 614, 1118, 1353, 2054, 2379, 2541 and 2792: seven
`aggregated amount differs`, two `aggregated posting set differs`. Eight are
observed through a later bounded overdraft; seed 1118 through a `world` funding
followed by a plain draw, where both balances are positive but differ by the
floored amount. Seeds 510 and 2541 start from a negative balance. Nothing in
the generator names `save` and `overdraft` together: the block draws one of nine
setup kinds and one of six observer kinds uniformly, and this is one cell.
`go run ./cmd/rundiff -seed 214 -n 1 -verbose` in `internal/difftest` prints the
first script; it diverges against a checkout without the floor.

**What it cost.** On `save`, as on `kept`, the oracle is no longer independent
of numscript. A `save` regression in the interpreter will not be caught here.
`TestOracleSaveFloorsAtZero` and `TestOracleSaveOnNegativeBalanceFloorsAtZero`
cover it instead, with ledger's numbers in the comments.

**Open question:** ledger adopting the floor. Ledger PR #2109
(https://github.com/formancehq/ledger/pull/2109) makes the same change with the
same two cases as tests. If it lands, delete the edit rather than keep it.

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
interpreter's ground-truth behaviour, not just catching up to ledger. With a
fallback source covering the shortfall, as here, the interpreter commits and
the oracle fails outright on `OP_TAKE_MAX`'s guard; `Compare` tolerates that by
name, `negative max clause`, the same tolerance as the destination-side twin
below. Without a fallback the interpreter instead fails on missing funds, a
different classification than the oracle's rejection, which `Compare` still
flags as a mismatch (`TestMissingFundsClassificationMismatchStillCaught`).

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

The destination-side version -- `max [COIN 0] - [COIN 10] to @a` -- is **not**
fixed, although an earlier note here said it was. The interpreter still clamps
the clause to zero (`sendTo`, `*parser.DestinationInorder`) and routes the
amount through `remaining`; ledger rejects the script in `OP_TAKE_MAX`.
`Compare` does not see it: neither side reports missing funds, and this exact
shape -- the other side's `RunErr` is `OP_TAKE_MAX`'s guard rejecting a
negative `max` clause, `SideResult.NegativeMaxReject` -- is tolerated by name,
`negative max clause`, and counted. Any other one-sided runtime failure is a
mismatch: the tolerance does not cover a one-sided failure in general, only
this one. It fired 0 times over 3000 seeds.
`TestDestinationSideNegativeMaxTolerated` pins the exact asymmetry.

**Open question:** whether numscript should reject.

## 5. Negative bounded overdraft cap

```numscript
send [COIN *] (
  source = @acc0 allowing overdraft up to [COIN 0] - [COIN 10]    // acc0 has 50
  destination = @acc1
)
```

| engine | postings |
|---|---|
| ledger | `acc0->acc1 40` |
| oracle | same |
| numscript | `acc0->acc1 50` |

numscript clamps the cap to zero (`tryTakingUpTo` and `takeAll`,
`*parser.SourceOverdraft`), so the clause behaves as no overdraft at all.
Ledger adds the negative cap to the balance in `withdrawAll`, so 10 of the 50
are out of reach, and after the draw sets the balance to the negated cap, +10.

Found by reading, not by the sweep: `monetary()` in `internal/gen` never emits
a negative amount in cap position (see its doc comment, and #4), so the
generator does not reach it. Pinned in `TestKnownOpenDivergences`.

**Open question:** which of the two is wanted. Left open on both sides.

## 6. Allotment portions above 100% with `remaining`

**Ledger rejects an allotment whose portions sum past 100%. numscript used to
compute a negative `remaining` portion and keep going; it rejects too since
2026-09-25** (`makeAllotment` refuses a negative remaining with
`InvalidAllotmentSum`, the same error its no-`remaining` path always raised).

```numscript
vars {
  portion $p
  portion $q
}

// bound to p = 2/3, q = 2/3

send [COIN 90] (
  source = @world
  destination = {
    $p to @acc1
    $q to @acc2
    remaining to @acc3
  }
)
```

| engine | result |
|---|---|
| ledger | `sum of portions exceeded 100%` |
| oracle | same |
| numscript before the fix | commits: `world->acc1 60`, `world->acc2 30` |
| numscript | `Invalid allotment: portions sum should be 1 (got 4/3 instead)` |

numscript's `makeAllotment` (interpreter.go) computed the remaining portion as
`1 - total` with no sign check, so the parts came out `[60, 60, -30]`. On the
destination side the receivers drained the 90 in order -- 60, then the 30 left,
then nothing for the negative part -- and the statement committed; money moved
that ledger refuses to move. On the source side the same parts reached
`tryTakingExact(-30)`, which failed as `Not enough funds. Needed [COIN -30]
(only [COIN 0] available)`: a missing-funds classification (and a negative
"needed" amount in the message) for a script that has no funds problem, against
the oracle's non-funds rejection. Both directions were Compare mismatches; both
engines now fail for a non-funds reason and Compare agrees.

The literal form -- `2/3 to @acc1` etc. written inline -- diverges identically,
but never reaches Compare: the oracle rejects it at compile time ("sum of known
portions is greater than 100%"), numscript's parser accepts it, and a b-side
compile rejection is a tolerated outcome. Only the var form, which compiles on
both engines and fails at `OP_MAKE_ALLOTMENT`, is visible to the sweep.

The machine checks the sum at runtime in `NewAllotment` (machine/allotment.go);
numscript used to check it only when there was no `remaining` clause. The two
agree on every sum at or below 100%, floored per-part with the leftover units
handed out front-first, including the `remaining` part -- that whole space is
generated and green (see the sweep's reach table).

The generator keeps every emitted sum strictly below 100% when a `remaining`
clause is present (portionsList): a sum of exactly 100% is an oracle-side
compile rejection ("known portions are already equal to 100%"), and above it
the literal form still compares nothing. `TestKnownBugRepros` pins both
directions of the var form, and the interpreter's own
`TestInvalidSourceAllotmentSumOverOneWithRemaining` /
`TestInvalidDestinationAllotmentSumOverOneWithRemaining` pin the rejection
without the harness.

---

## What the sweep compares

Only scripts both engines run to completion have their postings compared: 1032
of 3000. Of the rest, 250 are rejected by the oracle at compile time (the
generator's cleanup pass is best-effort; counted as `b-side compile rejection`),
85 are tolerated under #1, 535 are numscript-only (below) and skip the oracle
altogether, and the remainder fail on both engines for the same missing-funds
reason. A scenario block glued to a random program is often wasted this way,
which is why the generator has a scenario-only strategy.

Since 2026-09-25 the generator also reaches allotment `remaining` clauses,
portions written through vars (only ever inside a `remaining` block, summing
strictly below 100% — see #6) and origin vars chained through a meta-account
var, `balance($a, ...)` with `account $a = meta(...)` (the sweep's reach table
prints the live numbers). Percent-form portions, origin vars in cap positions
and a handful of other shared shapes the generator still cannot emit are pinned
by `TestUncoveredShapeAgreements` instead.

One specific case of one engine moving money the other refused to move is
tolerated by `Compare` and counted, by name, `negative max clause`; see #4.
Any other one-sided runtime failure is a mismatch.

A quarter of generated scripts may draw numscript-only shapes — `oneof`,
colored sources, division-expression portions (`$n/3`, any sign), caps in a
different asset than their statement — and any script actually containing one
skips the oracle legs (counted by name, `numscript-only script, oracle
skipped`) and is compared on `new vs vm` only, where nothing is tolerated.
Asset scaling and account interpolation still have no generator coverage;
they are checked by the fixture corpus (which both numscript engines run) and
`TestNumscriptOnlyShapeAgreements`.

---

## Where each one is pinned

| # | test |
|---|---|
| 1 | the sweep, as a named tolerance. `TestMissingFundsClassificationMismatchStillCaught` pins that the opposite direction is not tolerated |
| 2 | `TestKnownBugRepros` (the engines agree now) and `TestOracleKeptComplex` |
| 3 | `TestKnownBugRepros` (the engines agree now), `TestOracleSaveFloorsAtZero` and `TestOracleSaveOnNegativeBalanceFloorsAtZero` |
| 4 | `TestSourceSideNegativeMaxClauseTolerated`, `TestMissingFundsClassificationMismatchStillCaught` and, destination side, `TestDestinationSideNegativeMaxTolerated` |
| 5 | `TestKnownOpenDivergences` |
| 6 | `TestKnownBugRepros` (the engines agree now), both sides, and the interpreter's own over-100% tests |

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

Nine files differ. Three carry behaviour changes, five edits in all:

| file | why |
|---|---|
| `account.go`, `asset.go`, `internal/accounts/accounts.go`, `internal/assets/asset.go` | vendoring |
| `vm/oracle_types.go`, `vm/run.go` | vendoring |
| `vm/store.go` | vendoring, and the `StaticStore.GetBalances` multi-asset fix (`internal/difftest/regression_test.go`'s `TestKnownBugRepros`) |
| `vm/machine.go` | vendoring, the `save` floor (#3), and the `UnresolvedResourceBalances` duplicate-var fix (`TestKnownBugRepros`) |
| `script/compiler/destination.go` | `kept` (#2), two sites |

Anything else that shows up is undocumented drift. Either record it here or
remove it.
