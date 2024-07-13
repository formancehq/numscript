package analysis

import "numscript/parser"

type GotoDefinitionResult struct {
	Range parser.Range
}

func GotoDefinition(program parser.Program, position parser.Position, checkResult CheckResult) *GotoDefinitionResult {
	hover := HoverOn(program, position)

	if variableHover, ok := hover.(*VariableHover); ok {
		resolvedVar := checkResult.ResolveVar(variableHover.Node)
		if resolvedVar == nil {
			return nil
		}

		return &GotoDefinitionResult{
			Range: resolvedVar.Name.Range,
		}
	}

	return nil
}
