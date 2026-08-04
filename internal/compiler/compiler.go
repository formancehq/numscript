package compiler

import (
	"fmt"
	"maps"
	"math/big"
	"slices"

	"github.com/formancehq/numscript/internal/builtins"
	"github.com/formancehq/numscript/internal/flags"
	"github.com/formancehq/numscript/internal/ir"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/typecheck"
	"github.com/formancehq/numscript/internal/utils"
	"github.com/formancehq/numscript/internal/vm"
)

// Compile lowers a parsed program to the VarsEncoder that turns a json var
// payload into the vm.Vars the program expects, plus the vm.Program itself.
//
// featureFlags is the set of experimental features the host allows; a construct
// gated behind a flag that isn't in the set fails compilation. As in
// interpreter.RunProgram, the script's own #![feature(..)] declarations are
// unioned in.
func Compile(program parser.Program, featureFlags map[string]struct{}) (VarsEncoder, vm.Program, error) {
	compiled, cErr := compileProgramToIR(program, featureFlags)
	if cErr != nil {
		return VarsEncoder{}, vm.Program{}, fmt.Errorf("%v", cErr)
	}

	if err := ir.Typecheck(compiled.instructions); err != nil {
		return VarsEncoder{}, vm.Program{}, err
	}

	prog, err := ir.Assemble(compiled.instructions)
	if err != nil {
		return VarsEncoder{}, vm.Program{}, err
	}

	return compiled.varsEncoder, prog, nil
}

type compiledProgramIR struct {
	instructions []ir.Instr
	varsEncoder  VarsEncoder
}

type state struct {
	ir.Builder

	vars         map[string]value
	exprTypes    map[parser.ValueExpr]typecheck.Type
	featureFlags map[string]struct{}
	// set by compileSentValue before any source/destination is compiled; nil
	// until then, so compileCapAmount can't silently assert against register 0.
	currentAssetReg *ir.Reg

	nextIntVar int
	nextStrVar int
	varDecls   []varDecl

	// holds worldAccount; see pullFromAccount
	worldReg ir.Reg
}

// The unbounded account. The VM knows nothing about it: a source account is a
// register, so the comparison is compiled, not built in.
const worldAccount = "world"

// Every flag in flags.AllFlags has a check site below except
// ExperimentalScopedFunction: scoped() isn't in typecheck's builtin table, so a
// call to it is already rejected as an unknown function.
func (st *state) checkFeatureFlag(rng parser.Range, flag flags.FeatureFlag) CompilerError {
	if _, ok := st.featureFlags[flag]; ok {
		return nil
	}
	return ExperimentalFeature{Range: rng, FlagName: flag}
}

// pushInstructionWithDestErr is PushWithDest in the shape compileExpr returns.
func (st *state) pushInstructionWithDestErr(getInstr func(dest ir.Reg) ir.Instr) (ir.Reg, CompilerError) {
	return st.PushWithDest(getInstr), nil
}

func (st *state) compileAllot(amount ir.Reg, allotments []parser.AllotmentValue) ([]ir.Reg, CompilerError) {
	n := len(allotments)
	portions := make([]ir.Reg, n)
	remainingIdx := -1
	for i, al := range allotments {
		switch al := al.(type) {
		case *parser.ValueExprAllotment:
			p, err := st.compileExpr(al.Value)
			if err != nil {
				return nil, err
			}
			portions[i] = p
		case *parser.RemainingAllotment:
			if remainingIdx != -1 {
				return nil, DuplicateRemaining{Range: al.Range}
			}
			remainingIdx = i
		default:
			utils.NonExhaustiveMatchPanic[any](al)
		}
	}

	leftover := st.compilePortionOne()
	for i := range allotments {
		if i == remainingIdx {
			continue
		}
		prev, pi := leftover, portions[i]
		leftover = st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.BinaryOp{Op: ir.OpSubPortion{}, Left: prev, Right: pi, Dest: dest}
		})
	}

	st.Push(ir.AssertLeftover{Portion: leftover, Exact: remainingIdx == -1})
	if remainingIdx != -1 {
		portions[remainingIdx] = leftover
	}

	return st.compileAllotmentSplit(amount, portions), nil
}

// TODO properly review claude-generated compileAllotmentSplit

// compileAllotmentSplit writes the amount split across the portions: one int
// register per portion, summing exactly to amount. It expects the portions to
// sum to 1, which is what the assert_leftover emitted by the caller establishes.
//
// Each share is floor(portion * amount); flooring loses strictly less than one
// unit per share, so the shortfall is under len(portions) and a single
// front-to-back pass handing out one unit each closes it. That order is
// observable — 100 by thirds is 34/33/33, not 33/33/34.
func (st *state) compileAllotmentSplit(amount ir.Reg, portions []ir.Reg) []ir.Reg {
	n := len(portions)
	dest := make([]ir.Reg, n)

	amountPortion := st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.UnaryOp{Op: ir.OpIntToPortion{}, Arg: amount, Dest: dest}
	})

	// total accumulates the floored shares; it starts as a copy of the first one
	// rather than a zero, which saves a load
	var total ir.Reg
	for i, portion := range portions {
		product := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.BinaryOp{Op: ir.OpMulPortion{}, Left: portion, Right: amountPortion, Dest: dest}
		})
		dest[i] = st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{Op: ir.OpPortionToInt{}, Arg: product, Dest: dest}
		})

		if i == 0 {
			total = st.PushWithDest(func(t ir.Reg) ir.Instr {
				return ir.UnaryOp{Op: ir.OpIntCopy{}, Arg: dest[0], Dest: t}
			})
			continue
		}
		st.Push(ir.BinaryOp{Op: ir.OpAddInt{}, Left: total, Right: dest[i], Dest: total})
	}

	// The shortfall is at most n-1, so the last share never receives a unit and
	// its block would be dead. The jumps go forward to one shared exit, which is
	// what lets this be a straight line: the assembler rejects backward jumps.
	if n > 1 {
		one := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.LoadInt{Value: *big.NewInt(1), Dest: dest}
		})
		done := st.FreshLabel("allot_end")

		for i := 0; i < n-1; i++ {
			short := st.PushWithDest(func(dest ir.Reg) ir.Instr {
				return ir.BinaryOp{Op: ir.OpLtInt{}, Left: total, Right: amount, Dest: dest}
			})
			st.Push(ir.JmpIfFalse{Cond: short, Target: done})
			st.Push(ir.BinaryOp{Op: ir.OpAddInt{}, Left: dest[i], Right: one, Dest: dest[i]})
			st.Push(ir.BinaryOp{Op: ir.OpAddInt{}, Left: total, Right: one, Dest: total})
		}

		st.Push(ir.LabelMarker{Label: done})
	}

	return dest
}

