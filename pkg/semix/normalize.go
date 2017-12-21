package semix

import (
	"regexp"
	"strings"
)

// NormalizeString normalizes a given string.
// The normalization converts any non empty sequence of
// punctuation or whitespace characters to exactly one whitespace.
//
// If sourround is true, the result string is sourrounded
// with exactly one whitespace.
func NormalizeString(str string, sourround bool) string {
	str = strings.Trim(normalizeRegexp.ReplaceAllLiteralString(str, " "), " ")
	if sourround {
		str = " " + str + " "
	}
	return str
}

var normalizeRegexp = regexp.MustCompile(`[\s\pP\pS\pZ]+`)
