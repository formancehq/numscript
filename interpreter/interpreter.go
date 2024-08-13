package interpreter

import (
	"math/big"
	"numscript/analysis"
	"numscript/parser"
	"numscript/utils"
	"strconv"
	"strings"
)

type StaticStore map[string]map[string]*big.Int
type Metadata map[string]string

type ExecutionResult struct {
	Postings []Posting        `json:"postings"`
	TxMeta   map[string]Value `json:"txMeta"`
}

func parsePercentage(p string) big.Rat {
	num, den, err := parser.ParsePercentageRatio(p)
	if err != nil {
		panic(err)
	}
	return *big.NewRat(int64(num), int64(den))
}

func parseMonetary(source string) (Monetary, error) {
	parts := strings.Split(source, " ")
	if len(parts) != 2 {
		// TODO proper error handling
		panic("Invalid mon literal")
	}

	asset := parts[0]

	// TODO check original numscript impl
	rawAmount := parts[1]
	parsedAmount, err := strconv.ParseInt(rawAmount, 0, 64)
	if err != nil {
		return Monetary{}, err
	}
	mon := Monetary{
		Asset:  Asset(asset),
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
		return nil, InvalidTypeErr{Name: type_}
	}

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

	case "balance":
		monetary, err := balance(s, args)
		if err != nil {
			return nil, err
		}
		return *monetary, nil

	default:
		return nil, UnboundFunctionErr{Name: fnCall.Caller.Name}
	}

}

