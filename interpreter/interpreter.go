package interpreter

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/formancehq/numscript/analysis"
	"github.com/formancehq/numscript/parser"
	"github.com/formancehq/numscript/utils"
)

type StaticStore map[string]map[string]*big.Int
type Metadata map[string]string

type ExecutionResult struct {
	Postings     []Posting           `json:"postings"`
	TxMeta       map[string]Value    `json:"txMeta"`
	AccountsMeta map[string]Metadata `json:"accountsMeta"`
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
		return Monetary{}, fmt.Errorf("invalid monetary literal: %s", source)
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
	case analysis.FnVarOriginMeta:
		rawValue, err := meta(s, args)
		if err != nil {
			return nil, err
		}

		parsed, err := parseVar(type_, rawValue)
		if err != nil {
			return nil, err
		}

		return parsed, nil

	case analysis.FnVarOriginBalance:
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
		Postings:     postings,
		TxMeta:       st.TxMeta,
		AccountsMeta: st.Meta, // TODO clone the map
	}
	return res, nil
}

type programState struct {
	// Asset of the send statement currently being executed.
	//
	// it's value is undefined outside of send statements execution
	CurrentAsset string

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
		case analysis.FnSetAccountMeta:
			err := setAccountMeta(st, args)
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
	postings, err := Reconcile(st.CurrentAsset, st.Senders, st.Receivers)
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
		st.CurrentAsset = *asset
		sentAmt, err := st.trySendingAll(statement.Source)
		if err != nil {
			return nil, err
		}
		err = st.receiveAllFrom(statement.Destination, *sentAmt)
		if err != nil {
			return nil, err
		}
		return st.getPostings()

	case *parser.SentValueLiteral:
		monetary, err := evaluateLitExpecting(st, sentValue.Monetary, expectMonetary)
		if err != nil {
			return nil, err
		}
		st.CurrentAsset = string(monetary.Asset)

		monetaryAmt := big.Int(monetary.Amount)
		if monetaryAmt.Cmp(big.NewInt(0)) == -1 {
			return nil, NegativeAmountErr{Amount: monetary.Amount}
		}

		sentTotal, err := st.trySending(statement.Source, monetaryAmt)
		if err != nil {
			return nil, err
		}

		// sentTotal < monetary.Amount
		if sentTotal.Cmp((*big.Int)(&monetary.Amount)) == -1 {
			var missing big.Int
			missing.Sub((*big.Int)(&monetary.Amount), sentTotal)
			return nil, MissingFundsErr{
				Asset:   string(monetary.Asset),
				Missing: missing,
				Sent:    *sentTotal,
			}
		}

		// TODO simplify pointers
		amt := big.Int(monetary.Amount)
		_, err = st.receiveFrom(statement.Destination, &amt)
		if err != nil {
			return nil, err
		}

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

func (s *programState) trySendingAccount(name string, amount big.Int) (*big.Int, error) {
	var monetaryAmount big.Int
	monetaryAmount.Set(&amount)

	if name != "world" {
		balance := s.getBalance(name, s.CurrentAsset)

		// monetary = min(balance, monetary)
		if balance.Cmp(&monetaryAmount) == -1 /* balance < monetary */ {
			monetaryAmount.Set(balance)
		}
	}

	s.Senders = append(s.Senders, Sender{
		Name:     name,
		Monetary: &monetaryAmount,
	})

	return &monetaryAmount, nil
}

func (s *programState) trySendingAllFromAccount(name string) (*big.Int, error) {
	if name == "world" {
		return nil, InvalidUnboundedInSendAll{
			Name: name,
		}
	}

	var balanceClone big.Int

	// TODO err empty balance?

	balance := s.getBalance(name, s.CurrentAsset)
	s.Senders = append(s.Senders, Sender{
		Name:     name,
		Monetary: balanceClone.Set(balance),
	})

	return &balanceClone, nil
}

func (s *programState) trySendingAll(source parser.Source) (*big.Int, error) {
	switch source := source.(type) {
	case *parser.SourceAccount:
		account, err := evaluateLitExpecting(s, source.Literal, expectAccount)
		if err != nil {
			return nil, err
		}
		return s.trySendingAllFromAccount(string(*account))

	case *parser.SourceInorder:
		totalSent := big.NewInt(0)
		for _, subSource := range source.Sources {
			sent, err := s.trySendingAll(subSource)
			if err != nil {
				return nil, err
			}
			totalSent.Add(totalSent, sent)
		}
		return totalSent, nil

	case *parser.SourceCapped:
		monetary, err := evaluateLitExpecting(s, source.Cap, expectMonetaryOfAsset(s.CurrentAsset))
		if err != nil {
			return nil, err
		}

		// We switch to the default sending evaluation for this subsource
		return s.trySending(source.From, *monetary)

	case *parser.SourceAllotment:
		return nil, InvalidAllotmentInSendAll{}

	case *parser.SourceOverdraft:
		account, err := evaluateLitExpecting(s, source.Address, expectAccount)
		if err != nil {
			return nil, err
		}

		if source.Bounded == nil {
			return nil, InvalidUnboundedInSendAll{
				Name: *account,
			}
		}

		amount, err := evaluateLitExpecting(s, *source.Bounded, expectMonetaryOfAsset(s.CurrentAsset))
		if err != nil {
			return nil, err
		}

		return s.trySendingAccount(*account, *amount)

	default:
		utils.NonExhaustiveMatchPanic[error](source)
		return nil, nil
	}
}

func (s *programState) receiveAllFrom(destination parser.Destination, monetary big.Int) error {
	switch destination := destination.(type) {
	case *parser.DestinationAccount:
		account, err := evaluateLitExpecting(s, destination.Literal, expectAccount)
		if err != nil {
			return err
		}
		s.Receivers = append(s.Receivers, Receiver{
			Name:     *account,
			Monetary: nil,
		})
		return nil

	case *parser.DestinationInorder:
		for _, subDestination := range destination.Clauses {
			// TODO check asset
			cap, err := evaluateLitExpecting(s, subDestination.Cap, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return err
			}

			switch to := subDestination.To.(type) {
			case *parser.DestinationKept:
				s.Receivers = append(s.Receivers, Receiver{
					Name:     "<kept>",
					Monetary: cap,
				})

			case *parser.DestinationTo:
				_, err = s.receiveFrom(to.Destination, cap)
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
			err := s.receiveAllFrom(remaining.Destination, monetary)
			if err != nil {
				return err
			}
			return nil

		default:
			utils.NonExhaustiveMatchPanic[any](remaining)
			return nil
		}
	case *parser.DestinationAllotment:
		_, err := s.receiveFrom(destination, &monetary)
		return err

	default:
		return utils.NonExhaustiveMatchPanic[error](destination)
	}

}

func (s *programState) trySending(source parser.Source, amount big.Int) (*big.Int, error) {
	switch source := source.(type) {
	case *parser.SourceAccount:
		account, err := evaluateLitExpecting(s, source.Literal, expectAccount)
		if err != nil {
			return nil, err
		}
		return s.trySendingAccount(string(*account), amount)

	case *parser.SourceOverdraft:
		name, err := evaluateLitExpecting(s, source.Address, expectAccount)
		if err != nil {
			return nil, err
		}

		s.Senders = append(s.Senders, Sender{
			Name:     *name,
			Monetary: &amount,
		})

		return &amount, nil

	case *parser.SourceInorder:
		sentTotal := big.NewInt(0)
		for _, source := range source.Sources {
			var sendingMonetary big.Int
			sendingMonetary.Sub(&amount, sentTotal)
			sentAmt, err := s.trySending(source, sendingMonetary)
			if err != nil {
				return nil, err
			}
			sentTotal.Add(sentTotal, sentAmt)
		}
		return sentTotal, nil

	case *parser.SourceAllotment:
		receivedTotal := big.NewInt(0)
		var items []parser.AllotmentValue
		for _, i := range source.Items {
			items = append(items, i.Allotment)
		}
		allot, err := s.makeAllotment(amount.Int64(), items)
		if err != nil {
			return nil, err
		}
		for i, allotmentItem := range source.Items {
			source := allotmentItem.From
			received, err := s.trySending(source, *big.NewInt(allot[i]))
			if err != nil {
				return nil, err
			}
			receivedTotal.Add(receivedTotal, received)
		}
		return receivedTotal, nil

	case *parser.SourceCapped:
		cap, err := evaluateLitExpecting(s, source.Cap, expectMonetaryOfAsset(s.CurrentAsset))
		if err != nil {
			return nil, err
		}

		// TODO use utils.min
		var cappedAmount big.Int
		if amount.Cmp(cap) == -1 /* monetary < cap */ {
			cappedAmount.Set(&amount)
		} else {
			cappedAmount.Set(cap)
		}
		return s.trySending(source.From, cappedAmount)

	default:
		utils.NonExhaustiveMatchPanic[any](source)
		return nil, nil

	}

}

func (s *programState) receiveFrom(destination parser.Destination, amount *big.Int) (*big.Int, error) {
	switch destination := destination.(type) {
	case *parser.DestinationAccount:
		account, err := evaluateLitExpecting(s, destination.Literal, expectAccount)
		if err != nil {
			return nil, err
		}
		s.Receivers = append(s.Receivers, Receiver{
			Name:     *account,
			Monetary: amount,
		})
		return amount, nil

	case *parser.DestinationAllotment:

		var items []parser.AllotmentValue
		for _, i := range destination.Items {
			items = append(items, i.Allotment)
		}

		allot, err := s.makeAllotment(amount.Int64(), items)
		if err != nil {
			return nil, err
		}

		receivedTotal := big.NewInt(0)
		for i, allotmentItem := range destination.Items {
			amtToReceive := big.NewInt(allot[i])

			switch allotmentItem := allotmentItem.To.(type) {
			case *parser.DestinationTo:
				received, err := s.receiveFrom(allotmentItem.Destination, amtToReceive)
				if err != nil {
					return nil, err
				}
				receivedTotal.Add(receivedTotal, received)

			case *parser.DestinationKept:
				s.Receivers = append(s.Receivers, Receiver{
					Name:     "<kept>",
					Monetary: amtToReceive,
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
			var amountToReceive big.Int
			if capLit == nil {
				amountToReceive.Set(amount)
			} else {
				cap, err := evaluateLitExpecting(s, capLit, expectMonetaryOfAsset(s.CurrentAsset))
				if err != nil {
					return err
				}
				amountToReceive.Set(utils.MinBigInt(cap, amount))
			}

			switch destinationTarget := keptOrDest.(type) {
			case *parser.DestinationKept:
				s.Receivers = append(s.Receivers, Receiver{
					Name:     "<kept>",
					Monetary: &amountToReceive,
				})
				receivedTotal.Add(receivedTotal, &amountToReceive)
				return nil

			case *parser.DestinationTo:
				var remainingAmount big.Int
				remainingAmount.Sub(&amountToReceive, receivedTotal)

				if remainingAmount.Cmp(big.NewInt(0)) != 0 {
					// receivedTotal += destination.receive(monetary-receivedTotal, ctx)
					received, err := s.receiveFrom(destinationTarget.Destination, &remainingAmount)
					if err != nil {
						return err
					}
					receivedTotal.Add(receivedTotal, received)
				}

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
		return "", fmt.Errorf("account '@%s' doesn't have metadata associated to the '%s' key", *account, *key)
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
	if balance.Cmp(big.NewInt(0)) == -1 {
		return nil, NegativeBalanceError{
			Account: *account,
			Amount:  *balance,
		}
	}

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

func setAccountMeta(st *programState, args []Value) error {
	p := NewArgsParser(args)
	account := parseArg(p, expectAccount)
	key := parseArg(p, expectString)
	meta := parseArg(p, expectAnything)
	err := p.parse()
	if err != nil {
		return err
	}

	accountMeta := defaultMapGet(st.Meta, *account, func() Metadata {
		return make(Metadata)
	})

	accountMeta[*key] = (*meta).String()

	return nil
}

func defaultMapGet[T any](m map[string]T, key string, getDefault func() T) T {
	lookup, ok := m[key]
	if !ok {
		default_ := getDefault()
		m[key] = default_
		return default_
	}
	return lookup
}
