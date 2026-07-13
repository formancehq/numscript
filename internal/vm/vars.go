package vm

import (
	"fmt"
	"math/big"
)

type Vars struct {
	StringsPool []string
	IntsPool    []big.Int
}

func DecodeVars(buf []byte) (Vars, error) {
	if len(buf) < 4 || string(buf[0:4]) != "NVAR" {
		return Vars{}, fmt.Errorf("bad vars magic")
	}

	idx := 4

	dataSegment, err := readArr("data segment", buf, &idx, id)
	if err != nil {
		return Vars{}, err
	}

	stringsPool, err := readArr("strings pool", buf, &idx, func(buf []byte) ([]string, error) {
		return parseStringsPool(buf, dataSegment)
	})
	if err != nil {
		return Vars{}, err
	}

	intsPool, err := readArr("ints", buf, &idx, func(buf []byte) ([]big.Int, error) {
		return parseIntsPool(buf, dataSegment)
	})
	if err != nil {
		return Vars{}, err
	}

	return Vars{
		StringsPool: stringsPool,
		IntsPool:    intsPool,
	}, nil
}
