package compiler

import (
	"errors"

	"github.com/formancehq/numscript/internal/oracle/machine"
	"github.com/formancehq/numscript/internal/oracle/machine/script/parser"
	"github.com/formancehq/numscript/internal/oracle/machine/vm/program"
)

func (p *parseVisitor) VisitDestination(c parser.IDestinationContext) *CompileError {
	err := p.VisitDestinationRecursive(c)
	if err != nil {
		return err
	}
	p.AppendInstruction(program.OP_REPAY)
	return nil
}

func (p *parseVisitor) VisitDestinationRecursive(c parser.IDestinationContext) *CompileError {
	switch c := c.(type) {
	case *parser.DestAccountContext:
		p.AppendInstruction(program.OP_FUNDING_SUM)
		p.AppendInstruction(program.OP_TAKE)
		ty, _, err := p.VisitExpr(c.Expression(), true)
		if err != nil {
			return err
		}
		if ty != machine.TypeAccount {
			return LogicError(c,
				errors.New("wrong type: expected account as destination"),
			)
		}
		p.AppendInstruction(program.OP_SEND)
		return nil
	case *parser.DestInOrderContext:
		dests := c.DestinationInOrder().GetDests()
		amounts := c.DestinationInOrder().GetAmounts()
		n := len(dests)

		// initialize the `kept`/unsent-residual accumulator (an empty
		// Funding of the right asset, obtained via a no-op TakeMax(0) so
		// it starts out disjoint from the pool below it)
		p.AppendInstruction(program.OP_FUNDING_SUM)
		p.AppendInstruction(program.OP_ASSET)
		err := p.PushInteger(machine.NewNumber(0))
		if err != nil {
			return LogicError(c, err)
		}
		p.AppendInstruction(program.OP_MONETARY_NEW)
		p.AppendInstruction(program.OP_TAKE_MAX)
		err = p.Bump(2)
		if err != nil {
			return LogicError(c, err)
		}
		p.AppendInstruction(program.OP_DELETE)
		err = p.Bump(1)
		if err != nil {
			return LogicError(c, err)
		}

		for i := 0; i < n; i++ {
			ty, _, compErr := p.VisitExpr(amounts[i], true)
			if compErr != nil {
				return compErr
			}
			if ty != machine.TypeMonetary {
				return LogicError(c, errors.New("wrong type: expected monetary as max"))
			}
			p.AppendInstruction(program.OP_TAKE_MAX)
			err := p.Bump(2)
			if err != nil {
				return LogicError(c, err)
			}
			p.AppendInstruction(program.OP_DELETE)
			compErr = p.VisitKeptOrDestination(dests[i])
			if compErr != nil {
				return compErr
			}
			// fold whatever wasn't routed to a real account into the repay
			// accumulator, WITHOUT returning it to the pool: money that is
			// `kept` (or left unsent by a nested destination) must not be
			// visible to subsequent clauses in this same block.
			err = p.Bump(2)
			if err != nil {
				return LogicError(c, err)
			}
			err = p.PushInteger(machine.NewNumber(2))
			if err != nil {
				return LogicError(c, err)
			}
			p.AppendInstruction(program.OP_FUNDING_ASSEMBLE)
			err = p.Bump(1)
			if err != nil {
				return LogicError(c, err)
			}
		}
		cerr := p.VisitKeptOrDestination(c.DestinationInOrder().GetRemainingDest())
		if cerr != nil {
			return cerr
		}
		err = p.PushInteger(machine.NewNumber(2))
		if err != nil {
			return LogicError(c, err)
		}
		p.AppendInstruction(program.OP_FUNDING_ASSEMBLE)
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

	// initialize the `kept`/unsent-residual accumulator (an empty Funding
	// of the right asset, obtained via a no-op TakeMax(0) so it starts out
	// disjoint from the pool below it) — same trick as DestInOrderContext.
	p.AppendInstruction(program.OP_FUNDING_SUM)
	p.AppendInstruction(program.OP_ASSET)
	err = p.PushInteger(machine.NewNumber(0))
	if err != nil {
		return LogicError(dests[0], err)
	}
	p.AppendInstruction(program.OP_MONETARY_NEW)
	p.AppendInstruction(program.OP_TAKE_MAX)
	err = p.Bump(2)
	if err != nil {
		return LogicError(dests[0], err)
	}
	p.AppendInstruction(program.OP_DELETE)
	err = p.Bump(1)
	if err != nil {
		return LogicError(dests[0], err)
	}

	for _, dest := range dests {
		// +1 vs. before: the repay accumulator now permanently occupies
		// the slot right below the pool, so fetching the next part has to
		// reach one level deeper.
		err = p.Bump(2)
		if err != nil {
			return LogicError(dest, err)
		}
		p.AppendInstruction(program.OP_TAKE)
		compErr := p.VisitKeptOrDestination(dest)
		if compErr != nil {
			return compErr
		}
		// fold whatever wasn't routed to a real account into the repay
		// accumulator, WITHOUT returning it to the pool: money that is
		// `kept` must not be visible to subsequent clauses in this same
		// block (same reasoning as the DestInOrderContext fix).
		err = p.Bump(2)
		if err != nil {
			return LogicError(dest, err)
		}
		err = p.PushInteger(machine.NewNumber(2))
		if err != nil {
			return LogicError(dest, err)
		}
		p.AppendInstruction(program.OP_FUNDING_ASSEMBLE)
		err = p.Bump(1)
		if err != nil {
			return LogicError(dest, err)
		}
	}

	// merge the repay accumulator with whatever's left of the pool (empty
	// once portions sum to 1, but this stays correct even if they don't)
	// into the single Funding this function must leave on the stack.
	err = p.PushInteger(machine.NewNumber(2))
	if err != nil {
		return LogicError(dests[0], err)
	}
	p.AppendInstruction(program.OP_FUNDING_ASSEMBLE)
	return nil
}
