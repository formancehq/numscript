package analysis

import (
	"math/big"
	"slices"
	"strings"

	"github.com/formancehq/numscript/internal/flags"
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
	Params             []string
	Docs               string
	Return             string
	VersionConstraints []VersionClause
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
const FnVarOriginGetAsset = "get_asset"
const FnVarOriginGetAmount = "get_amount"

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
		VersionConstraints: []VersionClause{
			{
				Version:     parser.NewVersionInterpreter(0, 0, 15),
				FeatureFlag: flags.ExperimentalOverdraftFunctionFeatureFlag,
			},
		},
	},
	FnVarOriginGetAsset: VarOriginFnCallResolution{
		Params: []string{TypeMonetary},
		Return: TypeAsset,
		Docs:   "get the asset of the given monetary",
		VersionConstraints: []VersionClause{
			{
				Version:     parser.NewVersionInterpreter(0, 0, 16),
				FeatureFlag: flags.ExperimentalGetAssetFunctionFeatureFlag,
			},
		},
	},
	FnVarOriginGetAmount: VarOriginFnCallResolution{
		Params: []string{TypeMonetary},
		Return: TypeNumber,
		Docs:   "get the amount of the given monetary",
		VersionConstraints: []VersionClause{
			{
				Version:     parser.NewVersionInterpreter(0, 0, 16),
				FeatureFlag: flags.ExperimentalGetAmountFunctionFeatureFlag,
			},
		},
	},
}

type Diagnostic struct {
	Range parser.Range
	Kind  DiagnosticKind
	Id    int32
}

type CheckResult struct {
	nextDiagnosticId       int32
	unboundedAccountInSend parser.ValueExpr
	emptiedAccount         map[string]struct{}
	unboundedSend          bool
	declaredVars           map[string]parser.VarDeclaration
	unusedVars             map[string]parser.Range
	varResolution          map[*parser.Variable]parser.VarDeclaration
	fnCallResolution       map[*parser.FnCallIdentifier]FnCallResolution
	Diagnostics            []Diagnostic
	Program                parser.Program

	stmtType  Type
	ExprTypes map[parser.ValueExpr]Type
	VarTypes  map[parser.VarDeclaration]Type
}

func (r *CheckResult) getExprType(expr parser.ValueExpr) Type {
	exprType, ok := r.ExprTypes[expr]
	if !ok {
		t := TVar{}
		r.ExprTypes[expr] = &t
		return &t
	}
	return exprType
}

func (r *CheckResult) getVarDeclType(decl parser.VarDeclaration) Type {
	exprType, ok := r.VarTypes[decl]
	if !ok {
		t := TVar{}
		r.VarTypes[decl] = &t
		return &t
	}
	return exprType
}

func (r *CheckResult) unifyNodeWith(expr parser.ValueExpr, t Type) {
	exprT := r.getExprType(expr)
	r.unify(expr.GetRange(), exprT, t)
}

func (r *CheckResult) unify(rng parser.Range, t1 Type, t2 Type) {
	ok := Unify(t1, t2)
	if ok {
		return
	}

	r.Diagnostics = append(r.Diagnostics, Diagnostic{
		Range: rng,
		Kind: &AssetMismatch{
			Expected: TypeToString(t1),
			Got:      TypeToString(t2),
		},
	})
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
		Program: program,

		emptiedAccount:   make(map[string]struct{}),
		declaredVars:     make(map[string]parser.VarDeclaration),
		unusedVars:       make(map[string]parser.Range),
		varResolution:    make(map[*parser.Variable]parser.VarDeclaration),
		fnCallResolution: make(map[*parser.FnCallIdentifier]FnCallResolution),
		ExprTypes:        make(map[parser.ValueExpr]Type),
		VarTypes:         make(map[parser.VarDeclaration]Type),
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
				res.checkExpression(*varDecl.Origin, varDecl.Type.Name)
			}
		}
	}

	for _, statement := range res.Program.Statements {
		res.unboundedAccountInSend = nil
		res.checkStatement(statement)
	}

	// after static AST traversal is complete, check for unused vars
	for name, rng := range res.unusedVars {
		res.pushDiagnostic(rng, UnusedVar{Name: name})
	}
}

