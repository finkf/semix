package semix

// DFAMatcher uses a DFA to search for matches in a string.
type DFAMatcher struct {
	DFA DFA
}

// Match returns the MatchPos of the first encountered entry in the DFA.
// The MatchPos denotes the first encountered concept in the string or nil
// nothing could be matched.
func (m DFAMatcher) Match(str string) MatchPos {
	for i := 0; i < len(str); {
		pos, c, _ := m.matchFromHere(str[i:])
		// log.Printf("match from here %q:{%d %s}", str[i:], pos, c)
		if c != nil {
			return MatchPos{Concept: c, Begin: i + 1, End: i + pos}
		}
		i = next(i, pos, str)
	}
	return MatchPos{}
}

func next(i, pos int, str string) int {
	if pos > 0 {
		return i + pos
	}
	for i++; i < len(str); i++ {
		if str[i] == ' ' {
			return i
		}
	}
	return len(str)
}

func (m DFAMatcher) matchFromHere(str string) (int, *Concept, bool) {
	s := m.DFA.Initial()
	var concept *Concept
	var pos int
	for i := 0; i < len(str); i++ {
		s = m.DFA.Delta(s, str[i])
		if !s.Valid() {
			break
		}
		if c, f := m.DFA.Final(s); f {
			concept = c
			pos = i
		}
		if pos == 0 && str[i] == ' ' {
			pos = i
		}
	}
	return pos, concept, false
}