func (st *state) compileCapAmount(monExpr parser.ValueExpr) (ir.Reg, CompilerError) {
	mon, err := st.compileMonetaryExpr(monExpr)
	if err != nil {
		return 0, err
	}
	if st.currentAssetReg == nil {
		panic("compileCapAmount: no current asset (compileSentValue must run first)")
	}
	st.Push(ir.AssertSameAsset{Left: mon.Asset, Right: *st.currentAssetReg})
	return mon.Amount, nil
}

func (st *state) compilePortionOne() ir.Reg {
	one := st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.LoadInt{Value: *big.NewInt(1), Dest: dest}
	})
	return st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.BinaryOp{Op: ir.OpMakePortion{}, Left: one, Right: one, Dest: dest}
	})
}

// compileExpr compiles a non-monetary expression into the single register its
// type maps to. Monetary-typed expressions go to compileMonetaryExpr instead,
// since a monetary needs two registers.
func (st *state) compileExpr(expr parser.ValueExpr) (ir.Reg, CompilerError) {
	if st.exprTypes[expr] == typecheck.TypeMonetary {
		panic("compileExpr: monetary expression (use compileMonetaryExpr)")
	}

	switch expr := expr.(type) {
	case *parser.AssetLiteral:
		return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
			return ir.LoadStr{
				Value: expr.Asset,
				Dest:  dest,
			}
		})

	case *parser.StringLiteral:
		return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
			return ir.LoadStr{
				Value: expr.String,
				Dest:  dest,
			}
		})

	case *parser.NumberLiteral:
		return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
			return ir.LoadInt{
				Value: *expr.Number,
				Dest:  dest,
			}
		})

	case *parser.AccountInterpLiteral:
		var parts []ir.Reg
		hasVar := false
		for _, part := range expr.Parts {
			switch part := part.(type) {
			case parser.AccountTextPart:
				dest := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.LoadStr{
						Value: part.Name,
						Dest:  dest,
					}
				})
				parts = append(parts, dest)
			case *parser.Variable:
				if err := st.checkFeatureFlag(part.Range, flags.ExperimentalAccountInterpolationFlag); err != nil {
					return 0, err
				}
				hasVar = true
				// reject before compiling, so a non-castable part doesn't reach
				// compileExpr (which only handles non-monetary expressions)
				t := st.exprTypes[part]
				switch t {
				case typecheck.TypeAccount, typecheck.TypeString, typecheck.TypeNumber:
				default:
					return 0, CannotCastToString{Range: part.GetRange(), Type: t}
				}
				r, err := st.compileExpr(part)
				if err != nil {
					return 0, err
				}
				if t == typecheck.TypeNumber {
					r = st.PushWithDest(func(dest ir.Reg) ir.Instr {
						return ir.UnaryOp{Op: ir.OpIntToString{}, Arg: r, Dest: dest}
					})
				}
				parts = append(parts, r)
			}
		}

		acc := parts[0]
		for _, part := range parts[1:] {
			left, right := acc, part
			acc = st.PushWithDest(func(dest ir.Reg) ir.Instr {
				return ir.BinaryOp{Op: ir.OpAddString{}, Left: left, Right: right, Dest: dest}
			})
		}
		// an interpolated var can inject chars that make the name ill-formed;
		// all-text literals are valid by construction, so skip the check
		if hasVar {
			st.Push(ir.AssertValidAccount{Account: acc})
		}
		return acc, nil

	case *parser.Variable:
		v, ok := st.vars[expr.Name]
		if !ok {
			return 0, UnboundVar{Range: expr.Range, Var: expr.Name}
		}
		return v.Reg, nil

	case *parser.PercentageLiteral:
		// e.g. 50% -> portion 50/100; mk_portion reduces via SetFrac
		ratio := expr.ToRatio()
		numReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.LoadInt{Value: *ratio.Num(), Dest: dest}
		})
		denReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.LoadInt{Value: *ratio.Denom(), Dest: dest}
		})
		return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
			return ir.BinaryOp{Op: ir.OpMakePortion{}, Left: numReg, Right: denReg, Dest: dest}
		})

	case *parser.BinaryInfix:
		leftReg, err := st.compileExpr(expr.Left)
		if err != nil {
			return 0, err
		}
		rightReg, err := st.compileExpr(expr.Right)
		if err != nil {
			return 0, err
		}

		switch expr.Operator {
		case parser.InfixOperatorDiv:
			return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
				return ir.BinaryOp{Op: ir.OpMakePortion{}, Left: leftReg, Right: rightReg, Dest: dest}
			})

		case parser.InfixOperatorPlus:
			return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
				return ir.BinaryOp{Op: ir.OpAddInt{}, Left: leftReg, Right: rightReg, Dest: dest}
			})

		case parser.InfixOperatorMinus:
			return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
				return ir.BinaryOp{Op: ir.OpSubInt{}, Left: leftReg, Right: rightReg, Dest: dest}
			})

		default:
			panic("TODO compileExpr binary op " + string(expr.Operator))
		}

	case *parser.Prefix:
		switch expr.Operator {
		case parser.PrefixOperatorMinus:
			argReg, err := st.compileExpr(expr.Expr)
			if err != nil {
				return 0, err
			}
			return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
				return ir.UnaryOp{Op: ir.OpNegInt{}, Arg: argReg, Dest: dest}
			})

		default:
			panic("TODO compileExpr prefix op " + string(expr.Operator))
		}

	case *parser.FnCall:
		return st.compileFnCall(expr, false)

	default:
		return utils.NonExhaustiveMatchPanic[ir.Reg](expr), nil
	}
}

