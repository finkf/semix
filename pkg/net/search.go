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
	set := make(map[string]bool)
	var cs []*semix.Concept
	add := func(c *semix.Concept) {
		if !set[c.URL()] {
			cs = append(cs, c)
			set[c.URL()] = true
		}
	}
	if c, ok := g.FindByURL(str); ok {
		add(c)
	}
	if c, ok := d[" "+str+" "]; ok {
		add(c)
	}
	// iterate over concepts.
	for i := 0; i < g.ConceptsLen(); i++ {
		c := g.ConceptAt(i)
		if strings.Contains(c.URL(), str) {
			add(c)
		}
		if strings.Contains(c.Name, str) {
			add(c)
		}
	}
	// iterate over dictionary entries
	for e, c := range d {
		if strings.Contains(e, str) {
			add(c)
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
