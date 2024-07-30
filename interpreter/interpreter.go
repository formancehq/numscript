package interpreter

import (
	"fmt"
	"math/big"
	"numscript/analysis"
	"numscript/parser"
	"strconv"
	"strings"
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

func parseMonetary(source string) (Monetary, error) {
	stripBrackets := source[1 : len(source)-1]
	parts := strings.Split(stripBrackets, " ")
	if len(parts) != 2 {
		// TODO proper error handling
		panic("Invalid mon literal")
	}

	// TODO check original numscript impl
	rawAmount := parts[1]
	parsedAmount, err := strconv.ParseInt(rawAmount, 0, 64)
	if err != nil {
		return Monetary{}, err
	}
	mon := Monetary{
		Asset:  Asset(rawAmount),
		Amount: NewMonetaryInt(parsedAmount),
	}
	return mon, nil
}

func parseVar(type_ string, rawValue string) (Value, error) {
	switch type_ {
	// TODO why should the runtime depend on the static analysis module?
	case analysis.TypeMonetary:
		return parseMonetary(rawValue)
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
		asset, err := evaluateLitExpecting(st, literal.Asset, expectAsset)
		if err != nil {
			return nil, err
		}

		amount, err := evaluateLitExpecting(st, literal.Amount, expectNumber)
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

func evaluateLitExpecting[T any](st *programState, literal parser.Literal, expect func(Value) (*T, error)) (*T, error) {
	value, err := st.evaluateLit(literal)
	if err != nil {
		return nil, err
	}

	res, err := expect(value)
	if err != nil {
		return nil, err
	}

	return res, nil
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
		monetary, err := evaluateLitExpecting(st, sentValue.Monetary, expectMonetary)
		if err != nil {
			return nil, err
		}

		sentTotal := st.trySending(statement.Source, *monetary)

		// sentTotal < monetary.Amount
		if sentTotal.Cmp((*big.Int)(&monetary.Amount)) == -1 {
			var missing big.Int
			missing.Sub((*big.Int)(&monetary.Amount), &sentTotal)
			return nil, MissingFundsErr{
				Missing: missing,
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
		m := &AccountWithBalances{Balances: make(map[string]*big.Int)}
		s.Store[account] = m
		balance = m
	}

	assetBalance, ok := balance.Balances[asset]
	if !ok {
		zero := big.NewInt(0)
		balance.Balances[asset] = zero
		assetBalance = zero
	}
	return assetBalance
}

func (s *programState) trySendingAccount(name string, monetary Monetary) big.Int {
	var monetaryAmount big.Int
	amtRef := big.Int(monetary.Amount)
	monetaryAmount.Set(&amtRef)

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
		account, err := evaluateLitExpecting(s, source, expectAccount)
		if err != nil {
			// TODO proper error handling
			panic(err)
		}
		return s.trySendingAccount(string(*account), monetary)

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

	case *parser.SourceAllotment:
		monetaryAmount := big.Int(monetary.Amount)
		receivedTotal := big.NewInt(0)
		var items []parser.AllotmentValue
		for _, i := range source.Items {
			items = append(items, i.Allotment)
		}
		allot := s.makeAllotment(monetaryAmount.Int64(), items)
		for i, allotmentItem := range source.Items {
			source := allotmentItem.From
			receivedMon := monetary
			receivedMon.Amount = NewMonetaryInt(allot[i])
			received := s.trySending(source, receivedMon)
			receivedTotal.Add(receivedTotal, &received)
		}
		return *receivedTotal

	// case *parser.SourceCapped:
	// case *parser.SourceOverdraft:

	default:
		panic("TODO handle clause")

	}

}

func (s *programState) receiveFromAccount(name string, monetary Monetary) big.Int {
	mon := big.Int(monetary.Amount)

	balance := s.getBalance(name, string(monetary.Asset))
	balance.Add(balance, &mon)

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
		account, err := evaluateLitExpecting(s, destination, expectAccount)
		if err != nil {
			// TODO proper error handling
			panic(err)
		}
		return s.receiveFromAccount(string(*account), monetary)

	case *parser.DestinationAllotment:
		monetaryAmount := big.Int(monetary.Amount)
		var items []parser.AllotmentValue
		for _, i := range destination.Items {
			items = append(items, i.Allotment)
		}

		allot := s.makeAllotment(monetaryAmount.Int64(), items)

		receivedTotal := big.NewInt(0)
		for i, allotmentItem := range destination.Items {
			switch allotmentItem := allotmentItem.To.(type) {
			case *parser.DestinationTo:
				dest := allotmentItem.Destination
				receivedMon := monetary
				receivedMon.Amount = NewMonetaryInt(allot[i])
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

func (s *programState) makeAllotment(monetary int64, items []parser.AllotmentValue) []int64 {
	// TODO runtime error when totalAllotment != 1?
	totalAllotment := big.NewRat(0, 1)
	var allotments []big.Rat

	remainingAllotmentIndex := -1

	for i, item := range items {
		switch allotment := item.(type) {
		case *parser.RatioLiteral:
			rat := big.NewRat(int64(allotment.Numerator), int64(allotment.Denominator))
			totalAllotment.Add(totalAllotment, rat)
			allotments = append(allotments, *rat)
		case *parser.VariableLiteral:
			portion, err := evaluateLitExpecting(s, allotment, expectPortion)
			if err != nil {
				// TODO proper error handling
				panic(err)
			}

			rat := big.Rat(*portion)
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