// compileFnCall takes isVarOrigin to tell apart the two positions the interpreter
// distinguishes: a call that *is* a variable's origin expression, versus one
// nested anywhere else (which needs the mid-script-function-call flag).
func (st *state) compileFnCall(expr *parser.FnCall, isVarOrigin bool) (ir.Reg, CompilerError) {
	if !isVarOrigin {
		if err := st.checkFeatureFlag(expr.Range, flags.ExperimentalMidScriptFunctionCall); err != nil {
			return 0, err
		}
	}

	switch expr.Caller.Name {
	case builtins.GetAmount:
		if err := st.checkFeatureFlag(expr.Range, flags.ExperimentalGetAmountFunctionFeatureFlag); err != nil {
			return 0, err
		}
		mon, err := st.compileMonetaryExpr(expr.Args[0])
		if err != nil {
			return 0, err
		}
		return mon.Amount, nil

	case builtins.GetAsset:
		if err := st.checkFeatureFlag(expr.Range, flags.ExperimentalGetAssetFunctionFeatureFlag); err != nil {
			return 0, err
		}
		mon, err := st.compileMonetaryExpr(expr.Args[0])
		if err != nil {
			return 0, err
		}
		return mon.Asset, nil

	case builtins.Meta:
		return 0, InvalidMetaPosition{Range: expr.Range}

	default:
		panic("TODO compileExpr fn call " + expr.Caller.Name)
	}
}

// compileMonetaryExpr compiles a monetary-typed expression into the (asset,
// amount) register pair. Which expressions reach here is decided by
// st.exprTypes; compileExpr rejects monetary-typed ones.
func (st *state) compileMonetaryExpr(expr parser.ValueExpr) (monetaryValue, CompilerError) {
	switch expr := expr.(type) {
	case *parser.MonetaryLiteral:
		assetReg, err := st.compileExpr(expr.Asset)
		if err != nil {
			return monetaryValue{}, err
		}
		amtReg, err := st.compileExpr(expr.Amount)
		if err != nil {
			return monetaryValue{}, err
		}
		return monetaryValue{Asset: assetReg, Amount: amtReg}, nil

	case *parser.Variable:
		v, ok := st.vars[expr.Name]
		if !ok {
			return monetaryValue{}, UnboundVar{Range: expr.Range, Var: expr.Name}
		}
		if v.Mon == nil {
			panic("compileMonetaryExpr: $" + expr.Name + " is not a monetary")
		}
		return *v.Mon, nil

	case *parser.BinaryInfix:
		left, err := st.compileMonetaryExpr(expr.Left)
		if err != nil {
			return monetaryValue{}, err
		}
		right, err := st.compileMonetaryExpr(expr.Right)
		if err != nil {
			return monetaryValue{}, err
		}
		st.Push(ir.AssertSameAsset{Left: left.Asset, Right: right.Asset})

		var op ir.BinKind
		switch expr.Operator {
		case parser.InfixOperatorPlus:
			op = ir.OpAddInt{}
		case parser.InfixOperatorMinus:
			op = ir.OpSubInt{}
		default:
			panic("TODO compileMonetaryExpr binary op " + string(expr.Operator))
		}
		amount := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.BinaryOp{Op: op, Left: left.Amount, Right: right.Amount, Dest: dest}
		})
		// the assert above makes left vs right immaterial
		return monetaryValue{Asset: left.Asset, Amount: amount}, nil

	case *parser.Prefix:
		if expr.Operator != parser.PrefixOperatorMinus {
			panic("TODO compileMonetaryExpr prefix op " + string(expr.Operator))
		}
		arg, err := st.compileMonetaryExpr(expr.Expr)
		if err != nil {
			return monetaryValue{}, err
		}
		amount := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{Op: ir.OpNegInt{}, Arg: arg.Amount, Dest: dest}
		})
		return monetaryValue{Asset: arg.Asset, Amount: amount}, nil

	case *parser.FnCall:
		return st.compileMonetaryFnCall(expr, false)

	default:
		return utils.NonExhaustiveMatchPanic[monetaryValue](expr), nil
	}
}

