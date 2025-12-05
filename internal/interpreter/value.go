package interpreter

import (
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
)

type Value interface {
	value()
	String() string
}

type String string
type Asset string
type Portion big.Rat
type AccountAddress string
type MonetaryInt big.Int
type Monetary struct {
	Amount MonetaryInt
	Asset  Asset
}

func (String) value()         {}
func (AccountAddress) value() {}
func (MonetaryInt) value()    {}
func (Monetary) value()       {}
func (Portion) value()        {}
func (Asset) value()          {}

func NewAccountAddress(src string) (AccountAddress, InterpreterError) {
	if !checkAccountName(src) {
		return AccountAddress(""), InvalidAccountName{Name: src}
	}
	return AccountAddress(src), nil
}

func NewAsset(src string) (Asset, InterpreterError) {
	if !checkAssetName(src) {
		return Asset(""), InvalidAsset{Name: src}
	}
	return Asset(src), nil
}

func (v MonetaryInt) MarshalJSON() ([]byte, error) {
	bigInt := big.Int(v)
	s := fmt.Sprintf(`"%s"`, bigInt.String())
	return []byte(s), nil
}

func (v Portion) MarshalJSON() ([]byte, error) {
	r := big.Rat(v)
	s := fmt.Sprintf(`"%s"`, r.String())
	return []byte(s), nil
}

func (v Monetary) MarshalJSON() ([]byte, error) {
	m := fmt.Sprintf("\"%s %s\"", v.Asset, v.Amount.String())
	return []byte(m), nil
}

func (v String) String() string {
	return string(v)
}

func (v AccountAddress) String() string {
	return string(v)
}

func (v MonetaryInt) String() string {
	i := big.Int(v)
	return i.String()
}

func (v Monetary) String() string {
	return fmt.Sprintf("%s %s", v.Asset, v.Amount)
}

func (p Portion) String() string {
	r := big.Rat(p)
	return r.String()
}

func (v Asset) String() string {
	return string(v)
}

func expectMonetary(v Value, r parser.Range) (*Monetary, InterpreterError) {
	switch v := v.(type) {
	case Monetary:
		return &v, nil

	default:
		return nil, TypeError{Expected: analysis.TypeMonetary, Value: v, Range: r}
	}
}

func expectMonetaryOfAsset(expectedAsset string) func(Value, parser.Range) (*big.Int, InterpreterError) {
	return func(v Value, r parser.Range) (*big.Int, InterpreterError) {
		m, err := expectMonetary(v, r)
		if err != nil {
			return nil, err
		}

		asset := string(m.Asset)

		if asset != expectedAsset {
			return nil, MismatchedCurrencyError{Expected: expectedAsset, Got: asset}
		}

		i := big.Int(m.Amount)
		return &i, nil
	}
}

func expectNumber(v Value, r parser.Range) (*big.Int, InterpreterError) {
	switch v := v.(type) {
	case MonetaryInt:
		return (*big.Int)(&v), nil

	default:
		return nil, TypeError{Expected: analysis.TypeNumber, Value: v, Range: r}
	}
}

func expectString(v Value, r parser.Range) (*string, InterpreterError) {
	switch v := v.(type) {
	case String:
		return (*string)(&v), nil

	default:
		return nil, TypeError{Expected: analysis.TypeString, Value: v, Range: r}
	}
}

func expectAsset(v Value, r parser.Range) (*string, InterpreterError) {
	switch v := v.(type) {
	case Asset:
		return (*string)(&v), nil

	default:
		return nil, TypeError{Expected: analysis.TypeAsset, Value: v, Range: r}
	}
}

func expectAccount(v Value, r parser.Range) (*string, InterpreterError) {
	switch v := v.(type) {
	case AccountAddress:
		return (*string)(&v), nil

	default:
		return nil, TypeError{Expected: analysis.TypeAccount, Value: v, Range: r}
	}
}

func expectPortion(v Value, r parser.Range) (*big.Rat, InterpreterError) {
	switch v := v.(type) {
	case Portion:
		return (*big.Rat)(&v), nil

	default:
		return nil, TypeError{Expected: analysis.TypePortion, Value: v, Range: r}
	}
}

func expectAnything(v Value, _ parser.Range) (*Value, InterpreterError) {
	return &v, nil
}

func expectOneOf[T any](combinators ...func(v Value, r parser.Range) (*T, InterpreterError)) func(v Value, r parser.Range) (*T, InterpreterError) {
	return func(v Value, r parser.Range) (*T, InterpreterError) {
		if len(combinators) == 0 {
			// this should be unreachable
			panic("Invalid argument: no combinators given")
		}

		var errs []TypeError
		for _, combinator := range combinators {
			out, err := combinator(v, r)
			if err == nil {
				return out, nil
			}

			typeErr, ok := err.(TypeError)
			if !ok {
				return nil, err
			}
			errs = append(errs, typeErr)
		}

		// e.g. typeErr.map(e => e.Expected).join("|")
		expected := ""
		for index, typeErr := range errs {
			if index != 0 {
				expected += "|"
			}
			expected += typeErr.Expected
		}

		return nil, TypeError{
			Range:    r,
			Value:    v,
			Expected: expected,
		}
	}
}

func expectMapped[T any, U any](
	combinator func(v Value, r parser.Range) (*T, InterpreterError),
	mapper func(value T) U,
) func(v Value, r parser.Range) (*U, InterpreterError) {
	return func(v Value, r parser.Range) (*U, InterpreterError) {
		out, err := combinator(v, r)
		if err != nil {
			return nil, err
		}
		mapped := mapper(*out)
		return &mapped, nil
	}
}

func NewMonetary(asset string, n int64) Monetary {
	return Monetary{
		Asset:  Asset(asset),
		Amount: NewMonetaryInt(n),
	}
}

func NewMonetaryInt(n int64) MonetaryInt {
	bi := big.NewInt(n)
	return MonetaryInt(*bi)
}

func (m MonetaryInt) Add(other MonetaryInt) MonetaryInt {
	bi := big.Int(m)
	otherBi := big.Int(other)

	sum := new(big.Int).Add(&bi, &otherBi)
	return MonetaryInt(*sum)
}

func (m MonetaryInt) Sub(other MonetaryInt) MonetaryInt {
	bi := big.Int(m)
	otherBi := big.Int(other)

	sum := new(big.Int).Sub(&bi, &otherBi)
	return MonetaryInt(*sum)
}
