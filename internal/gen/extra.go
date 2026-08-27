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
// starting balances on the store, or both. Pre-set-only and both are new;
// seeds-only is kept as the majority case so the well-exercised original
// funding path stays dominant.
func genBalances(rng *rand.Rand, poolSize int, assets []string) (map[BalanceKey]*big.Int, Program) {
	balances := map[BalanceKey]*big.Int{}

	useSeeds := true
	usePreset := false
	switch rng.Intn(10) {
	case 0, 1, 2: // 30%: preset only
		useSeeds = false
		usePreset = true
	case 3: // 10%: both
		usePreset = true
	default: // 60%: seeds only (matches the original, well-tested behavior)
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

// genVarDecls generates 0-3 `monetary $name = balance(<account>, <asset>)`
// declarations. Accounts with a pre-set balance are preferred so the var
// actually observes something other than the default zero; "world" is
// picked at a deliberate elevated rate (~1 in 4) specifically to exercise
// balance(@world, ASSET) — confirmed directly against the oracle to be a
// legal, always-zero read (never an error).
func genVarDecls(rng *rand.Rand, poolSize int, balances map[BalanceKey]*big.Int) []VarDecl {
	n := rng.Intn(4) // 0..3
	if n == 0 {
		return nil
	}

	fundedKeys := make([]BalanceKey, 0, len(balances))
	for k := range balances {
		fundedKeys = append(fundedKeys, k)
	}

	out := make([]VarDecl, n)
	for i := range out {
		if rng.Intn(4) == 0 {
			out[i] = VarDecl{Account: "world", Asset: pickAsset(rng)}
			continue
		}
		if len(fundedKeys) > 0 && rng.Intn(2) == 0 {
			k := fundedKeys[rng.Intn(len(fundedKeys))]
			out[i] = VarDecl{Account: k.Account, Asset: k.Asset}
			continue
		}
		out[i] = VarDecl{Account: account(rng, poolSize), Asset: pickAsset(rng)}
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

// genExtraStatements generates 0-3 non-send statements (save/set_tx_meta/
// set_account_meta/a var-backed send), interspersed with the core
// send-only program by the caller.
func genExtraStatements(rng *rand.Rand, poolSize int, vars []VarDecl) []ExtraStatement {
	n := rng.Intn(4) // 0..3
	out := make([]ExtraStatement, 0, n)

	for range n {
		kind := ExtraStatementKind(rng.Intn(5))
		// ExtraSave/ExtraSendVar need a declared var to reference one; fall
		// back to a plain save/set_tx_meta if none were declared.
		if (kind == ExtraSave && rng.Intn(2) == 0 || kind == ExtraSendVar) && len(vars) == 0 {
			kind = ExtraSetTxMeta
		}

		switch kind {
		case ExtraSave:
			acc := account(rng, poolSize)
			if len(vars) > 0 && rng.Intn(2) == 0 {
				idx := rng.Intn(len(vars))
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
			idx := rng.Intn(len(vars))
			out = append(out, ExtraStatement{
				Kind:        ExtraSendVar,
				VarIdx:      &idx,
				Account:     account(rng, poolSize),
				Destination: account(rng, poolSize),
			})
		}
	}
	return out
}

// GenerateScriptAST orchestrates one full round of generation: picks a
// script-wide account-pool size, decides how balances are seeded
// (world-funding statements, pre-set store balances, or both), declares a
// handful of vars (including balance()-origin ones), generates the core
// send-only program, and generates a few extra non-send statements.
func GenerateScriptAST(rng *rand.Rand) Script {
	poolSize := pickPoolSize(rng)

	balances, seeds := genBalances(rng, poolSize, assetPool)
	vars := genVarDecls(rng, poolSize, balances)
	program := cleanupProgram(genProgram(rng, poolSize))
	extra := genExtraStatements(rng, poolSize, vars)

	return Script{
		Vars:     vars,
		Seeds:    seeds,
		Program:  program,
		Extra:    extra,
		Balances: balances,
	}
}