// compileMonetaryFnCall handles the builtins that return a monetary. isVarOrigin
// carries the same meaning as in compileFnCall.
func (st *state) compileMonetaryFnCall(expr *parser.FnCall, isVarOrigin bool) (monetaryValue, CompilerError) {
	if !isVarOrigin {
		if err := st.checkFeatureFlag(expr.Range, flags.ExperimentalMidScriptFunctionCall); err != nil {
			return monetaryValue{}, err
		}
	}

	switch expr.Caller.Name {
	case builtins.Balance:
		accountReg, err := st.compileExpr(expr.Args[0])
		if err != nil {
			return monetaryValue{}, err
		}
		assetReg, err := st.compileExpr(expr.Args[1])
		if err != nil {
			return monetaryValue{}, err
		}
		balReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.FetchBalance{Dest: dest, Account: accountReg, Asset: assetReg}
		})
		st.Push(ir.AssertNonNegativeBalance{Balance: balReg, Account: accountReg})
		return monetaryValue{Asset: assetReg, Amount: balReg}, nil

	case builtins.Overdraft:
		if err := st.checkFeatureFlag(expr.Range, flags.ExperimentalOverdraftFunctionFeatureFlag); err != nil {
			return monetaryValue{}, err
		}
		accountReg, err := st.compileExpr(expr.Args[0])
		if err != nil {
			return monetaryValue{}, err
		}
		assetReg, err := st.compileExpr(expr.Args[1])
		if err != nil {
			return monetaryValue{}, err
		}
		balReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.FetchBalance{Dest: dest, Account: accountReg, Asset: assetReg}
		})
		zeroReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.LoadInt{Value: *big.NewInt(0), Dest: dest}
		})
		// overdraft = max(0, -balance) = -min(balance, 0)
		minReg := st.minInt(balReg, zeroReg)
		negReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{Op: ir.OpNegInt{}, Arg: minReg, Dest: dest}
		})
		return monetaryValue{Asset: assetReg, Amount: negReg}, nil

	case builtins.Meta:
		return monetaryValue{}, InvalidMetaPosition{Range: expr.Range}

	default:
		panic("TODO compileMonetaryExpr fn call " + expr.Caller.Name)
	}
}

// compileColor returns nil when the source has no color clause: PullAccount with
// no color pulls the uncolored balance, same as an empty color string.
func (st *state) compileColor(colorExpr parser.ValueExpr) (*ir.Reg, CompilerError) {
	if colorExpr == nil {
		return nil, nil
	}
	if err := st.checkFeatureFlag(colorExpr.GetRange(), flags.ExperimentalAssetColors); err != nil {
		return nil, err
	}
	reg, err := st.compileExpr(colorExpr)
	if err != nil {
		return nil, err
	}
	st.Push(ir.AssertValidColor{Color: reg})
	return &reg, nil
}

// pullFromAccount emits the pull of a source account, including the @world
// check. The account is a register — it can come from a var, an interpolation or
// metadata — so the check cannot be decided here and becomes a run-time branch:
//
//	$eq = str_eq($account, $world)
//	jmp_if_false($eq, #not_world)
//	  $pulled = pull_account(...)              // no overdraft operand: unbounded
//	  jmp(#pull_end)
//	#not_world
//	  $pulled = pull_account(..., overdraft: $od)
//	#pull_end
//
// Both arms write the same dest, which the register typechecker allows because
// the type doesn't change. When overdraftReg is nil the source is unbounded for
// every account, so the two arms would be identical and the branch is skipped.
//
// Deliberately unconditional otherwise, even for a literal @world: one code path
// is easier to trust than a compile-time-folded one, and collapsing it is a
// peephole's job (const-fold str_eq, then drop the dead arm).
func (st *state) pullFromAccount(accReg ir.Reg, capReg, overdraftReg, colorReg *ir.Reg) ir.Reg {
	pull := func(dest ir.Reg, overdraft *ir.Reg) ir.Instr {
		return ir.PullAccount{
			Dest:      dest,
			Account:   accReg,
			Cap:       capReg,
			Overdraft: overdraft,
			Color:     colorReg,
		}
	}

	if overdraftReg == nil {
		return st.PushWithDest(func(dest ir.Reg) ir.Instr { return pull(dest, nil) })
	}

	isWorld := st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.BinaryOp{Op: ir.OpStrEq{}, Left: accReg, Right: st.worldReg, Dest: dest}
	})
	notWorldLabel := st.FreshLabel("not_world")
	endLabel := st.FreshLabel("pull_end")

	st.Push(ir.JmpIfFalse{Cond: isWorld, Target: notWorldLabel})
	// an uncapped context reaches this arm with no cap and no overdraft, which is
	// the InvalidUncappedSource case: taking *all* of an unbounded source
	pulledReg := st.PushWithDest(func(dest ir.Reg) ir.Instr { return pull(dest, nil) })
	st.Push(ir.Jmp{Target: endLabel})

	st.Push(ir.LabelMarker{Label: notWorldLabel})
	st.Push(pull(pulledReg, overdraftReg))
	st.Push(ir.LabelMarker{Label: endLabel})

	return pulledReg
}

// minInt writes min(leftReg, rightReg) into a fresh register.
//
// There is no min opcode: a min is a comparison and a copy, so it is one of each
// plus a branch. Speculatively copying the left operand first saves the `jmp`
// the else arm would otherwise need:
//
//	$min = int_copy($left)
//	$lt = lt_int($left, $right)
//	jmp_if_true($lt, #min_end)      ; left is already the answer
//	$min = int_copy($right)
//	#min_end
//
// That form is only correct because the dest is freshly allocated: an aliased
// dest would clobber $right before the else arm reads it.
func (st *state) minInt(leftReg, rightReg ir.Reg) ir.Reg {
	minReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.UnaryOp{Op: ir.OpIntCopy{}, Arg: leftReg, Dest: dest}
	})
	lt := st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.BinaryOp{Op: ir.OpLtInt{}, Left: leftReg, Right: rightReg, Dest: dest}
	})

	endLabel := st.FreshLabel("min_end")
	st.Push(ir.JmpIfTrue{Cond: lt, Target: endLabel})
	st.Push(ir.UnaryOp{Op: ir.OpIntCopy{}, Arg: rightReg, Dest: minReg})
	st.Push(ir.LabelMarker{Label: endLabel})

	return minReg
}

