package rest

import (
	"fmt"
	"sort"
	"strings"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/searcher"
	"bitbucket.org/fflo/semix/pkg/semix"
)

// ConceptInfo holds information about a concept.
type ConceptInfo struct {
	Concept *semix.Concept
	Entries []string
}

// Predicates returns a map of the targets ordered by the predicates.
func (info ConceptInfo) Predicates() map[*semix.Concept][]*semix.Concept {
	m := make(map[*semix.Concept][]*semix.Concept)
	if info.Concept == nil {
		return m
	}
	info.Concept.EachEdge(func(edge semix.Edge) {
		m[edge.P] = append(m[edge.P], edge.O)
	})
	return m
}

// Token mimics semix.Token
type Token struct {
	Token, Path, RelationURL string
	Concept                  *semix.Concept
	Begin, End, L            int
}

// NewTokens converts a semix.Token to an array of tokens.
// This function returns a slice, because ambiguous Concept tokens
// are specially resolved.
func NewTokens(t semix.Token) []Token {
	ts := []Token{Token{
		Token:   t.Token,
		Path:    t.Path,
		Concept: t.Concept,
		Begin:   t.Begin,
		End:     t.End,
	}}
	if t.Concept.Ambiguous() {
		ts[0].L = -1 // set -1 for ambiguous concepts.
		c := t.Concept
		n := t.Concept.EdgesLen()
		for i := 0; i < n; i++ {
			e := c.EdgeAt(i)
			ts = append(ts, Token{
				Token:   t.Token,
				Path:    t.Path,
				Concept: e.O,
				Begin:   t.Begin,
				End:     t.End,
				L:       e.L,
			})
		}
	}
	return ts
}

// Tokens represents an array of tokens.
type Tokens struct {
	Tokens []Token
}

// Counts returns a sorted slice of Counts ordered by the according predicates.
func (ts Tokens) Counts() map[*semix.Concept][]Count {
	m := make(map[*semix.Concept]map[*semix.Concept]int)
	var n int
	for _, t := range ts.Tokens {
		n++
		if t.Concept.Ambiguous() {
			continue
		}
		t.Concept.EachEdge(func(edge semix.Edge) {
			p := edge.P
			o := edge.O
			if m[p] == nil {
				m[p] = make(map[*semix.Concept]int)
			}
			m[p][o]++
		})
	}
	counts := make(map[*semix.Concept][]Count, len(m))
	for p := range m {
		for o, count := range m[p] {
			c := Count{Concept: o, Total: n, N: count}
			counts[p] = append(counts[p], c)
		}
		sort.Slice(counts[p], func(i, j int) bool {
			return counts[p][i].N > counts[p][j].N
		})
	}
	return counts
}

// Count represent the count of concept in an array of Tokens.
type Count struct {
	Concept  *semix.Concept
	Total, N int
}

// RelativeFrequency calculates the relative frequency of a count.
func (c Count) RelativeFrequency() float32 {
	return float32(c.N) / float32(c.Total)
}

// NewTokenFromEntry creates a Token from an index.Entry
func NewTokenFromEntry(s searcher.Searcher, e index.Entry) (Token, error) {
	c, ok := s.FindByURL(e.ConceptURL)
	if !ok {
		return Token{}, fmt.Errorf("invalid url %q", e.ConceptURL)
	}
	return Token{
		Token:       e.Token,
		Path:        e.Path,
		Begin:       e.Begin,
		End:         e.End,
		RelationURL: e.RelationURL,
		L:           e.L,
		Concept:     c,
	}, nil
}

// Context specifies the context of a match
type Context struct {
	Before, Match, After, URL string
	Begin, End, Len           int
}

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

// SearchParents searches the all the parent concepts of a given URL.
func SearchParents(g *semix.Graph, url string) []*semix.Concept {
	c, ok := g.FindByURL(url)
	if !ok {
		return nil
	}
	var cs []*semix.Concept
	for i := 0; i < g.ConceptsLen(); i++ {
		p := g.ConceptAt(i)
	edges:
		for i := 0; i < p.EdgesLen(); i++ {
			e := p.EdgeAt(i)
			if e.O.URL() != c.URL() {
				continue edges
			}
			cs = append(cs, p)
			break edges
		}
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