func (s *programState) parseVars(varDeclrs []parser.VarDeclaration, rawVars map[string]string) error {
	for _, varsDecl := range varDeclrs {
		if varsDecl.Origin == nil {
			raw, ok := rawVars[varsDecl.Name.Name]
			if !ok {
				return MissingVariableErr{Name: varsDecl.Name.Name}
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

	postings := make([]Posting, 0)
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

		return Monetary{Asset: Asset(*asset), Amount: MonetaryInt(*amount)}, nil

	case *parser.VariableLiteral:
		value, ok := st.Vars[literal.Name]
		if !ok {
			return nil, UnboundVariableErr{Name: literal.Name}
		}
		return value, nil
	default:
		utils.NonExhaustiveMatchPanic[any](literal)
		return nil, nil
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
		case analysis.FnSetTxMeta:
			err := setTxMeta(st, args)
			if err != nil {
				return nil, err
			}
		default:
			return nil, UnboundFunctionErr{Name: statement.Caller.Name}
		}
		return nil, nil

	case *parser.SendStatement:
		return st.runSendStatement(*statement)
	default:
		utils.NonExhaustiveMatchPanic[any](statement)
		return nil, nil
	}
}

func (st *programState) getPostings() ([]Posting, error) {
	postings, err := Reconcile(st.Senders, st.Receivers)
	if err != nil {
		return nil, err
	}

	for _, posting := range postings {
		srcBalance := st.getBalance(posting.Source, posting.Asset)
		srcBalance.Sub(srcBalance, posting.Amount)

		destBalance := st.getBalance(posting.Destination, posting.Asset)
		destBalance.Add(destBalance, posting.Amount)
	}
	return postings, nil
}

func (st *programState) runSendStatement(statement parser.SendStatement) ([]Posting, error) {
	switch sentValue := statement.SentValue.(type) {
	case *parser.SentValueAll:
		asset, err := evaluateLitExpecting(st, sentValue.Asset, expectAsset)
		if err != nil {
			return nil, err
		}
		err = st.trySendingAll(statement.Source, *asset)
		if err != nil {
			return nil, err
		}
		err = st.receiveAllFrom(statement.Destination, *asset)
		if err != nil {
			return nil, err
		}
		return st.getPostings()

	case *parser.SentValueLiteral:
		monetary, err := evaluateLitExpecting(st, sentValue.Monetary, expectMonetary)
		if err != nil {
			return nil, err
		}

		monetaryAmt := big.Int(monetary.Amount)
		if monetaryAmt.Cmp(big.NewInt(0)) == -1 {
			return nil, NegativeAmountErr{Amount: monetary.Amount}
		}

		sentTotal, err := st.trySending(statement.Source, *monetary)
		if err != nil {
			return nil, err
		}

		// sentTotal < monetary.Amount
		if sentTotal.Cmp((*big.Int)(&monetary.Amount)) == -1 {
			var missing big.Int
			missing.Sub((*big.Int)(&monetary.Amount), sentTotal)
			return nil, MissingFundsErr{
				Missing: missing,
				Sent:    *sentTotal,
			}
		}

		st.receiveFrom(statement.Destination, *monetary)

		return st.getPostings()
	default:
		utils.NonExhaustiveMatchPanic[any](sentValue)
		return nil, nil
	}

}

func (s *programState) getBalance(account string, asset string) *big.Int {
	balance, ok := s.Store[account]
	if !ok {
		m := make(map[string]*big.Int)
		s.Store[account] = m
		balance = m
	}

	assetBalance, ok := balance[asset]
	if !ok {
		zero := big.NewInt(0)
		balance[asset] = zero
		assetBalance = zero
	}
	return assetBalance
}

func (s *programState) trySendingAccount(name string, monetary Monetary) (*big.Int, error) {
	var monetaryAmount big.Int
	amtRef := big.Int(monetary.Amount)
	monetaryAmount.Set(&amtRef)

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

	return &monetaryAmount, nil
}

func (s *programState) trySendingAll(source parser.Source, asset string) error {
	switch source := source.(type) {
	case *parser.AccountLiteral:
		// TODO error when unbounded
		// TODO err empty balance?

		balance := s.getBalance(source.Name, asset)
		var sendMonetary big.Int
		sendMonetary.Set(balance)

		s.Senders = append(s.Senders, Sender{
			Name:     source.Name,
			Monetary: sendMonetary.Set(balance),
			Asset:    asset,
		})
		return nil

	case *parser.VariableLiteral:
		panic("TODO SEND ALL FROM VAR LIT")

	case *parser.SourceInorder:
		for _, subSource := range source.Sources {
			err := s.trySendingAll(subSource, asset)
			if err != nil {
				return err
			}
		}
		return nil

	case *parser.SourceCapped:
		monetary, err := evaluateLitExpecting(s, source.Cap, expectMonetary)
		if err != nil {
			return err
		}

		// We switch to the default sending evaluation for this subsource
		// We ignore the sent value, as it's ok for the source not to have enough funds
		_, err = s.trySending(source.From, *monetary)
		if err != nil {
			return err
		}

		return nil

	default:
		panic("TODO handle branch")
	}
}

func (s *programState) receiveAllFrom(destination parser.Destination, asset string) error {
	switch destination := destination.(type) {
	case *parser.AccountLiteral:
		s.Receivers = append(s.Receivers, Receiver{
			Name:     destination.Name,
			Monetary: nil,
			Asset:    asset,
		})
		return nil

	case *parser.VariableLiteral:
		panic("TODO variable lit in rcv")

	case *parser.DestinationInorder:
		for _, subDestination := range destination.Clauses {
			cap, err := evaluateLitExpecting(s, subDestination.Cap, expectMonetary)
			if err != nil {
				return err
			}

			switch to := subDestination.To.(type) {
			case *parser.DestinationKept:
				panic("TODO destination kept in send*")

			case *parser.DestinationTo:
				_, err = s.receiveFrom(to.Destination, *cap)
				if err != nil {
					return err
				}
			default:
				utils.NonExhaustiveMatchPanic[any](to)
			}
		}

		switch remaining := destination.Remaining.(type) {
		case *parser.DestinationKept:
			return nil

		case *parser.DestinationTo:
			err := s.receiveAllFrom(remaining.Destination, asset)
			if err != nil {
				return err
			}
			return nil

		default:
			utils.NonExhaustiveMatchPanic[any](remaining)
			return nil
		}

	default:
		panic("TODO handle dest branch")
	}

}

func (s *programState) trySending(source parser.Source, monetary Monetary) (*big.Int, error) {
	switch source := source.(type) {
	case *parser.VariableLiteral:
		account, err := evaluateLitExpecting(s, source, expectAccount)
		if err != nil {
			return nil, err
		}
		return s.trySendingAccount(string(*account), monetary)

	case *parser.AccountLiteral:
		return s.trySendingAccount(source.Name, monetary)

	case *parser.SourceOverdraft:
		var monetaryAmount big.Int
		amtRef := big.Int(monetary.Amount)
		monetaryAmount.Set(&amtRef)

		name, err := evaluateLitExpecting(s, source.Address, expectAccount)
		if err != nil {
			return nil, err
		}

		s.Senders = append(s.Senders, Sender{
			Name:     *name,
			Monetary: &monetaryAmount,
			Asset:    string(monetary.Asset),
		})

		return &monetaryAmount, nil

	case *parser.SourceInorder:
		sentTotal := big.NewInt(0)
		for _, source := range source.Sources {
			var sendingMonetary big.Int
			sendingMonetary.Sub((*big.Int)(&monetary.Amount), sentTotal)
			sentAmt, err := s.trySending(source, Monetary{
				Amount: MonetaryInt(sendingMonetary),
				Asset:  monetary.Asset,
			})
			if err != nil {
				return nil, err
			}
			sentTotal.Add(sentTotal, sentAmt)
		}
		return sentTotal, nil

	case *parser.SourceAllotment:
		monetaryAmount := big.Int(monetary.Amount)
		receivedTotal := big.NewInt(0)
		var items []parser.AllotmentValue
		for _, i := range source.Items {
			items = append(items, i.Allotment)
		}
		allot, err := s.makeAllotment(monetaryAmount.Int64(), items)
		if err != nil {
			return nil, err
		}
		for i, allotmentItem := range source.Items {
			source := allotmentItem.From
			receivedMon := monetary
			receivedMon.Amount = NewMonetaryInt(allot[i])
			received, err := s.trySending(source, receivedMon)
			if err != nil {
				return nil, err
			}
			receivedTotal.Add(receivedTotal, received)
		}
		return receivedTotal, nil

	case *parser.SourceCapped:
		monetaryAmount := big.Int(monetary.Amount)

		cap, err := evaluateLitExpecting(s, source.Cap, expectMonetary)
		if err != nil {
			return nil, err
		}
		// TODO check monetary asset
		capInt := big.Int(cap.Amount)

		var cappedAmount big.Int
		if monetaryAmount.Cmp(&capInt) == -1 /* monetary < cap */ {
			cappedAmount.Set(&monetaryAmount)
		} else {
			cappedAmount.Set(&capInt)
		}

		return s.trySending(source.From, Monetary{
			Amount: MonetaryInt(cappedAmount),
			Asset:  cap.Asset,
		})

	default:
		utils.NonExhaustiveMatchPanic[any](source)
		return nil, nil

	}

}

func (s *programState) receiveFromAccount(name string, monetary Monetary) *big.Int {
	mon := big.Int(monetary.Amount)

	s.Receivers = append(s.Receivers, Receiver{
		Name:     name,
		Monetary: &mon,
		Asset:    string(monetary.Asset),
	})
	return &mon
}

func (s *programState) receiveFrom(destination parser.Destination, monetary Monetary) (*big.Int, error) {
	switch destination := destination.(type) {
	case *parser.AccountLiteral:
		return s.receiveFromAccount(destination.Name, monetary), nil

	case *parser.VariableLiteral:
		account, err := evaluateLitExpecting(s, destination, expectAccount)
		if err != nil {
			return nil, err
		}
		return s.receiveFromAccount(string(*account), monetary), nil

	case *parser.DestinationAllotment:
		monetaryAmount := big.Int(monetary.Amount)
		var items []parser.AllotmentValue
		for _, i := range destination.Items {
			items = append(items, i.Allotment)
		}

		allot, err := s.makeAllotment(monetaryAmount.Int64(), items)
		if err != nil {
			return nil, err
		}

		receivedTotal := big.NewInt(0)
		for i, allotmentItem := range destination.Items {
			switch allotmentItem := allotmentItem.To.(type) {
			case *parser.DestinationTo:
				receivedMon := monetary
				receivedMon.Amount = NewMonetaryInt(allot[i])
				received, err := s.receiveFrom(allotmentItem.Destination, receivedMon)
				if err != nil {
					return nil, err
				}
				receivedTotal.Add(receivedTotal, received)

			case *parser.DestinationKept:
				mon := big.Int(monetary.Amount)
				s.Receivers = append(s.Receivers, Receiver{
					Name:     "<kept>",
					Monetary: &mon,
					Asset:    string(monetary.Asset),
				})
				// TODO Should I add this line?
				// receivedTotal.Add(receivedTotal, (*big.Int)(&monetary.Amount))
			}

		}

		return receivedTotal, nil

	case *parser.DestinationInorder:
		receivedTotal := big.NewInt(0)

		// TODO make this prettier
		handler := func(keptOrDest parser.KeptOrDestination, capLit parser.Literal) error {
			var amountToReceive *big.Int
			if capLit == nil {
				amountToReceive = (*big.Int)(&monetary.Amount)
			} else {
				cap, err := evaluateLitExpecting(s, capLit, expectMonetary)
				if err != nil {
					return err
				}

				amountToReceive = utils.MinBigInt((*big.Int)(&cap.Amount), (*big.Int)(&monetary.Amount))
			}

			switch destinationTarget := keptOrDest.(type) {
			case *parser.DestinationKept:
				s.Receivers = append(s.Receivers, Receiver{
					Name:     "<kept>",
					Monetary: amountToReceive,
					Asset:    string(monetary.Asset),
				})
				receivedTotal.Add(receivedTotal, amountToReceive)
				return nil

			case *parser.DestinationTo:
				var remainingAmount big.Int
				remainingAmount.Sub(amountToReceive, receivedTotal)
				remainingMonetary := Monetary{
					Amount: MonetaryInt(remainingAmount),
					Asset:  monetary.Asset,
				}
				// receivedTotal += destination.receive(monetary-receivedTotal, ctx)
				received, err := s.receiveFrom(destinationTarget.Destination, remainingMonetary)
				if err != nil {
					return err
				}
				receivedTotal.Add(receivedTotal, received)
				return nil

			default:
				utils.NonExhaustiveMatchPanic[any](destinationTarget)
				return nil
			}
		}

		for _, destinationClause := range destination.Clauses {
			handler(destinationClause.To, destinationClause.Cap)
			// TODO should I break if all the amount has been received?
		}
		handler(destination.Remaining, nil)
		return receivedTotal, nil

	default:
		utils.NonExhaustiveMatchPanic[any](destination)
		return nil, nil
	}
}

func (s *programState) makeAllotment(monetary int64, items []parser.AllotmentValue) ([]int64, error) {
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
			rat, err := evaluateLitExpecting(s, allotment, expectPortion)
			if err != nil {
				return nil, err
			}

			totalAllotment.Add(totalAllotment, rat)
			allotments = append(allotments, *rat)

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

	return parts, nil
}

// Builtins
func meta(
	s *programState,
	args []Value,
) (string, error) {
	p := NewArgsParser(args)
	account := parseArg(p, expectAccount)
	key := parseArg(p, expectString)
	err := p.parse()
	if err != nil {
		return "", err
	}

	// body
	accountMeta := s.Meta[*account]
	value, ok := accountMeta[*key]

	if !ok {
		// TODO err
		panic("META NOT FOUND")
	}

	return value, nil
}

func balance(
	s *programState,
	args []Value,
) (*Monetary, error) {
	p := NewArgsParser(args)
	account := parseArg(p, expectAccount)
	asset := parseArg(p, expectAsset)
	err := p.parse()
	if err != nil {
		return nil, err
	}

	// body
	balance := s.getBalance(*account, *asset)
	var balanceCopy big.Int
	balanceCopy.Set(balance)

	m := Monetary{
		Asset:  Asset(*asset),
		Amount: MonetaryInt(balanceCopy),
	}
	return &m, nil
}

func setTxMeta(st *programState, args []Value) error {
	p := NewArgsParser(args)
	key := parseArg(p, expectString)
	meta := parseArg(p, expectAnything)
	err := p.parse()
	if err != nil {
		return err
	}

	st.TxMeta[*key] = *meta
	return nil
}
