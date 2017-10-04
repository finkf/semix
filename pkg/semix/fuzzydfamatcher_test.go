package semix

import (
	"fmt"
	"sort"
	"testing"
)

func TestFuzzyDFAMatcher(t *testing.T) {
	tests := []struct {
		test, want string
		k          int
	}{
		{"", "{<nil> 0 0}", 3},
		{" ", "{<nil> 0 0}", 3},
		{" ", "{<nil> 0 0}", 3},
		{" match ", "{fuzzy-concept [{fuzzy-predicate match 0}] 1 6}", 0},
		{" match ", "{fuzzy-concept [{fuzzy-predicate match 0}] 1 6}", 3},
		{" mxtch ", "{<nil> 0 0}", 0},
		{" mxtch ", "{fuzzy-concept [{fuzzy-predicate match 1}] 1 6}", 3},
		{" mxtch bbbxxx ", "{fuzzy-concept [{fuzzy-predicate match 1}] 1 6}", 3},
		{" mxtch bbxxx ", "{fuzzy-concept [{fuzzy-predicate match 1}] 1 6}", 3},
		{" mxtch bxxx ", "{fuzzy-concept [{fuzzy-predicate match 1}] 1 6}", 3},
		{" mxtch bbb ", "{fuzzy-concept [{fuzzy-predicate match 1}] 1 10}", 3},
		{" XXXXXX mxtch ", "{fuzzy-concept [{fuzzy-predicate match 1}] 8 13}", 3},
		{" mxtch XXXXXX ", "{fuzzy-concept [{fuzzy-predicate match 1}] 1 6}", 3},
		{" XXXXXX mxtch XXXXXX ", "{fuzzy-concept [{fuzzy-predicate match 1}] 8 13}", 3},
		{" mxtch mxtch ", "{fuzzy-concept [{fuzzy-predicate match two 2}] 1 12}", 3},
		{" hxt ", "{fuzzy-concept [{fuzzy-predicate hit 2} {fuzzy-predicate hit two 1}] 1 4}", 3},
		{" XXXXXXXX XXXXXXXX XXXXXXXX ", "{<nil> 0 0}", 3},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			m := makeFuzzyDFAMatcher(t, tc.k)
			if k := m.DFA.MaxError(); k != tc.k {
				t.Errorf("expected k=%d; got %d", tc.k, k)
			}
			match := m.Match(tc.test)
			if str := fuzzyConceptToString(t, match); str != tc.want {
				t.Errorf("expeceted pos = %q; got %q", tc.want, str)
			}
		})
	}
}

func makeFuzzyDFAMatcher(t *testing.T, k int) FuzzyDFAMatcher {
	t.Helper()
	graph := NewGraph()
	t1 := graph.Add("match", "x", "y")
	t2 := graph.Add("match two", "x", "y")
	t3 := graph.Add("hit", "x", "y")
	t4 := graph.Add("hit two", "x", "y")
	dictionary := map[string]*Concept{
		" match ":       t1.S,
		" match bbb ":   t1.S,
		" mitch match ": t2.S,
		" hitt ":        t3.S,
		" hot ":         t4.S,
	}
	return FuzzyDFAMatcher{
		NewFuzzyDFA(k, NewDFA(dictionary, graph)),
	}
}

func fuzzyConceptToString(t *testing.T, m MatchPos) string {
	t.Helper()
	if m.Concept == nil {
		return fmt.Sprintf("%v", m)
	}
	if !m.Concept.Ambiguous() {
		t.Errorf("expected ambiguous concept: %s", m.Concept)
	}
	return fmt.Sprintf("{%s %v %v %v}",
		m.Concept, fuzzyEdgeToString(m.Concept.edges), m.Begin, m.End)
}

func fuzzyEdgeToString(es []Edge) string {
	sort.Slice(es, func(i, j int) bool {
		return es[i].O.URL() < es[j].O.URL()
	})
	str := "["
	for i, e := range es {
		if i > 0 {
			str += " "
		}
		str += fmt.Sprintf("{%s %s %d}", e.P, e.O, e.L)
	}
	return str + "]"
}
