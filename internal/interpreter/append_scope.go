package interpreter

import (
	"regexp"
	"strings"
)

// Precondition: scope is valid idenfitier
func appendScope(acc string, scope string) string {
	splits := strings.Split(acc, "/")

	acc = splits[0]
	// the old scope is splits[1]

	if scope == "" {
		return acc
	}

	return acc + "/" + scope
}

var scopeRegex = regexp.MustCompile(`^[a-z0-9_]*$`)

func validateScope(scope string) bool {
	return scopeRegex.MatchString(scope)
}
