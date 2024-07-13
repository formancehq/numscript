package analysis

import (
	"numscript/parser"
)

var AllowedTypes = []string{"monetary", "account", "portion", "asset", "number", "string"}

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
			res.checkLiteral(statement.Monetary)
			res.checkSource(statement.Source)
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

func (res *CheckResult) checkVarDecl(varDecl parser.VarDeclaration) {
	// check type is a valid type (e.g. portion, account, ...)
	isAllowed := false
	for _, allowedType := range AllowedTypes {
		if allowedType == varDecl.Type.Name {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		res.Diagnostics = append(res.Diagnostics, Diagnostic{
			Range: varDecl.Type.Range,
			Kind:  &InvalidType{Name: varDecl.Type.Name},
		})
	}

	// check there aren't duplicate variables
	if _, ok := res.declaredVars[varDecl.Name.Name]; ok {
		res.Diagnostics = append(res.Diagnostics, Diagnostic{
			Range: varDecl.Name.Range,
			Kind:  &DuplicateVariable{Name: varDecl.Name.Name},
		})
	} else {
		res.declaredVars[varDecl.Name.Name] = varDecl
		res.unusedVars[varDecl.Name.Name] = varDecl.Name.Range
	}
}

func (res *CheckResult) checkLiteral(lit parser.Literal) {
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
		return

	default:
		// TODO
		panic("TODO ")
	}
}

func (res *CheckResult) checkSource(source parser.Source) {
	switch source := source.(type) {
	case *parser.VariableLiteral:
		res.checkLiteral(source)

	case *parser.SourceSeq:
		for _, source := range source.Sources {
			res.checkSource(source)
		}

	case *parser.SourceCapped:
		res.checkLiteral(source.Cap)
		res.checkSource(source.From)

	case *parser.SourceAllotment:
		for _, allottedItem := range source.Items {
			res.checkSource(allottedItem.From)
		}
	}
}