// jmpIfAmountZero jumps to target when the quantity in amountReg is zero.
//
// The conditional jumps take a bool, so a quantity has to be projected onto one
// first: that costs an instruction, and is exactly what stops a monetary amount
// from being used as a condition by accident (ir.Typecheck rejects it).
func (st *state) jmpIfAmountZero(amountReg ir.Reg, target ir.Label) {
	isZero := st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.UnaryOp{Op: ir.OpIsZero{}, Arg: amountReg, Dest: dest}
	})
	st.Push(ir.JmpIfTrue{Cond: isZero, Target: target})
}

// capReg is the register containing the current cap (or nil if context is uncapped)
// returns (when there's no err) the register where we store the pulled amount of this source
func (st *state) compileSource(
	capReg *ir.Reg,
	src parser.Source,
) (ir.Reg, CompilerError) {
	switch src := src.(type) {
	case *parser.SourceAccount:
		accReg, err := st.compileExpr(src.ValueExpr)
		if err != nil {
			return 0, err
		}

		colorReg, err := st.compileColor(src.Color)
		if err != nil {
			return 0, err
		}

		overdraftReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.LoadInt{
				Value: *big.NewInt(0),
				Dest:  dest,
			}
		})

		return st.pullFromAccount(accReg, capReg, &overdraftReg, colorReg), nil

	case *parser.SourceOverdraft:
		if src.Bounded == nil && capReg == nil {
			return 0, InvalidUncappedSource{
				Range: src.GetRange(),
			}
		}

		accReg, err := st.compileExpr(src.Address)
		if err != nil {
			return 0, err
		}

		colorReg, err := st.compileColor(src.Color)
		if err != nil {
			return 0, err
		}

		var overdraftReg *ir.Reg
		if src.Bounded != nil {
			amtReg, err := st.compileCapAmount(*src.Bounded)
			if err != nil {
				return 0, err
			}
			overdraftReg = &amtReg
		}

		return st.pullFromAccount(accReg, capReg, overdraftReg, colorReg), nil

	case *parser.SourceCapped:
		clauseCapIntReg, err := st.compileCapAmount(src.Cap)
		if err != nil {
			return 0, err
		}

		var innerCapReg ir.Reg
		if capReg == nil {
			innerCapReg = clauseCapIntReg
		} else {
			innerCapReg = st.minInt(clauseCapIntReg, *capReg)
		}

		return st.compileSource(&innerCapReg, src.From)

	case *parser.SourceInorder:
		if capReg == nil {
			inorderTotalReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
				return ir.LoadInt{
					Value: *big.NewInt(0),
					Dest:  dest,
				}
			})
			for _, subSrc := range src.Sources {
				innerPulledAmtReg, err := st.compileSource(nil, subSrc)
				if err != nil {
					return 0, err
				}
				// inorderTotalReg += innerPulledAmtReg
				st.Push(ir.BinaryOp{
					Op:    ir.OpAddInt{},
					Dest:  inorderTotalReg,
					Left:  inorderTotalReg,
					Right: innerPulledAmtReg,
				})
			}
			return inorderTotalReg, nil
		}

		inorderTotalReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.LoadInt{
				Value: *big.NewInt(0),
				Dest:  dest,
			}
		})

		endLabel := st.FreshLabel("inorder_end")
		inorderCap := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{
				Op:   ir.OpIntCopy{},
				Arg:  *capReg,
				Dest: dest,
			}
		})

		for idx, subSrc := range src.Sources {
			innerPulledAmtReg, err := st.compileSource(&inorderCap, subSrc)
			if err != nil {
				return 0, err
			}

			// inorderTotalReg += innerPulledAmtReg
			st.Push(ir.BinaryOp{
				Op:    ir.OpAddInt{},
				Dest:  inorderTotalReg,
				Left:  inorderTotalReg,
				Right: innerPulledAmtReg,
			})

			isLast := idx == len(src.Sources)-1
			if !isLast {
				// inorderCap -= innerPulledAmtReg
				st.Push(ir.BinaryOp{
					Op:    ir.OpSubInt{},
					Dest:  inorderCap,
					Left:  inorderCap,
					Right: innerPulledAmtReg,
				})
				st.jmpIfAmountZero(inorderCap, endLabel)
			}
		}
		st.Push(ir.LabelMarker{
			Label: endLabel,
		})
		return inorderTotalReg, nil

	case *parser.SourceOneof:
		if err := st.checkFeatureFlag(src.GetRange(), flags.ExperimentalOneofFeatureFlag); err != nil {
			return 0, err
		}

		if capReg == nil || len(src.Sources) == 1 {
			return st.compileSource(capReg, src.Sources[0])
		}

		endLabel := st.FreshLabel("oneof_end")

		st.Push(ir.MarkPush{})

		// allocated at first use, not up front, to keep registers numbered in
		// emission order (see ir.Builder.PushWithDest)
		var resultReg ir.Reg

		for index, subSrc := range src.Sources {
			subPulledAmtReg, err := st.compileSource(capReg, subSrc)
			if err != nil {
				return 0, err
			}

			if index == 0 {
				resultReg = st.FreshReg()
			}

			st.Push(ir.UnaryOp{
				Op:   ir.OpIntCopy{},
				Arg:  subPulledAmtReg,
				Dest: resultReg,
			})

			isLast := index == len(src.Sources)-1
			if !isLast {
				// PRE: bounded capReg
				// $missing_amt = $cap - $pulled_amt
				missingAmt := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.BinaryOp{
						Op:    ir.OpSubInt{},
						Left:  *capReg,
						Right: subPulledAmtReg,
						Dest:  dest,
					}
				})

				st.jmpIfAmountZero(missingAmt, endLabel)
				// this branch fell short: undo it and reopen for the next one. There
				// is no rewind-without-closing, so a retry is a close plus a push —
				// and after the rollback the new mark is identical to the closed one.
				st.Push(ir.MarkEnd{Rewind: true})
				st.Push(ir.MarkPush{})
			}
		}

		st.Push(ir.LabelMarker{Label: endLabel})
		// every path into endLabel — the jumps from a branch that covered the cap,
		// and the fallthrough from the last branch — has exactly one region open, so
		// a single commit here closes it once on all of them. Keeping it
		// unconditional at the join is what keeps mark depth a function of position.
		st.Push(ir.MarkEnd{Rewind: false})

		return resultReg, nil

	case *parser.SourceAllotment:
		// an allotment source splits the cap among sub-sources, so it needs one
		if capReg == nil {
			return 0, InvalidUncappedSource{Range: src.GetRange()}
		}
		allotments := make([]parser.AllotmentValue, len(src.Items))
		for i, item := range src.Items {
			allotments[i] = item.Allotment
		}
		shares, err := st.compileAllot(*capReg, allotments)
		if err != nil {
			return 0, err
		}
		// pull exactly its share from each sub-source (tryTakingExact)
		for i, item := range src.Items {
			if _, err := st.compileSourceWithRequiredAmount(shares[i], item.From); err != nil {
				return 0, err
			}
		}
		return *capReg, nil

	case *parser.SourceWithScaling:
		if err := st.checkFeatureFlag(src.GetRange(), flags.AssetScaling); err != nil {
			return 0, err
		}
		return 0, FeatureNotImplemented{Range: src.GetRange(), Feature: "scaling"}

	default:
		return utils.NonExhaustiveMatchPanic[ir.Reg](src), nil
	}
}

