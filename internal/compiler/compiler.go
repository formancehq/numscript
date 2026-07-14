package compiler

import (
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/builtins"
	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/typecheck"
	"github.com/formancehq/numscript/internal/utils"
	"github.com/formancehq/numscript/internal/vm"
)

// Compile lowers a parsed program to the VarsEncoder that turns a json var
// payload into the vm.Vars the program expects, plus the vm.Program itself.
func Compile(program parser.Program) (VarsEncoder, vm.Program, error) {
	compiled, cErr := compileProgramToVirtual(program)
	if cErr != nil {
		return VarsEncoder{}, vm.Program{}, fmt.Errorf("%v", cErr)
	}

	prog, err := assembleProgram(compiled.instructions)
	if err != nil {
		return VarsEncoder{}, vm.Program{}, err
	}

	return compiled.varsEncoder, prog, nil
}

type compiledProgramVirtual struct {
	instructions []vInstr
	varsEncoder  VarsEncoder
}

type state struct {
	nextReg         int
	nextLabelId     int
	instructions    []vInstr
	vars            map[string]reg
	exprTypes       map[parser.ValueExpr]typecheck.Type
	currentAssetReg reg

	nextIntVar int
	nextStrVar int
	varDecls   []varDecl
}

func (st *state) getFreshReg() reg {
	id := st.nextReg
	st.nextReg++
	return reg(id)
}

func (st *state) pushInstruction(instr vInstr) {
	st.instructions = append(st.instructions, instr)
}

func (st *state) getFreshLabel(prefix string) label {
	l := label(fmt.Sprintf("%s_%d", prefix, st.nextLabelId))
	st.nextLabelId++
	return l
}

func (st *state) pushInstructionWithDest(getInstr func(dest reg) vInstr) reg {
	dest := st.getFreshReg()
	st.instructions = append(st.instructions, getInstr(dest))
	return dest
}

func (st *state) pushInstructionWithDestErr(getInstr func(dest reg) vInstr) (reg, CompilerError) {
	return st.pushInstructionWithDest(getInstr), nil
}

func (st *state) compileAllot(amount reg, allotments []parser.AllotmentValue) ([]reg, CompilerError) {
	n := len(allotments)
	portions := make([]reg, n)
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
		leftover = st.pushInstructionWithDest(func(dest reg) vInstr {
			return binaryOp{op: opSubPortion{}, left: prev, right: pi, dest: dest}
		})
	}

	st.pushInstruction(assertLeftover{portion: leftover, exact: remainingIdx == -1})
	if remainingIdx != -1 {
		portions[remainingIdx] = leftover
	}

	dest := make([]reg, n)
	for i := range dest {
		dest[i] = st.getFreshReg()
	}
	st.pushInstruction(makeAllotment{
		dest:     dest,
		amount:   amount,
		portions: portions,
	})
	return dest, nil
}

func (st *state) compileCapAmount(monExpr parser.ValueExpr) (reg, CompilerError) {
	monReg, err := st.compileExpr(monExpr)
	if err != nil {
		return 0, err
	}
	assetReg := st.pushInstructionWithDest(func(dest reg) vInstr {
		return unaryOp{op: opGetAsset{}, arg: monReg, dest: dest}
	})
	st.pushInstruction(assertSameAsset{left: assetReg, right: st.currentAssetReg})
	return st.pushInstructionWithDest(func(dest reg) vInstr {
		return unaryOp{op: opGetAmount{}, arg: monReg, dest: dest}
	}), nil
}

func (st *state) compilePortionOne() reg {
	one := st.pushInstructionWithDest(func(dest reg) vInstr {
		return loadInt{value: *big.NewInt(1), dest: dest}
	})
	return st.pushInstructionWithDest(func(dest reg) vInstr {
		return binaryOp{op: opMakePortion{}, left: one, right: one, dest: dest}
	})
}

