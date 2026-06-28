package compiler

import (
	"math/big"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
)

type type_ string

const (
	typeNumber   type_ = "number"
	typeString   type_ = "string"
	typeAsset    type_ = "asset"
	typeMonetary type_ = "monetary"
	typeAccount  type_ = "account"
	typePortion  type_ = "portion"
)

type compiledProgramVirtual struct {
	instructions []vInstr
}

type state struct {
	nextReg      int
	instructions []vInstr
}

func (st *state) getFreshReg() reg {
	id := st.nextReg
	st.nextReg++
	return reg(id)
}

func (st *state) pushInstruction(instr vInstr) {
	st.instructions = append(st.instructions, instr)
}

func (st *state) pushInstructionWithDest(getInstr func(dest reg) vInstr) reg {
	dest := st.getFreshReg()
	st.instructions = append(st.instructions, getInstr(dest))
	return dest
}

func (st *state) pushInstructionWithDestErr(getInstr func(dest reg) vInstr) (reg, CompilerError) {
	return st.pushInstructionWithDest(getInstr), nil
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
		panic("TODO compileExpr")

	case *parser.PercentageLiteral:
		panic("TODO compileExpr")

	case *parser.BinaryInfix:
		panic("TODO compileExpr")

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
		if capReg == nil {
			return 0, InvalidUncappedSource{
				Range: src.GetRange(),
			}
		}

		if src.Color != nil {
			panic("TODO impl color")
		}

		accReg, err := st.compileExpr(src.Address)
		if err != nil {
			return 0, err
		}

		var overdraftReg *reg
		if src.Bounded != nil {
			*overdraftReg, err = st.compileExpr(*src.Bounded)
			if err != nil {
				return 0, err
			}
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

	case *parser.SourceInorder:
		panic("TODO impl source")
	case *parser.SourceOneof:
		panic("TODO impl source")
	case *parser.SourceAllotment:
		panic("TODO impl source")
	case *parser.SourceCapped:
		panic("TODO impl source")
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
	case *parser.DestinationAccount:
		accReg, err := st.compileExpr(dest.ValueExpr)
		if err != nil {
			return err
		}

		var cap *reg
		if pulledAmtReg != currentCap {
			cap = &pulledAmtReg
		}
		st.pushInstruction(sendToAccount{
			account: &accReg,
			cap:     cap,
		})

	case *parser.DestinationInorder:
	case *parser.DestinationOneof:
	case *parser.DestinationAllotment:

	default:
		utils.NonExhaustiveMatchPanic[any](dest)
	}

	return nil
}

func (st *state) compileKeptOrDestination(keptOrDest parser.KeptOrDestination) CompilerError {
	switch keptOrDest := keptOrDest.(type) {
	case *parser.DestinationKept:
	case *parser.DestinationTo:
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
		panic("TODO save")
	case *parser.FnCall:
		panic("TODO fn call")

	default:
		return utils.NonExhaustiveMatchPanic[CompilerError](stmt)
	}
}

func compileProgramToVirtual(program parser.Program) (compiledProgramVirtual, CompilerError) {
	st := state{}
	for _, stmt := range program.Statements {
		st.compileStatements(stmt)
	}

	return compiledProgramVirtual{
		instructions: st.instructions,
	}, nil
}