func (st *state) compileSourceWithRequiredAmount(
	capReg ir.Reg,
	src parser.Source,
) (ir.Reg, CompilerError) {
	got, err := st.compileSource(&capReg, src)
	if err != nil {
		return 0, err
	}
	st.Push(ir.CheckEnoughFunds{
		Got:    got,
		Needed: capReg,
	})
	return got, nil
}

func (st *state) compileDestination(
	pulledAmtReg ir.Reg,
	currentCap ir.Reg,
	dest parser.Destination,
) CompilerError {
	switch dest := dest.(type) {
	case *parser.DestinationAllotment:
		allotments := make([]parser.AllotmentValue, len(dest.Items))
		for i, item := range dest.Items {
			allotments[i] = item.Allotment
		}
		// split the amount routed to this destination across the portions
		shares, err := st.compileAllot(currentCap, allotments)
		if err != nil {
			return err
		}
		// send each computed share to its target (capped by that exact amount)
		for i, item := range dest.Items {
			if err := st.compileKeptOrDestination(item.To, pulledAmtReg, shares[i]); err != nil {
				return err
			}
		}
		return nil

	case *parser.DestinationOneof:
		if err := st.checkFeatureFlag(dest.GetRange(), flags.ExperimentalOneofFeatureFlag); err != nil {
			return err
		}

		endLabel := st.FreshLabel("oneof_dest_end")

		clauseLabels := make([]ir.Label, len(dest.Clauses))
		for i, clause := range dest.Clauses {
			clauseLabels[i] = st.FreshLabel("oneof_dest_clause")

			capAmtReg, err := st.compileCapAmount(clause.Cap)
			if err != nil {
				return err
			}
			minReg := st.minInt(currentCap, capAmtReg)
			diff := st.PushWithDest(func(dest ir.Reg) ir.Instr {
				return ir.BinaryOp{Op: ir.OpSubInt{}, Left: currentCap, Right: minReg, Dest: dest}
			})
			st.jmpIfAmountZero(diff, clauseLabels[i])
		}

		if err := st.compileKeptOrDestination(dest.Remaining, pulledAmtReg, currentCap); err != nil {
			return err
		}
		st.Push(ir.Jmp{Target: endLabel})

		for i, clause := range dest.Clauses {
			st.Push(ir.LabelMarker{Label: clauseLabels[i]})
			if err := st.compileKeptOrDestination(clause.To, pulledAmtReg, currentCap); err != nil {
				return err
			}
			st.Push(ir.Jmp{Target: endLabel})
		}

		st.Push(ir.LabelMarker{Label: endLabel})
		return nil

	case *parser.DestinationAccount:
		accReg, err := st.compileExpr(dest.ValueExpr)
		if err != nil {
			return err
		}

		var cap *ir.Reg
		if pulledAmtReg != currentCap {
			cap = &currentCap
		}
		st.Push(ir.SendToAccount{
			Account: &accReg,
			Cap:     cap,
		})

	case *parser.DestinationInorder:
		remaining := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{Op: ir.OpIntCopy{}, Arg: currentCap, Dest: dest}
		})
		for _, clause := range dest.Clauses {
			capAmtReg, err := st.compileCapAmount(clause.Cap)
			if err != nil {
				return err
			}
			amtReg := st.minInt(remaining, capAmtReg)
			if err := st.compileKeptOrDestination(clause.To, pulledAmtReg, amtReg); err != nil {
				return err
			}
			st.Push(ir.BinaryOp{Op: ir.OpSubInt{}, Dest: remaining, Left: remaining, Right: amtReg})
		}

		return st.compileKeptOrDestination(dest.Remaining, pulledAmtReg, remaining)

	default:
		utils.NonExhaustiveMatchPanic[any](dest)
	}

	return nil
}

