package compiler

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/formancehq/numscript/internal/interpreter"
	"github.com/formancehq/numscript/internal/typecheck"
	"github.com/formancehq/numscript/internal/vm"
)

type VarsEncoder struct {
	decls []varDecl
	nStr  int
	nInt  int
}

type varDecl struct {
	name string
	typ  typecheck.Type
}

// TODO review AI blob
func (e VarsEncoder) Encode(vars map[string]string) (vm.Vars, error) {
	strs := make([]string, 0, e.nStr)
	ints := make([]big.Int, 0, e.nInt)

	for _, d := range e.decls {
		raw, ok := vars[d.name]
		if !ok {
			return vm.Vars{}, fmt.Errorf("missing variable: $%s", d.name)
		}

		var err error
		strs, ints, err = appendVar(strs, ints, d.typ, raw)
		if err != nil {
			return vm.Vars{}, fmt.Errorf("variable $%s: %w", d.name, err)
		}
	}

	return vm.Vars{StringsPool: strs, IntsPool: ints}, nil
}

// TODO review AI blob
func appendVar(strs []string, ints []big.Int, typ typecheck.Type, raw string) ([]string, []big.Int, error) {
	switch typ {
	case typecheck.TypeNumber:
		n, ok := new(big.Int).SetString(raw, 10)
		if !ok {
			return strs, ints, fmt.Errorf("invalid number: %q", raw)
		}
		ints = append(ints, *n)

	case typecheck.TypeString:
		strs = append(strs, raw)

	case typecheck.TypeAccount:
		acc, err := interpreter.NewAccountAddress(raw)
		if err != nil {
			return strs, ints, err
		}
		strs = append(strs, string(acc))

	case typecheck.TypeAsset:
		asset, err := interpreter.NewAsset(raw)
		if err != nil {
			return strs, ints, err
		}
		strs = append(strs, string(asset))

	case typecheck.TypePortion:
		r, err := interpreter.ParsePortionSpecific(raw)
		if err != nil {
			return strs, ints, err
		}
		ints = append(ints, *r.Num(), *r.Denom())

	case typecheck.TypeMonetary:
		asset, amount, err := parseMonetaryVar(raw)
		if err != nil {
			return strs, ints, err
		}
		strs = append(strs, asset)
		ints = append(ints, amount)

	default:
		panic("unexpected var type: " + typ)
	}

	return strs, ints, nil
}

// TODO review AI blob
func parseMonetaryVar(raw string) (string, big.Int, error) {
	parts := strings.Split(raw, " ")
	if len(parts) != 2 {
		return "", big.Int{}, fmt.Errorf("invalid monetary: %q", raw)
	}

	asset, err := interpreter.NewAsset(parts[0])
	if err != nil {
		return "", big.Int{}, err
	}

	n, ok := new(big.Int).SetString(parts[1], 10)
	if !ok {
		return "", big.Int{}, fmt.Errorf("invalid monetary amount: %q", parts[1])
	}

	return string(asset), *n, nil
}
