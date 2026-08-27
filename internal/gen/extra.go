package gen

import (
	"fmt"
	"math/big"
	"math/rand"
)

// seedAmount is used for `world -> accN` funding statements — always
// positive (funding is meant to give accounts something to spend; the
// "amounts can be zero/negative" exploration lives in monetary(), used
// elsewhere).
func seedAmount(rng *rand.Rand) *big.Int {
	return big.NewInt(int64(rng.Intn(1000)) + 1)
}

// genSeedStatements builds `world -> acc<i>` funding statements for every
// (account, asset) pair in the pool. Note: these run at EXECUTION time, so
// they do NOT feed a `balance()`-origin var (see genVarDecls) — vars-block
// origins are resolved once, before any statement in the script executes,
// on both engines (confirmed directly against the oracle). Pre-set starting
// balances (genPresetBalances) are the only way to make a balance()-origin
// var observe a non-zero value.
func genSeedStatements(rng *rand.Rand, poolSize int, assets []string) Program {
	stmts := make(Program, 0, poolSize*len(assets))
	for i := range poolSize {
		for _, asset := range assets {
			stmts = append(stmts, Statement{
				IsSendAll: false,
				Amount:    Monetary{Asset: asset, Amount: seedAmount(rng)},
				Source:    Source{Kind: SrcAccount, Account: "world"},
				Destination: Destination{
					Kind:    DestAccount,
					Account: fmt.Sprintf("acc%d", i),
				},
			})
		}
	}
	return stmts
}

// genPresetBalances populates balances with a starting balance for a random
// subset of (account, asset) pairs in the pool, including negative amounts
// — a precondition set directly on the store rather than by executing
// funding statements (see internal/difftest's runNew/runOracle, which feed
// this into each engine's StaticStore). "world" is deliberately never
// included: it's never balance-backed on either engine (the oracle's
// ResolveBalances never queries it; it must stay funding-only).
func genPresetBalances(rng *rand.Rand, poolSize int, assets []string, balances map[BalanceKey]*big.Int) {
	for i := range poolSize {
		acc := fmt.Sprintf("acc%d", i)
		for _, asset := range assets {
			if rng.Intn(2) != 0 {
				continue
			}
			// [-500, 1499]: comfortably covers both a plausible positive
			// starting balance and a negative one (the new/interesting case).
			amount := big.NewInt(int64(rng.Intn(2000) - 500))
			balances[BalanceKey{Account: acc, Asset: asset}] = amount
		}
	}
}

// genBalances decides, per script, how accounts get their starting
// balances: via `world ->` funding statements (as before), via pre-set
// starting balances on the store, or both. Weighted 50/25/25
// (seeds/preset/both) — "both" mode combined with multi-asset preset
// balances is close to the shape that surfaced a real bug this session
// (see DIFFTEST_HANDOFF.md), so it's weighted higher than a uniform split,
// while seeds-only stays the largest single bucket since it's the
// original, most-exercised path.
func genBalances(rng *rand.Rand, poolSize int, assets []string) (map[BalanceKey]*big.Int, Program) {
	balances := map[BalanceKey]*big.Int{}

	useSeeds := true
	usePreset := false
	switch rng.Intn(4) {
	case 0: // 25%: preset only
		useSeeds = false
		usePreset = true
	case 1: // 25%: both
		usePreset = true
	default: // 50%: seeds only
	}

	var seeds Program
	if useSeeds {
		seeds = genSeedStatements(rng, poolSize, assets)
	}
	if usePreset {
		genPresetBalances(rng, poolSize, assets, balances)
	}
	return balances, seeds
}

// genPresetMetadata populates metadata with a value for a random subset of
// (account, key) pairs in the pool — mirrors genPresetBalances. Values are
// plain decimal strings: both engines parse a `meta()`-origin var's raw
// stored string according to the var's declared type (confirmed directly:
// internal/interpreter's parseVar and the oracle's
// machine.NewValueFromString both just do a base-10 big.Int parse for a
// `number`-typed var), so a decimal string is all a `number`-typed
// meta()-origin var needs.
func genPresetMetadata(rng *rand.Rand, poolSize int, metadata map[MetaKey]string) {
	for i := range poolSize {
		acc := fmt.Sprintf("acc%d", i)
		for _, key := range metaKeyPool {
			if rng.Intn(2) != 0 {
				continue
			}
			metadata[MetaKey{Account: acc, Key: key}] = fmt.Sprintf("%d", rng.Intn(2000)-500)
		}
	}
}

// metaKeyPool is the small fixed set of metadata keys genPresetMetadata and
// genVarDecls pick from, mirroring assetPool's role for assets.
var metaKeyPool = []string{"k0", "k1", "k2"}

