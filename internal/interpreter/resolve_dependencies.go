package interpreter

import (
	"context"
	"maps"
	"math/big"
	"slices"

	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/parser"
)

// ResolvedDependencies holds the concrete data a script reads during resolution.
// The consumer can use this to preload volumes and detect input drift.
type ResolvedDependencies struct {
	// Volumes contains all (account, asset) → balance pairs read during resolution.
	Volumes map[string]map[string]*big.Int

	// Metadata contains all (account, key) → value pairs read during resolution.
	Metadata map[string]map[string]string
}

// ResolveDependenciesOptions configures ResolveDependencies behavior.
type ResolveDependenciesOptions struct {
	// FeatureFlags enables additional experimental features (same as RunWithFeatureFlags).
	FeatureFlags map[string]struct{}

	// ForbiddenFlags rejects scripts that declare any of these features.
	// This takes precedence over script-level #![feature("...")] directives.
	// Use this to block features that ResolveDependencies cannot fully resolve
	// (e.g. experimental-mid-script-function-call).
	ForbiddenFlags map[string]struct{}
}

// ResolveDependencies discovers which balances and metadata a script will read
// by performing variable resolution and balance preloading — the same two phases
// that RunProgram does before executing statements. It does NOT execute the
// statements themselves (no postings are produced).
//
// This covers all store reads for scripts that don't use
// experimental-mid-script-function-call. Scripts using that feature may trigger
// additional balance reads during execution (e.g. balance() called between two
// send statements, where the result depends on the first send's postings).
// Consumers that cannot tolerate incomplete dependency lists should forbid this
// feature via ForbiddenFlags.
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
		index := slices.Index(flags.AllFlags, flag.String)
		if index == -1 {
			return nil, InvalidFeature{Feature: flag.String}
		}

		if _, forbidden := opts.ForbiddenFlags[flag.String]; forbidden {
			return nil, ForbiddenFeature{FlagName: flag.String}
		}

		featureFlags[flag.String] = struct{}{}
	}

	// Replicate the initialization and preload phases of RunProgram,
	// but stop before statement execution.
	st := programState{
		ParsedVars:         make(map[string]Value),
		TxMeta:             make(map[string]Value),
		CachedAccountsMeta: AccountsMetadata{},
		CachedBalances:     Balances{},
		SetAccountsMeta:    AccountsMetadata{},
		Store:              recorder,
		Postings:           make([]Posting, 0),
		fundsQueue:         newFundsQueue(nil),

		CurrentBalanceQuery: BalanceQuery{},
		ctx:                 ctx,
		FeatureFlags:        featureFlags,
	}

	// Phase 1: parse variables — resolves meta(), balance(), overdraft() origins.
	st.varOriginPosition = true
	if program.Vars != nil {
		if err := st.parseVars(program.Vars.Declarations, vars); err != nil {
			return nil, err
		}
	}
	st.varOriginPosition = false

	// Phase 2: traverse statement ASTs to discover balance needs, then preload.
	for _, statement := range program.Statements {
		if err := st.findBalancesQueriesInStatement(statement); err != nil {
			return nil, err
		}
	}

	if err := st.runBalancesQuery(); err != nil {
		return nil, QueryBalanceError{WrappedError: err}
	}

	return &ResolvedDependencies{
		Volumes:  recorder.balanceReads,
		Metadata: recorder.metadataReads,
	}, nil
}
