package interpreter

import (
	"context"
	"math/big"
	"strconv"
	"strings"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

type VariablesMap map[string]string

// For each account, list of the needed assets
type BalanceQuery map[string][]string

// For each account, list of the needed keys
type MetadataQuery map[string][]string

type AccountBalance = map[string]*big.Int
type Balances map[string]AccountBalance

type AccountMetadata = map[string]string
type AccountsMetadata map[string]AccountMetadata

type Store interface {
	GetBalances(context.Context, BalanceQuery) (Balances, error)
	GetAccountsMetadata(context.Context, MetadataQuery) (AccountsMetadata, error)
}

type StaticStore struct {
	Balances Balances
	Meta     AccountsMetadata
}

func (s StaticStore) GetBalances(context.Context, BalanceQuery) (Balances, error) {
	if s.Balances == nil {
		s.Balances = Balances{}
	}
	return s.Balances, nil
}
func (s StaticStore) GetAccountsMetadata(context.Context, MetadataQuery) (AccountsMetadata, error) {
	if s.Meta == nil {
		s.Meta = AccountsMetadata{}
	}
	return s.Meta, nil
}

type InterpreterError interface {
	error
	parser.Ranged
}

type Metadata = map[string]Value

type ExecutionResult struct {
	Postings []Posting `json:"postings"`

	Metadata Metadata `json:"txMeta"`

	AccountsMetadata AccountsMetadata `json:"accountsMeta"`
}

func parsePercentage(p string) big.Rat {
	num, den, err := parser.ParsePercentageRatio(p)
	if err != nil {
		panic(err)
	}
	return *big.NewRat(int64(num), int64(den))
}

func parseMonetary(source string) (Monetary, InterpreterError) {
	parts := strings.Split(source, " ")
	if len(parts) != 2 {
		return Monetary{}, InvalidMonetaryLiteral{Source: source}
	}

	asset := parts[0]

	// TODO check original numscript impl
	rawAmount := parts[1]
	parsedAmount, err := strconv.ParseInt(rawAmount, 0, 64)
	if err != nil {
		return Monetary{}, InvalidMonetaryLiteral{Source: source}
	}
	mon := Monetary{
		Asset:  Asset(asset),
		Amount: NewMonetaryInt(parsedAmount),
	}
	return mon, nil
}

func parseVar(type_ string, rawValue string, r parser.Range) (Value, InterpreterError) {
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
			return nil, InvalidNumberLiteral{Source: rawValue}
		}
		return NewMonetaryInt(i), nil
	case analysis.TypeString:
		return String(rawValue), nil
	default:
		return nil, InvalidTypeErr{Name: type_, Range: r}
	}

}

func (s *programState) handleOrigin(type_ string, fnCall parser.FnCall) (Value, InterpreterError) {
	args, err := s.evaluateLiterals(fnCall.Args)
	if err != nil {
		return nil, err
	}

	switch fnCall.Caller.Name {
	case analysis.FnVarOriginMeta:
		rawValue, err := meta(s, fnCall.Range, args)
		if err != nil {
			return nil, err
		}

		parsed, err := parseVar(type_, rawValue, fnCall.Range)
		if err != nil {
			return nil, err
		}

		return parsed, nil

	case analysis.FnVarOriginBalance:
		monetary, err := balance(s, fnCall.Range, args)
		if err != nil {
			return nil, err
		}
		return *monetary, nil

	default:
		return nil, UnboundFunctionErr{Name: fnCall.Caller.Name}
	}

}

func (s *programState) parseVars(varDeclrs []parser.VarDeclaration, rawVars map[string]string) InterpreterError {
	for _, varsDecl := range varDeclrs {
		if varsDecl.Origin == nil {
			raw, ok := rawVars[varsDecl.Name.Name]
			if !ok {
				return MissingVariableErr{Name: varsDecl.Name.Name}
			}

			parsed, err := parseVar(varsDecl.Type.Name, raw, varsDecl.Type.Range)
			if err != nil {
				return err
			}
			s.ParsedVars[varsDecl.Name.Name] = parsed
		} else {
			value, err := s.handleOrigin(varsDecl.Type.Name, *varsDecl.Origin)
			if err != nil {
				return err
			}
			s.ParsedVars[varsDecl.Name.Name] = value
		}
	}
	return nil
}

