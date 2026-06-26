package interpreter

import (
	"context"
	"maps"
	"math/big"
	"regexp"
	"slices"
	"strings"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/runtime"
	"github.com/formancehq/numscript/internal/utils"
)

// zeroStore backs the runtime.RunState's lazy balance fallback. The interpreter
// pre-fetches every needed balance in one batched query (runBalancesQuery) and
// Prewarms it into the runtime, and treats any un-fetched (account, asset,
// color) as zero — exactly the semantics this store provides.
type zeroStore struct{}

func (zeroStore) GetBalance(account, asset, color string) *big.Int { return new(big.Int) }

type VariablesMap map[string]string

type InterpreterError interface {
	error
	parser.Ranged
}

type Metadata = map[string]Value

// Posting is an alias for runtime.Posting, which owns the definition (and the
// json serialization contract). Kept as an alias so the public name
// interpreter.Posting / numscript.Posting is preserved.
type Posting = runtime.Posting

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
		return monetary, nil

	case analysis.FnVarOriginOverdraft:
		monetary, err := overdraft(s, fnCall.Range, args)
		if err != nil {
			return nil, err
		}
		return monetary, nil

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
		SetAccountsMeta:    AccountsMetadata{},
		Store:              store,
		rs:                 runtime.New(zeroStore{}),

		CurrentBalanceQuery: BalanceQuery{},
		ctx:                 ctx,
		FeatureFlags:        maps.Clone(featureFlags),
	}

	if st.FeatureFlags == nil {
		st.FeatureFlags = make(map[string]struct{}, len(program.Flags))
	}

	for _, flag := range program.Flags {
		index := slices.Index(flags.AllFlags, flag.String)
		if index == -1 {
			return nil, InvalidFeature{
				Feature: flag.String,
			}
		}

		st.FeatureFlags[flag.String] = struct{}{}
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

	// GetPostings returns []runtime.Posting, which is []Posting (alias).
	postings := st.rs.GetPostings()
	for _, posting := range postings {
		err := checkPostingInvariants(posting)
		if err != nil {
			return nil, err
		}
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

	varOriginPosition bool

	// Asset of the send statement currently being executed.
	//
	// its value is undefined outside of send statements execution
	CurrentAsset Asset

	ParsedVars map[string]Value
	TxMeta     map[string]Value

	// rs owns the funds state: balances (write-through cache over the batched
	// Store fetch, seeded via Prewarm), the FIFO funding-source queue, and the
	// emitted postings. Replaces the former fundsQueue + CachedBalances + Postings.
	rs *runtime.RunState

	Store Store

	SetAccountsMeta AccountsMetadata

	CachedAccountsMeta AccountsMetadata

	CurrentBalanceQuery BalanceQuery

	FeatureFlags map[string]struct{}
}

// Append a posting without checking if account has enough balance.
// Updates both source and destination balances.
// Noop if the amount is zero
func (st *programState) forcePushPostingUncolored(
	source AccountAddress,
	destination AccountAddress,
	amount MonetaryInt,
	asset Asset,
) {
	amtBi := big.Int(amount)
	st.rs.ForcePosting(string(source), string(destination), string(asset), "", &amtBi)
}

func (st *programState) pushReceiver(name string, monetary *big.Int) {
	// color == nil: drain the queue regardless of color (PullAnything), each
	// posting keeping its source fund's own color.
	if name == KEPT_ADDR {
		// kept funds are refunded to their sources, emitting no posting
		st.rs.Send(nil, monetary, nil)
		return
	}
	dest := name
	st.rs.Send(&dest, monetary, nil)
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

	account, err := evaluateExprAs(st, saveStatement.Account, expectAccount)
	if err != nil {
		return err
	}

	// Do not allow negative saves
	if amt != nil && amt.Cmp(big.NewInt(0)) == -1 {
		return NegativeAmountErr{
			Range:  saveStatement.SentValue.GetRange(),
			Amount: MonetaryInt(*amt),
		}
	}

	// amt == nil -> "save all"; otherwise reduce by amt, floored at 0
	st.rs.Save(string(account), string(asset), "", amt)
	return nil
}

func (st *programState) runSendStatement(statement parser.SendStatement) InterpreterError {
	switch sentValue := statement.SentValue.(type) {
	case *parser.SentValueAll:
		asset, err := evaluateExprAs(st, sentValue.Asset, expectAsset)
		if err != nil {
			return err
		}
		st.CurrentAsset = asset
		st.rs.SetCurrentAsset(string(asset))
		sentAmt, err := st.takeAll(statement.Source)
		if err != nil {
			return err
		}
		return st.sendTo(statement.Destination, sentAmt)

	case *parser.SentValueLiteral:
		monetary, err := evaluateExprAs(st, sentValue.Monetary, expectMonetary)
		if err != nil {
			return err
		}
		st.CurrentAsset = monetary.Asset
		st.rs.SetCurrentAsset(string(monetary.Asset))

		amtBi := big.Int(monetary.Amount)
		if amtBi.Sign() == -1 {
			return NegativeAmountErr{Amount: monetary.Amount, Range: sentValue.Monetary.GetRange()}
		}
		err = st.tryTakingExact(statement.Source, monetary.Amount)
		if err != nil {
			return err
		}

		return st.sendTo(statement.Destination, &amtBi)
	default:
		utils.NonExhaustiveMatchPanic[any](sentValue)
		return nil
	}

}

// PRE: overdraft >= 0
func (s *programState) takeAllFromAccount(accountLiteral parser.ValueExpr, overdraft *big.Int, colorExpr parser.ValueExpr) (*big.Int, InterpreterError) {
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

	if account == "world" || overdraft == nil {
		return nil, InvalidUnboundedInSendAll{
			Name: string(account),
		}
	}

	color, err := s.evaluateColor(colorExpr)
	if err != nil {
		return nil, err
	}

	// PullUncapped queues balance+overdraft (== CalculateMaxSafeWithdraw),
	// debiting the (account, currentAsset, color) balance.
	sentAmt := s.rs.PullUncapped(string(account), overdraft, string(color))
	return sentAmt, nil
}

// Pull as much as possible (and return the sent amt)
func (s *programState) takeAll(source parser.Source) (*big.Int, InterpreterError) {
	switch source := source.(type) {
	case *parser.SourceAccount:
		return s.takeAllFromAccount(source.ValueExpr, big.NewInt(0), source.Color)

	case *parser.SourceOverdraft:
		var cap *big.Int
		if source.Bounded != nil {
			bounded, err := evaluateExprAs(s, *source.Bounded, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return nil, err
			}
			boundedBi := big.Int(bounded)
			cap = utils.NonNeg(&boundedBi)
		}
		return s.takeAllFromAccount(source.Address, cap, source.Color)

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

		baseAsset, assetScale := s.CurrentAsset.GetBaseAndScale()
		acc := toAccountBalances(s.rs.AccountBalances(string(account)))
		if len(acc) == 0 {
			return nil, InvalidUnboundedAddressInScalingAddress{Range: source.Range}
		}

		sol, totSent := findScalingSolution(
			nil,
			assetScale,
			getAssets(acc, baseAsset),
		)

		for _, convAmt := range sol {
			s.forcePushPostingUncolored(
				account,
				scalingAccount,
				MonetaryInt(*new(big.Int).Set(convAmt.amount)),
				Asset(buildScaledAsset(baseAsset, convAmt.scale)),
			)
		}

		s.forcePushPostingUncolored(
			scalingAccount,
			account,
			MonetaryInt(*new(big.Int).Set(totSent)),
			s.CurrentAsset,
		)

		return s.takeAllFromAccount(source.Address, big.NewInt(0), nil)

	case *parser.SourceInorder:
		totalSent := big.NewInt(0)
		for _, subSource := range source.Sources {
			sent, err := s.takeAll(subSource)
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
		return s.takeAll(first)

	case *parser.SourceCapped:
		monetary, err := evaluateExprAs(s, source.Cap, expectMonetaryOfAsset(s.CurrentAsset))
		if err != nil {
			return nil, err
		}
		monetaryBi := big.Int(monetary)
		// We switch to the default sending evaluation for this subsource
		return s.tryTakingUpTo(source.From, utils.NonNeg(&monetaryBi))

	case *parser.SourceAllotment:
		return nil, InvalidAllotmentInSendAll{}

	default:
		_ = utils.NonExhaustiveMatchPanic[error](source)
		return nil, nil
	}
}

// Fails if it doesn't manage to pull exactly "amount"
func (s *programState) tryTakingExact(source parser.Source, amount MonetaryInt) InterpreterError {
	amtBi := (*big.Int)(&amount)
	sentAmt, err := s.tryTakingUpTo(source, amtBi)
	if err != nil {
		return err
	}
	if sentAmt.Cmp(amtBi) != 0 {
		return MissingFundsErr{
			Asset:     string(s.CurrentAsset),
			Needed:    *amtBi,
			Available: *sentAmt,
			Range:     source.GetRange(),
		}
	}
	return nil
}

var colorRe = regexp.MustCompile("^[A-Z]*$")

// PRE: overdraft >= 0
func (s *programState) tryTakingFromAccount(accountLiteral parser.ValueExpr, amount *big.Int, overdraft *big.Int, colorExpr parser.ValueExpr) (*big.Int, InterpreterError) {
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
	if account == "world" {
		overdraft = nil
	}

	color, err := s.evaluateColor(colorExpr)
	if err != nil {
		return nil, err
	}

	// Pull computes the available amount (min(max(0, balance+overdraft), amount)
	// == CalculateSafeWithdraw; unbounded for world/overdraft==nil), debits the
	// (account, currentAsset, color) balance, and queues the funds. The
	// interpreter's overdraft convention (nil == unbounded) is exactly Pull's.
	actuallySentAmt := s.rs.Pull(string(account), amount, overdraft, string(color))
	return actuallySentAmt, nil
}

// cloneState returns an undo function for speculative source evaluation (oneof).
// Backtracking is a cheap source-queue snapshot: on undo, the runtime repays the
// funds pulled since the mark and truncates the queue — no map cloning.
func (s *programState) cloneState() func() {
	mark := s.rs.Snapshot()
	return func() {
		s.rs.Restore(mark)
	}
}

// Tries pulling up to "amount" and returns the actually pulled amt.
// Doesn't fail (unless nested sources fail)
func (s *programState) tryTakingUpTo(source parser.Source, amount *big.Int) (*big.Int, InterpreterError) {
	amount = utils.NonNeg(amount)

	switch source := source.(type) {
	case *parser.SourceAccount:
		return s.tryTakingFromAccount(source.ValueExpr, amount, big.NewInt(0), source.Color)

	case *parser.SourceWithScaling:
		// Note that scaled sources are colorless (for now). That's we we don't bother including
		// colors in the logic about scaling

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

		baseAsset, assetScale := s.CurrentAsset.GetBaseAndScale()

		acc := toAccountBalances(s.rs.AccountBalances(string(account)))
		if len(acc) == 0 {
			return nil, InvalidUnboundedAddressInScalingAddress{Range: source.Range}
		}

		sol, swappedAmt := findScalingSolution(
			amount,
			assetScale,
			getAssets(acc, baseAsset),
		)

		for _, pair := range sol {
			s.forcePushPostingUncolored(
				account,
				scalingAccount,
				NewMonetaryIntBig(pair.amount),
				Asset(buildScaledAsset(baseAsset, pair.scale)),
			)
		}

		s.forcePushPostingUncolored(
			scalingAccount,
			account,
			NewMonetaryIntBig(swappedAmt),
			s.CurrentAsset,
		)

		return s.tryTakingFromAccount(source.Address, amount, big.NewInt(0), nil)

	case *parser.SourceOverdraft:
		var cap *big.Int
		if source.Bounded != nil {
			upTo, err := evaluateExprAs(s, *source.Bounded, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return nil, err
			}
			bi := big.Int(upTo)
			cap = utils.NonNeg(&bi)
		}
		return s.tryTakingFromAccount(source.Address, amount, cap, source.Color)

	case *parser.SourceInorder:
		totalLeft := new(big.Int).Set(amount)
		for _, source := range source.Sources {
			sentAmt, err := s.tryTakingUpTo(source, totalLeft)
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
			// do not move this line below (as .tryTakingUpTo() will mutate the fundsQueue)
			undo := s.cloneState()

			sentAmt, err := s.tryTakingUpTo(source, amount)
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

		return s.tryTakingUpTo(source.Sources[len(source.Sources)-1], amount)

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
			err := s.tryTakingExact(allotmentItem.From, MonetaryInt(allot[i]))
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
		capBi := big.Int(cap)
		return s.tryTakingUpTo(source.From, utils.NonNeg(
			utils.MinBigInt(amount, &capBi),
		))

	default:
		utils.NonExhaustiveMatchPanic[any](source)
		return nil, nil

	}

}

func (s *programState) sendTo(destination parser.Destination, amount *big.Int) InterpreterError {
	switch destination := destination.(type) {
	case *parser.DestinationAccount:
		account, err := evaluateExprAs(s, destination.ValueExpr, expectAccount)
		if err != nil {
			return err
		}
		s.pushReceiver(string(account), amount)
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
			amtToReceive := &allot[i]
			err := s.sendToKeptOrDest(allotmentItem.To, amtToReceive)
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

			err := s.sendToKeptOrDest(keptOrDest, amountToReceive)
			if err != nil {
				return err
			}
			remainingAmount.Sub(remainingAmount, amountToReceive)
			return err
		}

		for _, destinationClause := range destination.Clauses {
			// If the remaining amt is zero, let's ignore the posting
			if remainingAmount.Cmp(big.NewInt(0)) == 0 {
				break
			}

			cap, err := evaluateExprAs(s, destinationClause.Cap, expectMonetaryOfAsset(s.CurrentAsset))
			if err != nil {
				return err
			}

			capBi := big.Int(cap)

			amountToReceive := utils.MaxBigInt(utils.MinBigInt(&capBi, remainingAmount), big.NewInt(0))
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

			capBi := big.Int(cap)

			// if the clause cap is >= the amount we're trying to receive, only go through this branch
			switch capBi.Cmp(amount) {
			case 0, 1:
				return s.sendToKeptOrDest(destinationClause.To, amount)
			}

			// otherwise try next clause (keep looping)
		}
		return s.sendToKeptOrDest(destination.Remaining, amount)

	default:
		utils.NonExhaustiveMatchPanic[any](destination)
		return nil
	}
}

const KEPT_ADDR = "<kept>"

func (s *programState) sendToKeptOrDest(keptOrDest parser.KeptOrDestination, amount *big.Int) InterpreterError {
	switch destinationTarget := keptOrDest.(type) {
	case *parser.DestinationKept:
		s.pushReceiver(KEPT_ADDR, amount)
		return nil

	case *parser.DestinationTo:
		return s.sendTo(destinationTarget.Destination, amount)

	default:
		utils.NonExhaustiveMatchPanic[any](destinationTarget)
		return nil
	}

}

func (s *programState) makeAllotment(monetary *big.Int, items []parser.AllotmentValue) ([]big.Int, InterpreterError) {
	totalAllotment := big.NewRat(0, 1)
	allotments := make([]big.Rat, 0, len(items))

	remainingAllotmentIndex := -1

	for i, item := range items {
		switch allotment := item.(type) {
		case *parser.ValueExprAllotment:
			rat, err := evaluateExprAs(s, allotment.Value, expectPortion)
			if err != nil {
				return nil, err
			}

			totalAllotment.Add(totalAllotment, rat)
			allotments = append(allotments, *rat)

		case *parser.RemainingAllotment:
			remainingAllotmentIndex = i
			allotments = append(allotments, big.Rat{})
			// TODO check there are not duplicate remaining clause
		}
	}

	if remainingAllotmentIndex != -1 {
		allotments[remainingAllotmentIndex] = *new(big.Rat).Sub(big.NewRat(1, 1), totalAllotment)
	} else if totalAllotment.Cmp(big.NewRat(1, 1)) != 0 {
		return nil, InvalidAllotmentSum{ActualSum: *totalAllotment}
	}

	// portions are resolved (remaining computed, sum validated) — delegate the
	// floor-then-distribute split to the runtime, filling one contiguous buffer.
	parts := make([]big.Int, len(allotments))
	runtime.MakeAllotment(parts, monetary, allotments)
	return parts, nil
}

// Utility function to get the balance
func getBalance(
	s *programState,
	account AccountAddress,
	asset Asset,
) (*big.Int, InterpreterError) {
	color := String("")

	s.batchQuery(account, asset, color)
	fetchBalanceErr := s.runBalancesQuery()
	if fetchBalanceErr != nil {
		return nil, QueryBalanceError{WrappedError: fetchBalanceErr}
	}
	balance := s.rs.GetAccountBalance(string(account), string(asset), string(color))
	return balance, nil

}

func (st *programState) evaluateSentAmt(sentValue parser.SentValue) (Asset, *big.Int, InterpreterError) {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueAll:
		asset, err := evaluateExprAs(st, sentValue.Asset, expectAsset)
		if err != nil {
			return "", nil, err
		}
		return asset, nil, nil

	case *parser.SentValueLiteral:
		monetary, err := evaluateExprAs(st, sentValue.Monetary, expectMonetary)
		if err != nil {
			return "", nil, err
		}

		bi := big.Int(monetary.Amount)
		return monetary.Asset, &bi, nil

	default:
		utils.NonExhaustiveMatchPanic[any](sentValue)
		return "", nil, nil
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
	// the Color column is shown only when at least one posting has a color
	hasColor := slices.ContainsFunc(postings, func(posting Posting) bool {
		return posting.Color != ""
	})

	var header []string
	if hasColor {
		header = []string{"Source", "Destination", "Asset", "Color", "Amount"}
	} else {
		header = []string{"Source", "Destination", "Asset", "Amount"}
	}

	var rows [][]string
	for _, posting := range postings {
		var row []string
		if hasColor {
			row = []string{posting.Source, posting.Destination, posting.Asset, posting.Color, posting.Amount.String()}
		} else {
			row = []string{posting.Source, posting.Destination, posting.Asset, posting.Amount.String()}
		}
		rows = append(rows, row)
	}
	return utils.CsvPretty(header, rows, false)
}

func PrettyPrintMeta(meta Metadata) string {
	m := map[string]string{}
	for k, v := range meta {
		m[k] = v.String()
	}

	return utils.CsvPrettyMap("Name", "Value", m)
}
