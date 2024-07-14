package analysis

import (
	"math/big"
	"numscript/parser"
)

const TypeMonetary = "monetary"
const TypeAccount = "account"
const TypePortion = "portion"
const TypeAsset = "asset"
const TypeNumber = "number"
const TypeString = "string"

var AllowedTypes = []string{TypeMonetary, TypeAccount, TypePortion, TypeAsset, TypeNumber, TypeString}

type Diagnostic struct {
	Range parser.Range
	Kind  DiagnosticKind
}

type CheckResult struct {
	declaredVars  map[string]parser.VarDeclaration
	unusedVars    map[string]parser.Range
	varResolution map[*parser.VariableLiteral]parser.VarDeclaration
	Diagnostics   []Diagnostic
}

func (r CheckResult) ResolveVar(v *parser.VariableLiteral) *parser.VarDeclaration {
	k, ok := r.varResolution[v]
	if !ok {
		return nil
	}
	return &k
}

func Check(program parser.Program) CheckResult {
	res := CheckResult{
		declaredVars:  make(map[string]parser.VarDeclaration),
		unusedVars:    make(map[string]parser.Range),
		varResolution: make(map[*parser.VariableLiteral]parser.VarDeclaration),
	}
	for _, varDecl := range program.Vars {
		res.checkVarDecl(varDecl)
	}
	for _, statement := range program.Statements {
		switch statement := statement.(type) {
		case *parser.SendStatement:
			res.checkLiteral(statement.Monetary, TypeMonetary)
			res.checkSource(statement.Source)
			res.checkDestination(statement.Destination)
		}
	}

	// after static AST traversal is complete, check for unused vars
	for name, rng := range res.unusedVars {
		res.Diagnostics = append(res.Diagnostics, Diagnostic{
			Range: rng,
			Kind:  &UnusedVar{Name: name},
		})
	}
	return res
}

func isTypeAllowed(typeName string) bool {
	isAllowed := false
	for _, allowedType := range AllowedTypes {
		if allowedType == typeName {
			isAllowed = true
			break
		}
	}
	return isAllowed
}

func (res *CheckResult) checkVarType(typeDecl parser.TypeDecl) {
	if !isTypeAllowed(typeDecl.Name) {
		res.Diagnostics = append(res.Diagnostics, Diagnostic{
			Range: typeDecl.Range,
			Kind:  &InvalidType{Name: typeDecl.Name},
		})
	}
}

func (res *CheckResult) checkDuplicateVars(variableName parser.VariableLiteral, decl parser.VarDeclaration) {
	// check there aren't duplicate variables
	if _, ok := res.declaredVars[variableName.Name]; ok {
		res.Diagnostics = append(res.Diagnostics, Diagnostic{
			Range: variableName.Range,
			Kind:  &DuplicateVariable{Name: variableName.Name},
		})
	} else {
		res.declaredVars[variableName.Name] = decl
		res.unusedVars[variableName.Name] = variableName.Range
	}
}

func (res *CheckResult) checkVarDecl(varDecl parser.VarDeclaration) {
	if varDecl.Type != nil {
		res.checkVarType(*varDecl.Type)
	}

	if varDecl.Name != nil {
		res.checkDuplicateVars(*varDecl.Name, varDecl)
	}
}

func (res *CheckResult) checkLiteral(lit parser.Literal, expectedType string) {
	switch lit := lit.(type) {
	case *parser.VariableLiteral:
		if varDeclaration, ok := res.declaredVars[lit.Name]; ok {
			res.varResolution[lit] = varDeclaration
		} else {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: lit.Range,
				Kind:  &UnboundVariable{Name: lit.Name},
			})
		}
		delete(res.unusedVars, lit.Name)

	case *parser.MonetaryLiteral:
	case *parser.AccountLiteral:
	case *parser.RatioLiteral:

	default:
		panic("unhandled clause")
	}

	res.assertType(lit, expectedType)
}

