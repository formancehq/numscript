package interpreter

import (
	"regexp"
)

var scopeRegex = regexp.MustCompile(`^[a-z0-9_]*$`)

func validateScope(scope string) bool {
	return scopeRegex.MatchString(scope)
}
