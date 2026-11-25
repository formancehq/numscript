package analysis

import (
	"fmt"
	"slices"
	"strings"

	"github.com/formancehq/numscript/internal/parser"
)

type CompletionItemKind uint8

const (
	CompletionKeyword CompletionItemKind = iota
	CompletionVariable
	CompletionFunction
	CompletionType
	CompletionAccount
	CompletionAsset
	CompletionSnippet
)

type CompletionItem struct {
	Label  string
	Kind   CompletionItemKind
	Detail string
	Docs   string

	// The range of text that the completion replaces
	Range parser.Range

	// The text inserted when snippets aren't supported
	InsertText string
	// The text inserted when snippets are supported. Empty when the item isn't a snippet.
	SnippetText string

	// Items with a lower priority are displayed first
	Priority int
}

const (
	priorityVariable = iota
	priorityLiteral
	priorityFunction
	priorityKeyword
	prioritySnippet
)

// Returns the completions that are valid at the given position.
// The program does not need to be valid: in fact it's usually being typed.
func Complete(source string, position parser.Position) []CompletionItem {
	parsed, ctx := parser.ParseForCompletion(source, position)
	if ctx == nil {
		return nil
	}

	c := completer{
		source:   source,
		position: position,
		ctx:      ctx,
		program:  parsed.Value,
		seen:     map[string]struct{}{},
	}
	return c.complete()
}

type completer struct {
	source   string
	position parser.Position
	ctx      *parser.CompletionContext
	program  parser.Program
	items    []CompletionItem
	seen     map[string]struct{}
}

func (c *completer) add(item CompletionItem) {
	key := fmt.Sprintf("%d:%s", item.Kind, item.Label)
	if _, ok := c.seen[key]; ok {
		return
	}
	c.seen[key] = struct{}{}

	if item.Range == (parser.Range{}) {
		item.Range = c.ctx.Range
	}
	if item.InsertText == "" {
		item.InsertText = item.Label
	}
	c.items = append(c.items, item)
}

func (c *completer) complete() []CompletionItem {
	for _, slot := range c.ctx.Slots {
		switch slot := slot.(type) {
		case *parser.ExprSlot:
			c.completeExpr(c.expectedTypes(slot))
		case *parser.AccountPartSlot:
			c.completeAccountPart(slot)
		case *parser.VarTypeSlot:
			c.completeVarType()
		case *parser.StatementSlot:
			c.completeStatement()
		}
	}

	for _, keyword := range c.ctx.Keywords {
		c.add(CompletionItem{
			Label:    keyword,
			Kind:     CompletionKeyword,
			Priority: priorityKeyword,
		})
	}

	// When a variable or an account is being typed, other kinds of completions are just noise
	if strings.HasPrefix(c.ctx.Prefix, "$") || strings.HasPrefix(c.ctx.Prefix, "@") {
		prefix := c.ctx.Prefix[:1]
		c.items = slices.DeleteFunc(c.items, func(item CompletionItem) bool {
			return !strings.HasPrefix(item.Label, prefix)
		})
	}

	slices.SortStableFunc(c.items, func(a, b CompletionItem) int {
		return a.Priority - b.Priority
	})
	return c.items
}

// The types the expression at the cursor is expected to have. An empty slice means any type.
func (c *completer) expectedTypes(slot *parser.ExprSlot) []string {
	if len(slot.Types) != 0 {
		return slot.Types
	}

	if len(c.ctx.FnCalls) == 0 {
		return nil
	}

	call := c.ctx.FnCalls[len(c.ctx.FnCalls)-1]
	resolution, ok := Builtins[call.Name]
	if !ok {
		return nil
	}

	params := resolution.GetParams()
	if call.ArgIndex >= len(params) || params[call.ArgIndex] == TypeAny {
		return nil
	}
	return []string{params[call.ArgIndex]}
}

func typeAllowed(t string, expected []string) bool {
	return len(expected) == 0 || slices.Contains(expected, t)
}

func (c *completer) completeExpr(expected []string) {
	if c.program.Vars != nil {
		for _, decl := range c.program.Vars.Declarations {
			if decl.Name == nil || decl.Type == nil || !typeAllowed(decl.Type.Name, expected) {
				continue
			}
			c.add(CompletionItem{
				Label:    "$" + decl.Name.Name,
				Kind:     CompletionVariable,
				Detail:   decl.Type.Name,
				Priority: priorityVariable,
			})
		}
	}

	if typeAllowed(TypeAccount, expected) {
		for _, account := range c.knownAccounts() {
			c.add(CompletionItem{
				Label:    account,
				Kind:     CompletionAccount,
				Detail:   TypeAccount,
				Priority: priorityLiteral,
			})
		}
	}

	if typeAllowed(TypeAsset, expected) && len(expected) != 0 {
		for _, asset := range c.knownAssets() {
			c.add(CompletionItem{
				Label:    asset,
				Kind:     CompletionAsset,
				Detail:   TypeAsset,
				Priority: priorityLiteral,
			})
		}
	}

	for _, name := range sortedBuiltins() {
		resolution, ok := Builtins[name].(VarOriginFnCallResolution)
		if !ok || (resolution.Return != TypeAny && !typeAllowed(resolution.Return, expected)) {
			continue
		}
		c.addFunction(name, resolution)
	}
}