func (st *state) compileKeptOrDestination(
	keptOrDest parser.KeptOrDestination,
	pulledAmtReg ir.Reg,
	currentCap ir.Reg,
) CompilerError {
	switch keptOrDest := keptOrDest.(type) {
	case *parser.DestinationTo:
		return st.compileDestination(pulledAmtReg, currentCap, keptOrDest.Destination)

	case *parser.DestinationKept:
		var cap *ir.Reg
		if pulledAmtReg != currentCap {
			cap = &currentCap
		}
		st.Push(ir.SendToAccount{
			Account: nil,
			Cap:     cap,
		})
		return nil

	default:
		utils.NonExhaustiveMatchPanic[any](keptOrDest)
	}

	return nil
}

func (st *state) compileSentValue(
	sentValue parser.SentValue,
	source parser.Source,
) (ir.Reg, CompilerError) {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueLiteral:
		mon, err := st.compileMonetaryExpr(sentValue.Monetary)
		if err != nil {
			return 0, err
		}
		st.Push(ir.SetCurrentAsset{
			Asset: mon.Asset,
		})
		st.currentAssetReg = &mon.Asset

		return st.compileSourceWithRequiredAmount(mon.Amount, source)

	case *parser.SentValueAll:
		assetReg, err := st.compileExpr(sentValue.Asset)
		if err != nil {
			return 0, err
		}
		st.Push(ir.SetCurrentAsset{
			Asset: assetReg,
		})
		st.currentAssetReg = &assetReg
		return st.compileSource(nil, source)

	default:
		return utils.NonExhaustiveMatchPanic[ir.Reg](sentValue), nil
	}

}

func (st *state) compileStatements(stmt parser.Statement) CompilerError {
	switch stmt := stmt.(type) {
	case *parser.SendStatement:
		pulledAmtReg, err := st.compileSentValue(stmt.SentValue, stmt.Source)
		if err != nil {
			return err
		}

		err = st.compileDestination(pulledAmtReg, pulledAmtReg, stmt.Destination)
		if err != nil {
			return err
		}

		return nil

	case *parser.SaveStatement:
		var assetReg ir.Reg
		var amountReg *ir.Reg
		switch sv := stmt.SentValue.(type) {
		case *parser.SentValueLiteral:
			mon, err := st.compileMonetaryExpr(sv.Monetary)
			if err != nil {
				return err
			}
			assetReg = mon.Asset
			amountReg = &mon.Amount
		case *parser.SentValueAll:
			r, err := st.compileExpr(sv.Asset)
			if err != nil {
				return err
			}
			assetReg = r
		default:
			utils.NonExhaustiveMatchPanic[any](stmt.SentValue)
		}

		accReg, err := st.compileExpr(stmt.Account)
		if err != nil {
			return err
		}
		st.Push(ir.Save{Account: accReg, Asset: assetReg, Amount: amountReg})
		return nil
	case *parser.FnCall:
		switch stmt.Caller.Name {
		case builtins.SetTxMeta:
			key, err := st.compileExpr(stmt.Args[0])
			if err != nil {
				return err
			}
			value, err := st.compileMetaValue(stmt.Args[1])
			if err != nil {
				return err
			}
			st.Push(ir.SetTxMeta{Key: key, Value: value})
			return nil

		case builtins.SetAccountMeta:
			account, err := st.compileExpr(stmt.Args[0])
			if err != nil {
				return err
			}
			key, err := st.compileExpr(stmt.Args[1])
			if err != nil {
				return err
			}
			value, err := st.compileMetaValue(stmt.Args[2])
			if err != nil {
				return err
			}
			st.Push(ir.SetAccountMeta{Account: account, Key: key, Value: value})
			return nil

		default:
			return utils.NonExhaustiveMatchPanic[CompilerError](stmt.Caller.Name)
		}

	default:
		return utils.NonExhaustiveMatchPanic[CompilerError](stmt)
	}
}

// compileMetaValue compiles a value into a string register (metadata is stored
// stringified). Strings/accounts/assets already live in string registers;
// numbers go through int_to_string.
func (st *state) compileMetaValue(expr parser.ValueExpr) (ir.Reg, CompilerError) {
	if st.exprTypes[expr] == typecheck.TypeMonetary {
		mon, err := st.compileMonetaryExpr(expr)
		if err != nil {
			return 0, err
		}
		return st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.BinaryOp{
				Op:    ir.OpMonetaryToString{},
				Left:  mon.Asset,
				Right: mon.Amount,
				Dest:  dest,
			}
		}), nil
	}

	r, err := st.compileExpr(expr)
	if err != nil {
		return 0, err
	}

	switch st.exprTypes[expr] {
	case typecheck.TypeString, typecheck.TypeAccount, typecheck.TypeAsset:
		return r, nil
	case typecheck.TypeNumber:
		return st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{Op: ir.OpIntToString{}, Arg: r, Dest: dest}
		}), nil
	case typecheck.TypePortion:
		return st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{Op: ir.OpPortionToString{}, Arg: r, Dest: dest}
		}), nil
	default:
		panic("TODO meta value of type " + st.exprTypes[expr])
	}
}

