package vm

import "math/big"

type Program struct {
	instructions []Instruction

	stringsPool []string
	intsPool    []big.Int
}
