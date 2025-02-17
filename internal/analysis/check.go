package analysis

import (
	"math/big"
	"slices"
	"strings"

	"github.com/formancehq/numscript/internal/parser"
	"github.com/formancehq/numscript/internal/utils"
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
	ContextName() string
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

func (VarOriginFnCallResolution) ContextName() string { return "variable origin" }
func (StatementFnCallResolution) ContextName() string { return "statement" }

func (VarOriginFnCallResolution) fnCallResolution() {}
func (StatementFnCallResolution) fnCallResolution() {}

func (r VarOriginFnCallResolution) GetParams() []string { return r.Params }
func (r StatementFnCallResolution) GetParams() []string { return r.Params }

const FnSetTxMeta = "set_tx_meta"
const FnSetAccountMeta = "set_account_meta"
const FnVarOriginMeta = "meta"
const FnVarOriginBalance = "balance"
const FnVarOriginOverdraft = "overdraft"

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
	FnVarOriginOverdraft: VarOriginFnCallResolution{
		Params: []string{TypeAccount, TypeAsset},
		Return: TypeMonetary,
		Docs:   "get absolute amount of the overdraft of an account. Returns zero if balance is not negative",
	},
}

type Diagnostic struct {
	Range parser.Range
	Kind  DiagnosticKind
}

type CheckResult struct {
	unboundedAccountInSend parser.ValueExpr
	emptiedAccount         map[string]struct{}
	unboundedSend          bool
	declaredVars           map[string]parser.VarDeclaration
	unusedVars             map[string]parser.Range
	varResolution          map[*parser.Variable]parser.VarDeclaration
	fnCallResolution       map[*parser.FnCallIdentifier]FnCallResolution
	Diagnostics            []Diagnostic
	Program                parser.Program
}

func (r CheckResult) GetErrorsCount() int {
	c := 0
	for _, d := range r.Diagnostics {
		if d.Kind.Severity() == ErrorSeverity {
			c++
		}
	}
	return c
}

func (r CheckResult) GetWarningsCount() int {
	c := 0
	for _, d := range r.Diagnostics {
		if d.Kind.Severity() == WarningSeverity {
			c++
		}
	}
	return c
}

