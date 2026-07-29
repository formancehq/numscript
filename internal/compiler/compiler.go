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

	vars            map[string]ir.Reg
	exprTypes       map[parser.ValueExpr]typecheck.Type
	featureFlags    map[string]struct{}
	currentAssetReg ir.Reg

	nextIntVar int
	nextStrVar int
	varDecls   []varDecl
}

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

	dest := make([]ir.Reg, n)
	for i := range dest {
		dest[i] = st.FreshReg()
	}
	st.Push(ir.MakeAllotment{
		Dest:     dest,
		Amount:   amount,
		Portions: portions,
	})
	return dest, nil
}

func (st *state) compileCapAmount(monExpr parser.ValueExpr) (ir.Reg, CompilerError) {
	monReg, err := st.compileExpr(monExpr)
	if err != nil {
		return 0, err
	}
	assetReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.UnaryOp{Op: ir.OpGetAsset{}, Arg: monReg, Dest: dest}
	})
	st.Push(ir.AssertSameAsset{Left: assetReg, Right: st.currentAssetReg})
	return st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.UnaryOp{Op: ir.OpGetAmount{}, Arg: monReg, Dest: dest}
	}), nil
}

func (st *state) compilePortionOne() ir.Reg {
	one := st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.LoadInt{Value: *big.NewInt(1), Dest: dest}
	})
	return st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.BinaryOp{Op: ir.OpMakePortion{}, Left: one, Right: one, Dest: dest}
	})
}

