package interpreter

import (
	"context"
	"math/big"
	"regexp"
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

func parseMonetary(source string) (Monetary, InterpreterError) {
	parts := strings.Split(source, " ")
	if len(parts) != 2 {
		return Monetary{}, InvalidMonetaryLiteral{Source: source}
	}

	asset := parts[0]

	rawAmount := parts[1]
	n, ok := new(big.Int).SetString(rawAmount, 10)
	if !ok {
		return Monetary{}, InvalidNumberLiteral{Source: rawAmount}
	}
	mon := Monetary{
		Asset:  Asset(asset),
		Amount: MonetaryInt(*n),
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
		bi, err := ParsePortionSpecific(rawValue)
		if err != nil {
			return nil, err
		}

		return Portion(*bi), nil
	case analysis.TypeAsset:
		return Asset(rawValue), nil
	case analysis.TypeNumber:
		n, ok := new(big.Int).SetString(rawValue, 10)
		if !ok {
			return nil, InvalidNumberLiteral{Source: rawValue}
		}
		return MonetaryInt(*n), nil
	case analysis.TypeString:
		return String(rawValue), nil
	default:
		return nil, InvalidTypeErr{Name: type_, Range: r}
	}

}

func (s *programState) handleOrigin(type_ string, fnCall parser.FnCall) (Value, InterpreterError) {
	args, err := s.evaluateExpressions(fnCall.Args)
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

	case analysis.FnVarOriginOverdraft:
		monetary, err := overdraft(s, fnCall.Range, args)
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

type FeatureFlag = string

const (
	ExperimentalOverdraftFunctionFeatureFlag FeatureFlag = "experimental-overdraft-function"
	ExperimentalOneofFeatureFlag             FeatureFlag = "experimental-oneof"
)

func RunProgram(
	ctx context.Context,
	program parser.Program,
	vars map[string]string,
	store Store,
	featureFlags map[string]struct{},
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
		FeatureFlags:        featureFlags,
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

	FeatureFlags map[string]struct{}
}

func (st *programState) pushSender(name string, monetary *big.Int) {
	if monetary.Cmp(big.NewInt(0)) == 0 {
		return
	}
	st.Senders = append(st.Senders, Sender{Name: name, Monetary: monetary})
}

func (st *programState) pushReceiver(name string, monetary *big.Int) {
	if monetary.Cmp(big.NewInt(0)) == 0 {
		return
	}
	st.Receivers = append(st.Receivers, Receiver{Name: name, Monetary: monetary})
}

func (st *programState) runStatement(statement parser.Statement) ([]Posting, InterpreterError) {
	st.Senders = nil
	st.Receivers = nil

	switch statement := statement.(type) {
	case *parser.FnCall:
		args, err := st.evaluateExpressions(statement.Args)
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

	case *parser.SaveStatement:
		return st.runSaveStatement(*statement)

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

func (st *programState) runSaveStatement(saveStatement parser.SaveStatement) ([]Posting, InterpreterError) {
	asset, amt, err := st.evaluateSentAmt(saveStatement.SentValue)
	if err != nil {
		return nil, err
	}

	account, err := evaluateExprAs(st, saveStatement.Amount, expectAccount)
	if err != nil {
		return nil, err
	}

	balance := st.getCachedBalance(*account, *asset)

	if amt == nil {
		balance.Set(big.NewInt(0))
	} else {
		// Do not allow negative saves
		if amt.Cmp(big.NewInt(0)) == -1 {
			return nil, NegativeAmountErr{
				Range:  saveStatement.SentValue.GetRange(),
				Amount: MonetaryInt(*amt),
			}
		}

		// we decrease the balance by "amt"
		balance.Sub(balance, amt)
		// without going under 0
		if balance.Cmp(big.NewInt(0)) == -1 {
			balance.Set(big.NewInt(0))
		}
	}

	return nil, nil
}

func (st *programState) runSendStatement(statement parser.SendStatement) ([]Posting, InterpreterError) {
	switch sentValue := statement.SentValue.(type) {
	case *parser.SentValueAll:
		asset, err := evaluateExprAs(st, sentValue.Asset, expectAsset)
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
		monetary, err := evaluateExprAs(st, sentValue.Monetary, expectMonetary)
		if err != nil {
			return nil, err
		}
		st.CurrentAsset = string(monetary.Asset)

		monetaryAmt := (*big.Int)(&monetary.Amount)
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

func (s *programState) sendAllToAccount(accountLiteral parser.ValueExpr, ovedraft *big.Int) (*big.Int, InterpreterError) {
	account, err := evaluateExprAs(s, accountLiteral, expectAccount)
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
	sentAmt := utils.MaxBigInt(new(big.Int).Add(balance, ovedraft), big.NewInt(0))
	s.pushSender(*account, sentAmt)
	return sentAmt, nil
}

// Send as much as possible (and return the sent amt)
func (s *programState) sendAll(source parser.Source) (*big.Int, InterpreterError) {
	switch source := source.(type) {
	case *parser.SourceAccount:
		return s.sendAllToAccount(source.ValueExpr, big.NewInt(0))

	case *parser.SourceOverdraft:
		var cap *big.Int
		if source.Bounded != nil {
			bounded, err := evaluateExprAs(s, *source.Bounded, expectMonetaryOfAsset(s.CurrentAsset))
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

	case *parser.SourceOneof:
		err := s.checkFeatureFlag(ExperimentalOneofFeatureFlag)
		if err != nil {
			return nil, err
		}

		// we can safely access the first one because empty oneof is parsing err
		first := source.Sources[0]
		return s.sendAll(first)

	case *parser.SourceCapped:
		monetary, err := evaluateExprAs(s, source.Cap, expectMonetaryOfAsset(s.CurrentAsset))
		if err != nil {
			return nil, err
		}
		if monetary.Cmp(big.NewInt(0)) == -1 {
			monetary.Set(big.NewInt(0))
		}
		// We switch to the default sending evaluation for this subsource
		return s.trySendingUpTo(source.From, monetary)

	case *parser.SourceAllotment:
		return nil, InvalidAllotmentInSendAll{}

	default:
		utils.NonExhaustiveMatchPanic[error](source)
		return nil, nil
	}
}

// Fails if it doesn't manage to send exactly "amount"
func (s *programState) trySendingExact(source parser.Source, amount *big.Int) InterpreterError {
	sentAmt, err := s.trySendingUpTo(source, amount)
	if err != nil {
		return err
	}
	if sentAmt.Cmp(amount) != 0 {
		return MissingFundsErr{
			Asset:     s.CurrentAsset,
			Needed:    *amount,
			Available: *sentAmt,
			Range:     source.GetRange(),
		}
	}
	return nil
}

func (s *programState) trySendingToAccount(accountLiteral parser.ValueExpr, amount *big.Int, overdraft *big.Int) (*big.Int, InterpreterError) {
	account, err := evaluateExprAs(s, accountLiteral, expectAccount)
	if err != nil {
		return nil, err
	}
	if *account == "world" {
		overdraft = nil
	}

	var actuallySentAmt *big.Int
	if overdraft == nil {
		// unbounded overdraft: we send the required amount
		actuallySentAmt = new(big.Int).Set(amount)
	} else {
		balance := s.getCachedBalance(*account, s.CurrentAsset)

		// that's the amount we are allowed to send (balance + overdraft)
		safeSendAmt := new(big.Int).Add(balance, overdraft)
		actuallySentAmt = utils.MinBigInt(safeSendAmt, amount)
	}

	s.pushSender(*account, actuallySentAmt)
	return actuallySentAmt, nil
}

// Tries sending "amount" and returns the actually sent amt.
// Doesn't fail (unless nested sources fail)
func (s *programState) trySendingUpTo(source parser.Source, amount *big.Int) (*big.Int, InterpreterError) {
	switch source := source.(type) {
	case *parser.SourceAccount:
		return s.trySendingToAccount(source.ValueExpr, amount, big.NewInt(0))

	case *parser.SourceOverdraft:
		var cap *big.Int
		if source.Bounded != nil {
			upTo, err := evaluateExprAs(s, *source.Bounded, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return nil, err
			}
			cap = upTo
		}
		return s.trySendingToAccount(source.Address, amount, cap)

	case *parser.SourceInorder:
		totalLeft := new(big.Int).Set(amount)
		for _, source := range source.Sources {
			sentAmt, err := s.trySendingUpTo(source, totalLeft)
			if err != nil {
				return nil, err
			}
			totalLeft.Sub(totalLeft, sentAmt)
		}
		return new(big.Int).Sub(amount, totalLeft), nil

	case *parser.SourceOneof:
		err := s.checkFeatureFlag(ExperimentalOneofFeatureFlag)
		if err != nil {
			return nil, err
		}

		// empty oneof is parsing err
		leadingSources := source.Sources[0 : len(source.Sources)-1]

		for _, source := range leadingSources {

			// do not move this line below (as .trySendingUpTo() will mutate senders' length)
			backtrackingIndex := len(s.Senders)

			sentAmt, err := s.trySendingUpTo(source, amount)
			if err != nil {
				return nil, err
			}

			// if this branch managed to sent all the required amount, return now
			if sentAmt.Cmp(amount) == 0 {
				return amount, nil
			}

			// else, backtrack to remove this branch's sendings
			s.Senders = s.Senders[0:backtrackingIndex]
		}

		return s.trySendingUpTo(source.Sources[len(source.Sources)-1], amount)

	case *parser.SourceAllotment:
		var items []parser.AllotmentValue
		for _, i := range source.Items {
			items = append(items, i.Allotment)
		}
		allot, err := s.makeAllotment(amount, items)
		if err != nil {
			return nil, err
		}
		for i, allotmentItem := range source.Items {
			err := s.trySendingExact(allotmentItem.From, allot[i])
			if err != nil {
				return nil, err
			}
		}
		return amount, nil

	case *parser.SourceCapped:
		cap, err := evaluateExprAs(s, source.Cap, expectMonetaryOfAsset(s.CurrentAsset))
		if err != nil {
			return nil, err
		}
		cappedAmount := utils.MinBigInt(amount, cap)
		if cappedAmount.Cmp(big.NewInt(0)) == -1 {
			cappedAmount.Set(big.NewInt(0))
		}
		return s.trySendingUpTo(source.From, cappedAmount)

	default:
		utils.NonExhaustiveMatchPanic[any](source)
		return nil, nil

	}

}

func (s *programState) receiveFrom(destination parser.Destination, amount *big.Int) InterpreterError {
	switch destination := destination.(type) {
	case *parser.DestinationAccount:
		account, err := evaluateExprAs(s, destination.ValueExpr, expectAccount)
		if err != nil {
			return err
		}
		s.pushReceiver(*account, amount)
		return nil

	case *parser.DestinationAllotment:
		var items []parser.AllotmentValue
		for _, i := range destination.Items {
			items = append(items, i.Allotment)
		}

		allot, err := s.makeAllotment(amount, items)
		if err != nil {
			return err
		}

		receivedTotal := big.NewInt(0)
		for i, allotmentItem := range destination.Items {
			amtToReceive := allot[i]
			err := s.receiveFromKeptOrDest(allotmentItem.To, amtToReceive)
			if err != nil {
				return err
			}

			receivedTotal.Add(receivedTotal, amtToReceive)
		}
		return nil

	case *parser.DestinationInorder:
		remainingAmount := new(big.Int).Set(amount)

		handler := func(keptOrDest parser.KeptOrDestination, amountToReceive *big.Int) InterpreterError {
			if amountToReceive.Cmp(big.NewInt(0)) == 0 {
				return nil
			}

			err := s.receiveFromKeptOrDest(keptOrDest, amountToReceive)
			if err != nil {
				return err
			}
			remainingAmount.Sub(remainingAmount, amountToReceive)
			return err
		}

		for _, destinationClause := range destination.Clauses {

			cap, err := evaluateExprAs(s, destinationClause.Cap, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return err
			}

			// If the remaining amt is zero, let's ignore the posting
			if remainingAmount.Cmp(big.NewInt(0)) == 0 {
				break
			}

			err = handler(destinationClause.To, utils.MinBigInt(cap, remainingAmount))
			if err != nil {
				return err
			}

		}

		remainingAmountCopy := new(big.Int).Set(remainingAmount)
		// passing "remainingAmount" directly breaks the code
		return handler(destination.Remaining, remainingAmountCopy)

	default:
		utils.NonExhaustiveMatchPanic[any](destination)
		return nil
	}
}

const KEPT_ADDR = "<kept>"

func (s *programState) receiveFromKeptOrDest(keptOrDest parser.KeptOrDestination, amount *big.Int) InterpreterError {
	switch destinationTarget := keptOrDest.(type) {
	case *parser.DestinationKept:
		s.pushReceiver(KEPT_ADDR, amount)
		return nil

	case *parser.DestinationTo:
		return s.receiveFrom(destinationTarget.Destination, amount)

	default:
		utils.NonExhaustiveMatchPanic[any](destinationTarget)
		return nil
	}

}

func (s *programState) makeAllotment(monetary *big.Int, items []parser.AllotmentValue) ([]*big.Int, InterpreterError) {
	totalAllotment := big.NewRat(0, 1)
	var allotments []*big.Rat

	remainingAllotmentIndex := -1

	for i, item := range items {
		switch allotment := item.(type) {
		case *parser.RatioLiteral:
			rat := allotment.ToRatio()
			totalAllotment.Add(totalAllotment, rat)
			allotments = append(allotments, rat)
		case *parser.Variable:
			rat, err := evaluateExprAs(s, allotment, expectPortion)
			if err != nil {
				return nil, err
			}

			totalAllotment.Add(totalAllotment, rat)
			allotments = append(allotments, rat)

		case *parser.RemainingAllotment:
			remainingAllotmentIndex = i
			allotments = append(allotments, new(big.Rat))
			// TODO check there are not duplicate remaining clause
		}
	}

	if remainingAllotmentIndex != -1 {
		allotments[remainingAllotmentIndex] = new(big.Rat).Sub(big.NewRat(1, 1), totalAllotment)
	} else if totalAllotment.Cmp(big.NewRat(1, 1)) != 0 {
		return nil, InvalidAllotmentSum{ActualSum: *totalAllotment}
	}

	parts := make([]*big.Int, len(allotments))

	totalAllocated := big.NewInt(0)

	for i, allot := range allotments {
		monetaryRat := new(big.Rat).SetInt(monetary)
		product := new(big.Rat).Mul(allot, monetaryRat)

		floored := new(big.Int).Div(product.Num(), product.Denom())

		parts[i] = floored
		totalAllocated.Add(totalAllocated, floored)

	}

	for i := range parts {
		if /* totalAllocated >= monetary */ totalAllocated.Cmp(monetary) != -1 {
			break
		}

		parts[i].Add(parts[i], big.NewInt(1))
		// totalAllocated++
		totalAllocated.Add(totalAllocated, big.NewInt(1))
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

// Utility function to get the balance
func getBalance(
	s *programState,
	account string,
	asset string,
) (*big.Int, InterpreterError) {
	s.batchQuery(account, asset)
	fetchBalanceErr := s.runBalancesQuery()
	if fetchBalanceErr != nil {
		return nil, QueryBalanceError{WrappedError: fetchBalanceErr}
	}
	balance := s.getCachedBalance(account, asset)
	return balance, nil

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

	balance, err := getBalance(s, *account, *asset)
	if err != nil {
		return nil, err
	}

	if balance.Cmp(big.NewInt(0)) == -1 {
		return nil, NegativeBalanceError{
			Account: *account,
			Amount:  *balance,
		}
	}

	balanceCopy := new(big.Int).Set(balance)

	m := Monetary{
		Asset:  Asset(*asset),
		Amount: MonetaryInt(*balanceCopy),
	}
	return &m, nil
}

func overdraft(
	s *programState,
	r parser.Range,
	args []Value,
) (*Monetary, InterpreterError) {
	err := s.checkFeatureFlag(ExperimentalOverdraftFunctionFeatureFlag)
	if err != nil {
		return nil, err
	}

	// TODO more precise args range location
	p := NewArgsParser(args)
	account := parseArg(p, r, expectAccount)
	asset := parseArg(p, r, expectAsset)
	err = p.parse()
	if err != nil {
		return nil, err
	}

	balance_, err := getBalance(s, *account, *asset)
	if err != nil {
		return nil, err
	}

	balanceIsPositive := balance_.Cmp(big.NewInt(0)) == 1
	if balanceIsPositive {
		return &Monetary{
			Amount: NewMonetaryInt(0),
			Asset:  Asset(*asset),
		}, nil
	}

	overdraft := new(big.Int).Neg(balance_)
	return &Monetary{
		Amount: MonetaryInt(*overdraft),
		Asset:  Asset(*asset),
	}, nil
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

func (st *programState) evaluateSentAmt(sentValue parser.SentValue) (*string, *big.Int, InterpreterError) {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueAll:
		asset, err := evaluateExprAs(st, sentValue.Asset, expectAsset)
		if err != nil {
			return nil, nil, err
		}
		return asset, nil, nil

	case *parser.SentValueLiteral:
		monetary, err := evaluateExprAs(st, sentValue.Monetary, expectMonetary)
		if err != nil {
			return nil, nil, err
		}
		s := string(monetary.Asset)
		bi := big.Int(monetary.Amount)
		return &s, &bi, nil

	default:
		utils.NonExhaustiveMatchPanic[any](sentValue)
		return nil, nil, nil
	}
}

var percentRegex = regexp.MustCompile(`^([0-9]+)(?:[.]([0-9]+))?[%]$`)
var fractionRegex = regexp.MustCompile(`^([0-9]+)\s?[/]\s?([0-9]+)$`)

// slightly edited copy-paste from:
// https://github.com/formancehq/ledger/blob/b188d0c80eadaab5024d74edc967c7005e155f7c/internal/machine/portion.go#L57

func ParsePortionSpecific(input string) (*big.Rat, InterpreterError) {
	var res *big.Rat
	var ok bool

	percentMatch := percentRegex.FindStringSubmatch(input)
	if len(percentMatch) != 0 {
		integral := percentMatch[1]
		fractional := percentMatch[2]
		res, ok = new(big.Rat).SetString(integral + "." + fractional)
		if !ok {
			return nil, BadPortionParsingErr{Reason: "invalid percent format", Source: input}
		}
		res.Mul(res, big.NewRat(1, 100))
	} else {
		fractionMatch := fractionRegex.FindStringSubmatch(input)
		if len(fractionMatch) != 0 {
			numerator := fractionMatch[1]
			denominator := fractionMatch[2]
			res, ok = new(big.Rat).SetString(numerator + "/" + denominator)
			if !ok {
				return nil, BadPortionParsingErr{Reason: "invalid fractional format", Source: input}
			}
		}
	}
	if res == nil {
		return nil, BadPortionParsingErr{Reason: "invalid format", Source: input}
	}

	if res.Cmp(big.NewRat(0, 1)) == -1 || res.Cmp(big.NewRat(1, 1)) == 1 {
		return nil, BadPortionParsingErr{Reason: "portion must be between 0% and 100% inclusive", Source: input}
	}

	return res, nil
}

func (s programState) checkFeatureFlag(flag string) InterpreterError {
	_, ok := s.FeatureFlags[flag]
	if ok {
		return nil
	} else {
		return ExperimentalFeature{FlagName: flag}
	}
}
