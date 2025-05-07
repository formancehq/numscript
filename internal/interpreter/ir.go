// An intermediate representation of the AST which is created after vars substitution occur
// and other data is validated into a proper data structure (e.g. no type errors)

package interpreter

import (
	"math/big"

	"github.com/formancehq/numscript/internal/parser"
)

type (
	Source interface {
		parser.Ranged
		irSource()
	}

	SourceAccount struct {
		parser.Range
		Color   string
		Account string

		// nil=unbounded
		// 0=no overdraft
		Overdraft *big.Int
	}

	SourceInorder struct {
		parser.Range
		Sources []Source
	}

	SourceOneof struct {
		parser.Range
		Sources []Source
	}

	SourceAllotment struct {
		parser.Range
		Items []SourceAllotmentItem
	}

	SourceAllotmentItem struct {
		parser.Range
		Allotment Portion
		From      Source
	}

	SourceCapped struct {
		parser.Range
		From Source
		Cap  *big.Int
	}
)

func (SourceAccount) irSource() {}

func (SourceInorder) irSource() {}

func (SourceOneof) irSource() {}

func (SourceAllotment) irSource() {}

func (SourceAllotmentItem) irSource() {}

func (SourceCapped) irSource() {}
