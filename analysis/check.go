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
	Diagnostics []Diagnostic
}

func Check(p parser.Program) CheckResult {
	res := CheckResult{}
	res.checkProgram(p)
	return res
}

func (res *CheckResult) checkProgram(program parser.Program) {
	declaredVars := make(map[string]struct{})

	for _, varDecl := range program.Vars {
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
		if _, ok := declaredVars[varDecl.Name.Name]; ok {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: varDecl.Name.Range,
				Kind:  &DuplicateVariable{Name: varDecl.Name.Name},
			})
		} else {
			declaredVars[varDecl.Name.Name] = struct{}{}
		}
	}
}
