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

func (e MissingFundsErr) Error() string {
	return fmt.Sprintf("Not enough funds. Missing %s (sent %s)", e.Missing.String(), e.Sent.String())
}

type TypeError struct {
	Expected string
	Value    Value
}

func (e TypeError) Error() string {
	return fmt.Sprintf("Invalid value received. Expecting value of type %s (got %#v instead)", e.Expected, e.Value)
}

func parsePercentage(p string) big.Rat {
	num, den, err := parser.ParsePercentageRatio(p)
	if err != nil {
		panic(err)
	}
	return *big.NewRat(int64(num), int64(den))
}

func parseVar(type_ string, rawValue string) (Value, error) {
	switch type_ {
	// TODO why should the runtime depend on the static analysis module?
	case analysis.TypeMonetary:
		panic("TODO handle parsing of: " + type_)
	case analysis.TypeAccount:
		return AccountAddress(rawValue), nil
	case analysis.TypePortion:
		return Portion(parsePercentage(rawValue)), nil
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

func meta(
	s *programState,
	args []Value,
) (string, error) {
	if len(args) < 2 {
		panic("TODO handle type error in meta")
	}

	account, err := expectAccount(args[0])
	if err != nil {
		return "", err
	}

	key, err := expectString(args[1])
	if err != nil {
		return "", err
	}

	// body
	accountMeta := s.Meta[account.String()]
	value, ok := accountMeta[string(*key)]

	if !ok {
		// TODO err
		panic("META NOT FOUND")
	}

	return value, nil
}

func (s *programState) handleOrigin(type_ string, fnCall parser.FnCall) (Value, error) {
	args, err := s.evaluateLiterals(fnCall.Args)
	if err != nil {
		return nil, err
	}

	switch fnCall.Caller.Name {
	case "meta":
		rawValue, err := meta(s, args)
		if err != nil {
			return nil, err
		}

		parsed, err := parseVar(type_, rawValue)
		if err != nil {
			return nil, err
		}

		return parsed, nil

	default:
		panic("TODO handle fn call: " + fnCall.Caller.Name)
	}

}

func (s *programState) parseVars(varDeclrs []parser.VarDeclaration, rawVars map[string]string) error {
	for _, varsDecl := range varDeclrs {
		if varsDecl.Origin == nil {
			raw, ok := rawVars[varsDecl.Name.Name]
			if !ok {
				panic("TODO handle var not found: " + varsDecl.Name.Name)
			}
			parsed, err := parseVar(varsDecl.Type.Name, raw)
			if err != nil {
				return err
			}
			s.Vars[varsDecl.Name.Name] = parsed
		} else {
			value, err := s.handleOrigin(varsDecl.Type.Name, *varsDecl.Origin)
			if err != nil {
				return err
			}
			s.Vars[varsDecl.Name.Name] = value
		}
	}
	return nil
}

func RunProgram(
	program parser.Program,
	vars map[string]string,
	store StaticStore,
	meta map[string]Metadata,
) (*ExecutionResult, error) {
	st := programState{
		Vars:   make(map[string]Value),
		TxMeta: make(map[string]Value),
		Store:  store,
		Meta:   meta,
	}

	err := st.parseVars(program.Vars, vars)
	if err != nil {
		return nil, err
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
	Meta      map[string]Metadata
}

func (st *programState) evaluateLit(literal parser.Literal) (Value, error) {
	switch literal := literal.(type) {
	case *parser.AssetLiteral:
		return Asset(literal.Asset), nil
	case *parser.AccountLiteral:
		return AccountAddress(literal.Name), nil
	case *parser.StringLiteral:
		return String(literal.String), nil
	case *parser.RatioLiteral:
		return Portion(*literal.ToRatio()), nil
	case *parser.NumberLiteral:
		return MonetaryInt(*big.NewInt(int64(literal.Number))), nil
	case *parser.MonetaryLiteral:
		assetValue, err := st.evaluateLit(literal.Asset)
		if err != nil {
			return Monetary{}, err
		}

		asset, err := expectAsset(assetValue)
		if err != nil {
			return nil, err
		}

		amountValue, err := st.evaluateLit(literal.Amount)
		if err != nil {
			return Monetary{}, err
		}

		amount, err := expectNumber(amountValue)
		if err != nil {
			return nil, err
		}

		return Monetary{Asset: *asset, Amount: *amount}, nil

	case *parser.VariableLiteral:
		value, ok := st.Vars[literal.Name]
		if !ok {
			panic("TODO err for unbound variable")
		}
		return value, nil
	default:
		panic("TODO handle literal evaluation")
	}
}

func (st *programState) evaluateLiterals(literals []parser.Literal) ([]Value, error) {
	var values []Value
	for _, argLit := range literals {
		value, err := st.evaluateLit(argLit)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func (st *programState) runStatement(statement parser.Statement) ([]Posting, error) {
	st.Senders = nil
	st.Receivers = nil

	switch statement := statement.(type) {
	case *parser.FnCall:
		args, err := st.evaluateLiterals(statement.Args)
		if err != nil {
			return nil, err
		}

		switch statement.Caller.Name {
		case "set_tx_meta":
			err := setTxMeta(st, args)
			if err != nil {
				return nil, err
			}
		default:
			panic("Invalid fn")
		}
		return nil, nil

	case *parser.SendStatement:
		return st.runSendStatement(*statement)
	default:
		panic("TODO unhandled clause")
	}
}

func setTxMeta(st *programState, args []Value) error {
	if len(args) != 2 {
		// TODO err
		panic("invalid args number")
	}

	k, err := expectString(args[0])
	if err != nil {
		return err
	}

	meta := args[1]
	st.TxMeta[string(*k)] = meta
	return nil
}

func (st *programState) runSendStatement(statement parser.SendStatement) ([]Posting, error) {
	switch sentValue := statement.SentValue.(type) {
	case *parser.SentValueAll:
		panic("TODO handle send*")
	case *parser.SentValueLiteral:
		sentValue_, err := st.evaluateLit(sentValue.Monetary)
		if err != nil {
			return nil, err
		}
		monetary, err := expectMonetary(sentValue_)
		if err != nil {
			return nil, err
		}

		sentTotal := st.trySending(statement.Source, *monetary)

		// sentTotal < monetary.Amount
		if sentTotal.Cmp((*big.Int)(&monetary.Amount)) == -1 {
			var missing big.Int
			return nil, MissingFundsErr{
				Missing: *missing.Sub((*big.Int)(&monetary.Amount), &sentTotal),
				Sent:    sentTotal,
			}
		}

		st.receiveFrom(statement.Destination, *monetary)

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

		assetBalance := s.getBalance(name, string(monetary.Asset))
		assetBalance.Sub(assetBalance, &monetaryAmount)
	}

	s.Senders = append(s.Senders, Sender{
		Name:     name,
		Monetary: &monetaryAmount,
		Asset:    string(monetary.Asset),
	})

	return monetaryAmount
}

func (s *programState) trySending(source parser.Source, monetary Monetary) big.Int {
	switch source := source.(type) {
	case *parser.VariableLiteral:
		account, err := expectAccountLit(source, s.Vars)
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
	monetaryAmount := big.Int(monetary.Amount)

	switch destination := destination.(type) {
	case *parser.AccountLiteral:
		return s.receiveFromAccount(destination.Name, monetary)
	case *parser.VariableLiteral:
		account, err := expectAccountLit(destination, s.Vars)
		if err != nil {
			// TODO return err
			panic(err)
		}

		return s.receiveFromAccount(string(account), monetary)

	case *parser.DestinationAllotment:
		// TODO runtime error when totalAllotment != 1?
		totalAllotment := big.NewRat(0, 1)
		receivedTotal := big.NewInt(0)
		var allotments []big.Rat

		remainingAllotmentIndex := -1

		for i, item := range destination.Items {
			switch allotment := item.Allotment.(type) {
			case *parser.RatioLiteral:
				rat := big.NewRat(int64(allotment.Numerator), int64(allotment.Denominator))
				totalAllotment.Add(totalAllotment, rat)
				allotments = append(allotments, *rat)
			case *parser.VariableLiteral:
				p, err := expectPortionLit(allotment, s.Vars)
				if err != nil {
					// TODO return err
					panic(err)
				}
				rat := big.Rat(p)
				totalAllotment.Add(totalAllotment, &rat)
				allotments = append(allotments, rat)

			case *parser.RemainingAllotment:
				remainingAllotmentIndex = i
				var rat big.Rat
				allotments = append(allotments, rat)
				// TODO check there are not duplicate remaining clause
			}
		}

		if remainingAllotmentIndex != -1 {
			var rat big.Rat
			rat.Sub(big.NewRat(1, 1), totalAllotment)
			allotments[remainingAllotmentIndex] = rat
		}

		allot := makeAllotment(monetaryAmount.Int64(), allotments)
		for i, allotmentItem := range destination.Items {
			allot_ := allot[i]

			switch allotmentItem := allotmentItem.To.(type) {
			case *parser.DestinationTo:
				dest := allotmentItem.Destination
				receivedMon := monetary
				receivedMon.Amount = NewMonetaryInt(allot_)
				received := s.receiveFrom(dest, receivedMon)
				receivedTotal.Add(receivedTotal, &received)

			case *parser.DestinationKept:
				panic("TODO handle kept destination")
			}

		}

		return *receivedTotal

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

	// case *parser.SourceCapped:
	// case *parser.SourceOverdraft:
	// case *parser.VariableLiteral:
	default:
		panic("TODO handle clause")

	}

}

func makeAllotment(monetary int64, allotments []big.Rat) [](int64) {
	parts := make([]int64, len(allotments))

	var totalAllocated int64

	for i, allot := range allotments {
		var product big.Rat
		product.Mul(&allot, big.NewRat(monetary, 1))

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
