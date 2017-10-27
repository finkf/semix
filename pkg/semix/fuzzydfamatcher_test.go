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
		ambiguous  bool
	}{
		{"", "{<nil> 0 0}", 3, true},
		{" ", "{<nil> 0 0}", 3, true},
		{" ", "{<nil> 0 0}", 3, true},
		{" match ", "{match [{x y 0}] 1 6}", 0, false},
		{" match ", "{match [{x y 0}] 1 6}", 3, false},
		{" mxtch ", "{<nil> 0 0}", 0, true},
		{" mxtch ", "{fuzzy-concept [{fuzzy-predicate match 1}] 1 6}", 3, true},
		{" mxtch bbbxxx ", "{<nil> 0 0}", 3, true},
		{" mxtch bbxxx ", "{fuzzy-concept [{fuzzy-predicate match 1}] 1 6}", 3, true},
		{" mxtch bxxx ", "{fuzzy-concept [{fuzzy-predicate match 1}] 1 6}", 3, true},
		{" mxtch bbb ", "{fuzzy-concept [{fuzzy-predicate match 1}] 1 10}", 3, true},
		{" XXXXXX mxtch ", "{fuzzy-concept [{fuzzy-predicate match 1}] 8 13}", 3, true},
		{" mxtch XXXXXX ", "{fuzzy-concept [{fuzzy-predicate match 1}] 1 6}", 3, true},
		{" XXXXXX mxtch XXXXXX ", "{fuzzy-concept [{fuzzy-predicate match 1}] 8 13}", 3, true},
		{" mxtch mxtch ", "{fuzzy-concept [{fuzzy-predicate match two 2}] 1 12}", 3, true},
		{" hxt hot ", "{fuzzy-concept [{fuzzy-predicate hit hit 2} {fuzzy-predicate hot hot 1}] 1 8}", 3, true},
		{" XXXXXXXX XXXXXXXX XXXXXXXX ", "{<nil> 0 0}", 3, true},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			m := makeFuzzyDFAMatcher(t, tc.k)
			if k := m.DFA.MaxError(); k != tc.k {
				t.Errorf("expected k=%d; got %d", tc.k, k)
			}
			match := m.Match(string(tc.test))
			// log.Printf("match: %v", match)
			if str := fuzzyConceptToString(t, match, tc.ambiguous); str != tc.want {
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
	t3 := graph.Add("hot hot", "x", "y")
	t4 := graph.Add("hit hit", "x", "y")
	dictionary := map[string]*Concept{
		" match ":       t1.S,
		" match bbb ":   t1.S,
		" mitch match ": t2.S,
		" hot hot ":     t3.S,
		" hit hit ":     t4.S,
	}
	return FuzzyDFAMatcher{
		NewFuzzyDFA(k, NewDFA(dictionary, graph)),
	}
}

func fuzzyConceptToString(t *testing.T, m MatchPos, a bool) string {
	t.Helper()
	if m.Concept == nil {
		return fmt.Sprintf("%v", m)
	}
	if aa := m.Concept.Ambiguous(); aa != a {
		t.Errorf("expected concept.Ambiguous()=%t; got %t", a, aa)
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