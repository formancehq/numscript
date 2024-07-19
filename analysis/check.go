package analysis

import (
	"fmt"
	"math/big"
	"numscript/parser"
	"slices"
)

const TypeMonetary = "monetary"
const TypeAccount = "account"
const TypePortion = "portion"
const TypeAsset = "asset"
const TypeNumber = "number"
const TypeString = "string"

const TypeAny = "any"

var AllowedTypes = []string{
	TypeMonetary,
	TypeAccount,
	TypePortion,
	TypeAsset,
	TypeNumber,
	TypeString,
}

type FnCallResolution interface {
	GetParams() []string
	fnCallResolution()
}

type VarOriginFnCallResolution struct {
	Params []string
	Docs   string
	Return string
}
type StatementFnCallResolution struct {
	Params []string
	Docs   string
}

func (VarOriginFnCallResolution) fnCallResolution() {}
func (StatementFnCallResolution) fnCallResolution() {}

func (r VarOriginFnCallResolution) GetParams() []string { return r.Params }
func (r StatementFnCallResolution) GetParams() []string { return r.Params }

const FnSetTxMeta = "set_tx_meta"
const FnSetAccountMeta = "set_account_meta"
const FnVarOriginMeta = "meta"
const FnVarOriginBalance = "balance"

var Builtins = map[string]FnCallResolution{
	FnSetTxMeta: StatementFnCallResolution{
		Params: []string{TypeString, TypeAny},
		Docs:   "set transaction metadata",
	},
	FnSetAccountMeta: StatementFnCallResolution{
		Params: []string{TypeAccount, TypeString, TypeAny},
		Docs:   "set account metadata",
	},
	FnVarOriginMeta: VarOriginFnCallResolution{
		Params: []string{TypeAccount, TypeString},
		Return: TypeAny,
		Docs:   "fetch account metadata",
	},
	FnVarOriginBalance: VarOriginFnCallResolution{
		Params: []string{TypeAccount, TypeAsset},
		Return: TypeMonetary,
		Docs:   "fetch account balance",
	},
}

type Diagnostic struct {
	Range parser.Range
	Kind  DiagnosticKind
}

type CheckResult struct {
	declaredVars     map[string]parser.VarDeclaration
	unusedVars       map[string]parser.Range
	varResolution    map[*parser.VariableLiteral]parser.VarDeclaration
	fnCallResolution map[*parser.FnCallIdentifier]FnCallResolution
	Diagnostics      []Diagnostic
}

func (r CheckResult) ResolveVar(v *parser.VariableLiteral) *parser.VarDeclaration {
	k, ok := r.varResolution[v]
	if !ok {
		return nil
	}
	return &k
}

func (r CheckResult) ResolveBuiltinFn(v *parser.FnCallIdentifier) FnCallResolution {
	k, ok := r.fnCallResolution[v]
	if !ok {
		return nil
	}
	return k
}

