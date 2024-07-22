package interpreter

import (
	"math/big"
	"numscript/parser"
)

type Metadata map[string]string

type ExecutionResult struct {
	Postings []Posting
	TxMeta   map[string]Value
}

type MissingFundsErr struct {
	error
	Missing big.Int
}

func RunProgram(
	program parser.Program,
	vars map[string]string,
	store StaticStore,
) (*ExecutionResult, error) {
	st := programState{}

	res := &ExecutionResult{
		Postings: nil,
		TxMeta:   map[string]Value{},
	}

	for _, statement := range program.Statements {
		postings, err := st.runStatement(statement)
		if err != nil {
			return nil, err
		}
		res.Postings = append(res.Postings, postings...)
	}

	postings, err := Reconcile(st.Senders, st.Receivers)
	if err != nil {
		return res, err
	}
	res.Postings = postings
	return res, nil
}

type programState struct {
	Vars      map[string]string
	Store     StaticStore
	Senders   []Sender
	Receivers []Receiver
}

func (st *programState) runStatement(statement parser.Statement) ([]Posting, error) {
	st.Senders = nil
	st.Receivers = nil

	switch statement := statement.(type) {
	case *parser.FnCall:
		panic("TODO handle fn call")
	case *parser.SendStatement:
		switch sentValue := statement.SentValue.(type) {
		case *parser.SentValueAll:
			panic("TODO handle send*")
		case *parser.SentValueLiteral:
			monetary, err := expectMonetary(sentValue.Monetary)
			if err != nil {
				return nil, err
			}

			sentTotal := st.trySending(statement.Source, monetary)

			// sentTotal < monetary.Amount
			if sentTotal.Cmp((*big.Int)(&monetary.Amount)) == -1 {
				var missing big.Int
				return nil, MissingFundsErr{Missing: *missing.Sub((*big.Int)(&monetary.Amount), &sentTotal)}
			}

			st.receiveFrom(statement.Destination, monetary)

			postings, err := Reconcile(st.Senders, st.Receivers)
			if err != nil {
				return nil, err
			}
			return postings, nil
		}

	}

	panic("TODO unhandled clause")
}

func (s *programState) trySending(source parser.Source, monetary Monetary) big.Int {
	switch source := source.(type) {
	case *parser.AccountLiteral:
		// if s.Name != "world" {
		// 	monetary = min(ctx.Balances[s.Name], monetary)
		// }

		mon := big.Int(monetary.Amount)
		s.Senders = append(s.Senders, Sender{
			Name:     source.Name,
			Monetary: &mon,
			Asset:    string(monetary.Asset),
		})

		// if ctx.Balances != nil {
		// 	ctx.Balances[s.Name] -= monetary
		// }

		return mon

	// case *parser.SourceAllotment:
	// case *parser.SourceCapped:
	// case *parser.SourceInorder:
	// case *parser.SourceOverdraft:
	// case *parser.VariableLiteral:
	default:
		panic("TODO handle clause")

	}

}

func (s *programState) receiveFrom(destination parser.Destination, monetary Monetary) big.Int {
	switch destination := destination.(type) {
	case *parser.AccountLiteral:
		mon := big.Int(monetary.Amount)
		s.Receivers = append(s.Receivers, Receiver{
			Name:     destination.Name,
			Monetary: &mon,
			Asset:    string(monetary.Asset),
		})
		return mon

	// case *parser.SourceAllotment:
	// case *parser.SourceCapped:
	// case *parser.SourceInorder:
	// case *parser.SourceOverdraft:
	// case *parser.VariableLiteral:
	default:
		panic("TODO handle clause")

	}

}

func makeAllotment[T interface{}](monetary int64, allotments []Allotment[T]) []int64 {
	parts := make([]int64, len(allotments))

	var totalAllocated int64

	for i, allot := range allotments {
		var product big.Rat
		product.Mul(&allot.Ratio, big.NewRat(monetary, 1))

		floored := product.Num().Int64() / product.Denom().Int64()

		parts[i] = floored
		totalAllocated += floored
	}

	for i := range parts {
		if totalAllocated >= monetary {
			break
		}

		parts[i]++
		totalAllocated++
	}

	return parts
}
