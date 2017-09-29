package net

import (
	"strings"

	"bitbucket.org/fflo/semix/pkg/semix"
)

// Search searches the concept for a given string.
// If a concept can be found the function returns the concept and true.
// If nothing could be found (nil, false) is returned.
//
// First it tries to find the concept with a simple URL lookup in the Graph.
// Then it tries a lookup in the dictionary.
// Then it iterates over all URLs and dictionary entries and returns the first
// matching Concept.
func Search(g *semix.Graph, d map[string]*semix.Concept, str string) (*semix.Concept, bool) {
	if c, ok := g.FindByURL(str); ok {
		return c, true
	}
	if c, ok := d[" "+str+" "]; ok {
		return c, true
	}
	// iterate over concepts.
	for i := 0; i < g.ConceptsLen(); i++ {
		c := g.ConceptAt(i)
		if strings.Contains(c.URL(), str) {
			return c, true
		}
		if strings.Contains(c.Name, str) {
			return c, true
		}
	}
	// iterate over dictionary entries
	for e, c := range d {
		if strings.Contains(e, str) {
			return c, true
		}
		// no need to check the concept, since we did this already.
	}
	return nil, false
}

// SearchDictionaryEntries iterates over a dictionary and returns
// all entries in the dictionary that reference the given concept.
func SearchDictionaryEntries(d map[string]*semix.Concept, c *semix.Concept) []string {
	if c == nil || d == nil {
		return nil
	}
	var entries []string
	for e, cc := range d {
		if cc.URL() == c.URL() {
			entries = append(entries, strings.Trim(e, " "))
		}
	}
	return entries
}
