# Bytecode Verifier

`vm.Verify` is a static pass over a `vm.Program`. A nil result means the
execution loop cannot read out of bounds or crash on that program.

It is **opt-in**. `Exec` does not call it, and neither does `compiler.Compile`:
the VM assumes its own compiler's output is well formed, and paying for a full
static pass on every compile would make that assumption cost something. The
assumption is earned by tests instead — see [What keeps it
honest](#what-keeps-it-honest).

Run it on any program that did not come out of `ir.Assemble` in this process:
anything read from a file, a wire, or a cache.

```go
program, err := vm.DecodeProgram(bytes)
if err != nil { ... }
if err := vm.VerifyWithVars(program, vars); err != nil { ... }
result, execErr := vm.Exec(ctx, vm.NewVm(program), vars, store)
```

`Verify(p)` checks the program alone. `VerifyWithVars(p, vars)` also checks that
`vars` carries every variable the program loads — a program that reads a
variable is only safe against the vars it will actually be given, so a caller
that passes vars should use the second form.

## What it checks

| Check | What it prevents |
|---|---|
| Instruction stream decodes end to end | A truncated multi-word instruction, whose ext word `Exec` reads past the end of the stream |
| Every opcode is known | `Exec` falling through to its `default` arm |
| Const-pool indices are in range | `Op_LoadInt`/`Op_LoadStr` indexing past the pool |
| Jumps land on an instruction boundary | Landing on an ext word, whose ignored opcode byte would then be decoded as an instruction |
| Register indices are below the bank's declared `MaxReg` | Indexing past a register bank — the banks are sized from those counts, so this is what makes the sizing safe |
| `nilReg` only in optional operands | Reading register 255 where `Exec` dereferences unconditionally |
| Flag operands are 0 or 1 | `Op_MarkEnd` tests `A == 1`, so a 2 silently commits a region that meant to rewind |
| Definite assignment | Reading a register not written on every path reaching the instruction |
| The current asset is set before it is used | A `send` or `pull` before any `set_current_asset` |
| (`VerifyWithVars`) Var-pool indices are in range | `Op_LoadVar*` indexing past the pool, or dereferencing a nil `*Vars` |

Two properties come for free rather than as their own pass:

- **Type confusion.** A register's bank is part of its identity, so a slot
  written as an int and later read as a string is a read of a register that was
  never written. `ir.Typecheck` covers the same ground one level up, on the IR,
  where a register still has a name.
- **Termination.** Jump deltas are unsigned and relative to the following
  instruction, so a backward jump cannot be encoded. This is also what makes the
  definite-assignment dataflow a single ordered pass rather than a worklist:
  every predecessor of a step is earlier in the stream.

## What it does not check

### Mark discipline

`Op_MarkPush`/`Op_MarkEnd` take no operand, so mark depth is a function of
position in the instruction stream and a verifier could decide statically that:

- pushes and ends balance on every path,
- no `Op_MarkEnd` runs at depth 0,
- no `Op_SendToAccount`, `Op_SetCurrentAsset` or `Op_Save` sits at depth > 0.

None of that is implemented. The VM enforces all three at execution time, via
`runstate.HasOpenMark()` and the `errSendWhileMarkOpen` /
`errSetAssetWhileMarkOpen` / `errSaveWhileMarkOpen` sentinels in `vm.go`.
Landing the static pass would let those five lines go.

The shape is the same forward dataflow the definite-assignment pass already
uses: carry a depth instead of an assigned-set, and require predecessors to
agree at a join. `ir/instr.go` already asks emitters to keep this decidable
("never emit a mark op on only one side of a branch").

**`internal/funds` must not be relaxed along with it.** The tree-walking
interpreter shares `RunState.MarkEnd` (see `interpreter.go`), and it is not
verified, so `ErrNoOpenMark` and the INVARIANT documented on `MarkEnd` stay.

### Unbounded sources

`Op_PullAccount` with both cap and overdraft nil is statically decidable, but
it is a legitimate user-facing error (`InvalidUncappedSource`, "unbounded source
is not allowed here"), not a malformed program. Deliberately left to run time.

### Anything semantic

A verified program can still produce nonsense: allotment portions that don't sum
to 1, assets that don't line up across a pull and a send, postings that make no
business sense. The verifier is about the VM's own memory safety, and the fuzz
tests tolerate garbage output on purpose (`_, _ = vm.Exec(...)`).

### Totality of `Exec`

`Exec` is total only for callers that verified first. Making it unconditionally
total would mean verifying inside `Exec`, and paying for it on a path where the
bytecode is nearly always the compiler's own.

### Cost

Definite assignment is O(steps × registers), with a map allocated per step.
Fine at current program sizes; if programs grow, the assigned-sets want to be
bitsets over a dense register numbering.

## What keeps it honest

The verifier is a second, independent model of what each opcode reads and
writes — `decodeInstr` mirrors, operand for operand, the matching arm of `Exec`.
Two models drift. Four things push back:

| Test | Property |
|---|---|
| `compiler.TestCompiledCorpusPassesVerify` | Every script in the corpus compiles to bytecode the verifier accepts |
| `vm.assembleIR` (all of `ir_test.go`) | Every IR-driven test program verifies, including sequences the compiler never emits |
| `vm.FuzzExec` | Arbitrary bytes: whatever verifies, `Exec` runs without panicking |
| `compiler.FuzzMutatedBytecode` | Corrupted real bytecode: same property, but close enough to valid to reach the deep checks |

An opcode added to `instruction.go` but not to `decodeInstr` is rejected as
unknown, so the corpus and IR tests fail loudly rather than silently skipping
it. That is the intended failure mode.