// genVarDecls generates 0-3 vars-block declarations, each either
// `monetary $name = balance(<account>, <asset>)` or
// `number $name = meta(<account>, "<key>")`. Accounts with a pre-set
// balance/metadata entry are preferred so the var actually observes
// something other than the default-zero/missing-key path; "world" is
// picked at a deliberate elevated rate (~1 in 4) specifically to exercise
// balance(@world, ASSET) — confirmed directly against the oracle to be a
// legal, always-zero read (never an error).
func genVarDecls(rng *rand.Rand, poolSize int, balances map[BalanceKey]*big.Int, metadata map[MetaKey]string) []VarDecl {
	n := rng.Intn(4) // 0..3
	if n == 0 {
		return nil
	}

	fundedKeys := make([]BalanceKey, 0, len(balances))
	for k := range balances {
		fundedKeys = append(fundedKeys, k)
	}
	metaKeys := make([]MetaKey, 0, len(metadata))
	for k := range metadata {
		metaKeys = append(metaKeys, k)
	}

	out := make([]VarDecl, n)
	for i := range out {
		if len(metaKeys) > 0 && rng.Intn(3) == 0 {
			k := metaKeys[rng.Intn(len(metaKeys))]
			out[i] = VarDecl{Kind: VarFromMeta, Account: k.Account, Key: k.Key}
			continue
		}
		if rng.Intn(4) == 0 {
			out[i] = VarDecl{Kind: VarFromBalance, Account: "world", Asset: pickAsset(rng)}
			continue
		}
		if len(fundedKeys) > 0 && rng.Intn(2) == 0 {
			k := fundedKeys[rng.Intn(len(fundedKeys))]
			out[i] = VarDecl{Kind: VarFromBalance, Account: k.Account, Asset: k.Asset}
			continue
		}
		out[i] = VarDecl{Kind: VarFromBalance, Account: account(rng, poolSize), Asset: pickAsset(rng)}
	}

	// Deliberately bias toward the exact collision shape that produced two
	// real bugs this session (see DIFFTEST_HANDOFF.md): two balance()-origin
	// vars on the same account, either the same asset or a different one.
	// Left to chance alone, this only happens incidentally.
	if len(out) >= 2 && rng.Intn(3) == 0 {
		i := rng.Intn(len(out))
		j := rng.Intn(len(out))
		if j == i {
			j = (j + 1) % len(out)
		}
		out[j].Kind = out[i].Kind
		out[j].Account = out[i].Account
		if out[i].Kind == VarFromBalance && rng.Intn(2) == 0 {
			out[j].Asset = out[i].Asset // exact duplicate (same account+asset)
		}
		if out[i].Kind == VarFromMeta && rng.Intn(2) == 0 {
			out[j].Key = out[i].Key
		}
	}

	return out
}

// genNumExpr generates a small arithmetic expression (literal, or +/- of
// two smaller expressions), used for set_tx_meta/set_account_meta values.
// depth bounds recursion (same hard-cap rationale as maxRecursionDepth).
// Leaf literals are always non-negative: a bare negative NUMBER literal
// can't be rendered legally (the oracle's `expression` grammar has no unary
// minus production — only ExprAddSub/ExprLiteral/ExprVariable — so a
// negative leaf would just make the oracle reject every such script, always
// trivially "ok" via Compare's tolerated-rejection path and never real
// coverage). NumSub compositions can still legally produce a negative
// runtime value (e.g. `5 - 10`), which both engines parse and evaluate.
func genNumExpr(rng *rand.Rand, depth int) NumExpr {
	if depth <= 0 || rng.Intn(3) != 0 {
		return NumExpr{Kind: NumLit, Lit: big.NewInt(int64(rng.Intn(1000)))}
	}
	left := genNumExpr(rng, depth-1)
	right := genNumExpr(rng, depth-1)
	if rng.Intn(2) == 0 {
		return NumExpr{Kind: NumAdd, Left: &left, Right: &right}
	}
	return NumExpr{Kind: NumSub, Left: &left, Right: &right}
}