func (r CheckResult) ResolveVar(v *parser.Variable) *parser.VarDeclaration {
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

func newCheckResult(program parser.Program) CheckResult {
	return CheckResult{
		emptiedAccount:   make(map[string]struct{}),
		declaredVars:     make(map[string]parser.VarDeclaration),
		unusedVars:       make(map[string]parser.Range),
		varResolution:    make(map[*parser.Variable]parser.VarDeclaration),
		fnCallResolution: make(map[*parser.FnCallIdentifier]FnCallResolution),
		Program:          program,
	}
}

func (res *CheckResult) check() {
	if res.Program.Vars != nil {
		for _, varDecl := range res.Program.Vars.Declarations {
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
	}

	for _, statement := range res.Program.Statements {
		res.unboundedAccountInSend = nil
		res.checkStatement(statement)
	}

	// after static AST traversal is complete, check for unused vars
	for name, rng := range res.unusedVars {
		res.Diagnostics = append(res.Diagnostics, Diagnostic{
			Range: rng,
			Kind:  &UnusedVar{Name: name},
		})
	}
}

func (res *CheckResult) checkStatement(statement parser.Statement) {
	res.emptiedAccount = make(map[string]struct{})

	switch statement := statement.(type) {
	case *parser.SaveStatement:
		res.checkSentValue(statement.SentValue)
		res.checkExpression(statement.Amount, TypeAccount)

	case *parser.SendStatement:
		_, isSendAll := statement.SentValue.(*parser.SentValueAll)
		res.unboundedSend = isSendAll

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

func CheckProgram(program parser.Program) CheckResult {
	res := newCheckResult(program)
	res.check()
	return res
}

func CheckSource(source string) CheckResult {
	result := parser.Parse(source)
	res := newCheckResult(result.Value)
	for _, parserError := range result.Errors {
		res.Diagnostics = append(res.Diagnostics, parsingErrorToDiagnostic(parserError))
	}
	res.check()
	return res
}

func parsingErrorToDiagnostic(parserError parser.ParserError) Diagnostic {
	return Diagnostic{
		Range: parserError.Range,
		Kind:  &Parsing{Description: parserError.Msg},
	}
}

func (res *CheckResult) checkFnCallArity(fnCall *parser.FnCall) {
	resolution, resolved := res.fnCallResolution[fnCall.Caller]

	var validArgs []parser.ValueExpr
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
			res.checkExpression(arg, type_)
		}
	} else {
		for _, arg := range validArgs {
			res.checkExpression(arg, TypeAny)
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

func (res *CheckResult) checkDuplicateVars(variableName parser.Variable, decl parser.VarDeclaration) {
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

func (res *CheckResult) checkExpression(lit parser.ValueExpr, requiredType string) {
	actualType := res.checkTypeOf(lit, requiredType)
	res.assertHasType(lit, requiredType, actualType)
}

func (res *CheckResult) checkTypeOf(lit parser.ValueExpr, typeHint string) string {
	switch lit := lit.(type) {
	case *parser.Variable:
		if varDeclaration, ok := res.declaredVars[lit.Name]; ok {
			res.varResolution[lit] = varDeclaration
		} else {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: lit.Range,
				Kind:  &UnboundVariable{Name: lit.Name, Type: typeHint},
			})
		}
		delete(res.unusedVars, lit.Name)

		resolved := res.ResolveVar(lit)
		if resolved == nil || resolved.Type == nil || !isTypeAllowed(resolved.Type.Name) {
			return TypeAny
		}
		return resolved.Type.Name

	case *parser.MonetaryLiteral:
		res.checkExpression(lit.Asset, TypeAsset)
		res.checkExpression(lit.Amount, TypeNumber)
		return TypeMonetary

	case *parser.BinaryInfix:
		switch lit.Operator {
		case parser.InfixOperatorPlus:
			return res.checkInfixOverload(lit, []string{TypeNumber, TypeMonetary})

		case parser.InfixOperatorMinus:
			return res.checkInfixOverload(lit, []string{TypeNumber, TypeMonetary})

		case parser.InfixOperatorDiv:
			res.checkExpression(lit.Left, TypeNumber)
			res.checkExpression(lit.Right, TypeNumber)
			return TypePortion

		default:
			// we should never get here
			// but just to be sure
			res.checkExpression(lit.Left, TypeAny)
			res.checkExpression(lit.Right, TypeAny)
			return TypeAny
		}

	case *parser.AccountInterpLiteral:
		for _, part := range lit.Parts {
			if v, ok := part.(*parser.Variable); ok {
				res.checkExpression(v, TypeAny)
			}
		}
		return TypeAccount
	case *parser.PercentageLiteral:
		return TypePortion
	case *parser.AssetLiteral:
		return TypeAsset
	case *parser.NumberLiteral:
		return TypeNumber
	case *parser.StringLiteral:
		return TypeString

	default:
		return TypeAny
	}
}

func (res *CheckResult) checkInfixOverload(bin *parser.BinaryInfix, allowed []string) string {
	leftType := res.checkTypeOf(bin.Left, allowed[0])

	if leftType == TypeAny || slices.Contains(allowed, leftType) {
		res.checkExpression(bin.Right, leftType)
		return leftType
	}

	res.Diagnostics = append(res.Diagnostics, Diagnostic{
		Range: bin.Left.GetRange(),
		Kind: &TypeMismatch{
			Expected: strings.Join(allowed, "|"),
			Got:      leftType,
		},
	})
	return TypeAny
}

func (res *CheckResult) assertHasType(lit parser.ValueExpr, requiredType string, actualType string) {
	if requiredType == TypeAny || actualType == TypeAny || requiredType == actualType {
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
		res.checkExpression(sentValue.Asset, TypeAsset)
	case *parser.SentValueLiteral:
		res.checkExpression(sentValue.Monetary, TypeMonetary)
	}
}

func (res *CheckResult) checkSource(source parser.Source) {
	if source == nil {
		return
	}

	if res.unboundedAccountInSend != nil {
		res.Diagnostics = append(res.Diagnostics, Diagnostic{
			Range: source.GetRange(),
			Kind:  &UnboundedAccountIsNotLast{},
		})
	}

	switch source := source.(type) {
	case *parser.SourceAccount:
		res.checkExpression(source.ValueExpr, TypeAccount)
		if account, ok := source.ValueExpr.(*parser.AccountInterpLiteral); ok {
			if account.IsWorld() && res.unboundedSend {
				res.Diagnostics = append(res.Diagnostics, Diagnostic{
					Range: source.GetRange(),
					Kind:  &InvalidUnboundedAccount{},
				})
			} else if account.IsWorld() {
				res.unboundedAccountInSend = account
			}

			if _, emptied := res.emptiedAccount[account.String()]; emptied && !account.IsWorld() {
				res.Diagnostics = append(res.Diagnostics, Diagnostic{
					Kind:  &EmptiedAccount{Name: account.String()},
					Range: account.Range,
				})
			}

			res.emptiedAccount[account.String()] = struct{}{}
		}

	case *parser.SourceOverdraft:
		if accountLiteral, ok := source.Address.(*parser.AccountInterpLiteral); ok && accountLiteral.IsWorld() {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: accountLiteral.Range,
				Kind:  &InvalidWorldOverdraft{},
			})
		}

		if source.Bounded == nil {
			res.unboundedAccountInSend = source.Address
		}

		if res.unboundedSend {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: source.Address.GetRange(),
				Kind:  &InvalidUnboundedAccount{},
			})
		}

		res.checkExpression(source.Address, TypeAccount)
		if source.Bounded != nil {
			res.checkExpression(*source.Bounded, TypeMonetary)
		}

	case *parser.SourceInorder:
		for _, source := range source.Sources {
			res.checkSource(source)
		}

	case *parser.SourceOneof:
		for _, source := range source.Sources {
			res.checkSource(source)
		}

	case *parser.SourceCapped:
		onExit := res.enterCappedSource()

		res.checkExpression(source.Cap, TypeMonetary)
		res.checkSource(source.From)

		onExit()

	case *parser.SourceAllotment:
		if res.unboundedSend {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Kind:  &NoAllotmentInSendAll{},
				Range: source.Range,
			})
		}

		var remainingAllotment *parser.RemainingAllotment = nil
		var variableLiterals []parser.ValueExpr

		sum := big.NewRat(0, 1)
		for i, allottedItem := range source.Items {
			isLast := i == len(source.Items)-1

			switch allotment := allottedItem.Allotment.(type) {
			case *parser.ValueExprAllotment:
				res.checkExpression(allotment.Value, TypePortion)
				rat := res.tryEvaluatingPortionExpr(allotment.Value)
				if rat == nil {
					variableLiterals = append(variableLiterals, allotment.Value)
				} else {
					sum.Add(sum, rat)
				}

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

			onExit := res.enterCappedSource()
			res.checkSource(allottedItem.From)
			onExit()
		}

		res.checkHasBadAllotmentSum(*sum, source.Range, remainingAllotment, variableLiterals)

	default:
		utils.NonExhaustiveMatchPanic[any](source)
	}
}

