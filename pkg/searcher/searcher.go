package searcher

import (
	"strings"

	"bitbucket.org/fflo/semix/pkg/semix"
)

// New create a new Searcher instance.
func New(g *semix.Graph, d semix.Dictionary) Searcher {
	return Searcher{graph: g, dict: d}
}

// Searcher holds a graph and a dictionary to search for concepts
// by name and URL.
type Searcher struct {
	graph *semix.Graph
	dict  semix.Dictionary
}

// FindByID mimics the graph searching interface.
func (s Searcher) FindByID(id int) (*semix.Concept, bool) {
	return s.graph.FindByID(int32(id))
}

// FindByURL mimics the graph searching interface.
func (s Searcher) FindByURL(url string) (*semix.Concept, bool) {
	return s.graph.FindByURL(url)
}

// SearchConcepts searches for maximal n concepts whose
// dictionary entries or URLs match a given query string.
// If n < 0, all matching concepts are returned.
func (s Searcher) SearchConcepts(q string, n int) []*semix.Concept {
	if n == 0 {
		return nil
	}
	if c, ok := s.graph.FindByURL(q); ok {
		return []*semix.Concept{c}
	}
	if id, ok := s.dict[q]; ok {
		if c, ok := s.graph.FindByID(id); ok {
			return []*semix.Concept{c}
		}
	}
	return s.searchMatchingConcepts(q, n)
}

// SearchParents searches maximal n parent concepts of a given URL.
// If n < 0, all matching concepts are returned.
func (s Searcher) SearchParents(url string, n int) []*semix.Concept {
	if n == 0 {
		return nil
	}
	c, ok := s.graph.FindByURL(url)
	if !ok {
		return nil
	}
	var res []*semix.Concept
	done := func() bool {
		return n > 0 && len(res) >= n
	}
	for i := 0; i < s.graph.ConceptsLen() && !done(); i++ {
		p := s.graph.ConceptAt(i)
	edges:
		for i := 0; i < p.EdgesLen(); i++ {
			e := p.EdgeAt(i)
			if e.O.URL() != c.URL() {
				continue edges
			}
			res = append(res, p)
			break edges
		}
	}
	return res
}

// SearchDictionaryEntries iterates over a dictionary and returns
// all entries in the dictionary that reference the concept with the given URL.
func (s Searcher) SearchDictionaryEntries(url string) []string {
	c, ok := s.graph.FindByURL(url)
	if !ok {
		return nil
	}
	var res []string
	for entry, id := range s.dict {
		if abs(id) == abs(c.ID()) {
			res = append(res, entry)
		}
	}
	return res
}

func abs(id int32) int32 {
	if id < 0 {
		return -id
	}
	return id
}

func (s Searcher) searchMatchingConcepts(q string, n int) []*semix.Concept {
	set := make(map[string]bool)
	var res []*semix.Concept
	addToResults := func(c *semix.Concept) {
		if !set[c.URL()] {
			res = append(res, c)
			set[c.URL()] = true
		}
	}
	done := func() bool {
		return n > 0 && len(res) >= n
	}
	// iterate over concepts.
	for i := 0; i < s.graph.ConceptsLen() && !done(); i++ {
		c := s.graph.ConceptAt(i)
		if strings.Contains(c.URL(), q) {
			addToResults(c)
		}
		if strings.Contains(c.Name, q) {
			addToResults(c)
		}
	}
	// iterate over dictionary entries
	for entry, id := range s.dict {
		if done() {
			break
		}
		if strings.Contains(entry, q) {
			if c, ok := s.graph.FindByID(id); ok {
				addToResults(c)
			}
		}
		// no need to check the concept, since we did this already.
	}
	return res
}
