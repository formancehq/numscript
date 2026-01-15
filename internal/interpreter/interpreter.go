package interpreter

import (
	"context"
	"math/big"
	"regexp"
	"strings"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/flags"
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

func (s StaticStore) GetBalances(_ context.Context, q BalanceQuery) (Balances, error) {
	if s.Balances == nil {
		s.Balances = Balances{}
	}

	outputBalance := Balances{}
	for queriedAccount, queriedCurrencies := range q {
		outputAccountBalance := AccountBalance{}
		outputBalance[queriedAccount] = outputAccountBalance

		accountBalanceLookup := utils.MapGetOrPutDefault(s.Balances, queriedAccount, func() AccountBalance {
			return AccountBalance{}
		})

		for _, curr := range queriedCurrencies {
			baseAsset, isCatchAll := strings.CutSuffix(curr, "/*")
			if isCatchAll {

				for k, v := range accountBalanceLookup {
					matchesAsset := k == baseAsset || strings.HasPrefix(k, baseAsset+"/")
					if !matchesAsset {
						continue
					}
					outputAccountBalance[k] = new(big.Int).Set(v)
				}

			} else {
				n := new(big.Int)
				outputAccountBalance[curr] = n

				if i, ok := accountBalanceLookup[curr]; ok {
					n.Set(i)
				}
			}

		}
	}

	return outputBalance, nil
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

type Posting struct {
	Source      string   `json:"source"`
	Destination string   `json:"destination"`
	Amount      *big.Int `json:"amount"`
	Asset       string   `json:"asset"`
}

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

	parsedAsset, err := NewAsset(asset)
	if err != nil {
		return Monetary{}, err
	}
	mon := Monetary{
		Asset:  parsedAsset,
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
		return NewAccountAddress(rawValue)
	case analysis.TypePortion:
		bi, err := ParsePortionSpecific(rawValue)
		if err != nil {
			return nil, err
		}

		return Portion(*bi), nil
	case analysis.TypeAsset:
		return NewAsset(rawValue)
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

func (s *programState) handleFnOrigin(type_ string, expr parser.ValueExpr) (Value, InterpreterError) {
	// Special case for top-level meta() call
	if fnCall, ok := expr.(*parser.FnCall); ok && fnCall.Caller.Name == analysis.FnVarOriginMeta {
		return s.handleFnCall(&type_, *fnCall)
	}

	if _, isFnCall := expr.(*parser.FnCall); !isFnCall {
		err := s.checkFeatureFlag(flags.ExperimentalMidScriptFunctionCall)
		if err != nil {
			return nil, err
		}
	}

	return s.evaluateExpr(expr)
}

func (s *programState) handleFnCall(type_ *string, fnCall parser.FnCall) (Value, InterpreterError) {
	args, err := s.evaluateExpressions(fnCall.Args)
	if err != nil {
		return nil, err
	}

	switch fnCall.Caller.Name {
	case analysis.FnVarOriginMeta:
		if type_ == nil {
			return nil, InvalidNestedMeta{}
		}

		rawValue, err := meta(s, fnCall.Range, args)
		if err != nil {
			return nil, err
		}
		return parseVar(*type_, rawValue, fnCall.Range)

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

	case analysis.FnVarOriginGetAsset:
		return getAsset(s, fnCall.Range, args)
	case analysis.FnVarOriginGetAmount:
		return getAmount(s, fnCall.Range, args)

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
			value, err := s.handleFnOrigin(varsDecl.Type.Name, *varsDecl.Origin)
			if err != nil {
				return err
			}
			s.ParsedVars[varsDecl.Name.Name] = value
		}
	}
	return nil
}

const accountSegmentRegex = "[a-zA-Z0-9_-]+"

var accountNameRegex = regexp.MustCompile("^" + accountSegmentRegex + "(:" + accountSegmentRegex + ")*$")

// https://github.com/formancehq/ledger/blob/main/pkg/accounts/accounts.go
func checkAccountName(addr string) bool {
	return accountNameRegex.Match([]byte(addr))
}

var assetNameRegexp = regexp.MustCompile(`^[A-Z][A-Z0-9]{0,16}(_[A-Z]{1,16})?(\/\d{1,6})?$`)

// https://github.com/formancehq/ledger/blob/main/pkg/assets/asset.go
func checkAssetName(v string) bool {
	return assetNameRegexp.Match([]byte(v))
}

// Check the following invariants:
//   - no negative postings
//   - no invalid account names
//   - no invalid asset names
func checkPostingInvariants(posting Posting) InterpreterError {
	isAmtNegative := posting.Amount.Cmp(big.NewInt(0)) == -1

	isInvalidPosting := (isAmtNegative ||
		!checkAssetName(posting.Asset) ||
		!checkAccountName(posting.Source) ||
		!checkAccountName(posting.Destination))

	if isInvalidPosting {
		return InternalError{Posting: posting}
	}

	return nil
}

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
		Postings:           make([]Posting, 0),
		fundsStack:         newFundsStack(nil),

		CurrentBalanceQuery: BalanceQuery{},
		ctx:                 ctx,
		FeatureFlags:        featureFlags,
	}

	st.varOriginPosition = true
	if program.Vars != nil {
		err := st.parseVars(program.Vars.Declarations, vars)
		if err != nil {
			return nil, err
		}
	}
	st.varOriginPosition = false

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

	for _, statement := range program.Statements {
		err := st.runStatement(statement)
		if err != nil {
			return nil, err
		}
	}

	for _, posting := range st.Postings {
		err := checkPostingInvariants(posting)
		if err != nil {
			return nil, err
		}
	}

	res := &ExecutionResult{
		Postings:         st.Postings,
		Metadata:         st.TxMeta,
		AccountsMetadata: st.SetAccountsMeta,
	}
	return res, nil
}