func (c *completer) addFunction(name string, resolution FnCallResolution) {
	var placeholders []string
	for index, param := range resolution.GetParams() {
		placeholders = append(placeholders, fmt.Sprintf("${%d:%s}", index+1, param))
	}

	c.add(CompletionItem{
		Label:       name,
		Kind:        CompletionFunction,
		Detail:      FnSignature(name, resolution),
		Docs:        fnDocs(resolution),
		InsertText:  name + "(",
		SnippetText: name + "(" + strings.Join(placeholders, ", ") + ")",
		Priority:    priorityFunction,
	})
}

func (c *completer) completeAccountPart(slot *parser.AccountPartSlot) {
	// The completion replaces the whole account literal typed so far
	end := c.ctx.Range.End
	if c.position.GtEq(end) {
		end = c.position
	}
	rng := parser.Range{Start: slot.LiteralStart, End: end}

	for _, account := range c.knownAccounts() {
		c.add(CompletionItem{
			Label:    account,
			Kind:     CompletionAccount,
			Detail:   TypeAccount,
			Range:    rng,
			Priority: priorityLiteral,
		})
	}

	// After a ':', suggest interpolating a variable (e.g. `@users:$id`)
	typed := textInRange(c.source, parser.Range{Start: slot.LiteralStart, End: c.position})
	lastColon := strings.LastIndex(typed, ":")
	if lastColon == -1 || c.program.Vars == nil {
		return
	}
	for _, decl := range c.program.Vars.Declarations {
		if decl.Name == nil || decl.Type == nil {
			continue
		}
		c.add(CompletionItem{
			Label:    typed[:lastColon+1] + "$" + decl.Name.Name,
			Kind:     CompletionVariable,
			Detail:   decl.Type.Name,
			Range:    rng,
			Priority: priorityVariable,
		})
	}
}

func (c *completer) completeVarType() {
	for _, t := range AllowedTypes {
		c.add(CompletionItem{
			Label:    t,
			Kind:     CompletionType,
			Priority: priorityVariable,
		})
	}
}

func (c *completer) completeStatement() {
	for _, name := range sortedBuiltins() {
		if resolution, ok := Builtins[name].(StatementFnCallResolution); ok {
			c.addFunction(name, resolution)
		}
	}

	if slices.Contains(c.ctx.Keywords, "send") {
		c.add(CompletionItem{
			Label:       "send",
			Kind:        CompletionSnippet,
			Detail:      "send statement",
			InsertText:  "send ",
			SnippetText: "send ${1:[ASSET amount]} (\n\tsource = ${2:@source}\n\tdestination = ${3:@destination}\n)",
			Priority:    prioritySnippet,
		})
	}

	if slices.Contains(c.ctx.Keywords, "vars") {
		c.add(CompletionItem{
			Label:       "vars",
			Kind:        CompletionSnippet,
			Detail:      "variables declaration",
			InsertText:  "vars ",
			SnippetText: "vars {\n\t${1:type} \\$${2:name}\n}",
			Priority:    prioritySnippet,
		})
	}
}

// The account literals (without interpolated variables) of the program, and @world
func (c *completer) knownAccounts() []string {
	accounts := []string{"@world"}
	walkExprs(c.program, func(expr parser.ValueExpr) {
		account, ok := expr.(*parser.AccountInterpLiteral)
		if !ok || account.Contains(c.position) {
			return
		}
		for _, part := range account.Parts {
			if _, isVar := part.(*parser.Variable); isVar {
				return
			}
		}
		name := "@" + account.String()
		if !slices.Contains(accounts, name) {
			accounts = append(accounts, name)
		}
	})
	return accounts
}

func (c *completer) knownAssets() []string {
	var assets []string
	walkExprs(c.program, func(expr parser.ValueExpr) {
		asset, ok := expr.(*parser.AssetLiteral)
		if ok && !asset.Contains(c.position) && !slices.Contains(assets, asset.Asset) {
			assets = append(assets, asset.Asset)
		}
	})
	return assets
}

type SignatureHelp struct {
	Label       string
	Params      []string
	ActiveParam int
	Docs        string
}

// Returns the signature of the builtin function whose arguments enclose the given position
func SignatureHelpAt(source string, position parser.Position) *SignatureHelp {
	_, ctx := parser.ParseForCompletion(source, position)
	if ctx == nil || len(ctx.FnCalls) == 0 {
		return nil
	}

	call := ctx.FnCalls[len(ctx.FnCalls)-1]
	resolution, ok := Builtins[call.Name]
	if !ok {
		return nil
	}

	return &SignatureHelp{
		Label:       FnSignature(call.Name, resolution),
		Params:      resolution.GetParams(),
		ActiveParam: call.ArgIndex,
		Docs:        fnDocs(resolution),
	}
}

