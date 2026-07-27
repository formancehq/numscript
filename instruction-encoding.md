# VM Bytecode Specification (proposal)

Instructions are **4 bytes** wide: `[Opcode: 8] [A: 8] [B: 8] [C: 8]`.

- Registers are split into **per-type banks** (`int_regs`, `str_regs`, `mon_regs`, `por_regs`, …); an operand indexes the bank implied by the opcode.
- `0xFF` in a register slot means **nil** (absent optional operand).
- **`Bx`** = a `u16` formed by slots `B`,`C` (little-endian); used for pool indices and jump targets. **`sBx`** is its signed form.
- Most instructions are one word. A few extend into **continuation words** (shown as `↳ cont.`); an instruction's length is fixed by its opcode.
- There is **no `HALT`**: programs terminate by design (jumps are forward-only).
- Opcodes are grouped by category with gaps, so new instructions slot into a category without renumbering. Unused values are reserved (users can't emit them, so we stay free to define them later).

> Opcode numbers are a proposal and don't yet match the `iota` values in `instruction.go`.

---

## 1. State & Assertions

<table width="100%">
  <thead>
    <tr>
      <th>Opcode</th><th>Hex</th><th>Name</th>
      <th width="10%">A</th><th width="10%">B</th><th width="10%">C</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>0</td><td><code>0x00</code></td><td><strong>SET_CURRENT_ASSET</strong></td>
      <td>asset</td><td>-</td><td>-</td>
      <td>Sets the current asset (used by <code>PULL_ACCOUNT</code> / <code>SEND_TO_ACCOUNT</code>) from <code>str_regs[A]</code></td>
    </tr>
    <tr>
      <td>1</td><td><code>0x01</code></td><td><strong>ASSERT_SAME_ASSET</strong></td>
      <td>x</td><td>y</td><td>-</td>
      <td>Traps unless <code>str_regs[A]</code> and <code>str_regs[B]</code> are the same asset</td>
    </tr>
    <tr>
      <td>2</td><td><code>0x02</code></td><td><strong>ASSERT_VALID_ACCOUNT</strong></td>
      <td>acc</td><td>-</td><td>-</td>
      <td>Traps if the account name in <code>str_regs[A]</code> is malformed</td>
    </tr>
    <tr>
      <td>3</td><td><code>0x03</code></td><td><strong>ASSERT_NON_NEGATIVE_BALANCE</strong></td>
      <td>mon</td><td>acc</td><td>-</td>
      <td>Traps if the monetary in <code>mon_regs[A]</code> is negative; <code>B</code> = account (for the error)</td>
    </tr>
    <tr>
      <td>4</td><td><code>0x04</code></td><td><strong>ASSERT_LEFTOVER</strong></td>
      <td>por</td><td>exact</td><td>-</td>
      <td>Traps if <code>por_regs[A]</code> is negative; when <code>B == 1</code> (no <code>remaining</code>) also traps if non-zero</td>
    </tr>
    <tr>
      <td>5</td><td><code>0x05</code></td><td><strong>CHECK_ENOUGH_FUNDS</strong></td>
      <td>pulled</td><td>target</td><td>-</td>
      <td>Traps if <code>int_regs[A] &lt; int_regs[B]</code> (missing funds)</td>
    </tr>
    <tr>
      <td colspan="7" align="center"><em>0x06..0x0F reserved</em></td>
    </tr>
  </tbody>
</table>

## 2. Constants & Variables

<table width="100%">
  <thead>
    <tr>
      <th>Opcode</th><th>Hex</th><th>Name</th>
      <th width="10%">A</th><th width="10%">B</th><th width="10%">C</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>16</td><td><code>0x10</code></td><td><strong>LOAD_INT</strong></td>
      <td>dest</td><td colspan="2" align="center">Bx (const idx)</td>
      <td><code>int_regs[A] = int_pool[Bx]</code></td>
    </tr>
    <tr>
      <td>17</td><td><code>0x11</code></td><td><strong>LOAD_STR</strong></td>
      <td>dest</td><td colspan="2" align="center">Bx (const idx)</td>
      <td><code>str_regs[A] = str_pool[Bx]</code></td>
    </tr>
    <tr>
      <td>18</td><td><code>0x12</code></td><td><strong>LOAD_VAR_INT</strong></td>
      <td>dest</td><td colspan="2" align="center">Bx (var idx)</td>
      <td><code>int_regs[A] = vars.int_pool[Bx]</code></td>
    </tr>
    <tr>
      <td>19</td><td><code>0x13</code></td><td><strong>LOAD_VAR_STR</strong></td>
      <td>dest</td><td colspan="2" align="center">Bx (var idx)</td>
      <td><code>str_regs[A] = vars.str_pool[Bx]</code></td>
    </tr>
    <tr>
      <td>20</td><td><code>0x14</code></td><td><strong>LOAD_INT_IMMEDIATE</strong></td>
      <td>dest</td><td colspan="2" align="center">sBx (i16 value)</td>
      <td><code>int_regs[A] = (big.Int)sBx</code> — small literals inline, no pool entry. <strong>Reserved; not implemented</strong></td>
    </tr>
    <tr>
      <td colspan="7" align="center"><em>0x15..0x1F reserved</em></td>
    </tr>
  </tbody>
</table>

## 3. Metadata

<table width="100%">
  <thead>
    <tr>
      <th>Opcode</th><th>Hex</th><th>Name</th>
      <th width="10%">A</th><th width="10%">B</th><th width="10%">C</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>32</td><td><code>0x20</code></td><td><strong>SET_TX_META</strong></td>
      <td>key</td><td>val</td><td>-</td>
      <td>Sets transaction metadata <code>str_regs[A] = str_regs[B]</code></td>
    </tr>
    <tr>
      <td>33</td><td><code>0x21</code></td><td><strong>SET_ACCOUNT_META</strong></td>
      <td>acc</td><td>key</td><td>val</td>
      <td>Sets account metadata: account <code>A</code>, key <code>B</code>, value <code>C</code></td>
    </tr>
    <tr>
      <td>34</td><td><code>0x22</code></td><td><strong>META_STR</strong></td>
      <td>dest</td><td>acc</td><td>key</td>
      <td><code>str_regs[A] = meta(account B, key C)</code></td>
    </tr>
    <tr>
      <td>35</td><td><code>0x23</code></td><td><strong>META_INT</strong></td>
      <td>dest</td><td>acc</td><td>key</td>
      <td>as <code>META_STR</code>, typed <code>int</code></td>
    </tr>
    <tr>
      <td>36</td><td><code>0x24</code></td><td><strong>META_PORTION</strong></td>
      <td>dest</td><td>acc</td><td>key</td>
      <td>as <code>META_STR</code>, typed <code>portion</code></td>
    </tr>
    <tr>
      <td>37</td><td><code>0x25</code></td><td><strong>META_MONETARY</strong></td>
      <td>dest</td><td>acc</td><td>key</td>
      <td>as <code>META_STR</code>, typed <code>monetary</code></td>
    </tr>
    <tr>
      <td colspan="7" align="center"><em>0x26..0x2F reserved</em></td>
    </tr>
  </tbody>
</table>

## 4. Arithmetic & Constructors (binary)

<table width="100%">
  <thead>
    <tr>
      <th>Opcode</th><th>Hex</th><th>Name</th>
      <th width="10%">A</th><th width="10%">B</th><th width="10%">C</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>48</td><td><code>0x30</code></td><td><strong>ADD_INT</strong></td>
      <td>dest</td><td>left</td><td>right</td>
      <td><code>int_regs[A] = int_regs[B] + int_regs[C]</code></td>
    </tr>
    <tr>
      <td>49</td><td><code>0x31</code></td><td><strong>SUB_INT</strong></td>
      <td>dest</td><td>left</td><td>right</td>
      <td><code>int_regs[A] = int_regs[B] - int_regs[C]</code></td>
    </tr>
    <tr>
      <td>50</td><td><code>0x32</code></td><td><strong>MIN_INT</strong></td>
      <td>dest</td><td>left</td><td>right</td>
      <td><code>int_regs[A] = min(int_regs[B], int_regs[C])</code></td>
    </tr>
    <tr>
      <td>51</td><td><code>0x33</code></td><td><strong>SUB_PORTION</strong></td>
      <td>dest</td><td>left</td><td>right</td>
      <td><code>por_regs[A] = por_regs[B] - por_regs[C]</code></td>
    </tr>
    <tr>
      <td>52</td><td><code>0x34</code></td><td><strong>MK_PORTION</strong></td>
      <td>dest</td><td>num</td><td>den</td>
      <td><code>por_regs[A] = int_regs[B] / int_regs[C]</code></td>
    </tr>
    <tr>
      <td>53</td><td><code>0x35</code></td><td><strong>MK_MONETARY</strong></td>
      <td>dest</td><td>asset</td><td>amount</td>
      <td><code>mon_regs[A] = { str_regs[B], int_regs[C] }</code></td>
    </tr>
    <tr>
      <td>54</td><td><code>0x36</code></td><td><strong>ADD_STRING</strong></td>
      <td>dest</td><td>left</td><td>right</td>
      <td><code>str_regs[A] = str_regs[B] + str_regs[C]</code></td>
    </tr>
    <tr>
      <td colspan="7" align="center"><em>0x37..0x3F reserved</em></td>
    </tr>
  </tbody>
</table>

## 5. Unary & Conversions

<table width="100%">
  <thead>
    <tr>
      <th>Opcode</th><th>Hex</th><th>Name</th>
      <th width="10%">A</th><th width="10%">B</th><th width="10%">C</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>64</td><td><code>0x40</code></td><td><strong>GET_AMOUNT</strong></td>
      <td>dest</td><td>mon</td><td>-</td>
      <td><code>int_regs[A] = mon_regs[B].amount</code></td>
    </tr>
    <tr>
      <td>65</td><td><code>0x41</code></td><td><strong>GET_ASSET</strong></td>
      <td>dest</td><td>mon</td><td>-</td>
      <td><code>str_regs[A] = mon_regs[B].asset</code></td>
    </tr>
    <tr>
      <td>66</td><td><code>0x42</code></td><td><strong>INT_COPY</strong></td>
      <td>dest</td><td>src</td><td>-</td>
      <td><code>int_regs[A] = int_regs[B]</code> (fresh copy)</td>
    </tr>
    <tr>
      <td>67</td><td><code>0x43</code></td><td><strong>PORTION_COPY</strong></td>
      <td>dest</td><td>src</td><td>-</td>
      <td><code>por_regs[A] = por_regs[B]</code> (fresh copy)</td>
    </tr>
    <tr>
      <td>68</td><td><code>0x44</code></td><td><strong>NEG_INT</strong></td>
      <td>dest</td><td>src</td><td>-</td>
      <td><code>int_regs[A] = -int_regs[B]</code></td>
    </tr>
    <tr>
      <td>69</td><td><code>0x45</code></td><td><strong>INT_TO_STRING</strong></td>
      <td>dest</td><td>src</td><td>-</td>
      <td><code>str_regs[A] = str(int_regs[B])</code></td>
    </tr>
    <tr>
      <td>70</td><td><code>0x46</code></td><td><strong>PORTION_TO_STRING</strong></td>
      <td>dest</td><td>src</td><td>-</td>
      <td><code>str_regs[A] = str(por_regs[B])</code></td>
    </tr>
    <tr>
      <td>71</td><td><code>0x47</code></td><td><strong>MONETARY_TO_STRING</strong></td>
      <td>dest</td><td>src</td><td>-</td>
      <td><code>str_regs[A] = str(mon_regs[B])</code></td>
    </tr>
    <tr>
      <td colspan="7" align="center"><em>0x48..0x4F reserved</em></td>
    </tr>
  </tbody>
</table>

## 6. Funds & Postings

<table width="100%">
  <thead>
    <tr>
      <th>Opcode</th><th>Hex</th><th>Name</th>
      <th width="10%">A</th><th width="10%">B</th><th width="10%">C</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>80</td><td><code>0x50</code></td><td><strong>PULL_ACCOUNT</strong></td>
      <td>dest</td><td>acc</td><td>cap</td>
      <td>Pulls funds from account <code>B</code> capped by <code>int_regs[C]</code> (<code>0xFF</code> = uncapped); pulled amount → <code>int_regs[A]</code>. 2 words:</td>
    </tr>
    <tr>
      <td>&#8203;</td><td>&#8203;</td><td><strong>&#8627; cont.</strong></td>
      <td>overdraft</td><td>color</td><td>-</td>
      <td>Overdraft cap reg and color reg (<code>0xFF</code> = none)</td>
    </tr>
    <tr>
      <td>81</td><td><code>0x51</code></td><td><strong>SEND_TO_ACCOUNT</strong></td>
      <td>acc</td><td>cap</td><td>color</td>
      <td>Emits a posting to account <code>A</code> (<code>0xFF</code> = world), each operand optional (<code>0xFF</code> = none)</td>
    </tr>
    <tr>
      <td>82</td><td><code>0x52</code></td><td><strong>SAVE</strong></td>
      <td>acc</td><td>asset</td><td>amount</td>
      <td>Reduce balance of account <code>A</code> for asset <code>B</code> by <code>int_regs[C]</code> (<code>C = 0xFF</code> ⇒ save all), floored at 0</td>
    </tr>
    <tr>
      <td>83</td><td><code>0x53</code></td><td><strong>MK_ALLOTMENT</strong></td>
      <td>dest0</td><td>in0</td><td>size</td>
      <td>Splits the current amount across <code>size</code> portions in <code>por_regs[in0..]</code>, writing shares to <code>int_regs[dest0..]</code></td>
    </tr>
    <tr>
      <td>84</td><td><code>0x54</code></td><td><strong>BALANCE</strong></td>
      <td>dest</td><td>acc</td><td>asset</td>
      <td><code>int_regs[A] = balance(account B, asset C)</code> from the run-state</td>
    </tr>
    <tr>
      <td>85</td><td><code>0x55</code></td><td><strong>SNAPSHOT</strong></td>
      <td>dest</td><td>-</td><td>-</td>
      <td><code>int_regs[A] =</code> current source-queue mark (<code>len(sources)</code>), for <code>oneof</code> backtracking</td>
    </tr>
    <tr>
      <td>86</td><td><code>0x56</code></td><td><strong>RESTORE</strong></td>
      <td>snap</td><td>-</td><td>-</td>
      <td>Rolls the source queue back to the mark in <code>int_regs[A]</code> (repays debited balances, then truncates)</td>
    </tr>
    <tr>
      <td colspan="7" align="center"><em>0x57..0x8F reserved (e.g. PULL_ACCOUNT specializations)</em></td>
    </tr>
  </tbody>
</table>

## 7. Control Flow

<table width="100%">
  <thead>
    <tr>
      <th>Opcode</th><th>Hex</th><th>Name</th>
      <th width="10%">A</th><th width="10%">B</th><th width="10%">C</th>
      <th>Description</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>144</td><td><code>0x90</code></td><td><strong>JMP_IF_ZERO</strong></td>
      <td>cond</td><td colspan="2" align="center">Bx (target instr)</td>
      <td>Jump to instruction <code>Bx</code> if <code>int_regs[A] == 0</code>. Forward-only (guarantees termination)</td>
    </tr>
    <tr>
      <td colspan="7" align="center"><em>0x91..0xFF reserved</em></td>
    </tr>
  </tbody>
</table>
