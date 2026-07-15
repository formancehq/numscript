package interpreter

import (
	"context"
	"errors"
	"math/big"

	"github.com/formancehq/numscript/internal/analysis"
	"github.com/formancehq/numscript/internal/parser"
)

var ErrScalingNotSupported = errors.New("scaling is not supported")

type AccountDependency struct {
	Account string
	Scope   string
	Color   string
	Asset   string
}

type MetaDependency struct {
	Account string
	Scope   string
	Key     string
}

type ResolvedDependencies struct {
	AccountsReads  map[AccountDependency]struct{}
	AccountsWrites map[AccountDependency]struct{}
	MetaReads      map[MetaDependency]struct{}
	MetaWrites     map[MetaDependency]struct{}
	TxMetaWrites   map[string]struct{}
}

// recordingStore wraps a Store and records every balance/metadata it is queried
// for. Evaluating an expression through an evalEnv backed by this store is what
// makes balance()/overdraft()/meta() reads show up as dependencies.
type recordingStore struct {
	inner     Store
	reads     map[AccountDependency]struct{}
	metaReads map[MetaDependency]struct{}
}

func (s *recordingStore) GetBalances(ctx context.Context, q BalanceQuery) (Balances, error) {
	for _, item := range q {
		s.reads[AccountDependency{
			Account: item.Account,
			Scope:   item.Scope,
			Color:   item.Color,
			Asset:   item.Asset,
		}] = struct{}{}
	}
	return s.inner.GetBalances(ctx, q)
}

func (s *recordingStore) GetAccountsMetadata(ctx context.Context, q MetadataQuery) (AccountsMetadata, error) {
	for _, item := range q {
		for _, key := range item.Keys {
			s.metaReads[MetaDependency{Account: item.Account, Scope: item.Scope, Key: key}] = struct{}{}
		}
	}
	return s.inner.GetAccountsMetadata(ctx, q)
}

type resolveDependenciesState struct {
	env  *evalEnv
	deps ResolvedDependencies
}

func ResolveDependencies(ctx context.Context, store Store, vars map[string]string, program parser.Program) (ResolvedDependencies, error) {
	deps := ResolvedDependencies{
		AccountsReads:  map[AccountDependency]struct{}{},
		AccountsWrites: map[AccountDependency]struct{}{},
		MetaReads:      map[MetaDependency]struct{}{},
		MetaWrites:     map[MetaDependency]struct{}{},
		TxMetaWrites:   map[string]struct{}{},
	}
	recording := &recordingStore{inner: store, reads: deps.AccountsReads, metaReads: deps.MetaReads}

	// binding the vars evaluates their origins, so balance()/overdraft()/meta()
	// origins already get recorded through the store here.
	//
	// dep resolution only needs the read to be recorded, not its value (a
	// balance() yields a Monetary, which can't name an account), so getBalance
	// just hits the recording store and returns zero — no funds engine involved.
	getBalance := func(account AccountAddress, asset Asset) (*big.Int, InterpreterError) {
		_, err := recording.GetBalances(ctx, BalanceQuery{
			{Account: account.Name, Asset: string(asset), Color: "", Scope: account.Scope},
		})
		if err != nil {
			return nil, QueryBalanceError{WrappedError: err}
		}
		return new(big.Int), nil
	}
	env, err := newEvalEnv(
		nil,
		getBalance,
		newMetadataGetter(ctx, recording, InternalAccountsMetadata{}),
		program.Vars, vars,
	)
	if err != nil {
		return ResolvedDependencies{}, err
	}

	st := &resolveDependenciesState{env: &env, deps: deps}
	for _, statement := range program.Statements {
		if err := st.resolveStatement(statement); err != nil {
			return ResolvedDependencies{}, err
		}
	}

	return st.deps, nil
}

func addAccountDep(deps map[AccountDependency]struct{}, account AccountAddress, asset Asset, color String) {
	deps[AccountDependency{
		Account: account.Name,
		Scope:   account.Scope,
		Color:   string(color),
		Asset:   string(asset),
	}] = struct{}{}
}

func addMetaDep(deps map[MetaDependency]struct{}, account AccountAddress, key string) {
	deps[MetaDependency{
		Account: account.Name,
		Scope:   account.Scope,
		Key:     key,
	}] = struct{}{}
}

// eval evaluates an expression only for its side effects: any balance()/meta()
// inside it is recorded through the store.
func (st *resolveDependenciesState) eval(expr parser.ValueExpr) error {
	_, err := evaluateExpr(st.env, expr)
	return err
}

func (st *resolveDependenciesState) resolveStatement(statement parser.Statement) error {
	switch statement := statement.(type) {
	case *parser.SendStatement:
		asset, _, err := evaluateSentAmt(st.env, statement.SentValue)
		if err != nil {
			return err
		}
		// funds keep their source color all the way to the destination, so a
		// destination can be credited in any of the send's source colors
		srcColors := map[String]struct{}{}
		if err := st.resolveSource(statement.Source, asset, srcColors); err != nil {
			return err
		}
		return st.resolveDestination(statement.Destination, asset, srcColors)

	case *parser.SaveStatement:
		asset, _, err := evaluateSentAmt(st.env, statement.SentValue)
		if err != nil {
			return err
		}
		account, err := evaluateExprAs(st.env, statement.Account, expectAccount)
		if err != nil {
			return err
		}
		addAccountDep(st.deps.AccountsReads, account, asset, "")
		return nil

	case *parser.FnCall:
		return st.resolveFnCallStatement(statement)

	default:
		return unhandledErr(statement)
	}
}

