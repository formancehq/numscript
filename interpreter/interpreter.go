package interpreter

import (
	"fmt"
	"math/big"
	"numscript/analysis"
	"numscript/parser"
	"strconv"
)

type Metadata map[string]string

type ExecutionResult struct {
	Postings []Posting
	TxMeta   map[string]Value
}

type MissingFundsErr struct {
	Missing big.Int
	Sent    big.Int
}

func (m MissingFundsErr) Error() string {
	return fmt.Sprintf("Not enough funds. Missing %s (sent %s)", m.Missing.String(), m.Sent.String())
}

func parseVar(type_ string, rawValue string) (Value, error) {
	switch type_ {
	// TODO why should the runtime depend on the static analysis module?
	case analysis.TypeMonetary:
		panic("TODO handle parsing of: " + type_)
	case analysis.TypeAccount:
		return AccountAddress(rawValue), nil
	case analysis.TypePortion:
		panic("TODO handle parsing of: " + type_)
	case analysis.TypeAsset:
		return Asset(rawValue), nil
	case analysis.TypeNumber:
		// TODO check original numscript impl
		i, err := strconv.ParseInt(rawValue, 0, 64)
		if err != nil {
			return nil, err
		}
		return NewMonetaryInt(i), nil

	case analysis.TypeString:
		return String(rawValue), nil
	default:
		panic("TODO invalid type: " + type_)
	}

}

func parseVars(varDeclrs []parser.VarDeclaration, rawVars map[string]string) (map[string]Value, error) {
	parsedVars := make(map[string]Value)
	for _, varsDecl := range varDeclrs {
		raw, ok := rawVars[varsDecl.Name.Name]
		if !ok {
			panic("TODO handle var not found")
		}
		parsed, err := parseVar(varsDecl.Type.Name, raw)
		if err != nil {
			return nil, err
		}
		parsedVars[varsDecl.Name.Name] = parsed

	}
	return parsedVars, nil
}

func RunProgram(
	program parser.Program,
	vars map[string]string,
	store StaticStore,
) (*ExecutionResult, error) {
	parsedVars, err := parseVars(program.Vars, vars)
	if err != nil {
		return nil, err
	}

	st := programState{
		Vars:   parsedVars,
		TxMeta: make(map[string]Value),
		Store:  store,
	}

	var postings []Posting
	for _, statement := range program.Statements {
		statementPostings, err := st.runStatement(statement)
		if err != nil {
			return nil, err
		}
		postings = append(postings, statementPostings...)
	}

	res := &ExecutionResult{
		Postings: postings,
		TxMeta:   st.TxMeta,
	}
	return res, nil
}

type programState struct {
	Vars      map[string]Value
	TxMeta    map[string]Value
	Store     StaticStore
	Senders   []Sender
	Receivers []Receiver
}

func (st *programState) runStatement(statement parser.Statement) ([]Posting, error) {
	st.Senders = nil
	st.Receivers = nil

	switch statement := statement.(type) {
	case *parser.FnCall:
		return nil, st.runFnCall(*statement)
	case *parser.SendStatement:
		return st.runSendStatement(*statement)
	default:
		panic("TODO unhandled clause")
	}
}

func (st *programState) runFnCall(f parser.FnCall) error {
	switch f.Caller.Name {
	case "set_tx_meta":
		if len(f.Args) != 2 {
			// TODO err
			panic("invalid args number")
		}

		k, err := expectString(f.Args[0], st.Vars)
		if err != nil {
			return err
		}

		meta, err := expectAnything(f.Args[1], st.Vars)
		if err != nil {
			return err
		}

		st.TxMeta[string(k)] = meta
		return nil
	default:
		panic("TODO handle unknown caller")
	}

}

func (st *programState) runSendStatement(statement parser.SendStatement) ([]Posting, error) {
	switch sentValue := statement.SentValue.(type) {
	case *parser.SentValueAll:
		panic("TODO handle send*")
	case *parser.SentValueLiteral:
		monetary, err := expectMonetary(sentValue.Monetary, st.Vars)
		if err != nil {
			return nil, err
		}

		sentTotal := st.trySending(statement.Source, monetary)

		// sentTotal < monetary.Amount
		if sentTotal.Cmp((*big.Int)(&monetary.Amount)) == -1 {
			var missing big.Int
			return nil, MissingFundsErr{
				Missing: *missing.Sub((*big.Int)(&monetary.Amount), &sentTotal),
				Sent:    sentTotal,
			}
		}

		st.receiveFrom(statement.Destination, monetary)

		postings, err := Reconcile(st.Senders, st.Receivers)
		if err != nil {
			return nil, err
		}
		return postings, nil
	default:
		panic("TODO handle")
	}

}

