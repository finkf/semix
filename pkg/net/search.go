package net

import (
	"strings"

	"bitbucket.org/fflo/semix/pkg/semix"
)

// Search searches the all the concepts for a given string.
// It returns a slice of all the found concepts.
//
// First it tries to find the concept with a simple URL lookup in the Graph.
// Then it tries a lookup in the dictionary.
// Then it iterates over all URLs and dictionary entries.
func Search(g *semix.Graph, d map[string]*semix.Concept, str string) []*semix.Concept {
	var cs []*semix.Concept
	if c, ok := g.FindByURL(str); ok {
		cs = append(cs, c)
	}
	if c, ok := d[" "+str+" "]; ok {
		cs = append(cs, c)
	}
	// iterate over concepts.
	for i := 0; i < g.ConceptsLen(); i++ {
		c := g.ConceptAt(i)
		if strings.Contains(c.URL(), str) {
			cs = append(cs, c)
		}
		if strings.Contains(c.Name, str) {
			cs = append(cs, c)
		}
	}
	// iterate over dictionary entries
	for e, c := range d {
		if strings.Contains(e, str) {
			cs = append(cs, c)
		}
		// no need to check the concept, since we did this already.
	}
	return cs
}

// SearchURL searches for the concept by a given URL.
func SearchURL(d map[string]*semix.Concept, url string) (*semix.Concept, bool) {
	for _, c := range d {
		if c.URL() == url {
			return c, true
		}
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
