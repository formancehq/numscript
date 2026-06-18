// Builder to format numscript document
package builder

import (
	"fmt"
	"strings"
)

const identStr = "  "

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

	env.builder.Grow(w * len(identStr))
	for range w {
		env.builder.WriteString(identStr)
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

// TODO double check this one (do we need to handle vars?)
func BuildProgram(statements ...Statement) (any, string) {
	env := newEnv()
	for _, stmt := range statements {
		stmt(env, 0)
	}

	// TODO!! vars needs to be returned
	return nil, env.builder.String()
}