// Try evaluating an expression, if it can be done statically.
//
// Returns nil when the expression contains variables, fn calls, or anything
// that cannot be computed statically.
//
// For example:
//
//	1 + 2 => 3
//	1 + $x => nil
func (res CheckResult) tryEvaluatingNumberExpr(expr parser.ValueExpr) *big.Int {
	switch expr := expr.(type) {

	case *parser.NumberLiteral:
		return big.NewInt(int64(expr.Number))

	case *parser.BinaryInfix:
		switch expr.Operator {
		case parser.InfixOperatorPlus:
			left := res.tryEvaluatingNumberExpr(expr.Left)
			if left == nil {
				return nil
			}
			right := res.tryEvaluatingNumberExpr(expr.Right)
			if right == nil {
				return nil
			}
			return new(big.Int).Add(left, right)

		case parser.InfixOperatorMinus:
			left := res.tryEvaluatingNumberExpr(expr.Left)
			if left == nil {
				return nil
			}
			right := res.tryEvaluatingNumberExpr(expr.Right)
			if right == nil {
				return nil
			}
			return new(big.Int).Sub(left, right)

		default:
			return nil
		}

	default:
		return nil
	}
}

// Same as analysis.tryEvaluatingNumberExpr, for portion
func (res *CheckResult) tryEvaluatingPortionExpr(expr parser.ValueExpr) *big.Rat {
	switch expr := expr.(type) {
	case *parser.PercentageLiteral:
		return expr.ToRatio()

	case *parser.BinaryInfix:
		switch expr.Operator {
		case parser.InfixOperatorDiv:
			right := res.tryEvaluatingNumberExpr(expr.Right)
			if right == nil {
				return nil
			}

			if right.Cmp(big.NewInt(0)) == 0 {
				res.Diagnostics = append(res.Diagnostics, Diagnostic{
					Kind:  &DivByZero{},
					Range: expr.Range,
				})
				return nil
			}

			left := res.tryEvaluatingNumberExpr(expr.Left)
			if left == nil {
				return nil
			}

			return new(big.Rat).SetFrac(left, right)

		default:
			return nil
		}

	default:
		return nil
	}
}

