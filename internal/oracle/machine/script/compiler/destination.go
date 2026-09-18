package compiler

import (
	"errors"

	"github.com/formancehq/numscript/internal/oracle/machine"
	"github.com/formancehq/numscript/internal/oracle/machine/script/parser"
	"github.com/formancehq/numscript/internal/oracle/machine/vm/program"
)

func (p *parseVisitor) VisitDestination(c parser.IDestinationContext) *CompileError {
	// <funding>
	err := p.VisitDestinationRecursive(c)
	if err != nil {
		return err
	}
	p.AppendInstruction(program.OP_REPAY)
	return nil
}

// should start with: <funding>
// should return: <funding> (whatever was kept)
func (p *parseVisitor) VisitDestinationRecursive(c parser.IDestinationContext) *CompileError {
	// STACK: <funding>
	switch c := c.(type) {
	case *parser.DestAccountContext:
		// we take everything (keep nothing): build an empty funding to return
		p.AppendInstruction(program.OP_FUNDING_SUM)
		// <funding> <sum: number>
		p.AppendInstruction(program.OP_TAKE)
		// <empty remaining: funding> <funding>
		ty, _, err := p.VisitExpr(c.Expression(), true)
		if err != nil {
			return err
		}
		if ty != machine.TypeAccount {
			return LogicError(c,
				errors.New("wrong type: expected account as destination"),
			)
		}
		// <empty remaining: funding> <f: funding> <account>
		p.AppendInstruction(program.OP_SEND)
		// <empty remaining: funding>
		return nil
	case *parser.DestInOrderContext:
		dests := c.DestinationInOrder().GetDests()
		amounts := c.DestinationInOrder().GetAmounts()
		n := len(dests)

		// initialize the `kept` accumulator
		p.AppendInstruction(program.OP_FUNDING_SUM)
		p.AppendInstruction(program.OP_TAKE)
		// <kept_acc: funding> <funding>

		for i := 0; i < n; i++ {
			ty, _, compErr := p.VisitExpr(amounts[i], true)
			if compErr != nil {
				return compErr
			}
			if ty != machine.TypeMonetary {
				return LogicError(c, errors.New("wrong type: expected monetary as max"))
			}
			// <kept_acc: funding> <funding> <max: monetary>
			p.AppendInstruction(program.OP_TAKE_MAX)
			err := p.Bump(2)
			if err != nil {
				return LogicError(c, err)
			}
			p.AppendInstruction(program.OP_DELETE)
			// <kept_acc: funding> <remaining: funding> <capped: funding>
			compErr = p.VisitKeptOrDestination(dests[i])
			if compErr != nil {
				return compErr
			}
			// <kept_acc: funding> <remaining: funding> <subkept: funding>
			// DIVERGES FROM LEDGER: ledger re-attributes <subkept> to the bottom of
			// the pool here (reassemble + reverse + take + reverse). <subkept> is kept
			// as-is instead, so the front of the pool keeps. See DIVERGENCES.md §2 ②.
			err = p.Bump(2)
			if err != nil {
				return LogicError(c, err)
			}
			// <remaining: funding> <subkept: funding> <kept_acc: funding>
			err = p.PushInteger(machine.NewNumber(2))
			if err != nil {
				return LogicError(c, err)
			}
			// <remaining: funding> <subkept: funding> <kept_acc: funding> <2>
			p.AppendInstruction(program.OP_FUNDING_ASSEMBLE)
			// <remaining: funding> <kept_acc: funding>
			err = p.Bump(1)
			if err != nil {
				return LogicError(c, err)
			}
			// <kept_acc: funding> <remaining: funding>
		}
		cerr := p.VisitKeptOrDestination(c.DestinationInOrder().GetRemainingDest())
		if cerr != nil {
			return cerr
		}
		// <kept_acc: funding> <subkept: funding>
		err := p.Bump(1)
		if err != nil {
			return LogicError(c, err)
		}
		// <subkept: funding> <kept_acc: funding>
		err = p.PushInteger(machine.NewNumber(2))
		if err != nil {
			return LogicError(c, err)
		}
		p.AppendInstruction(program.OP_FUNDING_ASSEMBLE)
		// <kept_acc: funding>
		return nil
	case *parser.DestAllotmentContext:
		err := p.VisitDestinationAllotment(c.DestinationAllotment())
		return err
	default:
		return InternalError(c)
	}
}

func (p *parseVisitor) VisitKeptOrDestination(c parser.IKeptOrDestinationContext) *CompileError {
	switch c := c.(type) {
	case *parser.IsKeptContext:
		return nil
	case *parser.IsDestinationContext:
		err := p.VisitDestinationRecursive(c.Destination())
		return err
	default:
		return InternalError(c)
	}
}

func (p *parseVisitor) VisitDestinationAllotment(c parser.IDestinationAllotmentContext) *CompileError {
	p.AppendInstruction(program.OP_FUNDING_SUM)
	err := p.VisitAllotment(c, c.GetPortions())
	if err != nil {
		return err
	}
	p.AppendInstruction(program.OP_ALLOC)
	err = p.VisitAllocDestination(c.GetDests())
	if err != nil {
		return err
	}
	return nil
}

func (p *parseVisitor) VisitAllocDestination(dests []parser.IKeptOrDestinationContext) *CompileError {
	err := p.Bump(int64(len(dests)))
	if err != nil {
		return LogicError(dests[0], err)
	}
	for _, dest := range dests {
		err = p.Bump(1)
		if err != nil {
			return LogicError(dest, err)
		}
		p.AppendInstruction(program.OP_TAKE)
		compErr := p.VisitKeptOrDestination(dest)
		if compErr != nil {
			return compErr
		}
		// DIVERGES FROM LEDGER: ledger bumps here, so <subkept> is assembled at the
		// FRONT of the pool and the next portion spends it again. Assembled at the
		// back instead, it survives as leftover and is repaid. See DIVERGENCES.md §2 ②.
		err = p.PushInteger(machine.NewNumber(2))
		if err != nil {
			return LogicError(dest, err)
		}
		p.AppendInstruction(program.OP_FUNDING_ASSEMBLE)
	}
	return nil
}
