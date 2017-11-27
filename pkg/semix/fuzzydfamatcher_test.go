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
		{" mxtch ", "{split-url [{split-url match 1}] 1 6}", 3, true},
		{" mxtch bbbxxx ", "{<nil> 0 0}", 3, true},
		{" mxtch bbxxx ", "{split-url [{split-url match 1}] 1 6}", 3, true},
		{" mxtch bxxx ", "{split-url [{split-url match 1}] 1 6}", 3, true},
		{" mxtch bbb ", "{split-url [{split-url match 1}] 1 10}", 3, true},
		{" XXXXXX mxtch ", "{split-url [{split-url match 1}] 8 13}", 3, true},
		{" mxtch XXXXXX ", "{split-url [{split-url match 1}] 1 6}", 3, true},
		{" XXXXXX mxtch XXXXXX ", "{split-url [{split-url match 1}] 8 13}", 3, true},
		{" mxtch mxtch ", "{split-url [{split-url match two 2}] 1 12}", 3, true},
		{" hxt hot ", "{split-url [{split-url hit hit 2} {split-url hot hot 1}] 1 8}", 3, true},
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
	s1, _, _ := graph.Add("match", "x", "y")
	s2, _, _ := graph.Add("match two", "x", "y")
	s3, _, _ := graph.Add("hot hot", "x", "y")
	s4, _, _ := graph.Add("hit hit", "x", "y")
	dictionary := map[string]int32{
		"match":       s1.ID(),
		"match bbb":   s1.ID(),
		"mitch match": s2.ID(),
		"hot hot":     s3.ID(),
		"hit hit":     s4.ID(),
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
	sort.Slice(m.Concept.edges, func(i, j int) bool {
		return m.Concept.edges[i].O.url < m.Concept.edges[j].O.url
	})
	return fmt.Sprintf("{%s %v %d %d}", m.Concept.ShortURL(), m.Concept.edges, m.Begin, m.End)
}
