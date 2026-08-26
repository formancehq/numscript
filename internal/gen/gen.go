package gen

import (
	"fmt"
	"math/big"
	"math/rand"
)

// accountsNumber matches Gen.hs: accounts are drawn from a fixed pool of
// "acc0".."acc15" (16 accounts) plus "world".
const accountsNumber = 15

// ratio mirrors Haskell's `Ratio Int`: the odds "num/denom" of an event.
type ratio struct {
	num, denom int
}

// smallerRat mirrors Gen.hs: (num+1)/(denom+1), converging toward 1 as it's
// repeatedly applied — used to make recursion/list-length increasingly
// likely to stop the deeper generation goes.
func smallerRat(r ratio) ratio {
	return ratio{r.num + 1, r.denom + 1}
}

// weightedCoin returns true with probability r.num/r.denom.
func weightedCoin(rng *rand.Rand, r ratio) bool {
	return rng.Intn(r.denom) < r.num
}

// nonUniform mirrors Gen.hs: picks an unbounded integer via a non-uniform
// distribution starting at n, recursing (n+1, smallerRat(r)) with
// probability (1 - r.num/r.denom).
func nonUniform(rng *rand.Rand, r ratio, n int) int {
	if weightedCoin(rng, r) {
		return n
	}
	return nonUniform(rng, smallerRat(r), n+1)
}

// nonUniformListOf mirrors Gen.hs's nonUniformListOf: a list whose length is
// drawn from nonUniform(1/10, 1).
func nonUniformListOf[T any](rng *rand.Rand, g func() T) []T {
	n := nonUniform(rng, ratio{1, 10}, 1)
	out := make([]T, n)
	for i := range out {
		out[i] = g()
	}
	return out
}

// portionsList mirrors Gen.hs: a non-uniform-length list of random positive
// weights, normalized into rationals summing to exactly 1.
func portionsList(rng *rand.Rand) []*big.Rat {
	xs := nonUniformListOf(rng, func() int64 {
		return int64(rng.Intn(100) + 1)
	})

	total := int64(0)
	for _, x := range xs {
		total += x
	}

	out := make([]*big.Rat, len(xs))
	for i, x := range xs {
		out[i] = big.NewRat(x, total)
	}
	return out
}

func monetary(rng *rand.Rand) Monetary {
	// Gen.hs hardcodes the asset to the literal "COIN" always.
	return Monetary{Asset: "COIN", Amount: big.NewInt(int64(rng.Intn(1000)))}
}

func addMonetary(x *big.Int, m Monetary) Monetary {
	return Monetary{Asset: m.Asset, Amount: new(big.Int).Add(x, m.Amount)}
}

func account(rng *rand.Rand) string {
	return fmt.Sprintf("acc%d", rng.Intn(accountsNumber+1))
}

func zeroFreqIf(weight int, cond bool) int {
	if cond {
		return 0
	}
	return weight
}

// weighted and pick mirror QuickCheck's `frequency`: weighted random
// selection proportional to weight, only ever evaluating the chosen branch
// (branches are thunks since they may recurse and consume more randomness).
type weighted[T any] struct {
	weight int
	build  func() T
}

func pick[T any](rng *rand.Rand, branches []weighted[T]) T {
	total := 0
	for _, b := range branches {
		total += b.weight
	}
	if total <= 0 {
		panic("gen: pick called with all-zero weights")
	}
	r := rng.Intn(total)
	for _, b := range branches {
		if r < b.weight {
			return b.build()
		}
		r -= b.weight
	}
	panic("gen: unreachable")
}

type sourceOptions struct {
	sentAmt                *big.Int
	isToplevel             bool
	isUnbounded            bool
	keepNestingProbability ratio
}

func defaultSourceOptions(sentAmt *big.Int, unbounded bool) sourceOptions {
	return sourceOptions{
		sentAmt:                sentAmt,
		isToplevel:             true,
		isUnbounded:            unbounded,
		keepNestingProbability: ratio{1, 15},
	}
}

