# The verify query language

`numscript verify` checks whether something is true about a script, for **every
possible** starting balance and input — or finds a concrete example where it
happens. You write the "something" as a short query.

```
numscript verify script.num '<query>'
```

## The two questions you can ask

Every query is one of two kinds:

| Prefix   | Question                        | Answer you get                          |
|----------|---------------------------------|-----------------------------------------|
| `prove:` | Is this **always** true?        | ✅ Proved, or ❌ a counterexample        |
| `find:`  | Can this **ever** happen?       | ✅ a witness example, or ❌ Impossible   |

If you don't write a prefix, it defaults to `prove:`.

```
numscript verify script.num 'prove: !fail => received("dest", "USD/2") == 10'
numscript verify script.num 'find: fail'
```

When the answer is a counterexample or a witness, you also get the exact
starting balances that produce it.

## What you can talk about

These read values out of the script's execution. Each takes an **account** and
an **asset**, both in quotes.

| Predicate                         | Meaning                                     |
|-----------------------------------|---------------------------------------------|
| `start_balance("acc", "USD/2")`   | balance before the script runs              |
| `end_balance("acc", "USD/2")`     | balance after the script runs               |
| `sent("acc", "USD/2")`            | how much the account paid out               |
| `received("acc", "USD/2")`        | how much the account got                    |
| `volumes("acc", "USD/2")`         | net change (`received - sent`)              |
| `fail`                            | true if the script aborts (e.g. not enough funds) |

## How to combine them

Numbers:

```
==   !=   <   <=   >   >=        (compare)
+    -    *                      (arithmetic)
```

Logic:

```
!        not
&&       and
||       or
=>       implies      (if ... then ...)
<=>      if and only if
```

Use parentheses to group. Numbers are plain integers.

## Examples

```
# The destination always ends up with exactly 100 (when the script succeeds):
prove: !fail => received("dest", "USD/2") == 100

# Money is conserved: what the source pays equals what the destination gets:
prove: sent("src", "USD/2") == received("dest", "USD/2")

# The script can fail (there's some balance where it aborts):
find: fail

# Nobody's balance ever goes negative:
prove: end_balance("alice", "USD/2") >= 0

# A split always divides 10 into 4 + 6:
prove: received("d1", "USD/2") == 4 && received("d2", "USD/2") == 6

# When it fails, nothing moves:
prove: fail => received("dest", "USD/2") == 0
```

## A note on failing scripts

If a script aborts, it moves no money — so `sent`, `received` and `volumes` are
all `0` on a failed run. That's why you'll often guard a claim with
`!fail => ...` ("if it doesn't fail, then ...").

## Reading the result

- ✅ **Proved** — true for every possible input. Done.
- ❌ **Counterexample** — not always true; you get one input where it breaks.
- ✅ **Witness found** — yes it can happen; you get an input that does it.
- ❌ **Impossible** — it can never happen.
- ⚠️ **Unknown** — the solver couldn't decide (rare; usually a timeout).

Add `--smt` to any command to see the exact math sent to the solver.