func (s *programState) getBalance(account string, asset string) *big.Int {
	balance, ok := s.Store[account]
	if !ok {
		panic(fmt.Sprintf("balance for '%s' not found (given: %v)", account, s.Store))
	}

	assetBalance, ok := balance.Balances[asset]
	if !ok {
		panic("balance not found for the given currency")
	}
	return assetBalance
}

func (s *programState) trySendingAccount(name string, monetary Monetary) big.Int {
	monetaryAmount := big.Int(monetary.Amount)

	if name != "world" {
		balance := s.getBalance(name, string(monetary.Asset))

		// monetary = min(balance, monetary)
		if balance.Cmp(&monetaryAmount) == -1 /* balance < monetary */ {
			monetaryAmount.Set(balance)
		}
	}

	s.Senders = append(s.Senders, Sender{
		Name:     name,
		Monetary: &monetaryAmount,
		Asset:    string(monetary.Asset),
	})

	assetBalance := s.getBalance(name, string(monetary.Asset))
	assetBalance.Sub(assetBalance, &monetaryAmount)
	return monetaryAmount
}

func (s *programState) trySending(source parser.Source, monetary Monetary) big.Int {
	switch source := source.(type) {
	case *parser.VariableLiteral:
		account, err := expectAccount(source, s.Vars)
		if err != nil {
			// TODO return err
			panic(err)
		}
		return s.trySendingAccount(string(account), monetary)

	case *parser.AccountLiteral:
		return s.trySendingAccount(source.Name, monetary)

	case *parser.SourceInorder:
		sentTotal := big.NewInt(0)
		for _, source := range source.Sources {
			var sendingMonetary big.Int
			sendingMonetary.Sub((*big.Int)(&monetary.Amount), sentTotal)
			sentAmt := s.trySending(source, Monetary{
				Amount: MonetaryInt(sendingMonetary),
				Asset:  monetary.Asset,
			})
			sentTotal.Add(sentTotal, &sentAmt)
		}
		return *sentTotal

	// case *parser.SourceAllotment:
	// case *parser.SourceCapped:
	// case *parser.SourceOverdraft:
	// case *parser.VariableLiteral:
	default:
		panic("TODO handle clause")

	}

}

func (s *programState) receiveFromAccount(name string, monetary Monetary) big.Int {
	mon := big.Int(monetary.Amount)
	s.Receivers = append(s.Receivers, Receiver{
		Name:     name,
		Monetary: &mon,
		Asset:    string(monetary.Asset),
	})
	return mon
}

func (s *programState) receiveFrom(destination parser.Destination, monetary Monetary) big.Int {
	switch destination := destination.(type) {
	case *parser.AccountLiteral:
		return s.receiveFromAccount(destination.Name, monetary)
	case *parser.VariableLiteral:
		account, err := expectAccount(destination, s.Vars)
		if err != nil {
			// TODO return err
			panic(err)
		}

		return s.receiveFromAccount(string(account), monetary)

	// case *parser.DestinationInorder:
	// sentTotal := big.NewInt(0)
	// for _, source := range source.Sources {
	// 	var sendingMonetary big.Int
	// 	sendingMonetary.Sub((*big.Int)(&monetary.Amount), sentTotal)
	// 	sentAmt := s.trySending(source, Monetary{
	// 		Amount: MonetaryInt(sendingMonetary),
	// 		Asset:  monetary.Asset,
	// 	})
	// 	sentTotal.Add(sentTotal, &sentAmt)
	// }
	// return *sentTotal

	// receivedTotal := big.NewInt(0)
	// for _, destination := range d.Destinations {
	// 	receivedTotal += destination.receive(monetary-receivedTotal, ctx)
	// 	// if receivedTotal >= monetary {
	// 	// 	break
	// 	// }
	// }

	// return receivedTotal

	// case *parser.SourceAllotment:
	// case *parser.SourceCapped:
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
