# The numscript prover

`numscript prove` checks a boolean property about one script for every initial balance and answers `valid`, `counterexample` or `unknown`. The script is compiled to the IR (see [ir-textual-format.md](ir-textual-format.md)), the IR is encoded into one quantifier-free SMT problem, and an external solver decides it. Everything lives under [internal/prove/](internal/prove/).

## What it proves, and what it does not

A property is a formula over the quantities of one run: the initial balance of each (account, asset) pair, the totals moved during the run, and whether the run succeeded. The prover checks it for every assignment of initial balances, unbounded integers with negatives included, not for a sample. The answer is one of three: `valid` (no initial balance falsifies the property), `counterexample` (concrete initial balances on which the property is false, replayed through the real VM before being reported), or `unknown` (the solver hit its time limit).

The model carries the run outcome, success or the first failing instruction and its kind, and per (account, asset) pair three totals: `sent`, the sum of postings out of the account, `received`, the sum of postings into it, and `moved(a→b)`, the sum of postings from `a` to `b`. `after` is defined as `before + received − sent`. What a property cannot name is not modeled: the number, order or shape of the postings, metadata, script variables, asset colors, scopes and asset scaling. The exact rejection list is `prescan` in [internal/prove/encode/encode.go](internal/prove/encode/encode.go) and is reproduced in [Supported scripts](#supported-scripts-v1).

| | |
| --- | --- |
| checked for | every initial balance of every (account, asset) pair the script or the property names, any integer |
| answers | `valid`; `counterexample` with initial balances the VM confirms; `unknown` on solver timeout |
| modeled | success or which instruction fails and why; `before`, `sent`, `received`, `moved`, `after` per pair |
| not modeled | posting list shape and order, metadata, script vars, colors, scopes, scaling |

## Pipeline and where things live

A `.num` file is parsed and lowered with `compiler.CompileToIR`; a `.ir` file is read with `ir.Parse`. `encode.Encode` first runs `ir.Typecheck`, `ir.Assemble` and `vm.Verify` on the stream, so every register read is defined, every jump is forward and marks balance on every path, then emits one SMT-LIB2 model of the run. For each property, `prop.Lower` turns the formula into a Bool term, its negation is asserted on top of the model, and the solver subprocess answers `(check-sat)`. `unsat` is `valid`. `sat` comes with a model: the initial balances are read off it, `encode.Replay` runs the same IR on `vm.Exec` against a static store holding those balances, `encode.Compare` checks every output of the model against the VM, and `prop.Eval` evaluates the formula on the VM run. Only when the VM agrees and the formula is false on it is the result a `counterexample`.

| what | where |
| --- | --- |
| `.num` → `[]ir.Instr` | `compiler.CompileToIR` in [internal/compiler/compiler.go](internal/compiler/compiler.go) |
| `.ir` → `[]ir.Instr` | `ir.Parse` in [internal/ir/parse.go](internal/ir/parse.go) |
| precondition on the IR | `ir.Typecheck`, `ir.Assemble`, `vm.Verify`, called at the top of `encode.Encode` |
| IR → SMT model | `encode.Encode` in [internal/prove/encode/encode.go](internal/prove/encode/encode.go); merging at labels in `state.go`, strings and portions in `values.go`, pull/send/save/balance/marks in `funds.go`, outputs and `UnsupportedError` in `problem.go` |
| property text → AST | `prop.ParseFile`, `prop.ParseInline` in [internal/prove/prop/parser.go](internal/prove/prop/parser.go), tokens in `lexer.go` |
| AST checks | `prop.Typecheck`, `prop.Atoms` in [internal/prove/prop/typecheck.go](internal/prove/prop/typecheck.go) |
| AST → SMT term; AST on concrete values | `prop.Lower` in `lower.go`; `prop.Eval` in `eval.go` |
| terms with folding, printer, script, solver driver, evaluator | [internal/prove/smt/](internal/prove/smt/) (`term.go`, `print.go`, `script.go`, `solver.go`, `eval.go`) |
| counterexample replay | `encode.Replay`, `encode.Compare` in [internal/prove/encode/replay.go](internal/prove/encode/replay.go) |
| orchestration, statuses, exit codes, reports | `prove.Check` in [internal/prove/prove.go](internal/prove/prove.go), `result.go`, `report.go` |
| CLI | `numscript prove` in [internal/cmd/prove.go](internal/cmd/prove.go), `numscript ir` in [internal/cmd/ir.go](internal/cmd/ir.go) |
| fixtures | [internal/prove/testdata/cases/](internal/prove/testdata/cases/): `<name>.num` or `<name>.ir`, plus `.props` and `.expected` |

Both commands are registered as hidden experimental commands in `internal/cmd/root.go`: they always run, and `NUMSCRIPT_EXPERIMENTAL_CLI=1` lists them in `--help`. `numscript ir <file.num> [-o path]` prints the IR the prover sees, with `-` for stdin, so `numscript ir x.num | numscript assemble -` also works.

## The property language

### Lexical structure

| token | form | notes |
| --- | --- | --- |
| ACCOUNT | `@` then segments of `[a-zA-Z0-9_-]+` joined by `:` | as numscript: `@src`, `@users:001`, `@a-b` |
| ASSET | `[A-Z][A-Z0-9]*` optionally `/[0-9]+` | as numscript: `USD/2`, `COIN` |
| INT | `[0-9]+` with `_` between digit groups | `1_000_000`; no sign, `-` is an operator |
| NAME | `[a-z_][a-z0-9_]*` | keywords and property names |
| operators | `( ) , : + - * / % == != < <= > >= => <=>` | a lone `=` is "unexpected character '='" |
| comments | `//` to end of line, `/* ... */` | an unterminated block comment is an error |

Whitespace and newlines are insignificant. Keywords: `before after sent received moved delta success fails and or not true false prop min max`. Three token kinds are lexed only to be rejected with a pointed message: `$name` ("script variables are not supported in properties yet"), `"string"` ("strings are not supported in properties"), and `[` where an expression is expected ("monetary literals are not allowed in properties; amounts are plain integers in the asset's minor unit: write 100"). The names `forall exists in accounts assets` ("quantifiers are not supported yet"), `sum` ("sums are not supported yet") and `meta` ("metadata is not supported yet") are reserved the same way.

### Grammar

```
file        = { "prop" NAME ":" formula } ;
inline      = [ NAME ":" ] formula ;                  the --prop form
formula     = iff ;
iff         = implies [ "<=>" implies ] ;             non-associative
implies     = disj [ "=>" implies ] ;                 right-associative
disj        = conj { "or" conj } ;
conj        = negation { "and" negation } ;
negation    = "not" negation | comparison ;
comparison  = sum [ ("==" | "!=" | "<" | "<=" | ">" | ">=") sum ] ;   non-associative
sum         = term { ("+" | "-") term } ;
term        = unary { ("*" | "/" | "%") unary } ;
unary       = "-" unary | primary ;
primary     = INT | "true" | "false" | "success" | "fails" | "(" formula ")"
            | ("before" | "after" | "sent" | "received" | "delta") "(" ACCOUNT "," ASSET ")"
            | "moved" "(" ACCOUNT "," ACCOUNT "," ASSET ")"
            | ("min" | "max") "(" sum "," sum { "," sum } ")" ;
```

| level | operators | associativity |
| --- | --- | --- |
| 1 (loosest) | `<=>` | none: `a <=> b <=> c` is an error, write `(a <=> b) <=> c` |
| 2 | `=>` | right: `a => b => c` is `a => (b => c)` |
| 3 | `or` | left |
| 4 | `and` | left |
| 5 | `not` | prefix, over one comparison: `not a < b` is `not (a < b)` |
| 6 | `== != < <= > >=` | none: `1 < 2 < 3` is an error, write `(1 < 2) and (2 < 3)` |
| 7 | `+ -` | left |
| 8 | `* / %` | left |
| 9 (tightest) | unary `-` | prefix |

### Atoms

| atom | type | value |
| --- | --- | --- |
| `before(@a, X)` | int | initial balance of `a` in `X` |
| `after(@a, X)` | int | `before + received − sent` |
| `sent(@a, X)` | int | total posted out of `a` in `X` |
| `received(@a, X)` | int | total posted into `a` in `X` |
| `moved(@a, @b, X)` | int | total posted from `a` to `b` in `X` |
| `delta(@a, X)` | int | sugar for `after(@a, X) - before(@a, X)` |
| `success`, `fails` | bool | the run committed / aborted; each is the negation of the other |
| `min(x, y, ...)`, `max(x, y, ...)` | int | at least two integer arguments |

Arithmetic is linear. `*` needs a literal operand, `/` and `%` a positive literal divisor, where a literal is any atom-free constant expression (`2 * 3 * sent(@a, USD/2)` and `sent(@a, USD/2) / min(2, 3)` typecheck). The messages are "multiplication must have a literal operand (linear arithmetic only)", "division must be by a positive literal" and "modulo must be by a positive literal". `/` and `%` are Euclidean, like SMT-LIB `div`/`mod` and Go's `big.Int.Div`/`Mod`: `-7 / 2 == -4` and `-7 % 2 == 1`, the remainder is never negative. Amounts are plain integers in the asset's minor unit, `100` where numscript writes `[USD/2 100]`; a monetary literal is rejected. The formula must be boolean, and the typechecker reports every sort error ("expected a boolean expression, found an integer", "expected an integer expression, found a boolean").

An account or asset the script never names is a typecheck error listing the known ones: `account @dset is not referenced by the script (known: @dest, @src, @world)`, `asset EUR/2 is not referenced by the script (known: USD/2)`. It is an error and not a silent zero because a pair the script never touches has `sent == received == 0` and `after == before` by construction, so a property about it is either trivially true or a typo.

### Files and inline properties

A `.props` file is a sequence of `prop name: formula`, comments anywhere; names must be unique ("property "a" is defined twice"), and an empty file is "no properties found". `--prop "[name:] formula"` is one property; without a name it is called `prop1`, `prop2`, ... in order. `--prop` and `--props` combine; with neither, `<script>.props` next to the script is read (`simple-send.num.props` for `simple-send.num`).

Errors print `line:col: msg` (1-based) followed by the source excerpt from `parser.Range.ShowOnSource`, prefixed with the file path for a `.props` file and with `--prop[N]` for the N-th inline property:

```
--prop[1]: 1:12: expected an asset but found '['
  0 | sent(@src, [USD/2 100]) == 0
    |            ~
```

The command then prints `invalid property` (or `invalid properties file`) and exits 3. Typecheck errors, unknown accounts or assets, nonlinear arithmetic and sort errors, are found after encoding and print as `property "name": line:col: msg` without an excerpt.

## Semantics of a run

For initial balances `B`:

| atom | run succeeds | run fails |
| --- | --- | --- |
| `before(@a, X)` | `B(a, X)` | `B(a, X)` |
| `sent(@a, X)` | Σ postings out of `a` in `X` | 0 |
| `received(@a, X)` | Σ postings into `a` in `X` | 0 |
| `moved(@a, @b, X)` | Σ postings `a → b` in `X` | 0 |
| `after(@a, X)` | `before − sent + received` | `before` |
| `delta(@a, X)` | `after − before` | 0 |
| `success` / `fails` | true / false | false / true |

A failed run applies nothing. The ledger commits a transaction or rejects it whole, so on failure every balance stays as it was and every total is 0. For `send [USD/2 100]` from `@src` to `@dest`, `after(@dest, USD/2) == before(@dest, USD/2) + 100` has the counterexample `before(@src) == -1` (the run fails, nothing moves); `success => after(@dest, USD/2) == before(@dest, USD/2) + 100` is the property about the committed run, and `before(@src, USD/2) >= 100 => success` the one about when it commits (`simple-send.num`).

A self-transfer counts on both sides: a posting `a → a` adds to `sent(@a)`, `received(@a)` and `moved(@a, @a, X)`, and leaves `after(@a)` unchanged.

`kept` funds are refunded, not sent. `remaining kept` returns the undistributed remainder to its source before the run ends, so `sent(@src)` counts only what reached a destination. In `kept.num` (`send [USD/2 100]`, `1/2 to @dest`, `remaining kept`) `success => sent(@src, USD/2) == 50` and `before(@src, USD/2) == 60 => fails` are both valid: 50 moves, but 100 must be available.

`save` only reserves. It lowers what a later `pull_account` or `balance()` sees, it is never a posting, so `after` ignores it. In `save-then-send.num` (`save [USD/2 50] from @src`, then `send [USD/2 100]` from `@src`) `success => before(@src, USD/2) >= 150` is valid and `before(@src, USD/2) >= 100 => success` has the counterexample `before(@src) == 100`.

Negative initial balances are in the universe. A `pull_account` with overdraft 0 sees `max(0, balance)`, so a negative balance sources nothing, but it is still a balance the property is checked on: `after(@src, USD/2) >= 0` is false at `before(@src) == -1` because the run fails and `after == before`. Exclude them with an antecedent, `before(@src, USD/2) >= 0 => ...`, as `send-all.num`'s `drains` does.

Properties in a file are independent checks: each is its own `check-sat`, and none assumes another.

Vacuity: `success => φ` is valid for every φ when no initial balance makes the run succeed. `vacuous.num` sends `2/3 from @a` and `2/3 from @b`, which fails `assert_leftover` on every balance, so `success => received(@dest, USD/2) == 1_000_000` is valid. The report header's `success reachable: no` is the tell; it also says whether failure is reachable (`allotment.num` and `send-all.num` never fail).

Three more fixtures show the atoms at work. `capped-source.num` (`send [USD/2 100]` from `max [USD/2 30] from @a` then `@b`, to `@dest`): `success => sent(@a, USD/2) == min(30, max(0, before(@a, USD/2)))` and `success => moved(@a, @dest, USD/2) + moved(@b, @dest, USD/2) == 100` are valid; `sent(@a, USD/2) == 30` has the counterexample `before(@a) == -1, before(@b) == -1`, where the run fails and `sent(@a) == 0`. `oneof-source.num` (`source = oneof { @a @b }`): `success <=> (before(@a, USD/2) >= 100 or before(@b, USD/2) >= 100)` and `before(@a, USD/2) >= 100 => sent(@a, USD/2) == 100` are valid, `before(@b, USD/2) >= 100 => sent(@b, USD/2) == 100` is refuted at `before(@a) == 100, before(@b) == 101` (the first branch that can pay wins), and `before(@a, USD/2) == 50 and before(@b, USD/2) >= 100 => after(@a, USD/2) == 50` is valid because the 50 pulled from `@a` on the branch that fell short is repaid by `mark_rewind`. `allotment.num` (`send [USD/2 *]` from `@src`, `1/2 to @a`, `remaining to @b`): `success => received(@a, USD/2) == sent(@src, USD/2) / 2 + sent(@src, USD/2) % 2` is valid, each share being floored and the missing unit added back to the first clause, and `received(@a, USD/2) == received(@b, USD/2)` is refuted at `before(@src) == 3` (`@a` gets 2, `@b` gets 1).

## Supported scripts (v1)

Supported IR: int, string and bool constants; `add_int`, `sub_int`, `$d += $r`, `$d -= $r`, `neg_int`, `lt_int`, `eq_int`, `is_zero`, `not`; the four copies; `add_string`, `str_eq`, `int_to_string`, `portion_to_string`, `monetary_to_string` (a string is tracked as a constant when its inputs are); `mk_portion`, `add_portion`, `sub_portion`, `mul_portion`, `lt_portion`, `eq_portion`, `int_to_portion`, `portion_to_int`; `jmp`, `jmp_if_true`, `jmp_if_false` and labels; `set_current_asset`, `pull_account`, `send_to_account`, `save`, `balance`; `mark_push`, `mark_commit`, `mark_rewind`; `check_enough_funds`, `assert_non_negative_amount`, `assert_non_negative_balance`, `assert_non_negative_portion`, `assert_leftover`, `assert_leftover_exact`, `assert_same_asset`, `assert_valid_account`. In numscript terms: `send` with literal amounts and `*`; sources `@account`, `@world`, `max [..] from`, inorder blocks, `oneof` (behind its feature flag, which the script declares itself), allotments with literal portions and `remaining`; destinations with inorder, allotments and `kept`; `save`; `balance()` and arithmetic in amounts.

Ignored: `set_tx_meta`, `set_account_meta`. They change no balance.

Rejected with `UnsupportedError{Index, Text, Reason}`, printed as `cannot prove: unsupported instruction N `text`: reason`, exit 3. `prescan` walks the program in text order before encoding and names the first hit, reachable or not:

| instruction | reason |
| --- | --- |
| `load_var<int>`, `load_var<str>` | script variables are not supported |
| `meta<str>`, `meta<int>`, `meta<portion>`, `meta_monetary` | metadata reads are not supported |
| `assert_valid_color`; a `color:` operand on `pull_account` | asset colors are not supported |
| `assert_valid_scope`, `assert_unscoped`; a `scope:` operand on `pull_account`, `send_to_account`, `save`, `balance` | scoped accounts are not supported |

The pass itself rejects what it cannot keep linear or constant:

| situation | reason |
| --- | --- |
| a non-constant string reaches an account or asset operand, `str_eq`, `assert_same_asset` or `assert_valid_account` | the value of string register $rN is not a constant |
| `set_current_asset` of a non-constant, then a pull, send or rewind | the current asset is not a constant |
| `mul_portion` of two non-constant portions | nonlinear portion arithmetic |
| `mk_portion` with a non-constant denominator | nonlinear portion arithmetic: the denominator is not a constant |
| a string register, a portion register or the current asset holding different constants on paths that join at a label | string register $rN takes different values on joining paths (likewise `portion register`, `the current asset differs`) |
| more than 255 live registers in one bank, or a program too large to assemble | the `ir.Assemble` error text, at index -1 |
| any other instruction | not modeled |

The compiler never emits the joining-paths cases from numscript. Asset scaling is rejected by the compiler (`feature not implemented: scaling`) before the prover sees it.

The known accounts and assets, which the typechecker holds property atoms to, are collected syntactically: `prescan` follows string constants through `str_copy` and `add_string` in text order, ignoring jumps, and records every (account, asset) pair reaching a `pull_account`, `send_to_account`, `save` or `balance` operand (the current asset for pulls and sends, the explicit asset operand for `save` and `balance`); the pass adds whatever it touches. An account on a branch the model folds away, such as a run that always fails before reaching it, is still known.

## Quantifiers

The implicit universal over initial balances is the whole meaning of "holds". Every `before(@a, X)` is a free `Int`; the prover asserts `model ∧ ¬φ` and asks whether it is satisfiable. `unsat` means no balance falsifies φ; a model is one concrete counterexample. The user never writes this quantifier.

Over accounts and assets: possible without SMT quantifiers, useful, not in v1. Without vars the touched (account, asset) set is finite and fixed at encode time, so `forall @x in accounts: ψ` is a finite conjunction over it, `exists` a finite disjunction, each plus one "untouched" witness whose delta is 0, and `sum(@x in accounts: delta(@x, USD/2)) == 0` is a finite sum. The two properties that want this, conservation and "nobody but @world ends negative", are aggregates over that set. The syntax is reserved (`forall exists sum in accounts assets`) and the report header prints the touched set, so the hand-written expansion is easy today: `simple-send.num`'s `conservation` is `delta(@src, USD/2) + delta(@dest, USD/2) == 0`.

Over integers: possible in principle, quantified linear integer arithmetic is decidable, and not needed. The implicit universal already covers "for any balance", and an existential over an amount is a fresh free variable. The cost would be robustness: a quantified problem leaves `QF_LIA`, and `unknown` stops meaning "timeout".

With vars later: account-typed vars make the touched set symbolic. The finite expansion survives by case-splitting on aliasing, `$dest` is one of the k literal accounts or a fresh (k+1)-th, which is also why an SMT quantifier will still not be needed then.

v1 ships only the implicit universal.

## Reading the output

`numscript prove internal/prove/testdata/cases/simple-send.num`:

```
internal/prove/testdata/cases/simple-send.num: touches @dest, @src in USD/2; success reachable: yes; failure reachable: yes

  ✓ enough_means_success         valid          10ms
  ✓ dest_credited                valid          8ms
  × src_never_negative           counterexample 8ms
  ✓ src_not_negative_on_success  valid          8ms
  ✓ conservation                 valid          6ms
  ✓ moved_all                    valid          6ms
  ✓ fails_below_100              valid          6ms

× src_never_negative: after(@src, USD/2) >= 0
  initial balances:
    | Account | Asset | Balance |
    | dest    | USD/2 | 0       |
    | src     | USD/2 | -1      |
  outcome: fails at instruction 20 `check_enough_funds($r9, $r2)`: missing funds
  atoms:
    after(@src, USD/2) = -1

6 valid, 1 counterexample
```

The header names the touched accounts and assets and answers two extra `check-sat`s, the model with `success` asserted and with `fails` asserted (`yes`, `no`, or `unknown` when the solver could not tell). Then one line per property with its status and the solver plus replay time, one block per property that is neither valid nor unknown, and the counts. A counterexample block shows the formula, the initial balances of every pair the model knows (property-only pairs included), the outcome (`success`, or the failing IR instruction and the failure kind), and the value of each atom of the formula on the VM run.

| status | icon | meaning | exit |
| --- | --- | --- | --- |
| `valid` | ✓ | `unsat`: no initial balance falsifies the property | 0 |
| `counterexample` | × | `sat`; the VM replay agrees with the model on outcome and every total, and the property is false on the VM run | 1 |
| `unknown` | ? | the solver answered `unknown` (its time limit) | 2 |
| `tool-bug` | ! | the model and the VM disagree, or the property is true on the VM run: a bug in the prover, never a finding about the script | 3 |
| `solver-error` | ! | the solver could not be run or rejected the problem | 3 |

The exit code is the worst status over all properties. A `tool-bug` block prints "the prover's model does not match the vm; this is a bug in the prover, not in the script:" followed by each disagreement (`sent(@a,USD/2): model 30, vm 0`) or "the property holds on the vm run". Errors before any check (an unsupported script, an invalid property, no properties, a missing solver) go to stderr with exit 3.

`-o json` renders the same report for machines (trimmed here to two results):

```json
{
  "script": "internal/prove/testdata/cases/simple-send.num",
  "accounts": ["dest", "src"],
  "assets": ["USD/2"],
  "successReachable": true,
  "failureReachable": true,
  "results": [
    { "name": "dest_credited", "formula": "success => after(@dest, USD/2) == before(@dest, USD/2) + 100", "status": "valid", "durationMs": 8 },
    {
      "name": "src_never_negative", "formula": "after(@src, USD/2) >= 0", "status": "counterexample", "durationMs": 8,
      "counterexample": {
        "balances": [ { "account": "dest", "asset": "USD/2", "amount": 0 }, { "account": "src", "asset": "USD/2", "amount": -1 } ],
        "outcome": "fails",
        "failure": { "kind": "missing funds", "instruction": 20 },
        "atoms": { "after(@src, USD/2)": "-1" }
      }
    }
  ]
}
```

`successReachable`/`failureReachable` are `null` when unknown; `failure` is present only when `outcome` is `"fails"`; balance amounts are JSON numbers, atom values are strings; a `tool-bug` result carries `modelVmDiffs`, and a solver or replay error carries `error`.

Flags: `--prop` (repeatable) and `--props <file>`; `--solver z3|cvc5|<path>` (default `z3`; the executable's basename picks the arguments, `z3 -in -smt2 -t:<ms>` or `cvc5 --lang smt2 --tlimit=<ms>`, and any other executable is run with no arguments and must read SMT-LIB2 on stdin); `--timeout` (default `10s` per property, passed as the solver's own soft limit so it answers `unknown` instead of dying, with the process killed 2 s later; `0` disables both); `--emit-smt <dir>` writes each property's complete problem to `<dir>/<name>.smt2`, `--emit-smt -` prints them to stdout under `; ---- <name>` headers. The emitted text ends at `(assert (not property))`; append `(check-sat)` to reproduce a verdict outside the tool.

## How the encoding works

Every initial balance is a free `Int` named `|before(@src,USD/2)|`. The IR has only forward jumps, so text order is execution order and one pass over the instructions suffices: each state write (a register, a running balance, a queue slot, a total) becomes a `define-fun` named after what it writes and the instruction index, `$r9@18` or `|bal(@src,USD/2)@18|`, so the script is a DAG that reads like the IR. A conditional jump clones the state with the condition conjoined into its reach guard and parks the clone on the target label; a label merges the fallthrough and the parked edges, every differing value becoming an `ite` chain over the edges' reach guards. Constant folding in the term constructors is load-bearing: `str_eq("src", "world")` folds to `false`, which kills the `@world` arm outright, and `lt_int(0, 100)` folds to `true`, which collapses the `max(0, amount)` diamond, so the simple send has no merge at all.

`pull_account` with an overdraft reads `available = max(0, balance + max(0, overdraft))` and takes `available`, or `max(0, cap)` when the cap is smaller; without an overdraft (the `@world` arm) it takes `max(0, cap)` without reading any balance; with neither it is an unconditional failure. What it takes becomes a queue slot and is subtracted from the running balance. `send_to_account` drains the slots in pull order, each `take = ite(rem > 0, min(slot, rem), 0)` for a capped send and the whole slot otherwise, and adds the takes to `sent` of the slot's account, `received` of the destination and `moved(a→d)`; a send without a destination (`kept`) refunds the takes to the running balances and touches no total. `save` clamps the running balance (`max(0, bal − amount)`, or `min(bal, 0)` for `save all`) and nothing else; `balance` reads the running balance. `mark_push` opens a region and `mark_rewind` repays every slot pulled inside it, guarded by which push opened the region when a region's end is reached from several pushes (`oneof`). A failing instruction defines a flag `fail@20 = alive ∧ cond` and continues with `alive ∧ ¬cond`, so the flags of different sites are never true together and the counterexample names the first failure; `fails` is their disjunction and every output is masked by it: `sent(k) = ite(fails, 0, …)`, `after(k) = before(k) + (received(k) − sent(k))`. Portions are exact rational linear forms over Int terms, `c₁·t₁ + … + c₀` with rational coefficients, so `mul_portion` by a constant, `add_portion` and comparisons stay linear and `portion_to_int` is `(div (D·form) D)` with `D` the common denominator; the whole problem is `QF_LIA`.

The compiled `send [USD/2 10] (source = @src destination = @dest)` has the overdraft pull at index 18, `check_enough_funds($r9, $r2)` at 20 and `send_to_account(account: $r10)` at 22, and encodes to exactly this (`internal/prove/encode/__snapshots__/encode_test.snap`):

```smt2
(set-logic QF_LIA)
(set-option :produce-models true)
(declare-const |before(@dest,USD/2)| Int)
(declare-const |before(@src,USD/2)| Int)
; 18: $r9 = pull_account(account: $r6, cap: $r4, overdraft: $r7)
(define-fun available@18 () Int (ite (< |before(@src,USD/2)| 0) 0 |before(@src,USD/2)|))
(define-fun $r9@18 () Int (ite (< 10 available@18) 10 available@18))
(define-fun |bal(@src,USD/2)@18| () Int (- |before(@src,USD/2)| $r9@18))
(define-fun fail@20 () Bool (not (= $r9@18 10)))
(define-fun alive@20 () Bool (= $r9@18 10))
; 22: send_to_account(account: $r10)
(define-fun |bal(@dest,USD/2)@22| () Int (+ |before(@dest,USD/2)| $r9@18))

; outputs: the ledger applies nothing on failure
(define-fun fails () Bool fail@20)
(define-fun success () Bool (not fails))
(define-fun |sent(@dest,USD/2)| () Int 0)
(define-fun |received(@dest,USD/2)| () Int (ite fails 0 $r9@18))
(define-fun |after(@dest,USD/2)| () Int (+ |before(@dest,USD/2)| (- |received(@dest,USD/2)| |sent(@dest,USD/2)|)))
(define-fun |sent(@src,USD/2)| () Int (ite fails 0 $r9@18))
(define-fun |received(@src,USD/2)| () Int 0)
(define-fun |after(@src,USD/2)| () Int (+ |before(@src,USD/2)| (- |received(@src,USD/2)| |sent(@src,USD/2)|)))
(define-fun |moved(@src,@dest,USD/2)| () Int (ite fails 0 $r9@18))
```

Per property, `prove.Problem` appends `; property <name>: <formula>`, `(define-fun property () Bool …)` and `(assert (not property))`; the driver adds `(check-sat)` and a `(get-value …)` of every input and output. A pair a property names and the script never touches gets a free `before`, `after` equal to it and zero totals (`Problem.AddKeys`).

Three mechanisms keep the model honest. Every counterexample is replayed on the real VM, and a model the VM does not reproduce, or on which the property is not actually false, is a `tool-bug`, never a counterexample. Differential tests evaluate the model at concrete balances with `smt.Eval`, without a solver, and compare every output with `vm.Exec`: `encode/sweep_test.go` runs the spec corpus in `internal/interpreter/testdata/script-tests`, where 166 of 234 scripts encode (the rest use vars, colors, scopes or scaling) and all 172 of their test cases agree, and `encode/funds_ir_test.go` does the same on hand-written IR for shapes the compiler never emits (a capped send smaller than the queue, a negative cap, an uncapped overdraft pull, an asset switch between pull and send, a queue left undrained, `kept` then send, `save` then pull). Finally `prove_test.go` cross-checks every `valid` verdict of the fixtures on 64 seeded random balance assignments (values in `[-200, 400]`, one in four drawn from `{0, 50, 100, 150}`) run through the VM with the formula evaluated by `prop.Eval`; counterexample models are not pinned, since they depend on the solver, but each must replay-confirm.

## Extension path

Vars come first, in three steps. Bind `load_var` to constants given on the command line, an options argument to `Encode` that turns `load_var<str>` into a string constant and `load_var<int>` into an int literal; that alone unlocks `TestGeneratedScriptsAgreeWithVM`, which today encodes 0 of 200 generated scripts because 189 use vars (the other 11 overflow a register bank). Then symbolic ints: a `load_var<int>` becomes a free `Int` input, declared and reported like a balance, so a property can quantify over the amount. Then account vars by case split: `$dest` is one of the k literal accounts or a fresh (k+1)-th, each key guarded by the alias it falls under, which keeps the touched set finite.

Metadata reads become free symbols: `meta<int>` a free `Int`, `meta<str>` a free string that only `str_eq` against constants inspects, a finite case split per comparison. `forall`, `exists` and `sum` expand over the known accounts and assets as described under [Quantifiers](#quantifiers). Colors and scopes become extra dimensions of `encode.Key` (`problem.go` notes they are `""` in v1), with the property atoms growing matching arguments. Tagged monetary literals, `[USD/2 100]` as an int whose asset must match the atom it is compared with, are a typecheck rule rather than a new value.