// varIndicesOfKind returns the indices into vars whose Kind matches.
func varIndicesOfKind(vars []VarDecl, kind VarDeclKind) []int {
	var idxs []int
	for i, v := range vars {
		if v.Kind == kind {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

// genExtraStatements generates 0-3 non-send statements (save/set_tx_meta/
// set_account_meta/a var-backed send/a meta()-var-backed set_tx_meta),
// interspersed with the core send-only program by the caller (see
// riffleOrder).
func genExtraStatements(rng *rand.Rand, poolSize int, vars []VarDecl) []ExtraStatement {
	n := rng.Intn(4) // 0..3
	out := make([]ExtraStatement, 0, n)

	balanceVarIdxs := varIndicesOfKind(vars, VarFromBalance)
	metaVarIdxs := varIndicesOfKind(vars, VarFromMeta)

	for range n {
		kind := ExtraStatementKind(rng.Intn(6))
		// ExtraSave/ExtraSendVar need a declared balance()-origin var to
		// reference; ExtraSetTxMetaVar needs a declared meta()-origin var.
		// Fall back to a plain set_tx_meta if the needed kind isn't
		// available.
		if (kind == ExtraSave && rng.Intn(2) == 0 || kind == ExtraSendVar) && len(balanceVarIdxs) == 0 {
			kind = ExtraSetTxMeta
		}
		if kind == ExtraSetTxMetaVar && len(metaVarIdxs) == 0 {
			kind = ExtraSetTxMeta
		}

		switch kind {
		case ExtraSave:
			acc := account(rng, poolSize)
			if len(balanceVarIdxs) > 0 && rng.Intn(2) == 0 {
				idx := balanceVarIdxs[rng.Intn(len(balanceVarIdxs))]
				out = append(out, ExtraStatement{Kind: ExtraSave, VarIdx: &idx, Account: acc})
			} else {
				asset := pickAsset(rng)
				m := monetary(rng, asset)
				out = append(out, ExtraStatement{Kind: ExtraSave, Monetary: &m, Account: acc})
			}

		case ExtraSaveAll:
			out = append(out, ExtraStatement{
				Kind:    ExtraSaveAll,
				Asset:   pickAsset(rng),
				Account: account(rng, poolSize),
			})

		case ExtraSetTxMeta:
			out = append(out, ExtraStatement{
				Kind:  ExtraSetTxMeta,
				Key:   fmt.Sprintf("k%d", rng.Intn(5)),
				Value: genNumExpr(rng, 3),
			})

		case ExtraSetAccountMeta:
			out = append(out, ExtraStatement{
				Kind:    ExtraSetAccountMeta,
				Account: account(rng, poolSize),
				Key:     fmt.Sprintf("k%d", rng.Intn(5)),
				Value:   genNumExpr(rng, 3),
			})

		case ExtraSendVar:
			idx := balanceVarIdxs[rng.Intn(len(balanceVarIdxs))]
			out = append(out, ExtraStatement{
				Kind:        ExtraSendVar,
				VarIdx:      &idx,
				Account:     account(rng, poolSize),
				Destination: account(rng, poolSize),
			})

		case ExtraSetTxMetaVar:
			idx := metaVarIdxs[rng.Intn(len(metaVarIdxs))]
			out = append(out, ExtraStatement{
				Kind:   ExtraSetTxMetaVar,
				VarIdx: &idx,
				Key:    fmt.Sprintf("k%d", rng.Intn(5)),
			})
		}
	}
	return out
}

// riffleOrder returns a random interleaving of two sequences of lengths a
// and b, as a slice of booleans (true = take the next element from the
// first sequence). Used to interleave Extra statements throughout Program
// instead of only appending them at the end — see Script.Order.
func riffleOrder(rng *rand.Rand, a, b int) []bool {
	order := make([]bool, 0, a+b)
	for a > 0 && b > 0 {
		if rng.Intn(a+b) < a {
			order = append(order, true)
			a--
		} else {
			order = append(order, false)
			b--
		}
	}
	for ; a > 0; a-- {
		order = append(order, true)
	}
	for ; b > 0; b-- {
		order = append(order, false)
	}
	return order
}

// GenerateScriptAST orchestrates one full round of generation: picks a
// script-wide account-pool size, decides how balances/metadata are seeded
// (world-funding statements, pre-set store state, or both), declares a
// handful of vars (balance()- and meta()-origin), generates the core
// send-only program, generates a few extra non-send statements, and decides
// how to interleave the two.
func GenerateScriptAST(rng *rand.Rand) Script {
	poolSize := pickPoolSize(rng)

	balances, seeds := genBalances(rng, poolSize, assetPool)
	metadata := map[MetaKey]string{}
	genPresetMetadata(rng, poolSize, metadata)
	vars := genVarDecls(rng, poolSize, balances, metadata)
	program := cleanupProgram(genProgram(rng, poolSize))
	extra := genExtraStatements(rng, poolSize, vars)
	order := riffleOrder(rng, len(program), len(extra))

	return Script{
		Vars:     vars,
		Seeds:    seeds,
		Program:  program,
		Extra:    extra,
		Order:    order,
		Balances: balances,
		Metadata: metadata,
	}
}