type programState struct {
	ctx context.Context

	varOriginPosition bool

	// Asset of the send statement currently being executed.
	//
	// its value is undefined outside of send statements execution
	CurrentAsset string

	ParsedVars map[string]Value
	TxMeta     map[string]Value
	Postings   []Posting
	fundsStack fundsStack

	Store Store

	SetAccountsMeta AccountsMetadata

	CachedAccountsMeta AccountsMetadata
	CachedBalances     Balances

	CurrentBalanceQuery BalanceQuery

	FeatureFlags map[string]struct{}
}

func (st *programState) pushSender(name string, monetary *big.Int, color string) {
	if monetary.Cmp(big.NewInt(0)) == 0 {
		return
	}

	balance := st.CachedBalances.fetchBalance(name, st.CurrentAsset, color)
	balance.Sub(balance, monetary)

	st.fundsStack.Push(Sender{Name: name, Amount: monetary, Color: color})
}

func (st *programState) pushReceiver(name string, monetary *big.Int) {
	if monetary.Cmp(big.NewInt(0)) == 0 {
		return
	}

	senders := st.fundsStack.PullAnything(monetary)

	for _, sender := range senders {
		postings := Posting{
			Source:      sender.Name,
			Destination: name,
			Asset:       coloredAsset(st.CurrentAsset, &sender.Color),
			Amount:      sender.Amount,
		}

		if name == KEPT_ADDR {
			// If funds are kept, give them back to senders
			srcBalance := st.CachedBalances.fetchBalance(postings.Source, st.CurrentAsset, sender.Color)
			srcBalance.Add(srcBalance, postings.Amount)

			continue
		}

		destBalance := st.CachedBalances.fetchBalance(postings.Destination, st.CurrentAsset, sender.Color)
		destBalance.Add(destBalance, postings.Amount)

		st.Postings = append(st.Postings, postings)
	}
}

func (st *programState) runStatement(statement parser.Statement) InterpreterError {
	switch statement := statement.(type) {
	case *parser.FnCall:
		args, err := st.evaluateExpressions(statement.Args)
		if err != nil {
			return err
		}

		switch statement.Caller.Name {
		case analysis.FnSetTxMeta:
			return setTxMeta(st, statement.Caller.Range, args)
		case analysis.FnSetAccountMeta:
			return setAccountMeta(st, statement.Caller.Range, args)
		default:
			return UnboundFunctionErr{Name: statement.Caller.Name}
		}

	case *parser.SendStatement:
		return st.runSendStatement(*statement)

	case *parser.SaveStatement:
		return st.runSaveStatement(*statement)

	default:
		utils.NonExhaustiveMatchPanic[any](statement)
		return nil
	}
}