func (res *CheckResult) checkStatement(statement parser.Statement) {
	res.emptiedAccount = make(map[string]struct{})
	res.stmtType = &TVar{}

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
		res.pushDiagnostic(parserError.Range, Parsing{Description: parserError.Msg})
	}
	res.check()
	return res
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
			res.pushDiagnostic(fnCall.Range, BadArity{
				Expected: expectedArgs,
				Actual:   actualArgs,
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

				res.pushDiagnostic(rng, BadArity{
					Expected: expectedArgs,
					Actual:   actualArgs,
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

		res.pushDiagnostic(fnCall.Caller.Range, UnknownFunction{
			Name: fnCall.Caller.Name,
		})
	}
}

func isTypeAllowed(typeName string) bool {
	return slices.Contains(AllowedTypes, typeName)
}

func (res *CheckResult) checkVarType(typeDecl parser.TypeDecl) {
	if !isTypeAllowed(typeDecl.Name) {
		res.pushDiagnostic(typeDecl.Range, InvalidType{Name: typeDecl.Name})
	}
}

func (res *CheckResult) checkDuplicateVars(variableName parser.Variable, decl parser.VarDeclaration) {
	// check there aren't duplicate variables
	if _, ok := res.declaredVars[variableName.Name]; ok {
		res.pushDiagnostic(variableName.Range, DuplicateVariable{Name: variableName.Name})
	} else {
		res.declaredVars[variableName.Name] = decl
		res.unusedVars[variableName.Name] = variableName.Range
	}
}

func (res *CheckResult) checkFnCall(fnCall parser.FnCall) string {
	returnType := TypeAny

	if resolution, ok := Builtins[fnCall.Caller.Name]; ok {
		if resolution, ok := resolution.(VarOriginFnCallResolution); ok {
			res.fnCallResolution[fnCall.Caller] = resolution
			returnType = resolution.Return

			res.requireVersion(fnCall.Range, resolution.VersionConstraints...)
		}
	}

	// this must come after resolution
	res.checkFnCallArity(&fnCall)

	return returnType
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
			res.unifyNodeWith(lit, res.getVarDeclType(varDeclaration))
		} else {
			res.pushDiagnostic(lit.Range, UnboundVariable{Name: lit.Name, Type: typeHint})
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
		/*
			we unify $mon and $asset in:
			`let $mon := [$asset 42]`
		*/
		res.unifyNodeWith(lit, res.getExprType(lit.Asset))
		return TypeMonetary

	case *parser.BinaryInfix:
		switch lit.Operator {
		case parser.InfixOperatorPlus:
			return res.checkInfixOverload(lit, []string{TypeNumber, TypeMonetary})

		case parser.InfixOperatorMinus:
			return res.checkInfixOverload(lit, []string{TypeNumber, TypeMonetary})

		case parser.InfixOperatorDiv:
			res.checkInfixVersion(*lit)

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
		res.checkAccountInterpolationVersion(*lit)

		for _, part := range lit.Parts {
			if v, ok := part.(*parser.Variable); ok {
				res.checkExpression(v, TypeAny)
			}
		}
		return TypeAccount
	case *parser.PercentageLiteral:
		return TypePortion
	case *parser.AssetLiteral:
		t := TAsset(lit.Asset)
		res.unifyNodeWith(lit, &t)
		return TypeAsset
	case *parser.NumberLiteral:
		return TypeNumber
	case *parser.StringLiteral:
		return TypeString

	case *parser.FnCall:
		return res.checkFnCall(*lit)

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

	res.pushDiagnostic(bin.Left.GetRange(), TypeMismatch{
		Expected: strings.Join(allowed, "|"),
		Got:      leftType,
	})
	return TypeAny
}

func (res *CheckResult) assertHasType(lit parser.ValueExpr, requiredType string, actualType string) {
	if requiredType == TypeAny || actualType == TypeAny || requiredType == actualType {
		return
	}

	res.pushDiagnostic(lit.GetRange(), TypeMismatch{
		Expected: requiredType,
		Got:      actualType,
	})
}

func (res *CheckResult) checkSentValue(sentValue parser.SentValue) {
	switch sentValue := sentValue.(type) {
	case *parser.SentValueAll:
		res.checkExpression(sentValue.Asset, TypeAsset)
	case *parser.SentValueLiteral:
		res.checkExpression(sentValue.Monetary, TypeMonetary)

		res.unifyNodeWith(sentValue.Monetary, res.stmtType)
		res.unifyNodeWith(
			sentValue.Monetary,
			res.stmtType,
		)
	}
}

func (res *CheckResult) checkSource(source parser.Source) {
	if source == nil {
		return
	}

	if res.unboundedAccountInSend != nil {
		res.pushDiagnostic(source.GetRange(), UnboundedAccountIsNotLast{})
	}

	switch source := source.(type) {
	case *parser.SourceAccount:
		res.checkExpression(source.ValueExpr, TypeAccount)
		res.checkExpression(source.Color, TypeString)
		if account, ok := source.ValueExpr.(*parser.AccountInterpLiteral); ok {
			if account.IsWorld() && res.unboundedSend {
				res.pushDiagnostic(source.GetRange(), InvalidUnboundedAccount{})
			} else if account.IsWorld() {
				res.unboundedAccountInSend = account
			}

			coloredAccountName := account.String()
			switch col := source.Color.(type) {
			case *parser.Variable:
				coloredAccountName += "\\$" + col.Name
			case *parser.StringLiteral:
				if col.String != "" {
					coloredAccountName += "\\\"" + col.String + "\""
				}
			}

			if _, emptied := res.emptiedAccount[coloredAccountName]; emptied && !account.IsWorld() {
				res.pushDiagnostic(account.Range, EmptiedAccount{Name: account.String()})
			}

			res.emptiedAccount[coloredAccountName] = struct{}{}
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
			if res.unboundedSend {
				res.pushDiagnostic(source.Address.GetRange(), InvalidUnboundedAccount{})
			}

		}

		res.checkExpression(source.Address, TypeAccount)
		res.checkExpression(source.Color, TypeString)
		if source.Bounded != nil {
			res.checkExpression(*source.Bounded, TypeMonetary)
		}

	case *parser.SourceInorder:
		for _, source := range source.Sources {
			res.checkSource(source)
		}

	case *parser.SourceOneof:
		res.checkOneofVersion(source.Range)

		for _, source := range source.Sources {
			res.checkSource(source)
		}

	case *parser.SourceCapped:
		onExit := res.enterCappedSource()

		res.unifyNodeWith(source.Cap, res.stmtType)

		res.checkExpression(source.Cap, TypeMonetary)
		res.checkSource(source.From)

		onExit()

	case *parser.SourceAllotment:
		if res.unboundedSend {
			res.pushDiagnostic(source.Range, NoAllotmentInSendAll{})
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
					res.pushDiagnostic(source.Range, RemainingIsNotLast{})
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
				res.pushDiagnostic(expr.Range, DivByZero{})
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
		res.checkOneofVersion(destination.Range)

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
					res.pushDiagnostic(destination.Range, RemainingIsNotLast{})
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
			res.pushDiagnostic(variableLiterals[0].GetRange(), FixedPortionVariable{
				Value: *value.Sub(big.NewRat(1, 1), &sum),
			})
		} else {
			res.pushDiagnostic(rng, BadAllotmentSum{Sum: sum})
		}

	// sum == 1
	case 0:
		for _, varLit := range variableLiterals {
			res.pushDiagnostic(varLit.GetRange(), FixedPortionVariable{
				Value: *big.NewRat(0, 1),
			})
		}
		if remaining != nil {
			res.pushDiagnostic(remaining.Range, RedundantRemaining{})
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

func (res *CheckResult) pushDiagnostic(rng parser.Range, kind DiagnosticKind) {
	id := res.nextDiagnosticId
	res.nextDiagnosticId++

	res.Diagnostics = append(res.Diagnostics, Diagnostic{
		Range: rng,
		Kind:  kind,
		Id:    id,
	})
}