func Check(program parser.Program) CheckResult {
	res := CheckResult{
		declaredVars:     make(map[string]parser.VarDeclaration),
		unusedVars:       make(map[string]parser.Range),
		varResolution:    make(map[*parser.VariableLiteral]parser.VarDeclaration),
		fnCallResolution: make(map[*parser.FnCallIdentifier]FnCallResolution),
	}
	for _, varDecl := range program.Vars {
		if varDecl.Type != nil {
			res.checkVarType(*varDecl.Type)
		}

		if varDecl.Name != nil {
			res.checkDuplicateVars(*varDecl.Name, varDecl)
		}

		if varDecl.Origin != nil {
			res.checkVarOrigin(*varDecl.Origin, varDecl)
		}
	}
	for _, statement := range program.Statements {
		switch statement := statement.(type) {
		case *parser.SendStatement:
			res.checkSentValue(statement.SentValue)
			res.checkSource(statement.Source)
			res.checkDestination(statement.Destination)
		case *parser.FnCall:
			resolution, ok := Builtins[statement.Caller.Name]
			if ok {
				if varOrigin, ok := resolution.(StatementFnCallResolution); ok {
					res.fnCallResolution[statement.Caller] = varOrigin
				}
			}

			// This must come after resolution
			res.checkFnCallArity(statement)
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

func (res *CheckResult) checkFnCallArity(fnCall *parser.FnCall) {
	resolution, resolved := res.fnCallResolution[fnCall.Caller]

	var validArgs []parser.Literal
	for _, lit := range fnCall.Args {
		if lit != nil {
			validArgs = append(validArgs, lit)
		}
	}

	if resolved {
		sig := resolution.GetParams()
		actualArgs := len(validArgs)
		expectedArgs := len(sig)

		if actualArgs < expectedArgs {
			// Too few args
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: fnCall.Range,
				Kind: &BadArity{
					Expected: expectedArgs,
					Actual:   actualArgs,
				},
			})
		} else if actualArgs > expectedArgs {
			// Too many args
			firstIllegalArg := validArgs[expectedArgs]
			lastIllegalArg := validArgs[len(validArgs)-1]

			if lastIllegalArg != nil {
				rng := parser.Range{
					Start: firstIllegalArg.GetRange().Start,
					End:   lastIllegalArg.GetRange().End,
				}

				res.Diagnostics = append(res.Diagnostics, Diagnostic{
					Range: rng,
					Kind: &BadArity{
						Expected: expectedArgs,
						Actual:   actualArgs,
					},
				})
			}

		}

		for index, arg := range validArgs {
			lastElemIndex := len(sig) - 1
			if index > lastElemIndex {
				break
			}

			type_ := sig[index]
			res.checkLiteral(arg, type_)
		}
	} else {
		for _, arg := range validArgs {
			res.checkLiteral(arg, TypeAny)
		}

		res.Diagnostics = append(res.Diagnostics, Diagnostic{
			Range: fnCall.Caller.Range,
			Kind: &UnknownFunction{
				Name: fnCall.Caller.Name,
			},
		})
	}
}

func isTypeAllowed(typeName string) bool {
	return slices.Contains(AllowedTypes, typeName)
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

func (res *CheckResult) checkVarOrigin(fnCall parser.FnCall, decl parser.VarDeclaration) {
	resolution, ok := Builtins[fnCall.Caller.Name]
	if ok {
		resolution, ok := resolution.(VarOriginFnCallResolution)
		if ok {
			res.fnCallResolution[decl.Origin.Caller] = resolution
			res.assertHasType(decl.Name, resolution.Return, decl.Type.Name)
		}
	}

	// this must come after resolution
	res.checkFnCallArity(&fnCall)
}

func (res *CheckResult) checkLiteral(lit parser.Literal, requiredType string) {
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

		resolved := res.ResolveVar(lit)
		if resolved == nil || resolved.Type == nil || !isTypeAllowed(resolved.Type.Name) {
			return
		}
		res.assertHasType(lit, requiredType, resolved.Type.Name)

	case *parser.MonetaryLiteral:
		res.assertHasType(lit, requiredType, TypeMonetary)
		res.checkLiteral(lit.Asset, TypeAsset)
		res.checkLiteral(lit.Amount, TypeNumber)
	case *parser.AccountLiteral:
		res.assertHasType(lit, requiredType, TypeAccount)
	case *parser.RatioLiteral:
		res.assertHasType(lit, requiredType, TypePortion)
	case *parser.AssetLiteral:
		res.assertHasType(lit, requiredType, TypeAsset)
	case *parser.NumberLiteral:
		res.assertHasType(lit, requiredType, TypeNumber)
	case *parser.StringLiteral:
		res.assertHasType(lit, requiredType, TypeString)
	}
}

func (res *CheckResult) assertHasType(lit parser.Literal, requiredType string, actualType string) {
	if requiredType == TypeAny || requiredType == actualType {
		return
	}

	res.Diagnostics = append(res.Diagnostics, Diagnostic{
		Range: lit.GetRange(),
		Kind: &TypeMismatch{
			Expected: requiredType,
			Got:      actualType,
		},
	})

}

func (res *CheckResult) checkSentValue(sentValue parser.SentValue) {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueAll:
		res.checkLiteral(sentValue.Asset, TypeAsset)
	case *parser.SentValueLiteral:
		res.checkLiteral(sentValue.Monetary, TypeMonetary)
	}
}

