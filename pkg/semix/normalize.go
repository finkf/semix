package semix

import "regexp"

// NormalizeString normalizes a given string.
// The normalization converts any non empty sequence of
// punctuation or whitespace characters to exactly one whitespace.
//
// If sourround is true, the result string is sourrounded
// with exactly one whitespace.
func NormalizeString(str string, sourround bool) string {
	if sourround {
		str = " " + str + " "
	}
	return normalizeRegexp.ReplaceAllLiteralString(str, " ")
}

var normalizeRegexp = regexp.MustCompile(`[\s\pP\pS\pZ]+`)
