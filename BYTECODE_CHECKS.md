# Bytecode checks

Two independent static passes validate the compiled program before it runs. Once
both pass, `vm.Exec` can index registers, pools and jump targets directly, with
no defensive guards in the execution loop.

## Pipeline

```
parser.Program
  └─ compileProgramToVirtual   → []vInstr        (virtual instructions)
       └─ typecheckInstructions → (1) TYPE CHECK  internal/compiler/bytecode_typecheck.go
            └─ assembleProgram   → vm.Program     (byte-encoded instructions)
                 └─ vm.Verify    → (2) VERIFIER    internal/vm/verify.go
```

- Both passes run inside `Compile` (`internal/compiler/compiler.go`), so the
  compiler *never* emits bytecode the VM would reject.
- The verifier runs **again** at runtime, lazily, on the first `vm.Exec`
  (gated by the `verified` flag in `internal/vm/vm.go`). This protects the VM
  from any bytecode not produced by our compiler (hand-written, decoded from an
  untrusted source, fuzzed, etc.).

The two passes are complementary: the typechecker works on the high-level
virtual instructions and enforces *types*; the verifier works on the final
encoded bytes and enforces *memory safety*.

---

## (1) Bytecode typechecker — `internal/compiler/bytecode_typecheck.go`

Runs on the virtual instruction stream. Every virtual register has exactly one
type for its whole life, drawn from the four VM banks: `int`, `string`,
`portion`, `monetary`.

Checks, per instruction, in program order:

- **Read-before-write** — a register read before it was ever written is rejected.
- **Type consistency on read** — a register must be read with the same type it
  was written with (`use`).
- **Type consistency on write** — a register cannot be rewritten with a
  different type than it already holds (`def`).
- **Operand types** — each instruction's operands must match the fixed type
  signature of its opcode (e.g. `AddInt` takes two `int`s and writes an `int`;
  `MakeMonetary` takes a `string` + `int` and writes a `monetary`).

Example rejections:

```
register r3 read as int before being written
register r5 read as string but holds int
register r2 written as portion but already holds int
```

This is a check on *our compiler*: a failure here means the code lowering AST →
virtual instructions produced something inconsistent. It is not reachable from
user input.

---

## (2) Bytecode verifier — `internal/vm/verify.go`

Runs on the final byte-encoded `vm.Program`. A `nil` result guarantees the
execution loop cannot read out of bounds or panic. It also computes the exact
register-bank sizes the program needs.

Checks:

- **Instruction framing** — multi-word instructions (`PullAccount`,
  `MkAllotment` take 2 words) must not be truncated at the end of the stream.
- **Known opcodes** — every opcode must decode; unknown opcodes are rejected
  (`decodeInstr`).
- **Register bounds** — banks are sized to `maxIndex + 1` per bank from the
  operands actually used, so every register access is in range by construction.
- **Constant-pool bounds** — every int/string constant index is `< len(pool)`.
- **Variable-pool bounds** — records the highest var index used; `Exec` later
  checks the caller-supplied `Vars` provides at least that many.
- **Jump validity** — every jump target must be:
  - **forward** (`target > current`), and
  - an **instruction boundary** (not the middle of a 2-word instruction, and not
    past the end — jumping to `len(instrs)` halts, which is allowed).
- **Definite assignment** — a register (and the `currentAsset` pseudo-slot) may
  only be read if it was written on *every* path reaching that instruction.
  Because all jumps are forward, predecessors are always earlier, so a single
  ordered pass intersecting predecessors' written-sets is sufficient.

Example rejections:

```
truncated instruction at 7: opcode 17 needs 2 words
at instruction 3: int constant 5 out of range (pool size 4)
at instruction 9: backward jump to 2
at instruction 4: jump to 6 is not an instruction boundary
at instruction 8: register (bank 0, index 2) read before being assigned on all paths
```

> Note: the forward-jump restriction is what keeps definite assignment a single
> linear pass. Numscript has no backward control flow, so this costs nothing.

---

## What is NOT checked here

These passes guarantee **structural and memory safety only**. They do *not*
validate numscript runtime semantics — those are genuine runtime errors that
depend on values only known during execution, and the VM checks them inline:

| Runtime check | Opcode | Error |
|---|---|---|
| Insufficient funds | `CheckEnoughFunds` | `MissingFundsError` |
| Allotment portions don't sum to 1 | `AssertLeftover` | `InvalidAllotmentSum` |
| Mismatched assets | `AssertSameAsset` | `AssetMismatchError` |
| Malformed account name | `AssertValidAccount` | `InvalidAccountName` |
| Negative balance | `AssertNonNegativeBalance` | `NegativeBalanceError` |
| Division by zero in a portion | `MkPortion` | `DivideByZeroError` |
| Unbounded pull with no cap/overdraft | `PullAccount` | `InvalidUncappedSource` |
| Non-numeric / malformed metadata | `Meta*` | `BadMetaValueError` |

A verified program is safe to *execute*; it can still fail with any of the above.
