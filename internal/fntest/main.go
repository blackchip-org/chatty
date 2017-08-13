package fntest

import (
	"strings"
)

func AnyOf(base string, matches ...string) string {
	for _, match := range matches {
		base = strings.Replace(base, match, "X", -1)
	}
	return base
}
