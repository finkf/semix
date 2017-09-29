package net

import (
	"sort"

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
	info.Concept.Edges(func(edge semix.Edge) {
		m[edge.P] = append(m[edge.P], edge.O)
	})
	return m
}

// Tokens represents an array of tokens.
type Tokens struct {
	Tokens []semix.Token
}

// Counts returns a sorted slice of Counts ordered by the according predicates.
func (ts Tokens) Counts() map[*semix.Concept][]Count {
	urls := make(map[string]*semix.Concept)
	register := func(c *semix.Concept) *semix.Concept {
		if _, ok := urls[c.URL()]; !ok {
			urls[c.URL()] = c
		}
		return urls[c.URL()]
	}
	m := make(map[*semix.Concept]map[*semix.Concept]int)
	var n int
	for _, t := range ts.Tokens {
		n++
		preds := make(map[*semix.Concept]bool)
		t.Concept.Edges(func(edge semix.Edge) {
			if m[register(edge.P)] == nil {
				m[register(edge.P)] = make(map[*semix.Concept]int)
			}
			m[register(edge.P)][register(edge.O)]++
			preds[register(edge.P)] = true
		})
		for p := range preds {
			m[p][register(t.Concept)]++
		}
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
