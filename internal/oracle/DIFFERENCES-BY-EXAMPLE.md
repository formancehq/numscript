# Oracle vs ledger vs numscript, by example

The exhaustive list lives in [`DIVERGENCES.md`](./DIVERGENCES.md). This file is
the short version: every remaining behavioural difference, each as one script
you can run, with what all three engines actually do.

Measured 2026-09-17 against:

- ledger `internal/machine` at main `8880965c2`
- numscript `test/difftest-oracle-harness`, `kept` aligned to the interpreter

Oracle and numscript results came from the difftest harness (`runOracle` /
`runNew`); ledger results from the real machine in a worktree at main. Not
recalled, not inferred.

There are **four** behavioural differences. They split into two kinds, and the
distinction is the whole point of the oracle:

- **two** where the oracle differs from ledger (§A, §B) — liabilities, kept on
  purpose
- **two** where numscript differs from both (§C, §D) — the oracle siding with
  ledger is it doing its job

§B changed on 2026-09-17: the oracle used to side with ledger on `kept` and now
sides with the interpreter, which is what took `TestDifferentialSweep` green.

Everything else in the `DIFFERS` list is vendoring plumbing: import rewrites and
local `Account`/`Zero` types. No semantics. See `DIVERGENCES.md` §1.

---

## A. Negative send amount

The only place the oracle is not faithful to ledger.

```numscript
send [COIN 0] - [COIN 90] (
  source = @alice        // alice has 1000
  destination = @bob
)
```

| engine | result |
|---|---|
| **ledger** | `account(s) @alice had/have insufficient funds` |
| **oracle** | `cannot send a monetary with a negative amount: [COIN -90]` |
| **numscript** | `Cannot send negative amount: -90` |

Alice has 1000 and is being asked for −90. The amount is invalid at *any*
balance, so reporting it as a funding problem points you at the wrong thing —
and ledger's API maps it to HTTP 400 `INSUFFICIENT_FUND`.

Upstream agrees it is wrong: `vm/machine_negative_amount_test.go` pins the
behaviour with a `BUG:` comment, and separately pins the four source shapes that
compile to `OP_TAKE_MAX` and *do* reject correctly. But the fix (ledger PR
#2060) was closed unmerged, so main still has no guard on `OP_TAKE` and the
oracle carries one alone.

**Why the oracle keeps it:** without the guard, 161 of 3000 swept scripts (5.4%)
diverge on this one classification and drown out everything else. Measured by
removing it and re-running the sweep. It costs a wrong error message, not wrong
postings — the send fails either way.

## B. `kept` consumes the funding, or doesn't

```numscript
send [COIN *] (
  source = { @acc1  @acc0 }        // both 1000
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

| engine | postings |
|---|---|
| **ledger** | `acc1→dst 1000`, `acc0→dst 600`, `acc1→sink 0` |
| **oracle** | `acc1→dst 600`, `acc0→dst 1000`, `acc1→sink 400` |
| **numscript** | same as oracle |

Both send 1600 to `@dst` — the first statement agrees on totals. They disagree
on *who kept the 400*. Ledger does not consume a kept funding: it goes back
into the pool, funds the rest of the destination, and the keeping lands on
`@acc0` (the bottom), draining `acc1`. numscript consumes it, so `@acc1` (the
front) keeps. The second statement reads `@acc1`, and the money is either there
or gone.

Ledger's own `TestKeptComplex` asserts its behaviour deliberately, so this is
intended upstream, not a bug.

**The oracle was changed to match numscript** (`DIVERGENCES.md` §2 ②) — two
sites in `script/compiler/destination.go`. That took `TestDifferentialSweep`
from 7 failing + 85 tolerated scripts to zero of each, and let `Compare`'s
`kept source attribution` tolerance be deleted outright.

The cost: on `kept`, the oracle is no longer an independent check on the
interpreter. `TestOracleKeptComplex` carries ledger's original expectations in
its comment so the delta stays visible.

> The `acc1→sink 0` in ledger's row is real: the legacy machine emits
> zero-amount postings and the rewrite does not, and `run_oracle.go` drops them
> for comparison. A harness normalization, not a divergence.

## C. `save` past the balance

```numscript
save [COIN 100] from @src        // src only has 50

send [COIN *] (
  source = @src allowing overdraft up to [COIN 100]
  destination = @dst
)
```

| engine | postings |
|---|---|
| **ledger** | `src→dst 50` |
| **oracle** | `src→dst 50` |
| **numscript** | `src→dst 100` |

numscript floors the save at the real balance (50), leaving 0 tracked, so the
100 overdraft is fully available. Ledger subtracts unfloored to −50, so the
overdraft only brings it back to 50. Same script, double the money moved.

The sweep does not reach this (the generator rarely produces a save-overdraw
followed by a bounded-overdraft draw on the same account), so it is pinned
directly in `TestKnownOpenDivergences`.

## D. Source-side negative `max`

```numscript
send [COIN 100] (
  source = {
    max [COIN 0] - [COIN 10] from @a    // both 1000
    @b
  }
  destination = @dst
)
```

| engine | result |
|---|---|
| **ledger** | `cannot send a monetary with a negative amount: [COIN -10]` |
| **oracle** | same |
| **numscript** | `b→dst 100` — the clause contributes nothing, `@b` covers it |

Known on the numscript side and deliberately unfixed: closing it means changing
the interpreter's ground-truth behaviour, not just catching up to ledger.
`Compare` tolerates it by matching missing-funds *classification* rather than
error text.

---

## Where each one stands

| | difference | oracle vs ledger | decision needed |
|---|---|---|---|
| A | negative send amount | **differs** | re-file upstream, or drop the guard and accept 161/3000 diverging |
| B | `kept` consumes the funding | **differs** | whether ledger should adopt it. Until then the oracle cannot check `kept` |
| C | `save` past the balance | agrees | whether numscript should stop flooring |
| D | source-side negative `max` | agrees | whether numscript should reject |

A and B are the two places the oracle is not an independent check, and both are
there on purpose. A is a documentation-and-filing question. B is settled inside
numscript and unsettled against ledger — the sweep is green either way now, so
nothing forces the question, which is exactly why it is written down here.

C and D are product questions about numscript, and the answer to either may be a
change to the interpreter rather than to the oracle.