func RunProgram(
	ctx context.Context,
	program parser.Program,
	vars map[string]string,
	store Store,
) (*ExecutionResult, InterpreterError) {
	st := programState{
		ParsedVars:         make(map[string]Value),
		TxMeta:             make(map[string]Value),
		CachedAccountsMeta: AccountsMetadata{},
		CachedBalances:     Balances{},
		SetAccountsMeta:    AccountsMetadata{},
		Store:              store,

		CurrentBalanceQuery: BalanceQuery{},
		ctx:                 ctx,
	}

	err := st.parseVars(program.Vars, vars)
	if err != nil {
		return nil, err
	}

	// preload balances before executing the script
	for _, statement := range program.Statements {
		err := st.findBalancesQueriesInStatement(statement)
		if err != nil {
			return nil, err
		}
	}

	genericErr := st.runBalancesQuery()
	if genericErr != nil {
		return nil, QueryBalanceError{WrappedError: genericErr}
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
		Postings:         postings,
		Metadata:         st.TxMeta,
		AccountsMetadata: st.SetAccountsMeta,
	}
	return res, nil
}

type programState struct {
	ctx context.Context

	// Asset of the send statement currently being executed.
	//
	// it's value is undefined outside of send statements execution
	CurrentAsset string

	ParsedVars map[string]Value
	TxMeta     map[string]Value
	Senders    []Sender
	Receivers  []Receiver

	Store Store

	SetAccountsMeta AccountsMetadata

	CachedAccountsMeta AccountsMetadata
	CachedBalances     Balances

	CurrentBalanceQuery BalanceQuery
}

