# The IR textual format

The compiler doesn't emit `vm.Instruction` directly: it emits a `[]ir.Instr` stream first (see [compiler-architecture.md](compiler-architecture.md)). This document specifies the **textual notation** for that stream — the thing you get when you dump a compiled program, and the thing you can write by hand to feed the assembler.

It is a real format, not just a pretty-printing convention: it has a grammar, a parser, and a round-trip guarantee.

The whole IR layer lives in [internal/ir/](internal/ir/), and that package is the entire API:

| what | where |
| --- | --- |
| grammar | [IR.g4](IR.g4) (ANTLR; generated into `internal/ir/internal/syntax/antlrParser/` by `just generate`) |
| text → `[]ir.Instr` | `ir.Parse` in [internal/ir/parse.go](internal/ir/parse.go) |
| `[]ir.Instr` → text | `ir.Dump` in [internal/ir/dump.go](internal/ir/dump.go) |
| `[]ir.Instr` → `vm.Program` | `ir.Assemble` in [internal/ir/assemble.go](internal/ir/assemble.go) |
| register typing | `ir.Typecheck` in [internal/ir/typecheck.go](internal/ir/typecheck.go) |
| round-trip tests | `TestRoundtripAllInstructions` in [internal/ir/parse_test.go](internal/ir/parse_test.go) |

`ir.Parse` is the only way in: the grammar's AST lives under `internal/ir/internal/syntax`, which Go's import rules make unreachable from anywhere outside `internal/ir`. Callers see instructions, never parse trees.

