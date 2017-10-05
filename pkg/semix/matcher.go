package semix

import (
	"regexp"
)

// MatchPos represents a matching position in a string.
// Concept is the associated concept of the match. It is nil if nothing
// can be matched. Begin and End mark the begin and end positions of the match
// if Concept is not nil.
type MatchPos struct {
	Concept    *Concept
	Begin, End int
}

// Matcher is a simple interface for searching a concept in a string.
type Matcher interface {
	// Match returns the MatchPos of the next concept in the given string.
	Match([]byte) MatchPos
}

// RegexMatcher uses a regex to search for a match in a string.
type RegexMatcher struct {
	Re      *regexp.Regexp
	Concept *Concept
}

// Match returns the MatchPos of the first occurence of the regex.
func (m RegexMatcher) Match(str []byte) MatchPos {
	pos := m.Re.FindIndex(str)
	if pos == nil {
		return MatchPos{}
	}
	return MatchPos{Concept: m.Concept, Begin: pos[0], End: pos[1]}
}