func (st *programState) runStatement(statement parser.Statement) ([]Posting, InterpreterError) {
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
			err := setTxMeta(st, statement.Caller.Range, args)
			if err != nil {
				return nil, err
			}
		case analysis.FnSetAccountMeta:
			err := setAccountMeta(st, statement.Caller.Range, args)
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

func (st *programState) getPostings() ([]Posting, InterpreterError) {
	postings, err := Reconcile(st.CurrentAsset, st.Senders, st.Receivers)
	if err != nil {
		return nil, err
	}

	for _, posting := range postings {
		srcBalance := st.getCachedBalance(posting.Source, posting.Asset)
		srcBalance.Sub(srcBalance, posting.Amount)

		destBalance := st.getCachedBalance(posting.Destination, posting.Asset)
		destBalance.Add(destBalance, posting.Amount)
	}
	return postings, nil
}

func (st *programState) runSendStatement(statement parser.SendStatement) ([]Posting, InterpreterError) {
	switch sentValue := statement.SentValue.(type) {
	case *parser.SentValueAll:
		asset, err := evaluateLitExpecting(st, sentValue.Asset, expectAsset)
		if err != nil {
			return nil, err
		}
		st.CurrentAsset = *asset
		sentAmt, err := st.sendAll(statement.Source)
		if err != nil {
			return nil, err
		}
		err = st.receiveFrom(statement.Destination, sentAmt)
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

		err = st.trySendingExact(statement.Source, monetaryAmt)
		if err != nil {
			return nil, err
		}

		// TODO simplify pointers
		amt := big.Int(monetary.Amount)
		err = st.receiveFrom(statement.Destination, &amt)
		if err != nil {
			return nil, err
		}

		return st.getPostings()
	default:
		utils.NonExhaustiveMatchPanic[any](sentValue)
		return nil, nil
	}

}

func (s *programState) getCachedBalance(account string, asset string) *big.Int {
	balance := defaultMapGet(s.CachedBalances, account, func() AccountBalance {
		return AccountBalance{}
	})
	assetBalance := defaultMapGet(balance, asset, func() *big.Int {
		return big.NewInt(0)
	})
	return assetBalance
}

func (s *programState) sendAllToAccount(accountLiteral parser.Literal, ovedraft *big.Int) (*big.Int, InterpreterError) {
	account, err := evaluateLitExpecting(s, accountLiteral, expectAccount)
	if err != nil {
		return nil, err
	}

	if *account == "world" || ovedraft == nil {
		return nil, InvalidUnboundedInSendAll{
			Name: *account,
		}
	}

	balance := s.getCachedBalance(*account, s.CurrentAsset)

	// we sent balance+overdraft
	sentAmt := new(big.Int).Add(balance, ovedraft)
	s.Senders = append(s.Senders, Sender{
		Name:     *account,
		Monetary: sentAmt,
	})
	return sentAmt, nil
}

// Send as much as possible (and return the sent amt)
func (s *programState) sendAll(source parser.Source) (*big.Int, InterpreterError) {
	switch source := source.(type) {
	case *parser.SourceAccount:
		return s.sendAllToAccount(source.Literal, big.NewInt(0))

	case *parser.SourceOverdraft:
		var cap *big.Int
		if source.Bounded != nil {
			bounded, err := evaluateLitExpecting(s, *source.Bounded, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return nil, err
			}
			cap = bounded
		}
		return s.sendAllToAccount(source.Address, cap)

	case *parser.SourceInorder:
		totalSent := big.NewInt(0)
		for _, subSource := range source.Sources {
			sent, err := s.sendAll(subSource)
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
		return s.trySendingUpTo(source.From, *monetary)

	case *parser.SourceAllotment:
		return nil, InvalidAllotmentInSendAll{}

	default:
		utils.NonExhaustiveMatchPanic[error](source)
		return nil, nil
	}
}

// Fails if it doesn't manage to send exactly "amount"
func (s *programState) trySendingExact(source parser.Source, amount big.Int) InterpreterError {
	sentAmt, err := s.trySendingUpTo(source, amount)
	if err != nil {
		return err
	}
	if sentAmt.Cmp(&amount) != 0 {
		return MissingFundsErr{
			Asset:     s.CurrentAsset,
			Needed:    amount,
			Available: *sentAmt,
			Range:     source.GetRange(),
		}
	}
	return nil
}

func (s *programState) trySendingToAccount(accountLiteral parser.Literal, amount big.Int, overdraft *big.Int) (*big.Int, InterpreterError) {
	account, err := evaluateLitExpecting(s, accountLiteral, expectAccount)
	if err != nil {
		return nil, err
	}
	if *account == "world" {
		overdraft = nil
	}

	var actuallySentAmt big.Int
	if overdraft == nil {
		// unbounded overdraft: we send the required amount
		actuallySentAmt.Set(&amount)
	} else {
		balance := s.getCachedBalance(*account, s.CurrentAsset)

		// that's the amount we are allowed to send (balance + overdraft)
		var safeSendAmt big.Int
		safeSendAmt.Add(balance, overdraft)

		actuallySentAmt = *utils.MinBigInt(&safeSendAmt, &amount)
	}

	s.Senders = append(s.Senders, Sender{
		Name:     *account,
		Monetary: &actuallySentAmt,
	})
	return &actuallySentAmt, nil
}

// Tries sending "amount" and returns the actually sent amt.
// Doesn't fail (unless nested sources fail)
func (s *programState) trySendingUpTo(source parser.Source, amount big.Int) (*big.Int, InterpreterError) {
	switch source := source.(type) {
	case *parser.SourceAccount:
		return s.trySendingToAccount(source.Literal, amount, big.NewInt(0))

	case *parser.SourceOverdraft:
		var cap *big.Int
		if source.Bounded != nil {
			upTo, err := evaluateLitExpecting(s, *source.Bounded, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return nil, err
			}
			cap = upTo
		}
		return s.trySendingToAccount(source.Address, amount, cap)

	case *parser.SourceInorder:
		var totalLeft big.Int
		totalLeft.Set(&amount)
		for _, source := range source.Sources {
			sentAmt, err := s.trySendingUpTo(source, totalLeft)
			if err != nil {
				return nil, err
			}
			totalLeft.Sub(&totalLeft, sentAmt)
		}

		var sentAmt big.Int
		sentAmt.Sub(&amount, &totalLeft)
		return &sentAmt, nil

	case *parser.SourceAllotment:
		var items []parser.AllotmentValue
		for _, i := range source.Items {
			items = append(items, i.Allotment)
		}
		allot, err := s.makeAllotment(amount.Int64(), items)
		if err != nil {
			return nil, err
		}
		for i, allotmentItem := range source.Items {
			err := s.trySendingExact(allotmentItem.From, *big.NewInt(allot[i]))
			if err != nil {
				return nil, err
			}
		}
		return &amount, nil

	case *parser.SourceCapped:
		cap, err := evaluateLitExpecting(s, source.Cap, expectMonetaryOfAsset(s.CurrentAsset))
		if err != nil {
			return nil, err
		}
		cappedAmount := utils.MinBigInt(&amount, cap)
		return s.trySendingUpTo(source.From, *cappedAmount)

	default:
		utils.NonExhaustiveMatchPanic[any](source)
		return nil, nil

	}

}

func (s *programState) receiveFrom(destination parser.Destination, amount *big.Int) InterpreterError {
	switch destination := destination.(type) {
	case *parser.DestinationAccount:
		account, err := evaluateLitExpecting(s, destination.Literal, expectAccount)
		if err != nil {
			return err
		}
		s.Receivers = append(s.Receivers, Receiver{
			Name:     *account,
			Monetary: amount,
		})
		return nil

	case *parser.DestinationAllotment:
		var items []parser.AllotmentValue
		for _, i := range destination.Items {
			items = append(items, i.Allotment)
		}

		allot, err := s.makeAllotment(amount.Int64(), items)
		if err != nil {
			return err
		}

		receivedTotal := big.NewInt(0)
		for i, allotmentItem := range destination.Items {
			amtToReceive := big.NewInt(allot[i])
			err := s.receiveFromKeptOrDest(allotmentItem.To, amtToReceive)
			if err != nil {
				return err
			}

			receivedTotal.Add(receivedTotal, amtToReceive)
		}
		return nil

	case *parser.DestinationInorder:
		var remainingAmount big.Int
		remainingAmount.Set(amount)

		handler := func(keptOrDest parser.KeptOrDestination, amountToReceive big.Int) InterpreterError {
			err := s.receiveFromKeptOrDest(keptOrDest, &amountToReceive)
			if err != nil {
				return err
			}
			remainingAmount.Sub(&remainingAmount, &amountToReceive)
			return err
		}

		for _, destinationClause := range destination.Clauses {

			cap, err := evaluateLitExpecting(s, destinationClause.Cap, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return err
			}

			// If the remaining amt is zero, let's ignore the posting
			if remainingAmount.Cmp(big.NewInt(0)) == 0 {
				break
			}

			err = handler(destinationClause.To, *utils.MinBigInt(cap, &remainingAmount))
			if err != nil {
				return err
			}

		}

		var cp big.Int // if remainingAmount bad things with pointers happen.. somehow
		cp.Set(&remainingAmount)
		return handler(destination.Remaining, cp)

	default:
		utils.NonExhaustiveMatchPanic[any](destination)
		return nil
	}
}

func (s *programState) receiveFromKeptOrDest(keptOrDest parser.KeptOrDestination, amount *big.Int) InterpreterError {
	switch destinationTarget := keptOrDest.(type) {
	case *parser.DestinationKept:
		s.Receivers = append(s.Receivers, Receiver{
			Name:     "<kept>",
			Monetary: amount,
		})
		return nil

	case *parser.DestinationTo:
		return s.receiveFrom(destinationTarget.Destination, amount)

	default:
		utils.NonExhaustiveMatchPanic[any](destinationTarget)
		return nil
	}

}

func (s *programState) makeAllotment(monetary int64, items []parser.AllotmentValue) ([]int64, InterpreterError) {
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
	} else if totalAllotment.Cmp(big.NewRat(1, 1)) != 0 {
		return nil, InvalidAllotmentSum{ActualSum: *totalAllotment}
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
	rng parser.Range,
	args []Value,
) (string, InterpreterError) {
	// TODO more precise location
	p := NewArgsParser(args)
	account := parseArg(p, rng, expectAccount)
	key := parseArg(p, rng, expectString)
	err := p.parse()
	if err != nil {
		return "", err
	}

	meta, fetchMetaErr := s.Store.GetAccountsMetadata(s.ctx, MetadataQuery{
		*account: []string{*key},
	})
	if fetchMetaErr != nil {
		return "", QueryMetadataError{WrappedError: fetchMetaErr}
	}
	s.CachedAccountsMeta = meta

	// body
	accountMeta := s.CachedAccountsMeta[*account]
	value, ok := accountMeta[*key]

	if !ok {
		return "", MetadataNotFound{Account: *account, Key: *key, Range: rng}
	}

	return value, nil
}

func balance(
	s *programState,
	r parser.Range,
	args []Value,
) (*Monetary, InterpreterError) {
	// TODO more precise args range location
	p := NewArgsParser(args)
	account := parseArg(p, r, expectAccount)
	asset := parseArg(p, r, expectAsset)
	err := p.parse()
	if err != nil {
		return nil, err
	}

	// body
	s.batchQuery(*account, *asset)
	fetchBalanceErr := s.runBalancesQuery()
	if fetchBalanceErr != nil {
		return nil, QueryBalanceError{WrappedError: fetchBalanceErr}
	}

	balance := s.getCachedBalance(*account, *asset)
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

func setTxMeta(st *programState, r parser.Range, args []Value) InterpreterError {
	p := NewArgsParser(args)
	key := parseArg(p, r, expectString)
	meta := parseArg(p, r, expectAnything)
	err := p.parse()
	if err != nil {
		return err
	}

	st.TxMeta[*key] = *meta
	return nil
}

func setAccountMeta(st *programState, r parser.Range, args []Value) InterpreterError {
	p := NewArgsParser(args)
	account := parseArg(p, r, expectAccount)
	key := parseArg(p, r, expectString)
	meta := parseArg(p, r, expectAnything)
	err := p.parse()
	if err != nil {
		return err
	}

	accountMeta := defaultMapGet(st.SetAccountsMeta, *account, func() AccountMetadata {
		return AccountMetadata{}
	})

	accountMeta[*key] = (*meta).String()

	return nil
}