func (res *CheckResult) assertType(lit parser.Literal, requiredType string) {
	varLit, ok := lit.(*parser.VariableLiteral)
	if !ok {
		return
	}

	resolved := res.ResolveVar(varLit)
	if resolved == nil || resolved.Type == nil || !isTypeAllowed(resolved.Type.Name) {
		return
	}

	if resolved.Type.Name != requiredType {
		res.Diagnostics = append(res.Diagnostics, Diagnostic{
			Range: varLit.Range,
			Kind: &TypeMismatch{
				Expected: requiredType,
				Got:      resolved.Type.Name,
			},
		})
	}
}

func (res *CheckResult) checkSource(source parser.Source) {
	switch source := source.(type) {
	case *parser.AccountLiteral:

	case *parser.VariableLiteral:
		res.checkLiteral(source, TypeAccount)

	case *parser.SourceOverdraft:
		res.checkLiteral(source.Address, TypeAccount)
		if source.Bounded != nil {
			res.checkLiteral(*source.Bounded, TypeMonetary)
		}

	case *parser.SourceSeq:
		for _, source := range source.Sources {
			res.checkSource(source)
		}

	case *parser.SourceCapped:
		res.checkLiteral(source.Cap, TypeMonetary)
		res.checkSource(source.From)

	case *parser.SourceAllotment:
		var remainingAllotment *parser.RemainingAllotment = nil

		sum := big.NewRat(0, 1)
		for i, allottedItem := range source.Items {
			isLast := i == len(source.Items)-1

			switch allotment := allottedItem.Allotment.(type) {
			case *parser.VariableLiteral:
				res.checkLiteral(allotment, TypePortion)
			case *parser.RatioLiteral:
				sum.Add(sum, allotment.ToRatio())
			case *parser.RemainingAllotment:
				if isLast {
					remainingAllotment = allotment
				} else {
					res.Diagnostics = append(res.Diagnostics, Diagnostic{
						Range: source.Range,
						Kind:  &RemainingIsNotLast{},
					})
				}
			}

			res.checkSource(allottedItem.From)
		}

		res.checkHasBadAllotmentSum(*sum, source.Range, remainingAllotment)

	default:
		panic("unhandled clause")
	}
}

func (res *CheckResult) checkDestination(source parser.Destination) {
	switch source := source.(type) {
	case *parser.VariableLiteral:
		res.checkLiteral(source, TypeAccount)

	case *parser.DestinationSeq:
		for _, dest := range source.Destinations {
			res.checkDestination(dest)
		}

	case *parser.DestinationAllotment:
		var remainingAllotment *parser.RemainingAllotment = nil
		sum := big.NewRat(0, 1)

		for i, allottedItem := range source.Items {
			isLast := i == len(source.Items)-1

			switch allotment := allottedItem.Allotment.(type) {
			case *parser.VariableLiteral:
				res.checkLiteral(allotment, TypePortion)
			case *parser.RatioLiteral:
				sum.Add(sum, allotment.ToRatio())
			case *parser.RemainingAllotment:
				if isLast {
					remainingAllotment = allotment
				} else {
					res.Diagnostics = append(res.Diagnostics, Diagnostic{
						Range: source.Range,
						Kind:  &RemainingIsNotLast{},
					})
				}
			}

			res.checkDestination(allottedItem.To)
		}

		res.checkHasBadAllotmentSum(*sum, source.Range, remainingAllotment)
	}
}

func (res *CheckResult) checkHasBadAllotmentSum(
	sum big.Rat, rng parser.Range, remaining *parser.RemainingAllotment,
) {
	cmp := sum.Cmp(big.NewRat(1, 1))
	switch cmp {
	case 1, -1:
		if cmp == -1 && remaining != nil {
			return
		}

		res.Diagnostics = append(res.Diagnostics, Diagnostic{
			Range: rng,
			Kind: &BadAllotmentSum{
				Sum: sum,
			},
		})

	// sum == 1
	case 0:
		if remaining != nil {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: remaining.Range,
				Kind:  &RedundantRemaining{},
			})
		}
	}
}
