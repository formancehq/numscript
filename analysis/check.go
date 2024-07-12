package analysis

import "numscript/parser"

var allowedTypes = []string{"monetary", "account", "portion", "asset", "number", "string"}

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
		isAllowed := false
		for _, allowedType := range allowedTypes {
			if allowedType == var_.Name.Name {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{
				Range: var_.Type.Range,
				Kind:  &InvalidType{Name: var_.Type.Name},
			})
		}
	}
}

func Example() {
	parsed := parser.Parse("")

	st := parsed.Value.Statements[0]

	switch st := st.(type) {
	case *parser.SendStatement:

		pointer := st

		var example map[*parser.SendStatement]int = make(map[*parser.SendStatement]int)
		example[pointer] = 0

		return
	}

}