func genSource(rng *rand.Rand, opts sourceOptions) Source {
	stopRecursion := weightedCoin(rng, opts.keepNestingProbability)

	nestedOpts := opts
	nestedOpts.isToplevel = false
	nestedOpts.keepNestingProbability = smallerRat(opts.keepNestingProbability)

	return pick(rng, []weighted[Source]{
		{
			zeroFreqIf(5, opts.isUnbounded),
			func() Source {
				return Source{Kind: SrcAccount, Account: "world"}
			},
		},
		{
			15,
			func() Source {
				return Source{Kind: SrcAccount, Account: account(rng)}
			},
		},
		{
			5,
			func() Source {
				m := monetary(rng)
				return Source{Kind: SrcAccountOverdraft, Account: account(rng), Overdraft: &m}
			},
		},
		{
			zeroFreqIf(5, opts.isUnbounded),
			func() Source {
				return Source{Kind: SrcAccountOverdraft, Account: account(rng), Overdraft: nil}
			},
		},
		{
			5,
			func() Source {
				cap := addMonetary(opts.sentAmt, monetary(rng))
				innerOpts := nestedOpts
				innerOpts.isUnbounded = false
				inner := genSource(rng, innerOpts)
				return Source{Kind: SrcCapped, Cap: &cap, Inner: &inner}
			},
		},
		{
			zeroFreqIf(10, stopRecursion),
			func() Source {
				list := nonUniformListOf(rng, func() Source { return genSource(rng, nestedOpts) })
				return Source{Kind: SrcInorder, Sources: list}
			},
		},
		{
			zeroFreqIf(15, stopRecursion || !opts.isToplevel || opts.isUnbounded),
			func() Source {
				innerOpts := nestedOpts
				innerOpts.isUnbounded = false
				portions := portionsList(rng)
				clauses := make([]SourceAllotmentClause, len(portions))
				for i, p := range portions {
					clauses[i] = SourceAllotmentClause{Portion: p, Source: genSource(rng, innerOpts)}
				}
				return Source{Kind: SrcAllotment, Clauses: clauses}
			},
		},
	})
}

type destinationOptions struct {
	keepNestingProbability ratio
}

func defaultDestinationOptions() destinationOptions {
	return destinationOptions{keepNestingProbability: ratio{1, 15}}
}

func genDestination(rng *rand.Rand, opts destinationOptions) Destination {
	stopRecursion := weightedCoin(rng, opts.keepNestingProbability)
	nestedOpts := destinationOptions{keepNestingProbability: smallerRat(opts.keepNestingProbability)}

	return pick(rng, []weighted[Destination]{
		{
			30,
			func() Destination {
				return Destination{Kind: DestAccount, Account: account(rng)}
			},
		},
		{
			zeroFreqIf(10, stopRecursion),
			func() Destination {
				clauses := nonUniformListOf(rng, func() DestInorderClause {
					return DestInorderClause{Max: monetary(rng), KeptOrDest: genKeptOrDest(rng, nestedOpts)}
				})
				remaining := genKeptOrDest(rng, nestedOpts)
				return Destination{Kind: DestInorder, InorderClauses: clauses, Remaining: &remaining}
			},
		},
		{
			zeroFreqIf(10, stopRecursion),
			func() Destination {
				portions := portionsList(rng)
				clauses := make([]DestAllotmentClause, len(portions))
				for i, p := range portions {
					clauses[i] = DestAllotmentClause{Portion: p, KeptOrDest: genKeptOrDest(rng, nestedOpts)}
				}
				return Destination{Kind: DestAllotment, AllotClauses: clauses}
			},
		},
	})
}

func genKeptOrDest(rng *rand.Rand, opts destinationOptions) KeptOrDest {
	return pick(rng, []weighted[KeptOrDest]{
		{1, func() KeptOrDest { return KeptOrDest{Kind: Kept} }},
		{3, func() KeptOrDest {
			d := genDestination(rng, opts)
			return KeptOrDest{Kind: To, Dest: &d}
		}},
	})
}

func genStatement(rng *rand.Rand) Statement {
	pickUnbounded := pick(rng, []weighted[bool]{
		{1, func() bool { return true }},
		{3, func() bool { return false }},
	})

	sent := monetary(rng)
	src := genSource(rng, defaultSourceOptions(sent.Amount, pickUnbounded))
	dest := genDestination(rng, defaultDestinationOptions())

	if pickUnbounded {
		return Statement{IsSendAll: true, Asset: sent.Asset, Source: src, Destination: dest}
	}
	return Statement{IsSendAll: false, Amount: sent, Source: src, Destination: dest}
}

func genProgram(rng *rand.Rand) Program {
	return Program(nonUniformListOf(rng, func() Statement { return genStatement(rng) }))
}

// GenerateProgram generates a random program and applies the cleanup pass
// (cleanup.go) that removes constructs the legacy machine would reject.
func GenerateProgram(rng *rand.Rand) Program {
	return cleanupProgram(genProgram(rng))
}

// GenerateSeeds generates a funding program: `world` -> each of `acc0`..
// `accN`, used to give the generated program's accounts a starting balance
// when run in the same script (see convert.go / the difftest harness).
func GenerateSeeds(rng *rand.Rand) Program {
	stmts := make(Program, accountsNumber+1)
	for i := 0; i <= accountsNumber; i++ {
		stmts[i] = Statement{
			IsSendAll: false,
			Amount:    monetary(rng),
			Source:    Source{Kind: SrcAccount, Account: "world"},
			Destination: Destination{
				Kind:    DestAccount,
				Account: fmt.Sprintf("acc%d", i),
			},
		}
	}
	return stmts
}
