package semix

import (
	"fmt"
	"testing"
)

func TestFuzzyDFAMatcher(t *testing.T) {
	tests := []struct {
		test, want string
		k          int
	}{
		{"", "{<nil> 0 0}", 3},
		// {" ", "{<nil> 0 0}", 3},
		// {" ", "{<nil> 0 0}", 3},
		// {" ", MatchPos{}, 3},
		// {"  ", MatchPos{}, 3},
		// {" nothing to find ", MatchPos{}, 3},
		{" match ", "{fuzzy-concept [{fuzzy-predicate match 0}] 1 6}", 0},
		/*
			{" here is the match ", MatchPos{Begin: 13, End: 18, Concept: c1}, 1},
			{" another match is here ", MatchPos{Begin: 9, End: 14, Concept: c1}, 1},
			{" here is mitch match ", MatchPos{Begin: 9, End: 20, Concept: c2}, 1},
			{" mitch match ", MatchPos{Begin: 1, End: 12, Concept: c2}, 1},
			{" mitch mitch match ", MatchPos{Begin: 7, End: 18, Concept: c2}, 1},
		*/
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			m := makeFuzzyDFAMatcher(t, tc.k)
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
	dictionary := map[string]*Concept{
		" match ":       t1.S,
		" mitch match ": t2.S,
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
	str := "["
	for i, e := range es {
		if i > 0 {
			str += " "
		}
		str += fmt.Sprintf("{%s %s %d}", e.P, e.O, e.L)
	}
	return str + "]"
}
