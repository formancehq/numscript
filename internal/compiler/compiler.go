package compiler

import (
	"fmt"
	"math/big"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/typecheck"
	"github.com/formancehq/numscript/internal/utils"
)

type compiledProgramVirtual struct {
	instructions []vInstr
}

type state struct {
	nextReg      int
	nextLabelId  int
	instructions []vInstr
	vars         map[string]reg
	exprTypes    map[parser.ValueExpr]typecheck.Type
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
	st.pushInstruction(checkEqCurrentAsset{got: assetReg})
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
				panic("TODO interp var")
			}
		}

		if len(parts) == 1 {
			return parts[0], nil
		}

		panic("TODO compileExpr interp of many segments")

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

			default:
				panic("TODO compileExpr + for non-int")

			}

		default:
			panic("TODO compileExpr binary op " + string(expr.Operator))
		}

	case *parser.Prefix:
		panic("TODO compileExpr")

	case *parser.FnCall:
		panic("TODO compileExpr")

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
			panic("TODO unbounded inorder")
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
		panic("TODO fn call")

	default:
		return utils.NonExhaustiveMatchPanic[CompilerError](stmt)
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
	}, nil
}

func (st *state) compileVarDeclaration(decl parser.VarDeclaration) CompilerError {
	if decl.Origin == nil {
		panic("TODO external vars")
	}
	r, err := st.compileExpr(*decl.Origin)
	if err != nil {
		return err
	}
	st.vars[decl.Name.Name] = r
	return nil
}
