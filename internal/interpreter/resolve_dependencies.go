package interpreter

import (
	"context"
	"maps"
	"slices"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

// ResolvedDependencies summarizes what a script reads from and writes to the
// store. The caller can use it to preload data and to detect input drift
// between successive runs.
type ResolvedDependencies struct {
	// Reads contains the data the script read from the store while resolving.
	Reads ResolvedReads

	// Writes contains the (account, asset, color) tuples whose balance can be
	// impacted by a posting emitted by the script.
	Writes ResolvedWrites
}

// ResolvedReads holds the data read from the store while resolving the
// script's dependencies.
type ResolvedReads struct {
	// Volumes contains every (account, asset, color) → balance row read from
	// the store, in the order it was returned.
	Volumes Balances

	// Metadata contains all (account, key) → value pairs read from the store.
	Metadata AccountsMetadata
}

// ResolvedWrites holds the data the script may write to the store.
type ResolvedWrites struct {
	// Volumes lists every (account, asset, color) tuple that may be impacted
	// by a posting emitted by the script.
	Volumes BalanceQuery
}

// ResolveDependenciesOptions configures ResolveDependencies behavior.
type ResolveDependenciesOptions struct {
	// FeatureFlags enables additional experimental features
	// (same semantics as RunWithFeatureFlags).
	FeatureFlags map[string]struct{}
}

// ResolveDependencies discovers which data a script reads from the store and
// which (account, asset, color) tuples it may write to, without executing any
// posting.
//
// It performs variable resolution and source preloading — the two phases that
// RunProgram runs before executing statements — then walks the send statements
// to collect the touched accounts. No transfers are simulated, so the call is
// cheap and does not depend on the script's runtime semantics (allotments,
// overdraft, etc.).
//
// Store calls (GetBalances/GetAccountsMetadata) are issued in a deterministic
// order across runs with identical inputs, so the caller can hash them to
// detect input drift.
func ResolveDependencies(
	ctx context.Context,
	program parser.Program,
	vars map[string]string,
	store Store,
	opts ResolveDependenciesOptions,
) (*ResolvedDependencies, InterpreterError) {
	recorder := newRecordingStore(store)

	featureFlags := maps.Clone(opts.FeatureFlags)
	if featureFlags == nil {
		featureFlags = make(map[string]struct{}, len(program.Flags))
	}
	for _, flag := range program.Flags {
		if slices.Index(flags.AllFlags, flag.String) == -1 {
			return nil, InvalidFeature{Feature: flag.String}
		}
		featureFlags[flag.String] = struct{}{}
	}

	st := programState{
		ParsedVars:          make(map[string]Value),
		TxMeta:              make(map[string]Value),
		CachedAccountsMeta:  AccountsMetadata{},
		CachedBalances:      InternalBalances{},
		SetAccountsMeta:     AccountsMetadata{},
		Store:               recorder,
		Postings:            make([]Posting, 0),
		fundsQueue:          newFundsQueue(nil),
		CurrentBalanceQuery: BalanceQuery{},
		ctx:                 ctx,
		FeatureFlags:        featureFlags,
	}

	st.varOriginPosition = true
	if program.Vars != nil {
		if err := st.parseVars(program.Vars.Declarations, vars); err != nil {
			return nil, err
		}
	}
	st.varOriginPosition = false

	for _, statement := range program.Statements {
		if err := st.findBalancesQueriesInStatement(statement); err != nil {
			return nil, err
		}
	}
	if err := st.runBalancesQuery(); err != nil {
		return nil, QueryBalanceError{WrappedError: err}
	}

	writes := BalanceQuery{}
	for _, statement := range program.Statements {
		send, ok := statement.(*parser.SendStatement)
		if !ok {
			continue
		}
		if err := st.collectSendWrites(*send, &writes); err != nil {
			return nil, err
		}
	}

	return &ResolvedDependencies{
		Reads: ResolvedReads{
			Volumes:  recorder.balanceReads,
			Metadata: recorder.metadataReads,
		},
		Writes: ResolvedWrites{Volumes: writes},
	}, nil
}

func (st *programState) collectSendWrites(
	send parser.SendStatement,
	writes *BalanceQuery,
) InterpreterError {
	asset, _, err := st.evaluateSentAmt(send.SentValue)
	if err != nil {
		return err
	}
	st.CurrentAsset = asset

	if err := st.collectSourceWrites(send.Source, writes); err != nil {
		return err
	}
	return st.collectDestinationWrites(send.Destination, writes)
}

func (st *programState) collectSourceWrites(
	source parser.Source,
	writes *BalanceQuery,
) InterpreterError {
	switch source := source.(type) {
	case *parser.SourceAccount:
		return st.touchAccount(source.ValueExpr, source.Color, writes)

	case *parser.SourceOverdraft:
		return st.touchAccount(source.Address, source.Color, writes)

	case *parser.SourceWithScaling:
		return st.touchAccount(source.Address, nil, writes)

	case *parser.SourceInorder:
		for _, sub := range source.Sources {
			if err := st.collectSourceWrites(sub, writes); err != nil {
				return err
			}
		}
		return nil

	case *parser.SourceOneof:
		for _, sub := range source.Sources {
			if err := st.collectSourceWrites(sub, writes); err != nil {
				return err
			}
		}
		return nil

	case *parser.SourceCapped:
		return st.collectSourceWrites(source.From, writes)

	case *parser.SourceAllotment:
		for _, item := range source.Items {
			if err := st.collectSourceWrites(item.From, writes); err != nil {
				return err
			}
		}
		return nil

	default:
		utils.NonExhaustiveMatchPanic[any](source)
		return nil
	}
}

func (st *programState) collectDestinationWrites(
	dest parser.Destination,
	writes *BalanceQuery,
) InterpreterError {
	switch dest := dest.(type) {
	case *parser.DestinationAccount:
		return st.touchAccount(dest.ValueExpr, nil, writes)

	case *parser.DestinationInorder:
		for _, clause := range dest.Clauses {
			if err := st.collectKeptOrDestWrites(clause.To, writes); err != nil {
				return err
			}
		}
		return st.collectKeptOrDestWrites(dest.Remaining, writes)

	case *parser.DestinationOneof:
		for _, clause := range dest.Clauses {
			if err := st.collectKeptOrDestWrites(clause.To, writes); err != nil {
				return err
			}
		}
		return st.collectKeptOrDestWrites(dest.Remaining, writes)

	case *parser.DestinationAllotment:
		for _, item := range dest.Items {
			if err := st.collectKeptOrDestWrites(item.To, writes); err != nil {
				return err
			}
		}
		return nil

	default:
		utils.NonExhaustiveMatchPanic[any](dest)
		return nil
	}
}

func (st *programState) collectKeptOrDestWrites(
	k parser.KeptOrDestination,
	writes *BalanceQuery,
) InterpreterError {
	switch k := k.(type) {
	case *parser.DestinationKept:
		return nil
	case *parser.DestinationTo:
		return st.collectDestinationWrites(k.Destination, writes)
	default:
		utils.NonExhaustiveMatchPanic[any](k)
		return nil
	}
}

func (st *programState) touchAccount(
	accountExpr parser.ValueExpr,
	colorExpr parser.ValueExpr,
	writes *BalanceQuery,
) InterpreterError {
	account, err := evaluateExprAs(st, accountExpr, expectAccount)
	if err != nil {
		return err
	}
	color, err := evaluateOptExprAs(st, colorExpr, expectString)
	if err != nil {
		return err
	}

	item := BalanceQueryItem{
		Account: string(account),
		Asset:   string(st.CurrentAsset),
		Color:   string(color),
	}
	if !slices.Contains(*writes, item) {
		*writes = append(*writes, item)
	}
	return nil
}