func FnSignature(name string, resolution FnCallResolution) string {
	signature := name + "(" + strings.Join(resolution.GetParams(), ", ") + ")"
	if resolution, ok := resolution.(VarOriginFnCallResolution); ok {
		signature += " -> " + resolution.Return
	}
	return signature
}

func fnDocs(resolution FnCallResolution) string {
	switch resolution := resolution.(type) {
	case VarOriginFnCallResolution:
		return resolution.Docs
	case StatementFnCallResolution:
		return resolution.Docs
	default:
		return ""
	}
}

func sortedBuiltins() []string {
	var names []string
	for name := range Builtins {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Pre: the range is valid for the source
func textInRange(source string, rng parser.Range) string {
	lines := strings.Split(source, "\n")
	if rng.Start.Line != rng.End.Line || rng.Start.Line >= len(lines) {
		return ""
	}
	line := []rune(lines[rng.Start.Line])
	start := min(rng.Start.Character, len(line))
	end := min(max(rng.End.Character, start), len(line))
	return string(line[start:end])
}

// Calls fn on every expression of the program (including nested ones)
func walkExprs(program parser.Program, fn func(parser.ValueExpr)) {
	var walkExpr func(expr parser.ValueExpr)
	walkExpr = func(expr parser.ValueExpr) {
		if expr == nil {
			return
		}
		fn(expr)
		switch expr := expr.(type) {
		case *parser.MonetaryLiteral:
			walkExpr(expr.Asset)
			walkExpr(expr.Amount)
		case *parser.BinaryInfix:
			walkExpr(expr.Left)
			walkExpr(expr.Right)
		case *parser.Prefix:
			walkExpr(expr.Expr)
		case *parser.FnCall:
			for _, arg := range expr.Args {
				walkExpr(arg)
			}
		}
	}

	walkAllotment := func(allotment parser.AllotmentValue) {
		if allotment, ok := allotment.(*parser.ValueExprAllotment); ok {
			walkExpr(allotment.Value)
		}
	}

	var walkSource func(source parser.Source)
	walkSource = func(source parser.Source) {
		switch source := source.(type) {
		case *parser.SourceAccount:
			walkExpr(source.ValueExpr)
			walkExpr(source.Color)
		case *parser.SourceOverdraft:
			walkExpr(source.Address)
			walkExpr(source.Color)
			if source.Bounded != nil {
				walkExpr(*source.Bounded)
			}
		case *parser.SourceWithScaling:
			walkExpr(source.Address)
			walkExpr(source.Through)
		case *parser.SourceCapped:
			walkExpr(source.Cap)
			walkSource(source.From)
		case *parser.SourceInorder:
			for _, s := range source.Sources {
				walkSource(s)
			}
		case *parser.SourceOneof:
			for _, s := range source.Sources {
				walkSource(s)
			}
		case *parser.SourceAllotment:
			for _, item := range source.Items {
				walkAllotment(item.Allotment)
				walkSource(item.From)
			}
		}
	}

	var walkDestination func(destination parser.Destination)
	walkKeptOrDestination := func(target parser.KeptOrDestination) {
		if to, ok := target.(*parser.DestinationTo); ok {
			walkDestination(to.Destination)
		}
	}
	walkDestination = func(destination parser.Destination) {
		switch destination := destination.(type) {
		case *parser.DestinationAccount:
			walkExpr(destination.ValueExpr)
		case *parser.DestinationInorder:
			for _, clause := range destination.Clauses {
				walkExpr(clause.Cap)
				walkKeptOrDestination(clause.To)
			}
			walkKeptOrDestination(destination.Remaining)
		case *parser.DestinationOneof:
			for _, clause := range destination.Clauses {
				walkExpr(clause.Cap)
				walkKeptOrDestination(clause.To)
			}
			walkKeptOrDestination(destination.Remaining)
		case *parser.DestinationAllotment:
			for _, item := range destination.Items {
				walkAllotment(item.Allotment)
				walkKeptOrDestination(item.To)
			}
		}
	}

	walkSentValue := func(sentValue parser.SentValue) {
		switch sentValue := sentValue.(type) {
		case *parser.SentValueLiteral:
			walkExpr(sentValue.Monetary)
		case *parser.SentValueAll:
			walkExpr(sentValue.Asset)
		}
	}

	if program.Vars != nil {
		for _, decl := range program.Vars.Declarations {
			if decl.Origin != nil {
				walkExpr(*decl.Origin)
			}
		}
	}

	for _, statement := range program.Statements {
		switch statement := statement.(type) {
		case *parser.SendStatement:
			walkSentValue(statement.SentValue)
			walkSource(statement.Source)
			walkDestination(statement.Destination)
		case *parser.SaveStatement:
			walkSentValue(statement.SentValue)
			walkExpr(statement.Account)
		case *parser.FnCall:
			walkExpr(statement)
		}
	}
}