func (st *state) compileExpr(expr parser.ValueExpr) (reg, CompilerError) {
	switch expr := expr.(type) {
	case *parser.AssetLiteral:
		return st.pushInstructionWithDestErr(func(reg reg) vInstr {
			return loadStr{
				value: expr.Asset,
				dest:  reg,
			}
		})

	case *parser.StringLiteral:
		return st.pushInstructionWithDestErr(func(reg reg) vInstr {
			return loadStr{
				value: expr.String,
				dest:  reg,
			}
		})

	case *parser.NumberLiteral:
		return st.pushInstructionWithDestErr(func(reg reg) vInstr {
			return loadInt{
				value: *expr.Number,
				dest:  reg,
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

		return st.pushInstructionWithDestErr(func(dest reg) vInstr {
			return binaryOp{
				op:    opMakeMonetary{},
				left:  assetReg,
				right: amtReg,
				dest:  dest,
			}
		})

	case *parser.AccountInterpLiteral:
		var parts []reg
		for _, part := range expr.Parts {
			switch part := part.(type) {
			case parser.AccountTextPart:
				dest := st.pushInstructionWithDest(func(dest reg) vInstr {
					return loadStr{
						value: part.Name,
						dest:  dest,
					}
				})
				parts = append(parts, dest)
			case *parser.Variable:
				r, err := st.compileExpr(part)
				if err != nil {
					return 0, err
				}
				switch t := st.exprTypes[part]; t {
				case typecheck.TypeAccount, typecheck.TypeString:
					parts = append(parts, r)
				case typecheck.TypeNumber:
					parts = append(parts, st.pushInstructionWithDest(func(dest reg) vInstr {
						return unaryOp{op: opIntToString{}, arg: r, dest: dest}
					}))
				default:
					return 0, CannotCastToString{Range: part.GetRange(), Type: t}
				}
			}
		}

		acc := parts[0]
		for _, part := range parts[1:] {
			left, right := acc, part
			acc = st.pushInstructionWithDest(func(dest reg) vInstr {
				return binaryOp{op: opAddString{}, left: left, right: right, dest: dest}
			})
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
		numReg := st.pushInstructionWithDest(func(dest reg) vInstr {
			return loadInt{value: *ratio.Num(), dest: dest}
		})
		denReg := st.pushInstructionWithDest(func(dest reg) vInstr {
			return loadInt{value: *ratio.Denom(), dest: dest}
		})
		return st.pushInstructionWithDestErr(func(dest reg) vInstr {
			return binaryOp{op: opMakePortion{}, left: numReg, right: denReg, dest: dest}
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
			return st.pushInstructionWithDestErr(func(dest reg) vInstr {
				return binaryOp{op: opMakePortion{}, left: leftReg, right: rightReg, dest: dest}
			})

		case parser.InfixOperatorPlus:
			switch st.exprTypes[expr.Left] {
			case typecheck.TypeNumber:
				return st.pushInstructionWithDestErr(func(dest reg) vInstr {
					return binaryOp{op: opAddInt{}, left: leftReg, right: rightReg, dest: dest}
				})

			case typecheck.TypeMonetary:
				lAsset := st.pushInstructionWithDest(func(dest reg) vInstr {
					return unaryOp{op: opGetAsset{}, arg: leftReg, dest: dest}
				})
				rAsset := st.pushInstructionWithDest(func(dest reg) vInstr {
					return unaryOp{op: opGetAsset{}, arg: rightReg, dest: dest}
				})
				st.pushInstruction(assertSameAsset{left: lAsset, right: rAsset})

				lAmt := st.pushInstructionWithDest(func(dest reg) vInstr {
					return unaryOp{op: opGetAmount{}, arg: leftReg, dest: dest}
				})
				rAmt := st.pushInstructionWithDest(func(dest reg) vInstr {
					return unaryOp{op: opGetAmount{}, arg: rightReg, dest: dest}
				})
				sum := st.pushInstructionWithDest(func(dest reg) vInstr {
					return binaryOp{op: opAddInt{}, left: lAmt, right: rAmt, dest: dest}
				})
				return st.pushInstructionWithDestErr(func(dest reg) vInstr {
					return binaryOp{op: opMakeMonetary{}, left: lAsset, right: sum, dest: dest}
				})

			default:
				panic("TODO compileExpr + for unexpected type")

			}

		case parser.InfixOperatorMinus:
			switch st.exprTypes[expr.Left] {
			case typecheck.TypeNumber:
				return st.pushInstructionWithDestErr(func(dest reg) vInstr {
					return binaryOp{op: opSubInt{}, left: leftReg, right: rightReg, dest: dest}
				})

			case typecheck.TypeMonetary:
				lAsset := st.pushInstructionWithDest(func(dest reg) vInstr {
					return unaryOp{op: opGetAsset{}, arg: leftReg, dest: dest}
				})
				rAsset := st.pushInstructionWithDest(func(dest reg) vInstr {
					return unaryOp{op: opGetAsset{}, arg: rightReg, dest: dest}
				})
				st.pushInstruction(assertSameAsset{left: lAsset, right: rAsset})

				lAmt := st.pushInstructionWithDest(func(dest reg) vInstr {
					return unaryOp{op: opGetAmount{}, arg: leftReg, dest: dest}
				})
				rAmt := st.pushInstructionWithDest(func(dest reg) vInstr {
					return unaryOp{op: opGetAmount{}, arg: rightReg, dest: dest}
				})
				diff := st.pushInstructionWithDest(func(dest reg) vInstr {
					return binaryOp{op: opSubInt{}, left: lAmt, right: rAmt, dest: dest}
				})
				return st.pushInstructionWithDestErr(func(dest reg) vInstr {
					return binaryOp{op: opMakeMonetary{}, left: lAsset, right: diff, dest: dest}
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
				return st.pushInstructionWithDestErr(func(dest reg) vInstr {
					return unaryOp{op: opNegInt{}, arg: argReg, dest: dest}
				})

			case typecheck.TypeMonetary:
				amt := st.pushInstructionWithDest(func(dest reg) vInstr {
					return unaryOp{op: opGetAmount{}, arg: argReg, dest: dest}
				})
				negAmt := st.pushInstructionWithDest(func(dest reg) vInstr {
					return unaryOp{op: opNegInt{}, arg: amt, dest: dest}
				})
				asset := st.pushInstructionWithDest(func(dest reg) vInstr {
					return unaryOp{op: opGetAsset{}, arg: argReg, dest: dest}
				})
				return st.pushInstructionWithDestErr(func(dest reg) vInstr {
					return binaryOp{op: opMakeMonetary{}, left: asset, right: negAmt, dest: dest}
				})

			default:
				panic("TODO compileExpr prefix - for unexpected type")
			}

		default:
			panic("TODO compileExpr prefix op " + string(expr.Operator))
		}

	case *parser.FnCall:
		switch expr.Caller.Name {
		case builtins.GetAmount:
			argReg, err := st.compileExpr(expr.Args[0])
			if err != nil {
				return 0, err
			}
			return st.pushInstructionWithDestErr(func(dest reg) vInstr {
				return unaryOp{op: opGetAmount{}, arg: argReg, dest: dest}
			})

		case builtins.GetAsset:
			argReg, err := st.compileExpr(expr.Args[0])
			if err != nil {
				return 0, err
			}
			return st.pushInstructionWithDestErr(func(dest reg) vInstr {
				return unaryOp{op: opGetAsset{}, arg: argReg, dest: dest}
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
			return st.pushInstructionWithDestErr(func(dest reg) vInstr {
				return fetchBalance{dest: dest, account: accountReg, asset: assetReg}
			})

		case builtins.Meta:
			return 0, InvalidMetaPosition{Range: expr.Range}

		default:
			panic("TODO compileExpr fn call " + expr.Caller.Name)
		}

	default:
		return utils.NonExhaustiveMatchPanic[reg](expr), nil
	}
}

// capReg is the register containing the current cap (or nil if context is uncapped)
// returns (when there's no err) the register where we store the pulled amount of this source
func (st *state) compileSource(
	capReg *reg,
	src parser.Source,
) (reg, CompilerError) {
	switch src := src.(type) {
	case *parser.SourceAccount:
		if src.Color != nil {
			panic("TODO impl color")
		}

		accReg, err := st.compileExpr(src.ValueExpr)
		if err != nil {
			return 0, err
		}

		overdraftReg := st.pushInstructionWithDest(func(dest reg) vInstr {
			return loadInt{
				value: *big.NewInt(0),
				dest:  dest,
			}
		})

		return st.pushInstructionWithDestErr(func(dest reg) vInstr {
			return pullAccount{
				dest:      dest,
				account:   accReg,
				cap:       capReg,
				overdraft: &overdraftReg,
				color:     nil,
			}
		})

	case *parser.SourceOverdraft:
		if src.Color != nil {
			panic("TODO impl color")
		}

		if src.Bounded == nil && capReg == nil {
			return 0, InvalidUncappedSource{
				Range: src.GetRange(),
			}
		}

		accReg, err := st.compileExpr(src.Address)
		if err != nil {
			return 0, err
		}

		var overdraftReg *reg
		if src.Bounded != nil {
			amtReg, err := st.compileCapAmount(*src.Bounded)
			if err != nil {
				return 0, err
			}
			overdraftReg = &amtReg
		}

		return st.pushInstructionWithDestErr(func(dest reg) vInstr {
			return pullAccount{
				dest:      dest,
				account:   accReg,
				cap:       capReg,
				overdraft: overdraftReg,
				color:     nil,
			}
		})

	case *parser.SourceCapped:
		clauseCapIntReg, err := st.compileCapAmount(src.Cap)
		if err != nil {
			return 0, err
		}

		var innerCapReg reg
		if capReg == nil {
			innerCapReg = clauseCapIntReg
		} else {
			minReg := st.pushInstructionWithDest(func(dest reg) vInstr {
				return binaryOp{
					op:    opMinInt{},
					left:  clauseCapIntReg,
					right: *capReg,
					dest:  dest,
				}
			})
			innerCapReg = minReg
		}

		return st.compileSource(&innerCapReg, src.From)

	case *parser.SourceInorder:
		if capReg == nil {
			inorderTotalReg := st.pushInstructionWithDest(func(dest reg) vInstr {
				return loadInt{
					value: *big.NewInt(0),
					dest:  dest,
				}
			})
			for _, subSrc := range src.Sources {
				innerPulledAmtReg, err := st.compileSource(nil, subSrc)
				if err != nil {
					return 0, err
				}
				// inorderTotalReg += innerPulledAmtReg
				st.pushInstruction(binaryOp{
					op:    opAddInt{},
					dest:  inorderTotalReg,
					left:  inorderTotalReg,
					right: innerPulledAmtReg,
				})
			}
			return inorderTotalReg, nil
		}

		inorderTotalReg := st.pushInstructionWithDest(func(dest reg) vInstr {
			return loadInt{
				value: *big.NewInt(0),
				dest:  dest,
			}
		})

		endLabel := st.getFreshLabel("inorder_end")
		inorderCap := st.pushInstructionWithDest(func(dest reg) vInstr {
			return unaryOp{
				op:   opIntCopy{},
				arg:  *capReg,
				dest: dest,
			}
		})

		for idx, subSrc := range src.Sources {
			innerPulledAmtReg, err := st.compileSource(&inorderCap, subSrc)
			if err != nil {
				return 0, err
			}

			// inorderTotalReg += innerPulledAmtReg
			st.pushInstruction(binaryOp{
				op:    opAddInt{},
				dest:  inorderTotalReg,
				left:  inorderTotalReg,
				right: innerPulledAmtReg,
			})

			isLast := idx == len(src.Sources)-1
			if !isLast {
				// inorderCap -= innerPulledAmtReg
				st.pushInstruction(binaryOp{
					op:    opSubInt{},
					dest:  inorderCap,
					left:  inorderCap,
					right: innerPulledAmtReg,
				})
				st.pushInstruction(jmpIfZero{
					cond:   inorderCap,
					target: endLabel,
				})
			}
		}
		st.pushInstruction(labelMarker{
			label: endLabel,
		})
		return inorderTotalReg, nil

	case *parser.SourceOneof:
		panic("TODO impl source")

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
		panic("TODO impl source")

	default:
		return utils.NonExhaustiveMatchPanic[reg](src), nil
	}
}

func (st *state) compileSourceWithRequiredAmount(
	capReg reg,
	src parser.Source,
) (reg, CompilerError) {
	got, err := st.compileSource(&capReg, src)
	if err != nil {
		return 0, err
	}
	st.pushInstruction(checkEnoughFunds{
		got:    got,
		needed: capReg,
	})
	return got, nil
}

func (st *state) compileDestination(
	pulledAmtReg reg,
	currentCap reg,
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
		panic("TODO unimplemented")

	case *parser.DestinationAccount:
		accReg, err := st.compileExpr(dest.ValueExpr)
		if err != nil {
			return err
		}

		var cap *reg
		if pulledAmtReg != currentCap {
			cap = &currentCap
		}
		st.pushInstruction(sendToAccount{
			account: &accReg,
			cap:     cap,
		})

	case *parser.DestinationInorder:
		remaining := st.pushInstructionWithDest(func(dest reg) vInstr {
			return unaryOp{op: opIntCopy{}, arg: currentCap, dest: dest}
		})
		for _, clause := range dest.Clauses {
			capAmtReg, err := st.compileCapAmount(clause.Cap)
			if err != nil {
				return err
			}
			amtReg := st.pushInstructionWithDest(func(dest reg) vInstr {
				return binaryOp{op: opMinInt{}, left: remaining, right: capAmtReg, dest: dest}
			})
			if err := st.compileKeptOrDestination(clause.To, pulledAmtReg, amtReg); err != nil {
				return err
			}
			st.pushInstruction(binaryOp{op: opSubInt{}, dest: remaining, left: remaining, right: amtReg})
		}

		return st.compileKeptOrDestination(dest.Remaining, pulledAmtReg, remaining)

	default:
		utils.NonExhaustiveMatchPanic[any](dest)
	}

	return nil
}

func (st *state) compileKeptOrDestination(
	keptOrDest parser.KeptOrDestination,
	pulledAmtReg reg,
	currentCap reg,
) CompilerError {
	switch keptOrDest := keptOrDest.(type) {
	case *parser.DestinationTo:
		return st.compileDestination(pulledAmtReg, currentCap, keptOrDest.Destination)

	case *parser.DestinationKept:
		var cap *reg
		if pulledAmtReg != currentCap {
			cap = &currentCap
		}
		st.pushInstruction(sendToAccount{
			account: nil,
			cap:     cap,
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
) (reg, CompilerError) {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueLiteral:
		monetaryReg, err := st.compileExpr(sentValue.Monetary)
		if err != nil {
			return 0, err
		}
		assetReg := st.pushInstructionWithDest(func(dest reg) vInstr {
			return unaryOp{
				op:   opGetAsset{},
				arg:  monetaryReg,
				dest: dest,
			}
		})
		st.pushInstruction(setCurrentAsset{
			asset: assetReg,
		})
		st.currentAssetReg = assetReg
		capReg := st.pushInstructionWithDest(func(dest reg) vInstr {
			return unaryOp{
				op:   opGetAmount{},
				arg:  monetaryReg,
				dest: dest,
			}
		})

		return st.compileSourceWithRequiredAmount(capReg, source)

	case *parser.SentValueAll:
		assetReg, err := st.compileExpr(sentValue.Asset)
		if err != nil {
			return 0, err
		}
		st.pushInstruction(setCurrentAsset{
			asset: assetReg,
		})
		st.currentAssetReg = assetReg
		return st.compileSource(nil, source)

	default:
		return utils.NonExhaustiveMatchPanic[reg](sentValue), nil
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
		var assetReg reg
		var amountReg *reg
		switch sv := stmt.SentValue.(type) {
		case *parser.SentValueLiteral:
			monReg, err := st.compileExpr(sv.Monetary)
			if err != nil {
				return err
			}
			assetReg = st.pushInstructionWithDest(func(dest reg) vInstr {
				return unaryOp{op: opGetAsset{}, arg: monReg, dest: dest}
			})
			amt := st.pushInstructionWithDest(func(dest reg) vInstr {
				return unaryOp{op: opGetAmount{}, arg: monReg, dest: dest}
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
		st.pushInstruction(save{account: accReg, asset: assetReg, amount: amountReg})
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
			st.pushInstruction(setTxMeta{key: key, value: value})
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
			st.pushInstruction(setAccountMeta{account: account, key: key, value: value})
			return nil

		default:
			panic("TODO fn call statement: " + stmt.Caller.Name)
		}

	default:
		return utils.NonExhaustiveMatchPanic[CompilerError](stmt)
	}
}

// compileMetaValue compiles a value into a string register (metadata is stored
// stringified). Strings/accounts/assets already live in string registers;
// numbers go through int_to_string.
func (st *state) compileMetaValue(expr parser.ValueExpr) (reg, CompilerError) {
	r, err := st.compileExpr(expr)
	if err != nil {
		return 0, err
	}

	switch st.exprTypes[expr] {
	case typecheck.TypeString, typecheck.TypeAccount, typecheck.TypeAsset:
		return r, nil
	case typecheck.TypeNumber:
		return st.pushInstructionWithDest(func(dest reg) vInstr {
			return unaryOp{op: opIntToString{}, arg: r, dest: dest}
		}), nil
	case typecheck.TypePortion:
		return st.pushInstructionWithDest(func(dest reg) vInstr {
			return unaryOp{op: opPortionToString{}, arg: r, dest: dest}
		}), nil
	case typecheck.TypeMonetary:
		return st.pushInstructionWithDest(func(dest reg) vInstr {
			return unaryOp{op: opMonetaryToString{}, arg: r, dest: dest}
		}), nil
	default:
		panic("TODO meta value of type " + st.exprTypes[expr])
	}
}

func compileProgramToVirtual(program parser.Program) (compiledProgramVirtual, CompilerError) {
	tc := typecheck.Check(program)
	if len(tc.Errors) > 0 {
		return compiledProgramVirtual{}, TypeError{Range: tc.Errors[0].Range, Kind: tc.Errors[0].Kind}
	}

	st := state{vars: map[string]reg{}, exprTypes: tc.ExprTypes}

	if program.Vars != nil {
		for _, decl := range program.Vars.Declarations {
			if err := st.compileVarDeclaration(decl); err != nil {
				return compiledProgramVirtual{}, err
			}
		}
	}

	for _, stmt := range program.Statements {
		if err := st.compileStatements(stmt); err != nil {
			return compiledProgramVirtual{}, err
		}
	}

	return compiledProgramVirtual{
		instructions: st.instructions,
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
	// meta() is only supported as a variable origin, statically dispatched on
	// the declared type; elsewhere compileExpr reports InvalidMetaPosition.
	if fnCall, ok := (*decl.Origin).(*parser.FnCall); ok && fnCall.Caller.Name == builtins.Meta {
		return st.compileMetaVar(decl, fnCall)
	}
	r, err := st.compileExpr(*decl.Origin)
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

	var typ metaType
	switch decl.Type.Name {
	case typecheck.TypeString, typecheck.TypeAccount, typecheck.TypeAsset:
		typ = metaStr{}
	case typecheck.TypeNumber:
		typ = metaInt{}
	case typecheck.TypePortion:
		typ = metaPortion{}
	case typecheck.TypeMonetary:
		typ = metaMonetary{}
	default:
		panic("unexpected meta var type: " + decl.Type.Name)
	}

	st.vars[decl.Name.Name] = st.pushInstructionWithDest(func(dest reg) vInstr {
		return metaVar{dest: dest, account: account, key: key, typ: typ}
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
		st.vars[name] = st.pushInstructionWithDest(func(dest reg) vInstr {
			return binaryOp{op: opMakePortion{}, left: num, right: den, dest: dest}
		})

	case typecheck.TypeMonetary:
		asset := st.loadStrVar()
		amount := st.loadIntVar()
		st.vars[name] = st.pushInstructionWithDest(func(dest reg) vInstr {
			return binaryOp{op: opMakeMonetary{}, left: asset, right: amount, dest: dest}
		})

	default:
		panic("unexpected var type: " + decl.Type.Name)
	}
}

func (st *state) loadIntVar() reg {
	index := uint16(st.nextIntVar)
	st.nextIntVar++
	return st.pushInstructionWithDest(func(dest reg) vInstr {
		return loadVar{dest: dest, typ: varInt{}, index: index}
	})
}

func (st *state) loadStrVar() reg {
	index := uint16(st.nextStrVar)
	st.nextStrVar++
	return st.pushInstructionWithDest(func(dest reg) vInstr {
		return loadVar{dest: dest, typ: varStr{}, index: index}
	})
}