func (st *programState) runSaveStatement(saveStatement parser.SaveStatement) InterpreterError {
	asset, amt, err := st.evaluateSentAmt(saveStatement.SentValue)
	if err != nil {
		return err
	}

	account, err := evaluateExprAs(st, saveStatement.Amount, expectAccount)
	if err != nil {
		return err
	}

	balance := st.CachedBalances.fetchBalance(*account, *asset, "")

	if amt == nil {
		if balance.Sign() > 0 {
			balance.Set(big.NewInt(0))
		}
	} else {
		// Do not allow negative saves
		if amt.Cmp(big.NewInt(0)) == -1 {
			return NegativeAmountErr{
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

	return nil
}

func (st *programState) runSendStatement(statement parser.SendStatement) InterpreterError {
	switch sentValue := statement.SentValue.(type) {
	case *parser.SentValueAll:
		asset, err := evaluateExprAs(st, sentValue.Asset, expectAsset)
		if err != nil {
			return err
		}
		st.CurrentAsset = *asset
		sentAmt, err := st.sendAll(statement.Source)
		if err != nil {
			return err
		}
		return st.receiveFrom(statement.Destination, sentAmt)

	case *parser.SentValueLiteral:
		monetary, err := evaluateExprAs(st, sentValue.Monetary, expectMonetary)
		if err != nil {
			return err
		}
		st.CurrentAsset = string(monetary.Asset)

		monetaryAmt := (*big.Int)(&monetary.Amount)
		if monetaryAmt.Cmp(big.NewInt(0)) == -1 {
			return NegativeAmountErr{Amount: monetary.Amount}
		}

		err = st.trySendingExact(statement.Source, monetaryAmt)
		if err != nil {
			return err
		}

		amt := big.Int(monetary.Amount)
		return st.receiveFrom(statement.Destination, &amt)
	default:
		utils.NonExhaustiveMatchPanic[any](sentValue)
		return nil
	}

}

// PRE: overdraft >= 0
func (s *programState) sendAllToAccount(accountLiteral parser.ValueExpr, overdraft *big.Int, colorExpr parser.ValueExpr) (*big.Int, InterpreterError) {
	if colorExpr != nil {
		err := s.checkFeatureFlag(flags.ExperimentalAssetColors)
		if err != nil {
			return nil, err
		}
	}

	account, err := evaluateExprAs(s, accountLiteral, expectAccount)
	if err != nil {
		return nil, err
	}

	if *account == "world" || overdraft == nil {
		return nil, InvalidUnboundedInSendAll{
			Name: *account,
		}
	}

	color, err := s.evaluateColor(colorExpr)
	if err != nil {
		return nil, err
	}

	balance := s.CachedBalances.fetchBalance(*account, s.CurrentAsset, *color)

	// we sent balance+overdraft
	sentAmt := CalculateMaxSafeWithdraw(balance, overdraft)

	s.pushSender(*account, sentAmt, *color)
	return sentAmt, nil
}

// Send as much as possible (and return the sent amt)
func (s *programState) sendAll(source parser.Source) (*big.Int, InterpreterError) {
	switch source := source.(type) {
	case *parser.SourceAccount:
		return s.sendAllToAccount(source.ValueExpr, big.NewInt(0), source.Color)

	case *parser.SourceOverdraft:
		var cap *big.Int
		if source.Bounded != nil {
			bounded, err := evaluateExprAs(s, *source.Bounded, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return nil, err
			}
			cap = utils.NonNeg(bounded)
		}
		return s.sendAllToAccount(source.Address, cap, source.Color)

	case *parser.SourceWithScaling:
		err := s.checkFeatureFlag(flags.AssetScaling)
		if err != nil {
			return nil, err
		}

		account, err := evaluateExprAs(s, source.Address, expectAccount)
		if err != nil {
			return nil, err
		}

		scalingAccount, err := evaluateExprAs(s, source.Through, expectAccount)
		if err != nil {
			return nil, err
		}

		baseAsset, assetScale := getAssetScale(s.CurrentAsset)
		acc, ok := s.CachedBalances[*account]
		if !ok {
			return nil, InvalidUnboundedAddressInScalingAddress{Range: source.Range}
		}

		sol, totSent := findSolution(
			nil,
			assetScale,
			getAssets(acc, baseAsset),
		)

		for _, convAmt := range sol {
			scale := convAmt.scale
			convAmt := convAmt.amount

			// here we manually emit postings based on the known solution,
			// and update balances accordingly
			asset := buildScaledAsset(baseAsset, scale)
			s.Postings = append(s.Postings, Posting{
				Source:      *account,
				Destination: *scalingAccount,
				Amount:      new(big.Int).Set(convAmt),
				Asset:       asset,
			})
			acc[asset].Sub(acc[asset], convAmt)
		}

		s.Postings = append(s.Postings, Posting{
			Source:      *scalingAccount,
			Destination: *account,
			Amount:      new(big.Int).Set(totSent),
			Asset:       s.CurrentAsset,
		})
		accBalance := utils.MapGetOrPutDefault(acc, s.CurrentAsset, func() *big.Int {
			return big.NewInt(0)
		})
		accBalance.Add(accBalance, totSent)

		return s.trySendingToAccount(source.Address, totSent, big.NewInt(0), source.Color)

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
		err := s.checkFeatureFlag(flags.ExperimentalOneofFeatureFlag)
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
		// We switch to the default sending evaluation for this subsource
		return s.trySendingUpTo(source.From, utils.NonNeg(monetary))

	case *parser.SourceAllotment:
		return nil, InvalidAllotmentInSendAll{}

	default:
		_ = utils.NonExhaustiveMatchPanic[error](source)
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

var colorRe = regexp.MustCompile("^[A-Z]*$")

// PRE: overdraft >= 0
func (s *programState) trySendingToAccount(accountLiteral parser.ValueExpr, amount *big.Int, overdraft *big.Int, colorExpr parser.ValueExpr) (*big.Int, InterpreterError) {
	if colorExpr != nil {
		err := s.checkFeatureFlag(flags.ExperimentalAssetColors)
		if err != nil {
			return nil, err
		}
	}

	account, err := evaluateExprAs(s, accountLiteral, expectAccount)
	if err != nil {
		return nil, err
	}
	if *account == "world" {
		overdraft = nil
	}

	color, err := s.evaluateColor(colorExpr)
	if err != nil {
		return nil, err
	}

	var actuallySentAmt *big.Int
	if overdraft == nil {
		// unbounded overdraft: we send the required amount
		actuallySentAmt = new(big.Int).Set(amount)
	} else {
		balance := s.CachedBalances.fetchBalance(*account, s.CurrentAsset, *color)

		// that's the amount we are allowed to send (balance + overdraft)
		actuallySentAmt = CalculateSafeWithdraw(balance, overdraft, amount)
	}
	s.pushSender(*account, actuallySentAmt, *color)
	return actuallySentAmt, nil
}

func (s *programState) cloneState() func() {
	fsBackup := s.fundsStack.Clone()
	balancesBackup := s.CachedBalances.DeepClone()

	return func() {
		s.fundsStack = fsBackup
		s.CachedBalances = balancesBackup
	}
}

// Tries sending "amount" and returns the actually sent amt.
// Doesn't fail (unless nested sources fail)
func (s *programState) trySendingUpTo(source parser.Source, amount *big.Int) (*big.Int, InterpreterError) {
	switch source := source.(type) {
	case *parser.SourceAccount:
		return s.trySendingToAccount(source.ValueExpr, amount, big.NewInt(0), source.Color)

	case *parser.SourceWithScaling:
		err := s.checkFeatureFlag(flags.AssetScaling)
		if err != nil {
			return nil, err
		}

		account, err := evaluateExprAs(s, source.Address, expectAccount)
		if err != nil {
			return nil, err
		}
		scalingAccount, err := evaluateExprAs(s, source.Through, expectAccount)
		if err != nil {
			return nil, err
		}

		baseAsset, assetScale := getAssetScale(s.CurrentAsset)
		acc, ok := s.CachedBalances[*account]
		if !ok {
			return nil, InvalidUnboundedAddressInScalingAddress{Range: source.Range}
		}

		sol, total := findSolution(
			amount,
			assetScale,
			getAssets(acc, baseAsset),
		)

		if sol == nil || amount.Cmp(total) == 1 {
			// we already know we are failing, but we're delegating to the "standard" (non-scaled) mode
			// so that we get a somewhat helpful (although limited) error message
			return s.trySendingToAccount(source.Address, amount, big.NewInt(0), source.Color)
		}

		for _, pair := range sol {
			scale := pair.scale
			sending := pair.amount
			// here we manually emit postings based on the known solution,
			// and update balances accordingly
			asset := buildScaledAsset(baseAsset, scale)
			s.Postings = append(s.Postings, Posting{
				Source:      *account,
				Destination: *scalingAccount,
				Amount:      new(big.Int).Set(sending),
				Asset:       asset,
			})
			acc[asset].Sub(acc[asset], sending)
		}

		s.Postings = append(s.Postings, Posting{
			Source:      *scalingAccount,
			Destination: *account,
			Amount:      new(big.Int).Set(amount),
			Asset:       s.CurrentAsset,
		})

		accBalance := utils.MapGetOrPutDefault(acc, s.CurrentAsset, func() *big.Int {
			return big.NewInt(0)
		})
		accBalance.Add(accBalance, amount)

		return s.trySendingToAccount(source.Address, amount, big.NewInt(0), source.Color)

	case *parser.SourceOverdraft:
		var cap *big.Int
		if source.Bounded != nil {
			upTo, err := evaluateExprAs(s, *source.Bounded, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return nil, err
			}
			cap = utils.NonNeg(upTo)
		}
		return s.trySendingToAccount(source.Address, amount, cap, source.Color)

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
		err := s.checkFeatureFlag(flags.ExperimentalOneofFeatureFlag)
		if err != nil {
			return nil, err
		}

		// empty oneof is parsing err
		leadingSources := source.Sources[0 : len(source.Sources)-1]

		for _, source := range leadingSources {
			// do not move this line below (as .trySendingUpTo() will mutate the fundsStack)
			undo := s.cloneState()

			sentAmt, err := s.trySendingUpTo(source, amount)
			if err != nil {
				return nil, err
			}

			// if this branch managed to sent all the required amount, return now
			if sentAmt.Cmp(amount) == 0 {
				return amount, nil
			}

			// else, backtrack to remove this branch's sendings
			undo()
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
		return s.trySendingUpTo(source.From, utils.NonNeg(
			utils.MinBigInt(amount, cap),
		))

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

			amountToReceive := utils.MaxBigInt(utils.MinBigInt(cap, remainingAmount), big.NewInt(0))
			err = handler(destinationClause.To, amountToReceive)
			if err != nil {
				return err
			}

		}

		remainingAmountCopy := new(big.Int).Set(remainingAmount)
		// passing "remainingAmount" directly breaks the code
		return handler(destination.Remaining, remainingAmountCopy)

	case *parser.DestinationOneof:
		err := s.checkFeatureFlag(flags.ExperimentalOneofFeatureFlag)
		if err != nil {
			return err
		}
		for _, destinationClause := range destination.Clauses {
			cap, err := evaluateExprAs(s, destinationClause.Cap, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return err
			}

			// if the clause cap is >= the amount we're trying to receive, only go through this branch
			switch cap.Cmp(amount) {
			case 0, 1:
				return s.receiveFromKeptOrDest(destinationClause.To, amount)
			}

			// otherwise try next clause (keep looping)
		}
		return s.receiveFromKeptOrDest(destination.Remaining, amount)

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
		case *parser.ValueExprAllotment:
			rat, err := evaluateExprAs(s, allotment.Value, expectPortion)
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

// Utility function to get the balance
func getBalance(
	s *programState,
	account string,
	asset string,
) (*big.Int, InterpreterError) {
	s.batchQuery(account, asset, nil)
	fetchBalanceErr := s.runBalancesQuery()
	if fetchBalanceErr != nil {
		return nil, QueryBalanceError{WrappedError: fetchBalanceErr}
	}
	balance := s.CachedBalances.fetchBalance(account, asset, "")
	return balance, nil

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

/*
PRE: ovedraft != nil, balance != nil
PRE: ovedraft >= 0
POST: $out >= 0
*/
func CalculateMaxSafeWithdraw(balance *big.Int, overdraft *big.Int) *big.Int {
	return utils.NonNeg(
		new(big.Int).Add(balance, overdraft),
	)
}

/*
PRE: ovedraft != nil, balance != nil
PRE: ovedraft >= 0
PRE: requestedAmount >= 0
POST: $out >= 0
*/
func CalculateSafeWithdraw(
	balance *big.Int,
	overdraft *big.Int,
	requestedAmount *big.Int,
) *big.Int {
	safe := CalculateMaxSafeWithdraw(balance, overdraft)
	return utils.MinBigInt(safe, requestedAmount)
}

func PrettyPrintPostings(postings []Posting) string {
	var rows [][]string
	for _, posting := range postings {
		row := []string{posting.Source, posting.Destination, posting.Asset, posting.Amount.String()}
		rows = append(rows, row)
	}
	return utils.CsvPretty([]string{"Source", "Destination", "Asset", "Amount"}, rows, false)
}

func PrettyPrintMeta(meta Metadata) string {
	m := map[string]string{}
	for k, v := range meta {
		m[k] = v.String()
	}

	return utils.CsvPrettyMap("Name", "Value", m)
}
