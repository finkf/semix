package net

import (
	"sort"

	"bitbucket.org/fflo/semix/pkg/semix"
)

// Tokens represents an array of tokens.
type Tokens struct {
	Tokens []semix.Token
}

// Counts returns a sorted slice of Counts.
func (ts Tokens) Counts() []Count {
	urls := make(map[string]*semix.Concept)
	m := make(map[string]int)
	var n int
	for _, t := range ts.Tokens {
		if _, ok := urls[t.Concept.URL()]; !ok {
			urls[t.Concept.URL()] = t.Concept
		}
		m[t.Concept.URL()]++
		t.Concept.Edges(func(edge semix.Edge) {
			m[edge.O.URL()]++
			if _, ok := urls[edge.O.URL()]; !ok {
				urls[edge.O.URL()] = edge.O
			}
		})
		n++
	}
	counts := make([]Count, 0, len(m))
	for url, count := range m {
		counts = append(counts, Count{Concept: urls[url], Total: n, N: count})
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].N > counts[j].N
	})
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
