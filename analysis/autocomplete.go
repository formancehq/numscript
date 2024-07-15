package analysis

import "numscript/parser"

type CompletionList struct {
	Items []CompletionItem
}
type CompletionItem struct {
	Label string
}

func HandleCompletion(program parser.Program, position parser.Position) CompletionList {
	hovered := HoverOn(program, position)

	if hovered == nil {
		return CompletionList{}
	}

	if varHover, ok := hovered.(*VariableHover); ok {

		name := varHover.Node.Name
		if name != "" {
			return CompletionList{}
		}

		var items []CompletionItem

		// TODO context-aware suggestion
		for _, varDecl := range program.Vars {
			items = append(items, CompletionItem{
				Label: varDecl.Name.Name,
			})
		}

		return CompletionList{
			Items: items,
		}
	}

	return CompletionList{}
}
