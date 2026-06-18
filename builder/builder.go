// Builder to format numscript document
package builder

import (
	"fmt"
	"strings"
)

const indentStr = "  "

func (p pool[T]) getItemId(elem T) int {
	previousLookup, isElemInPool := p.elems[elem]
	if !isElemInPool {
		elemId := len(p.elems)
		p.elems[elem] = elemId
		previousLookup = elemId
	}
	return previousLookup
}

type env struct {
	builder *strings.Builder

	accountsPool pool[string]
	assetsPool   pool[string]
	stringsPool  pool[string]
	// numbersPool  pool[*big.Int]
}

func writeIndentation(env env, w int) {
	if w == 0 {
		return
	}

	env.builder.Grow(w * len(indentStr))
	for range w {
		env.builder.WriteString(indentStr)
	}
}

func newEnv() env {
	var sb strings.Builder
	return env{
		builder:      &sb,
		accountsPool: newPool[string](),
		assetsPool:   newPool[string](),
		stringsPool:  newPool[string](),
		// numbersPool:  newPool[*big.Int](),
	}
}

// The underlying type of any a pretty printing document
type render = func(
	env env,

	// The current width
	w int,
)

func itemIdToName(id int, prefix string) string {
	return fmt.Sprintf("$%s_%d", prefix, id)
}
func accountToName(id int) string {
	return itemIdToName(id, "account")
}
func assetToName(id int) string {
	return itemIdToName(id, "asset")
}
func numberToName(id int) string {
	return itemIdToName(id, "number")
}
func stringToName(id int) string {
	return itemIdToName(id, "string")
}

func renderVars(env env) string {
	var sb strings.Builder

	hasVars := false

	renderVarsTyp := func(
		typ string,
		pool pool[string],
		getVarName func(id int) string,
	) {
		for id := range len(pool.elems) {
			hasVars = true
			sb.WriteString(indentStr)
			sb.WriteString(typ)
			sb.WriteString(" ")
			sb.WriteString(getVarName(id))
			sb.WriteByte('\n')
		}
	}

	sb.WriteString("vars {\n")
	renderVarsTyp("account", env.accountsPool, accountToName)
	renderVarsTyp("string", env.stringsPool, stringToName)
	renderVarsTyp("asset", env.assetsPool, assetToName)
	sb.WriteString("}\n\n")

	if !hasVars {
		return ""
	}

	return sb.String()
}

// TODO double check this one (do we need to handle vars?)
func BuildProgram(statements ...Statement) (any, string) {
	env := newEnv()
	for _, stmt := range statements {
		stmt(env, 0)
	}

	// AFTER we've rendered the whole program, we can render the vars block
	vars := renderVars(env)

	// TODO!! vars needs to be returned
	return nil, vars + env.builder.String()
}
