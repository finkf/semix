package semix

import "testing"

func makeDFA(t *testing.T) DFA {
	t.Helper()
	graph := NewGraph()
	graph.Add("match", "x", "y")
	graph.Add("match two", "x", "y")
	one, _ := graph.FindByURL("match")
	two, _ := graph.FindByURL("match two")
	dictionary := map[string]*Concept{
		" match ":       one,
		" mitch match ": two,
	}
	return NewDFA(dictionary, graph)
}

func TestDFAMatcher(t *testing.T) {
	dfa := makeDFA(t)
	c1, _ := dfa.graph.FindByURL("match")
	c2, _ := dfa.graph.FindByURL("match two")
	tests := []struct {
		test string
		want MatchPos
	}{
		{"", MatchPos{}},
		{" nothing to find", MatchPos{}},
		{" here is the match ", MatchPos{Begin: 13, End: 18, Concept: c1}},
		{" another match is here ", MatchPos{Begin: 9, End: 14, Concept: c1}},
		{" here is mitch match ", MatchPos{Begin: 9, End: 20, Concept: c2}},
		{" mitch match ", MatchPos{Begin: 1, End: 12, Concept: c2}},
		{" mitch mitch match ", MatchPos{Begin: 7, End: 18, Concept: c2}},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			m := DFAMatcher{DFA: dfa}
			if pos := m.Match(string(tc.test)); pos != tc.want {
				t.Errorf("expeceted pos = %v; got %v", tc.want, pos)
			}
		})
	}
}
