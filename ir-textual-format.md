# The IR textual format

The compiler doesn't emit `vm.Instruction` directly: it emits a `[]compiler.irInstr` stream first (see [compiler-architecture.md](compiler-architecture.md)). This document specifies the **textual notation** for that stream — the thing you get when you dump a compiled program, and the thing you can write by hand to feed the assembler.

It is a real format, not just a pretty-printing convention: it has a grammar, a parser, and a round-trip guarantee.

| what | where |
| --- | --- |
| grammar | [IR.g4](IR.g4) (ANTLR; generated code in [internal/irparser/antlrParser/](internal/irparser/antlrParser/)) |
| text → AST | `irparser.Parse` in [internal/irparser/parser.go](internal/irparser/parser.go) |
| AST → `[]irInstr` | `compiler.Transform` in [internal/compiler/ir_parser.go](internal/compiler/ir_parser.go) |
| `[]irInstr` → text | `dump` in [internal/compiler/ir_instr_dump.go](internal/compiler/ir_instr_dump.go) |
| round-trip tests | `TestRoundtripAllInstructions` in [internal/compiler/ir_parser_test.go](internal/compiler/ir_parser_test.go) |

**Round-trip property:** for every instruction the parser supports, `dump(Transform(Parse(text))) == text`. This is what makes the format usable for snapshot tests and for hand-writing IR fixtures. See [`snapshot` / `restore`](#not-yet-parseable-snapshot--restore) for the two instructions not yet covered, and [Round-trip caveats](#round-trip-caveats) for the (few) inputs that don't survive it.

## Lexical structure

```
REG          '$' [a-zA-Z_] [a-zA-Z0-9_]*      $r0, $r12, $pulled
LABEL        '#' [a-zA-Z_] [a-zA-Z0-9_]*      #inorder_end_0
INT          [0-9]+                           42                (no sign, no separators)
STRING       '"' ('\"' | ~["\r\n])* '"'       "USD/2", "a\"b"
IDENTIFIER   [a-z] [a-z0-9_]*                 mk_monetary, account
TYPE_KEYWORD 'int' | 'str' | 'portion' | 'monetary'
BOOL         'true' | 'false'
```

* Spaces, tabs and newlines are **skipped**, not significant. Statements are delimited by the grammar, not by line breaks — `$r0 = 1 $r1 = 2` is two valid instructions. Newlines are pure convention (a very useful one).
* **There are no comments.** Any `//` or `#`-style comment is a syntax error (`#foo` lexes as a label).
* `TYPE_KEYWORD` is matched before `IDENTIFIER`, so `int`, `str`, `portion` and `monetary` are reserved: they cannot be used as an instruction name or an argument label. Register names are unaffected (`$int` is fine, the `$` starts a `REG`).
* Instruction names and argument labels are lowercase-only (`IDENTIFIER` starts with `[a-z]`).

## Statements

A program is a flat sequence of two kinds of line: **label markers** and **instructions**.

```
#some_label                       ← label marker, flush left
  set_current_asset($r3)          ← instruction, indented 2 spaces
```

Indentation is cosmetic, but `dump` always emits labels at column 0 and instructions indented by two spaces.

Instructions come in five shapes:

```
  dest = name(args)         instruction call with a destination
  name(args)                instruction call with no destination
  dest = <literal>          constant load
  dest = $l + $r            infix int arithmetic       (only + and -)
  $d += $r                  compound assign            (only += and -=)
```

### Destinations

```
$r0                single register
[$r0, $r1, $r2]    register list  (only mk_allot uses this)
_                  discard
```

`_` discards the result, and exists **only in the text**: there is no discard at the `irInstr` level. `Transform` desugars each occurrence to a fresh register — one above every register the program refers to, so no statement can name it and each `_` gets its own (two discards that aliased would be forced to share a type). The write is still a write: the assembler gives that register a slot in its bank, so a discard costs a register even though nothing reads it.

Because the desugaring happens on the way in, `_` doesn't survive a dump: `_ = int_copy($r0)` comes back as `$r1 = int_copy($r0)`.

### Arguments

Arguments are comma-separated and either **positional** or **labeled**:

```
  check_enough_funds($r7, $r4)                            positional
  $r8 = pull_account(account: $r5, cap: $r4)              labeled
```

Which form an argument takes is fixed per instruction (see the reference below) — it is not a free choice. An argument value is one of:

```
$r0            register
#my_label      label reference   (jmp_if_zero only)
42             int literal       (load_var index only)
[$r0, $r1]     register list     (mk_allot portions only)
true / false   boolean           — parseable, but no instruction currently takes one
```

Labeled arguments are looked up **by name**, so their order is free: `pull_account(cap: $c, account: $a)` is the same instruction as `pull_account(account: $a, cap: $c)`. `dump` always emits them in the canonical order given below.

### Registers

Registers in the IR are "logical": an unbounded stream of unsigned indices (`reg` is a `uint`), later mapped onto the VM's 256-per-bank physical registers by the assembler's allocator. Each register has exactly one type for its whole lifetime (`int`, `str`, `portion`, `monetary`), checked by `typecheckInstructions` — the type is never written in the text, it is inferred from the instruction that writes the register.

`$r<N>` is the canonical spelling and maps to logical register `N`. Any other name (`$pulled`, `$src`) parses, but is **hashed** to an index — convenient for hand-written examples, but names are lost on dump and distinct names can collide:

```
$src = "acc"                        dumps back as    $r1186624 = "acc"
$pulled = pull_account(account: $src)                $r2265800 = pull_account(account: $r1186624)
```

Either way the index a name resolves to stays under `maxRegIndex` (`1 << 24`): `$r<N>` above that bound is hashed like any other name, and the hash is reduced into range. That's what guarantees room above the named registers for the fresh ones `_` desugars to.

Use `$r<N>` for anything that must round-trip.

## Instruction reference

Types are the register types of each operand; `?` marks an optional labeled argument.

### Constants and variables

| syntax | types |
| --- | --- |
| `$d = 42` | `int` |
| `$d = "USD/2"` | `str` |
| `$d = load_var<int>(0)` | `int`; the index is a literal in `0..65535` |
| `$d = load_var<str>(1)` | `str` |

`load_var` reads from the encoded `vm.Vars` pool at that index. There is no `load_var<portion>` / `load_var<monetary>`: composite vars are encoded as their int/str components.

### Pure arithmetic and constructors

| syntax | signature |
| --- | --- |
| `$d = min_int($l, $r)` | `(int, int) -> int` |
| `$d = add_int($l, $r)` | `(int, int) -> int` |
| `$d = sub_int($l, $r)` | `(int, int) -> int` |
| `$d = add_string($l, $r)` | `(str, str) -> str` |
| `$d = sub_portion($l, $r)` | `(portion, portion) -> portion` |
| `$d = mk_portion($num, $den)` | `(int, int) -> portion` |
| `$d = mk_monetary($asset, $amt)` | `(str, int) -> monetary` |

`add_int` and `sub_int` have infix sugar, which is what `dump` always prints:

```
  $r2 = $r0 + $r1        add_int
  $r2 = $r0 - $r1        sub_int
  $r0 += $r1             add_int where dest == left
  $r0 -= $r1             sub_int where dest == left
```

So `add_int($a, $b)` parses fine, but a dump never contains it. No other operator has infix syntax.

### Unary ops

| syntax | signature |
| --- | --- |
| `$d = int_copy($a)` | `int -> int` |
| `$d = portion_copy($a)` | `portion -> portion` |
| `$d = neg_int($a)` | `int -> int` |
| `$d = get_asset($m)` | `monetary -> str` |
| `$d = get_amount($m)` | `monetary -> int` |
| `$d = int_to_string($a)` | `int -> str` |
| `$d = portion_to_string($a)` | `portion -> str` |
| `$d = monetary_to_string($a)` | `monetary -> str` |

There is no register-to-register move: use `int_copy` / `portion_copy`. `$r0 = $r1` is not valid syntax.

### Run-state reads (impure)

| syntax | signature |
| --- | --- |
| `$d = balance($account, $asset)` | `(str, str) -> monetary` |
| `$d = meta<str>($account, $key)` | `(str, str) -> str` |
| `$d = meta<int>($account, $key)` | `(str, str) -> int` |
| `$d = meta<portion>($account, $key)` | `(str, str) -> portion` |
| `$d = meta<monetary>($account, $key)` | `(str, str) -> monetary` |

### Funds movement

```
  $pulled = pull_account(account: $a, cap: $c, overdraft: $o, color: $col)
```
`account: str` is required; `cap: int`, `overdraft: int`, `color: str` are optional. Writes the amount actually pulled (`int`) into the destination. No `cap` means uncapped. Canonical dump order: `account, cap, overdraft, color`.

```
  send_to_account(account: $a, cap: $c)
```
No destination. Both arguments are optional: no `cap` sends everything currently queued; **no `account` refunds the funds to their sources without emitting postings**.

```
  save(account: $a, asset: $as, amount: $amt)
```
No destination. `account: str` and `asset: str` are required, `amount: int` is optional — omitting it saves the whole balance.

```
  [$s1, $s2, $s3] = mk_allot($amount, [$p1, $p2, $p3])
```
Splits `$amount` (`int`) across `n` portions (`portion`), writing `n` shares (`int`). The destination **must** be a register list, and its length must match the portion list. Both lengths are statically known at compile time; the assembler materializes each list into contiguous registers.

### Assertions and checks

| syntax | operands |
| --- | --- |
| `check_enough_funds($got, $needed)` | `int, int` |
| `assert_leftover($portion)` | `portion` — leftover must be `>= 0` |
| `assert_leftover_exact($portion)` | `portion` — leftover must be exactly `0` |
| `assert_same_asset($l, $r)` | `str, str` |
| `assert_valid_account($a)` | `str` |
| `assert_non_negative_balance($bal, $account)` | `monetary, str` |
| `set_current_asset($asset)` | `str` — required before `pull_account` / `send_to_account` |

`assert_leftover` / `assert_leftover_exact` are two separate instruction names, not one instruction with an `exact:` flag.

### Metadata writes

| syntax | operands |
| --- | --- |
| `set_tx_meta($key, $value)` | `str, str` |
| `set_account_meta($account, $key, $value)` | `str, str, str` |

### Control flow

```
#my_label
  jmp_if_zero($cond, #my_label)
```

`$cond` is `int`; the target must be a label defined somewhere in the program (checked by `Transform`) and unique (duplicates are an error). The VM only permits **forward** jumps — that's what guarantees termination — but note that `Transform` itself does not enforce direction; a backward jump is caught later, so don't rely on the IR layer to reject it.

`labelMarker` is a pseudo-instruction: it emits no bytecode, it only feeds the assembler's symbol table.

### Not yet parseable: `snapshot` / `restore`

The `oneof` support added two instructions that `dump` prints but `Transform` does **not** accept yet:

```
  $mark = snapshot()      // int: marks the current source-queue position
  restore($mark)          // rewinds the source queue to a mark
```

They are syntactically valid (`snapshot()` is just a call with no arguments), but `Transform` rejects both names with `unknown instruction`. So a dump of a script using `oneof` does not currently round-trip. Adding them is two entries in the `transformCall` switch plus a parser function each.

## A full example

`send [USD/2 10] (source = @src destination = @dest)` compiles to:

```
  $r0 = "USD/2"
  $r1 = 10
  $r2 = mk_monetary($r0, $r1)
  $r3 = get_asset($r2)
  set_current_asset($r3)
  $r4 = get_amount($r2)
  $r5 = "src"
  $r6 = 0
  $r7 = pull_account(account: $r5, cap: $r4, overdraft: $r6)
  check_enough_funds($r7, $r4)
  $r8 = "dest"
  send_to_account(account: $r8)
```

## Round-trip caveats

Known asymmetries between what `dump` writes and what the parser accepts:

* **Named registers don't survive.** `$src` comes back as `$r<hash>` (see [Registers](#registers)).
* **`_` doesn't survive.** It's desugared to a fresh register on the way in, so it dumps as that register (see [Destinations](#destinations)).
* **Negative int literals are not expressible.** `INT` has no sign, so `$r0 = -1` is a syntax error, while `dump` would happily print it for a negative `loadInt`. This is not reachable today — the compiler emits `neg_int` for negative literals rather than a negative constant — but a constant-folding peephole could produce a dump that no longer parses.
* **Duplicate labeled arguments are silently ignored.** In `pull_account(account: $a, cap: $c1, cap: $c2)` the first `cap` wins and the second is dropped without an error.
* **`Parse` panics on some malformed input** instead of returning `ParserError`s: when ANTLR's error recovery yields a partial `instrCall` node, the AST builder dereferences a nil token (`buildInstrCall`, [internal/irparser/parser.go:257](internal/irparser/parser.go#L257)). A comment line (`// x`) and a bare `$r0 = $r1` both reproduce it. Well-formed input is unaffected; hand-written IR is worth double-checking.
