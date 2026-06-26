// Builder to format numscript document
package builder

import (
	"fmt"
	"math/big"
	"strings"
)

const indentStr = "  "

func (p *pool[T]) getFreshId() int {
	id := p.nextId
	p.nextId += 1
	return id
}

func (p *pool[T]) getItemId(elem T) int {
	previousLookup, isElemInPool := p.elems[elem]
	if !isElemInPool {
		elemId := p.getFreshId()
		p.elems[elem] = elemId
		previousLookup = elemId
	}
	return previousLookup
}

type env struct {
	builder      strings.Builder
	accountsPool pool[string]
	assetsPool   pool[string]
	stringsPool  pool[string]
	numbersPool  pool[*big.Int]
	varsEnv      VarsEnv
}

func writeIndentation(env *env, w int) {
	if w == 0 {
		return
	}

	env.builder.Grow(w * len(indentStr))
	for range w {
		env.builder.WriteString(indentStr)
	}
}

func newEnv() env {
	return env{
		accountsPool: newPool[string](),
		assetsPool:   newPool[string](),
		stringsPool:  newPool[string](),
		numbersPool:  newPool[*big.Int](),
		varsEnv:      VarsEnv{bindings: map[anyVar]string{}},
	}
}

// The underlying type of any a pretty printing document
type render = func(
	env *env,

	// The current width
	w int,
)

func itemIdToName(id int, prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, id)
}
func accountToName(id int) string {
	return itemIdToName(id, "account")
}
func assetToName(id int) string {
	return itemIdToName(id, "asset")
}
func stringToName(id int) string {
	return itemIdToName(id, "string")
}
func numberToName(id int) string {
	return itemIdToName(id, "number")
}

type varRenderState struct {
	hasVars       bool
	sb            strings.Builder
	knownBindings map[string]string
}

func renderVar[T comparable](
	st *varRenderState,

	typ string,
	pool pool[T],
	getVarName func(id int) string,
	stringifyValue func(value T) string,
) {
	for key, id := range pool.elems {
		varName := getVarName(id)
		st.knownBindings[varName] = stringifyValue(key)
	}

	for id := range pool.nextId {
		st.hasVars = true

		varName := getVarName(id)

		st.sb.WriteString(indentStr)
		st.sb.WriteString(typ)
		st.sb.WriteString(" $")
		st.sb.WriteString(varName)
		st.sb.WriteByte('\n')
	}

}

func stringId(x string) string { return x }

func renderVars(
	st *varRenderState,
	env *env,
) string {

	st.sb.WriteString("vars {\n")
	renderVar(st, "account", env.accountsPool, accountToName, stringId)
	renderVar(st, "string", env.stringsPool, stringToName, stringId)
	renderVar(st, "asset", env.assetsPool, assetToName, stringId)
	renderVar(st, "number", env.numbersPool, numberToName, func(bi *big.Int) string {
		return bi.String()
	})
	st.sb.WriteString("}\n\n")

	if !st.hasVars {
		return ""
	}

	return st.sb.String()
}

func BuildProgram(statements ...Statement) (map[string]string, VarsEnv, string) {
	env := newEnv()
	for _, stmt := range statements {
		stmt(&env, 0)
	}

	st := varRenderState{
		knownBindings: make(map[string]string),
	}
	// AFTER we've rendered the whole program, we can render the vars block
	vars := renderVars(&st, &env)

	return st.knownBindings, env.varsEnv, vars + env.builder.String()
}

// Check feature flag has only lower chars and "-" chars.
// Less declarative than a regex, but this way we don't need a more complex api user-wise
// just for the sake of perfs (this should be good enough)
func checkIsFlagValid(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if (ch < 'a' || ch > 'z') && ch != '-' {
			return false
		}
	}
	return true
}

// Same as `BuildProgram`, but accepts features flag.
//
// IMPORTANT NOTE: this function will panic if a feature flag doesn't match `^[a-z-]+$`. Flags are meant to be passed directly, and a sintactically
// incorrect flag is treated as an argument error panic right away, not returned as an "error".
// Also note we don't keep the list of valid flags here
func BuildProgramWithFeatureFlags(
	featureFlags []string,
	statements ...Statement,
) (map[string]string, VarsEnv, string) {
	knownBindings, varsEnv, script := BuildProgram(statements...)

	var flagsArgs strings.Builder
	for index, flag := range featureFlags {
		if !checkIsFlagValid(flag) {
			// Yes, we are panicking instead of returning an error here.
			// That's desidered: flags are meant to be passed manually. Not computed, created conditionally, etc.
			// If a feature flag is wrong we want to crash the thing immediately instead of having the user handle that, log that, or whatever.
			panic(fmt.Sprintf("Invalid argument: the `%s` feature flag is invalid. Only flags matching `^[a-z-]+$` are accepted.", flag))
		}

		if index != 0 {
			flagsArgs.WriteString(", ")
		}

		flagsArgs.WriteByte('"')
		flagsArgs.WriteString(flag)
		flagsArgs.WriteByte('"')

	}

	updatedScript := fmt.Sprintf("#![feature(%s)]\n%s", flagsArgs.String(), script)
	return knownBindings, varsEnv, updatedScript
}