func compileProgramToIR(program parser.Program, featureFlags map[string]struct{}) (compiledProgramIR, CompilerError) {
	tc := typecheck.Check(program)
	if len(tc.Errors) > 0 {
		return compiledProgramIR{}, TypeError{Range: tc.Errors[0].Range, Kind: tc.Errors[0].Kind}
	}

	flagSet := maps.Clone(featureFlags)
	if flagSet == nil {
		flagSet = make(map[string]struct{}, len(program.Flags))
	}
	for _, flag := range program.Flags {
		if !slices.Contains(flags.AllFlags, flag.String) {
			return compiledProgramIR{}, InvalidFeature{Range: flag.Range, Feature: flag.String}
		}
		flagSet[flag.String] = struct{}{}
	}

	st := state{vars: map[string]value{}, exprTypes: tc.ExprTypes, featureFlags: flagSet}

	// loaded once, up front, so that it dominates every pullFromAccount branch
	// regardless of the jumps those branches sit between
	st.worldReg = st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.LoadStr{Value: worldAccount, Dest: dest}
	})

	if program.Vars != nil {
		for _, decl := range program.Vars.Declarations {
			if err := st.compileVarDeclaration(decl); err != nil {
				return compiledProgramIR{}, err
			}
		}
	}

	for _, stmt := range program.Statements {
		if err := st.compileStatements(stmt); err != nil {
			return compiledProgramIR{}, err
		}
	}

	return compiledProgramIR{
		instructions: st.Instrs(),
		varsEncoder: VarsEncoder{
			decls: st.varDecls,
			nStr:  st.nextStrVar,
			nInt:  st.nextIntVar,
		},
	}, nil
}

func (st *state) compileVarDeclaration(decl parser.VarDeclaration) CompilerError {
	if decl.Origin == nil {
		st.compileExternalVar(decl)
		return nil
	}
	if decl.Type.Name == typecheck.TypeMonetary {
		if fnCall, ok := (*decl.Origin).(*parser.FnCall); ok {
			if fnCall.Caller.Name == builtins.Meta {
				return st.compileMetaVar(decl, fnCall)
			}
			mon, err := st.compileMonetaryFnCall(fnCall, true)
			if err != nil {
				return err
			}
			st.vars[decl.Name.Name] = monValue(mon)
			return nil
		}
		mon, err := st.compileMonetaryExpr(*decl.Origin)
		if err != nil {
			return err
		}
		st.vars[decl.Name.Name] = monValue(mon)
		return nil
	}

	var r ir.Reg
	var err CompilerError
	if fnCall, ok := (*decl.Origin).(*parser.FnCall); ok {
		// meta() is only supported as a variable origin, statically dispatched on
		// the declared type; elsewhere compileFnCall reports InvalidMetaPosition.
		if fnCall.Caller.Name == builtins.Meta {
			return st.compileMetaVar(decl, fnCall)
		}
		// a call that is the whole origin expression isn't a mid-script call
		r, err = st.compileFnCall(fnCall, true)
	} else {
		r, err = st.compileExpr(*decl.Origin)
	}
	if err != nil {
		return err
	}
	st.vars[decl.Name.Name] = scalarValue(r)
	return nil
}

func (st *state) compileMetaVar(decl parser.VarDeclaration, fnCall *parser.FnCall) CompilerError {
	account, err := st.compileExpr(fnCall.Args[0])
	if err != nil {
		return err
	}
	key, err := st.compileExpr(fnCall.Args[1])
	if err != nil {
		return err
	}

	// monetary is the one meta type whose single store read yields two values, so
	// it has its own two-destination instruction rather than a MetaType.
	if decl.Type.Name == typecheck.TypeMonetary {
		destAsset := st.FreshReg()
		destAmount := st.FreshReg()
		st.Push(ir.MetaMonetary{
			DestAsset:  destAsset,
			DestAmount: destAmount,
			Account:    account,
			Key:        key,
		})
		st.vars[decl.Name.Name] = monValue(monetaryValue{Asset: destAsset, Amount: destAmount})
		return nil
	}

	var typ ir.MetaType
	switch decl.Type.Name {
	case typecheck.TypeString, typecheck.TypeAccount, typecheck.TypeAsset:
		typ = ir.MetaStr{}
	case typecheck.TypeNumber:
		typ = ir.MetaInt{}
	case typecheck.TypePortion:
		typ = ir.MetaPortion{}
	default:
		panic("unexpected meta var type: " + decl.Type.Name)
	}

	st.vars[decl.Name.Name] = scalarValue(st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.MetaVar{Dest: dest, Account: account, Key: key, Typ: typ}
	}))
	return nil
}

// TODO review AI blob
func (st *state) compileExternalVar(decl parser.VarDeclaration) {
	name := decl.Name.Name
	st.varDecls = append(st.varDecls, varDecl{name: name, typ: decl.Type.Name})

	switch decl.Type.Name {
	case typecheck.TypeNumber:
		st.vars[name] = scalarValue(st.loadIntVar())

	case typecheck.TypeString, typecheck.TypeAsset, typecheck.TypeAccount:
		st.vars[name] = scalarValue(st.loadStrVar())

	case typecheck.TypePortion:
		num := st.loadIntVar()
		den := st.loadIntVar()
		st.vars[name] = scalarValue(st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.BinaryOp{Op: ir.OpMakePortion{}, Left: num, Right: den, Dest: dest}
		}))

	case typecheck.TypeMonetary:
		// the vars payload already carries a monetary as two scalars, so the pair
		// is the value — nothing to assemble
		asset := st.loadStrVar()
		amount := st.loadIntVar()
		st.vars[name] = monValue(monetaryValue{Asset: asset, Amount: amount})

	default:
		panic("unexpected var type: " + decl.Type.Name)
	}
}

func (st *state) loadIntVar() ir.Reg {
	index := uint16(st.nextIntVar)
	st.nextIntVar++
	return st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.LoadVar{Dest: dest, Typ: ir.VarInt{}, Index: index}
	})
}

func (st *state) loadStrVar() ir.Reg {
	index := uint16(st.nextStrVar)
	st.nextStrVar++
	return st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.LoadVar{Dest: dest, Typ: ir.VarStr{}, Index: index}
	})
}