func (res *CheckResult) checkDestination(destination parser.Destination) {
	if destination == nil {
		return
	}

	switch destination := destination.(type) {
	case *parser.DestinationAccount:
		res.checkExpression(destination.ValueExpr, TypeAccount)

	case *parser.DestinationInorder:
		for _, clause := range destination.Clauses {
			res.checkExpression(clause.Cap, TypeMonetary)
			res.checkKeptOrDestination(clause.To)
		}
		res.checkKeptOrDestination(destination.Remaining)

	case *parser.DestinationOneof:
		for _, clause := range destination.Clauses {
			res.checkExpression(clause.Cap, TypeMonetary)
			res.checkKeptOrDestination(clause.To)
		}
		res.checkKeptOrDestination(destination.Remaining)

	case *parser.DestinationAllotment:
		var remainingAllotment *parser.RemainingAllotment
		var variableLiterals []parser.ValueExpr
		sum := big.NewRat(0, 1)

		for i, allottedItem := range destination.Items {
			isLast := i == len(destination.Items)-1

			switch allotment := allottedItem.Allotment.(type) {
			case *parser.ValueExprAllotment:
				res.checkExpression(allotment.Value, TypePortion)
				rat := res.tryEvaluatingPortionExpr(allotment.Value)
				if rat == nil {
					variableLiterals = append(variableLiterals, allotment.Value)
				} else {
					sum.Add(sum, rat)
				}

			// 	res.checkExpression(allotment, TypePortion)
			// case *parser.PortionLiteral:
			// 	sum.Add(sum, allotment.ToRatio())
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
	variableLiterals []parser.ValueExpr,
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
				Range: variableLiterals[0].GetRange(),
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
				Range: varLit.GetRange(),
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

func (res *CheckResult) withCloneEmptyAccount() func() {
	initial := res.emptiedAccount
	res.emptiedAccount = make(map[string]struct{})
	for k, v := range initial {
		res.emptiedAccount[k] = v
	}
	return func() {
		res.emptiedAccount = initial
	}
}

func (res *CheckResult) withCloneUnboundedAccountInSend() func() {
	initial := res.unboundedAccountInSend
	res.unboundedAccountInSend = nil

	return func() {
		res.unboundedAccountInSend = initial
	}
}

func (res *CheckResult) withCloneUnboundedSend() func() {
	initial := res.unboundedSend
	res.unboundedSend = false

	return func() {
		res.unboundedSend = initial
	}
}

func (res *CheckResult) enterCappedSource() func() {
	exitCloneEmptyAccount := res.withCloneEmptyAccount()
	exitCloneUnboundeAccountInSend := res.withCloneUnboundedAccountInSend()
	exitCloneUnboundedSend := res.withCloneUnboundedSend()

	return func() {
		exitCloneEmptyAccount()
		exitCloneUnboundeAccountInSend()
		exitCloneUnboundedSend()
	}
}