func (res *CheckResult) checkSource(source parser.Source) {
	if source == nil {
		return
	}

	switch source := source.(type) {
	case *parser.AccountLiteral:

	case *parser.VariableLiteral:
		res.checkLiteral(source, TypeAccount)

	case *parser.SourceOverdraft:
		res.checkLiteral(source.Address, TypeAccount)
		if source.Bounded != nil {
			res.checkLiteral(*source.Bounded, TypeMonetary)
		}

	case *parser.SourceInorder:
		for _, source := range source.Sources {
			res.checkSource(source)
		}

	case *parser.SourceCapped:
		res.checkLiteral(source.Cap, TypeMonetary)
		res.checkSource(source.From)

	case *parser.SourceAllotment:
		var remainingAllotment *parser.RemainingAllotment = nil
		var variableLiterals []parser.VariableLiteral

		sum := big.NewRat(0, 1)
		for i, allottedItem := range source.Items {
			isLast := i == len(source.Items)-1

			switch allotment := allottedItem.Allotment.(type) {
			case *parser.VariableLiteral:
				variableLiterals = append(variableLiterals, *allotment)
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

		res.checkHasBadAllotmentSum(*sum, source.Range, remainingAllotment, variableLiterals)

	default:
		panic(fmt.Sprintf("unhandled clause: %+s", source))
	}
}

func (res *CheckResult) checkDestination(destination parser.Destination) {
	if destination == nil {
		return
	}

	switch destination := destination.(type) {
	case *parser.VariableLiteral:
		res.checkLiteral(destination, TypeAccount)

	case *parser.DestinationInorder:
		for _, clause := range destination.Clauses {
			res.checkLiteral(clause.Cap, TypeMonetary)
			res.checkKeptOrDestination(clause.To)
		}
		res.checkKeptOrDestination(destination.Remaining)

	case *parser.DestinationAllotment:
		var remainingAllotment *parser.RemainingAllotment
		var variableLiterals []parser.VariableLiteral
		sum := big.NewRat(0, 1)

		for i, allottedItem := range destination.Items {
			isLast := i == len(destination.Items)-1

			switch allotment := allottedItem.Allotment.(type) {
			case *parser.VariableLiteral:
				variableLiterals = append(variableLiterals, *allotment)
				res.checkLiteral(allotment, TypePortion)
			case *parser.RatioLiteral:
				sum.Add(sum, allotment.ToRatio())
			case *parser.RemainingAllotment:
				if isLast {
					remainingAllotment = allotment
				} else {
					res.Diagnostics = append(res.Diagnostics, Diagnostic{
						Range: destination.Range,
						Kind:  &RemainingIsNotLast{},
					})
				}
			}

			res.checkKeptOrDestination(allottedItem.To)
		}

		res.checkHasBadAllotmentSum(*sum, destination.Range, remainingAllotment, variableLiterals)
	}
}

func (res *CheckResult) checkKeptOrDestination(target parser.KeptOrDestination) {
	switch target := target.(type) {
	case *parser.DestinationTo:
		res.checkDestination(target.Destination)
	case *parser.DestinationKept:
		// nothing to do
	}
}

func (res *CheckResult) checkHasBadAllotmentSum(
	sum big.Rat,
	rng parser.Range,
	remaining *parser.RemainingAllotment,
	variableLiterals []parser.VariableLiteral,
) {
	cmp := sum.Cmp(big.NewRat(1, 1))
	switch cmp {
	case 1, -1:
		if (cmp == -1 && remaining != nil) || (cmp == -1 && len(variableLiterals) > 1) {
			return
		}

		if cmp == -1 && len(variableLiterals) == 1 {
			var value big.Rat
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: variableLiterals[0].Range,
				Kind: &FixedPortionVariable{
					Value: *value.Sub(big.NewRat(1, 1), &sum),
				},
			})
		} else {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: rng,
				Kind: &BadAllotmentSum{
					Sum: sum,
				},
			})
		}

	// sum == 1
	case 0:
		for _, varLit := range variableLiterals {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: varLit.Range,
				Kind: &FixedPortionVariable{
					Value: *big.NewRat(0, 1),
				},
			})
		}
		if remaining != nil {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: remaining.Range,
				Kind:  &RedundantRemaining{},
			})
		}
	}
}
