package analysis

import "numscript/parser"

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
	for _, var_ := range program.Vars {
		type_ := var_.Type
		isAllowed := false
		for _, allowedType := range AllowedTypes {
			if allowedType == type_.Name {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: type_.Range,
				Kind:  &InvalidType{Name: type_.Name},
			})
		}
	}
}
