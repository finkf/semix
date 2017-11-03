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
func Search(g *semix.Graph, d semix.Dictionary, str string) []*semix.Concept {
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
	if id, ok := d[str]; ok {
		if c, ok := g.FindByID(id); ok {
			add(c)
		}
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
	for entry, id := range d {
		if strings.Contains(entry, str) {
			if c, ok := g.FindByID(id); ok {
				add(c)
			}
		}
		// no need to check the concept, since we did this already.
	}
	return cs
}

// SearchDictionaryEntries iterates over a dictionary and returns
// all entries in the dictionary that reference the given concept.
func SearchDictionaryEntries(d semix.Dictionary, c *semix.Concept) []string {
	if c == nil || d == nil || c.ID() == 0 {
		return nil
	}
	var entries []string
	for entry, id := range d {
		if abs(id) == abs(c.ID()) {
			entries = append(entries, entry)
		}
	}
	return entries
}

func abs(id int32) int32 {
	if id < 0 {
		return -id
	}
	return id
}