func (st *state) compileExpr(expr parser.ValueExpr) (ir.Reg, CompilerError) {
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

	case *parser.MonetaryLiteral:
		assetReg, err := st.compileExpr(expr.Asset)
		if err != nil {
			return 0, err
		}

		amtReg, err := st.compileExpr(expr.Amount)
		if err != nil {
			return 0, err
		}

		return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
			return ir.BinaryOp{
				Op:    ir.OpMakeMonetary{},
				Left:  assetReg,
				Right: amtReg,
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
				r, err := st.compileExpr(part)
				if err != nil {
					return 0, err
				}
				switch t := st.exprTypes[part]; t {
				case typecheck.TypeAccount, typecheck.TypeString:
					parts = append(parts, r)
				case typecheck.TypeNumber:
					parts = append(parts, st.PushWithDest(func(dest ir.Reg) ir.Instr {
						return ir.UnaryOp{Op: ir.OpIntToString{}, Arg: r, Dest: dest}
					}))
				default:
					return 0, CannotCastToString{Range: part.GetRange(), Type: t}
				}
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
		r, ok := st.vars[expr.Name]
		if !ok {
			return 0, UnboundVar{Range: expr.Range, Var: expr.Name}
		}
		return r, nil

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
			switch st.exprTypes[expr.Left] {
			case typecheck.TypeNumber:
				return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
					return ir.BinaryOp{Op: ir.OpAddInt{}, Left: leftReg, Right: rightReg, Dest: dest}
				})

			case typecheck.TypeMonetary:
				lAsset := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpGetAsset{}, Arg: leftReg, Dest: dest}
				})
				rAsset := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpGetAsset{}, Arg: rightReg, Dest: dest}
				})
				st.Push(ir.AssertSameAsset{Left: lAsset, Right: rAsset})

				lAmt := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpGetAmount{}, Arg: leftReg, Dest: dest}
				})
				rAmt := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpGetAmount{}, Arg: rightReg, Dest: dest}
				})
				sum := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.BinaryOp{Op: ir.OpAddInt{}, Left: lAmt, Right: rAmt, Dest: dest}
				})
				return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
					return ir.BinaryOp{Op: ir.OpMakeMonetary{}, Left: lAsset, Right: sum, Dest: dest}
				})

			default:
				panic("TODO compileExpr + for unexpected type")

			}

		case parser.InfixOperatorMinus:
			switch st.exprTypes[expr.Left] {
			case typecheck.TypeNumber:
				return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
					return ir.BinaryOp{Op: ir.OpSubInt{}, Left: leftReg, Right: rightReg, Dest: dest}
				})

			case typecheck.TypeMonetary:
				lAsset := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpGetAsset{}, Arg: leftReg, Dest: dest}
				})
				rAsset := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpGetAsset{}, Arg: rightReg, Dest: dest}
				})
				st.Push(ir.AssertSameAsset{Left: lAsset, Right: rAsset})

				lAmt := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpGetAmount{}, Arg: leftReg, Dest: dest}
				})
				rAmt := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpGetAmount{}, Arg: rightReg, Dest: dest}
				})
				diff := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.BinaryOp{Op: ir.OpSubInt{}, Left: lAmt, Right: rAmt, Dest: dest}
				})
				return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
					return ir.BinaryOp{Op: ir.OpMakeMonetary{}, Left: lAsset, Right: diff, Dest: dest}
				})

			default:
				panic("TODO compileExpr - for unexpected type")

			}

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
			switch st.exprTypes[expr.Expr] {
			case typecheck.TypeNumber:
				return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpNegInt{}, Arg: argReg, Dest: dest}
				})

			case typecheck.TypeMonetary:
				amt := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpGetAmount{}, Arg: argReg, Dest: dest}
				})
				negAmt := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpNegInt{}, Arg: amt, Dest: dest}
				})
				asset := st.PushWithDest(func(dest ir.Reg) ir.Instr {
					return ir.UnaryOp{Op: ir.OpGetAsset{}, Arg: argReg, Dest: dest}
				})
				return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
					return ir.BinaryOp{Op: ir.OpMakeMonetary{}, Left: asset, Right: negAmt, Dest: dest}
				})

			default:
				panic("TODO compileExpr prefix - for unexpected type")
			}

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
		argReg, err := st.compileExpr(expr.Args[0])
		if err != nil {
			return 0, err
		}
		return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{Op: ir.OpGetAmount{}, Arg: argReg, Dest: dest}
		})

	case builtins.GetAsset:
		if err := st.checkFeatureFlag(expr.Range, flags.ExperimentalGetAssetFunctionFeatureFlag); err != nil {
			return 0, err
		}
		argReg, err := st.compileExpr(expr.Args[0])
		if err != nil {
			return 0, err
		}
		return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{Op: ir.OpGetAsset{}, Arg: argReg, Dest: dest}
		})

	case builtins.Balance:
		accountReg, err := st.compileExpr(expr.Args[0])
		if err != nil {
			return 0, err
		}
		assetReg, err := st.compileExpr(expr.Args[1])
		if err != nil {
			return 0, err
		}
		balReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.FetchBalance{Dest: dest, Account: accountReg, Asset: assetReg}
		})
		st.Push(ir.AssertNonNegativeBalance{Balance: balReg, Account: accountReg})
		return balReg, nil

	case builtins.Overdraft:
		if err := st.checkFeatureFlag(expr.Range, flags.ExperimentalOverdraftFunctionFeatureFlag); err != nil {
			return 0, err
		}
		accountReg, err := st.compileExpr(expr.Args[0])
		if err != nil {
			return 0, err
		}
		assetReg, err := st.compileExpr(expr.Args[1])
		if err != nil {
			return 0, err
		}
		balReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.FetchBalance{Dest: dest, Account: accountReg, Asset: assetReg}
		})
		amtReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{Op: ir.OpGetAmount{}, Arg: balReg, Dest: dest}
		})
		zeroReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.LoadInt{Value: *big.NewInt(0), Dest: dest}
		})
		minReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.BinaryOp{Op: ir.OpMinInt{}, Left: amtReg, Right: zeroReg, Dest: dest}
		})
		// overdraft = max(0, -balance) = -min(balance, 0)
		negReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{Op: ir.OpNegInt{}, Arg: minReg, Dest: dest}
		})
		return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
			return ir.BinaryOp{Op: ir.OpMakeMonetary{}, Left: assetReg, Right: negReg, Dest: dest}
		})

	case builtins.Meta:
		return 0, InvalidMetaPosition{Range: expr.Range}

	default:
		panic("TODO compileExpr fn call " + expr.Caller.Name)
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

		return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
			return ir.PullAccount{
				Dest:      dest,
				Account:   accReg,
				Cap:       capReg,
				Overdraft: &overdraftReg,
				Color:     colorReg,
			}
		})

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

		return st.pushInstructionWithDestErr(func(dest ir.Reg) ir.Instr {
			return ir.PullAccount{
				Dest:      dest,
				Account:   accReg,
				Cap:       capReg,
				Overdraft: overdraftReg,
				Color:     colorReg,
			}
		})

	case *parser.SourceCapped:
		clauseCapIntReg, err := st.compileCapAmount(src.Cap)
		if err != nil {
			return 0, err
		}

		var innerCapReg ir.Reg
		if capReg == nil {
			innerCapReg = clauseCapIntReg
		} else {
			minReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
				return ir.BinaryOp{
					Op:    ir.OpMinInt{},
					Left:  clauseCapIntReg,
					Right: *capReg,
					Dest:  dest,
				}
			})
			innerCapReg = minReg
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
				st.Push(ir.JmpIfZero{
					Cond:   inorderCap,
					Target: endLabel,
				})
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

		snapshotReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.Snapshot{Dest: dest}
		})

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

				st.Push(ir.JmpIfZero{
					Cond:   missingAmt,
					Target: endLabel,
				})
				st.Push(ir.Restore{
					Mark: snapshotReg,
				})
			}
		}

		st.Push(ir.LabelMarker{Label: endLabel})

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

		zero := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.LoadInt{Value: *big.NewInt(0), Dest: dest}
		})

		clauseLabels := make([]ir.Label, len(dest.Clauses))
		for i, clause := range dest.Clauses {
			clauseLabels[i] = st.FreshLabel("oneof_dest_clause")

			capAmtReg, err := st.compileCapAmount(clause.Cap)
			if err != nil {
				return err
			}
			minReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
				return ir.BinaryOp{Op: ir.OpMinInt{}, Left: currentCap, Right: capAmtReg, Dest: dest}
			})
			diff := st.PushWithDest(func(dest ir.Reg) ir.Instr {
				return ir.BinaryOp{Op: ir.OpSubInt{}, Left: currentCap, Right: minReg, Dest: dest}
			})
			st.Push(ir.JmpIfZero{Cond: diff, Target: clauseLabels[i]})
		}

		if err := st.compileKeptOrDestination(dest.Remaining, pulledAmtReg, currentCap); err != nil {
			return err
		}
		st.Push(ir.JmpIfZero{Cond: zero, Target: endLabel})

		for i, clause := range dest.Clauses {
			st.Push(ir.LabelMarker{Label: clauseLabels[i]})
			if err := st.compileKeptOrDestination(clause.To, pulledAmtReg, currentCap); err != nil {
				return err
			}
			st.Push(ir.JmpIfZero{Cond: zero, Target: endLabel})
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
			amtReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
				return ir.BinaryOp{Op: ir.OpMinInt{}, Left: remaining, Right: capAmtReg, Dest: dest}
			})
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
		monetaryReg, err := st.compileExpr(sentValue.Monetary)
		if err != nil {
			return 0, err
		}
		assetReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{
				Op:   ir.OpGetAsset{},
				Arg:  monetaryReg,
				Dest: dest,
			}
		})
		st.Push(ir.SetCurrentAsset{
			Asset: assetReg,
		})
		st.currentAssetReg = assetReg
		capReg := st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{
				Op:   ir.OpGetAmount{},
				Arg:  monetaryReg,
				Dest: dest,
			}
		})

		return st.compileSourceWithRequiredAmount(capReg, source)

	case *parser.SentValueAll:
		assetReg, err := st.compileExpr(sentValue.Asset)
		if err != nil {
			return 0, err
		}
		st.Push(ir.SetCurrentAsset{
			Asset: assetReg,
		})
		st.currentAssetReg = assetReg
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
			monReg, err := st.compileExpr(sv.Monetary)
			if err != nil {
				return err
			}
			assetReg = st.PushWithDest(func(dest ir.Reg) ir.Instr {
				return ir.UnaryOp{Op: ir.OpGetAsset{}, Arg: monReg, Dest: dest}
			})
			amt := st.PushWithDest(func(dest ir.Reg) ir.Instr {
				return ir.UnaryOp{Op: ir.OpGetAmount{}, Arg: monReg, Dest: dest}
			})
			amountReg = &amt
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
	case typecheck.TypeMonetary:
		return st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.UnaryOp{Op: ir.OpMonetaryToString{}, Arg: r, Dest: dest}
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

	st := state{vars: map[string]ir.Reg{}, exprTypes: tc.ExprTypes, featureFlags: flagSet}

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
	st.vars[decl.Name.Name] = r
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

	var typ ir.MetaType
	switch decl.Type.Name {
	case typecheck.TypeString, typecheck.TypeAccount, typecheck.TypeAsset:
		typ = ir.MetaStr{}
	case typecheck.TypeNumber:
		typ = ir.MetaInt{}
	case typecheck.TypePortion:
		typ = ir.MetaPortion{}
	case typecheck.TypeMonetary:
		typ = ir.MetaMonetary{}
	default:
		panic("unexpected meta var type: " + decl.Type.Name)
	}

	st.vars[decl.Name.Name] = st.PushWithDest(func(dest ir.Reg) ir.Instr {
		return ir.MetaVar{Dest: dest, Account: account, Key: key, Typ: typ}
	})
	return nil
}

// TODO review AI blob
func (st *state) compileExternalVar(decl parser.VarDeclaration) {
	name := decl.Name.Name
	st.varDecls = append(st.varDecls, varDecl{name: name, typ: decl.Type.Name})

	switch decl.Type.Name {
	case typecheck.TypeNumber:
		st.vars[name] = st.loadIntVar()

	case typecheck.TypeString, typecheck.TypeAsset, typecheck.TypeAccount:
		st.vars[name] = st.loadStrVar()

	case typecheck.TypePortion:
		num := st.loadIntVar()
		den := st.loadIntVar()
		st.vars[name] = st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.BinaryOp{Op: ir.OpMakePortion{}, Left: num, Right: den, Dest: dest}
		})

	case typecheck.TypeMonetary:
		asset := st.loadStrVar()
		amount := st.loadIntVar()
		st.vars[name] = st.PushWithDest(func(dest ir.Reg) ir.Instr {
			return ir.BinaryOp{Op: ir.OpMakeMonetary{}, Left: asset, Right: amount, Dest: dest}
		})

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
