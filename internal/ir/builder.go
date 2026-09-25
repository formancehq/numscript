package ir

import "fmt"

// Builder accumulates an instruction stream, handing out the registers and
// labels it needs.
type Builder struct {
	instrs      []Instr
	nextReg     Reg
	nextLabelID int
}

func (b *Builder) FreshReg() Reg {
	r := b.nextReg
	b.nextReg++
	return r
}

// FreshLabel suffixes prefix with a counter: "inorder_end" -> #inorder_end_0.
func (b *Builder) FreshLabel(prefix string) Label {
	l := Label(fmt.Sprintf("%s_%d", prefix, b.nextLabelID))
	b.nextLabelID++
	return l
}

func (b *Builder) Push(instr Instr) {
	b.instrs = append(b.instrs, instr)
}

// PushWithDest allocates the register the instruction writes to. Allocating here
// rather than up front keeps registers numbered in emission order, which is what
// makes a Dump of the result parse back to the same program.
func (b *Builder) PushWithDest(getInstr func(dest Reg) Instr) Reg {
	dest := b.FreshReg()
	b.Push(getInstr(dest))
	return dest
}

func (b *Builder) Instrs() []Instr {
	return b.instrs
}