**Round-trip property:** for every instruction, `ir.Dump` of what `ir.Parse` returns is the text it was given. This is what makes the format usable for snapshot tests and for hand-writing IR fixtures. See [Round-trip caveats](#round-trip-caveats) for the (few) inputs that don't survive it.

## Lexical structure

```
REG          '$' [a-zA-Z_] [a-zA-Z0-9_]*      $r0, $r12, $pulled
LABEL        '#' [a-zA-Z_] [a-zA-Z0-9_]*      #inorder_end_0
INT          [0-9]+                           42                (no sign, no separators)
STRING       '"' ('\"' | ~["\r\n])* '"'       "USD/2", "a\"b"
IDENTIFIER   [a-z] [a-z0-9_]*                 mk_monetary, account
TYPE_KEYWORD 'int' | 'str' | 'portion' | 'monetary'
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

Indentation is cosmetic, but `ir.Dump` always emits labels at column 0 and instructions indented by two spaces.

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

`_` discards the result, and exists **only in the text**: there is no discard at the `ir.Instr` level. `ir.Parse` desugars each occurrence to a fresh register — allocated from the same counter as named ones, but bound to no name, so nothing can refer to it and each `_` gets its own (two discards that aliased would be forced to share a type). The write is still a write: the assembler gives that register a slot in its bank, so a discard costs a register even though nothing reads it.

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
```

Labeled arguments are looked up **by name**, so their order is free: `pull_account(cap: $c, account: $a)` is the same instruction as `pull_account(account: $a, cap: $c)`. `ir.Dump` always emits them in the canonical order given below.

### Registers

Registers in the IR are "logical": an unbounded stream of unsigned indices (`ir.Reg` is a `uint`), later mapped onto the VM's 256-per-bank physical registers by the assembler's allocator. Each register has exactly one type for its whole lifetime (`int`, `str`, `portion`, `monetary`), checked by `ir.Typecheck` — the type is never written in the text, it is inferred from the instruction that writes the register.

A register name is just a name: `ir.Parse` keeps a symbol table and allocates registers in order of **first appearance**, reusing the same one every later time a name shows up. `$r<N>` is a convention, not an index — `$r7` is no more meaningful than `$src`.

```
$src = "acc"                        dumps back as    $r0 = "acc"
$pulled = pull_account(account: $src)                $r1 = pull_account(account: $r0)
```

This is why a dump round-trips: `ir.Dump` numbers registers `$r0`, `$r1`, … in the order they first appear, so re-parsing binds each name to the register it already had. Names of your own choosing are fine to write, they just come back as `$r<k>` in first-appearance order.

The compiler holds up its end by allocating registers in the order it emits them — `getCompiledOutput` in the compiler tests asserts the round-trip on every snapshot, so a change that breaks the ordering fails there.

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

`add_int` and `sub_int` have infix sugar, which is what `ir.Dump` always prints:

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
| `set_tx_meta<str>($key, $value)` | `str, str` |
| `set_tx_meta<int>($key, $value)` | `str, int` |
| `set_account_meta<monetary>($account, $key, $value)` | `str, str, monetary` |

Both take a required type parameter — one of `str`, `account`, `asset`, `int`,
`portion`, `monetary` — and it does two jobs at once:

* it selects the register bank `$value` is read from, exactly as it does on the
  read side (`meta<int>` writes an int reg, `set_tx_meta<int>` reads one). `str`,
  `account` and `asset` all read the string bank; they differ only in the type
  reported in the result. `Typecheck` enforces the agreement, so
  `set_tx_meta<int>($k, $v)` with a string `$v` is rejected.
* it is reported in the execution result, so a host can rebuild the typed value.
  Metadata is stored as text and the VM stringifies the value when it stores it —
  without the type, the string `"42"` and the number `42` would be
  indistinguishable in the output.

There is deliberately no conversion instruction on this path: stating the type on
the instruction and letting the VM stringify keeps it in one place, instead of
implying it with an `int_to_string` and then restating it.

### Control flow

```
#my_label
  jmp_if_zero($cond, #my_label)
```

`$cond` is `int`; the target must be a label that is defined in the program, unique, and **after** the jump. The VM only permits forward jumps — that's what guarantees termination — and `ir.Parse` enforces all three rules, so a program that assembles can't loop:

```
jmp_if_zero($r0, #nope)     → label #nope is not defined in the program
#back                       → label #back is behind the jump (jumps must go forward)
  jmp_if_zero($r0, #back)
```

`labelMarker` is a pseudo-instruction: it emits no bytecode, it only feeds the assembler's symbol table.

### Backtracking (`oneof`)

```
  $mark = snapshot()      // int: marks the current position of the source queue
  restore($mark)          // rewinds the source queue to a mark
```

`snapshot` takes no arguments and writes an `int` mark; `restore` reads one back. A `oneof` source compiles to a `snapshot` before the first branch and a `restore` before each retry.

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

Known asymmetries between what `ir.Dump` writes and what the parser accepts:

* **Register names don't survive.** `$src` comes back as `$r<k>`, numbered by first appearance (see [Registers](#registers)).
* **`_` doesn't survive.** It's desugared to a fresh register on the way in, so it dumps as that register (see [Destinations](#destinations)).
* **Negative int literals are not expressible.** `INT` has no sign, so `$r0 = -1` is a syntax error, while `ir.Dump` would happily print it for a negative `ir.LoadInt`. This is not reachable today — the compiler emits `neg_int` for negative literals rather than a negative constant — but a constant-folding peephole could produce a dump that no longer parses.

## Error handling

Text → `[]ir.Instr` never panics: it reports `ir.Error`s. Anything the grammar rejects comes back as a syntax error (and since ANTLR's error recovery leaves partial nodes behind, no AST is built at all in that case). On top of that, `ir.Parse` reports what the grammar can't express:

* unknown instruction names, and a type parameter on an instruction that doesn't take one
* wrong argument kinds or counts, unknown or duplicate labeled arguments
* duplicate labels, and jumps that don't resolve or don't go forward
* **a register that is read but never written**, reported under the name the text used:

```
  $a = 42
  $y = min_int($a, $b)      →  3:3: register $b is read but never written
```

Since jumps only go forward, text order is execution order, so a read with no earlier write can't be reached by any path — it would hand the VM whatever that register happens to hold. Note this is a linear check: a register written only inside a branch that may be skipped and read afterwards is *not* caught here, which is the job of the path-sensitive bytecode verifier.

Type errors are **not** checked by `ir.Parse`: writing a `str` register where an `int` is expected parses happily and is caught by `ir.Typecheck` afterwards.