func (st *resolveDependenciesState) resolveFnCallStatement(fnCall *parser.FnCall) error {
	args, err := evaluateExpressions(st.env, fnCall.Args)
	if err != nil {
		return err
	}

	switch fnCall.Caller.Name {
	case analysis.FnSetAccountMeta:
		p := NewArgsParser(args)
		account := parseArg(p, fnCall.Range, expectAccount)
		key := parseArg(p, fnCall.Range, expectString)
		_ = parseArg(p, fnCall.Range, expectAnything)
		if err := p.parse(); err != nil {
			return err
		}
		addMetaDep(st.deps.MetaWrites, account, string(key))
		return nil

	case analysis.FnSetTxMeta:
		p := NewArgsParser(args)
		key := parseArg(p, fnCall.Range, expectString)
		_ = parseArg(p, fnCall.Range, expectAnything)
		if err := p.parse(); err != nil {
			return err
		}
		st.deps.TxMetaWrites[string(key)] = struct{}{}
		return nil

	default:
		return UnboundFunctionErr{Name: fnCall.Caller.Name}
	}
}

func (st *resolveDependenciesState) resolveSource(source parser.Source, asset Asset, srcColors map[String]struct{}) error {
	switch source := source.(type) {
	case *parser.SourceAccount:
		account, err := evaluateExprAs(st.env, source.ValueExpr, expectAccount)
		if err != nil {
			return err
		}
		color, err := evaluateOptExprAs(st.env, source.Color, expectString)
		if err != nil {
			return err
		}
		srcColors[color] = struct{}{}
		st.recordSource(account, asset, color, account.Name == "world")
		return nil

	case *parser.SourceOverdraft:
		account, err := evaluateExprAs(st.env, source.Address, expectAccount)
		if err != nil {
			return err
		}
		color, err := evaluateOptExprAs(st.env, source.Color, expectString)
		if err != nil {
			return err
		}
		srcColors[color] = struct{}{}
		st.recordSource(account, asset, color, source.Bounded == nil)
		if source.Bounded != nil {
			return st.eval(*source.Bounded)
		}
		return nil

	case *parser.SourceInorder:
		for _, sub := range source.Sources {
			if err := st.resolveSource(sub, asset, srcColors); err != nil {
				return err
			}
		}
		return nil

	case *parser.SourceOneof:
		for _, sub := range source.Sources {
			if err := st.resolveSource(sub, asset, srcColors); err != nil {
				return err
			}
		}
		return nil

	case *parser.SourceCapped:
		if err := st.eval(source.Cap); err != nil {
			return err
		}
		return st.resolveSource(source.From, asset, srcColors)

	case *parser.SourceAllotment:
		for _, item := range source.Items {
			if err := st.resolveAllotment(item.Allotment); err != nil {
				return err
			}
			if err := st.resolveSource(item.From, asset, srcColors); err != nil {
				return err
			}
		}
		return nil

	case *parser.SourceWithScaling:
		return ErrScalingNotSupported

	default:
		return unhandledErr(source)
	}
}

// recordSource records a source account: always a write, and a read unless the
// account is unbounded (its balance is never consulted).
func (st *resolveDependenciesState) recordSource(account AccountAddress, asset Asset, color String, unbounded bool) {
	if !unbounded {
		addAccountDep(st.deps.AccountsReads, account, asset, color)
	}
	addAccountDep(st.deps.AccountsWrites, account, asset, color)
}

func (st *resolveDependenciesState) resolveDestination(destination parser.Destination, asset Asset, srcColors map[String]struct{}) error {
	switch destination := destination.(type) {
	case *parser.DestinationAccount:
		account, err := evaluateExprAs(st.env, destination.ValueExpr, expectAccount)
		if err != nil {
			return err
		}
		// the destination can be credited in any of the source colors
		for color := range srcColors {
			addAccountDep(st.deps.AccountsWrites, account, asset, color)
		}
		return nil

	case *parser.DestinationInorder:
		for _, clause := range destination.Clauses {
			if err := st.eval(clause.Cap); err != nil {
				return err
			}
			if err := st.resolveKeptOrDestination(clause.To, asset, srcColors); err != nil {
				return err
			}
		}
		return st.resolveKeptOrDestination(destination.Remaining, asset, srcColors)

	case *parser.DestinationOneof:
		for _, clause := range destination.Clauses {
			if err := st.eval(clause.Cap); err != nil {
				return err
			}
			if err := st.resolveKeptOrDestination(clause.To, asset, srcColors); err != nil {
				return err
			}
		}
		return st.resolveKeptOrDestination(destination.Remaining, asset, srcColors)

	case *parser.DestinationAllotment:
		for _, item := range destination.Items {
			if err := st.resolveAllotment(item.Allotment); err != nil {
				return err
			}
			if err := st.resolveKeptOrDestination(item.To, asset, srcColors); err != nil {
				return err
			}
		}
		return nil

	default:
		return unhandledErr(destination)
	}
}

func (st *resolveDependenciesState) resolveKeptOrDestination(kd parser.KeptOrDestination, asset Asset, srcColors map[String]struct{}) error {
	switch kd := kd.(type) {
	case *parser.DestinationKept:
		return nil
	case *parser.DestinationTo:
		return st.resolveDestination(kd.Destination, asset, srcColors)
	default:
		return unhandledErr(kd)
	}
}

func (st *resolveDependenciesState) resolveAllotment(allotment parser.AllotmentValue) error {
	switch allotment := allotment.(type) {
	case *parser.ValueExprAllotment:
		return st.eval(allotment.Value)
	case *parser.RemainingAllotment:
		return nil
	default:
		return unhandledErr(allotment)
	}
}
