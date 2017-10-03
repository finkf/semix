package semix

import (
	"log"
	"math"
)

// FuzzyDFAMatcher uses a FuzzyDFA to search for matches in a string.
type FuzzyDFAMatcher struct {
	DFA FuzzyDFA
}

// Match returns the MatchPos of the first encountered entry in the DFA.
// The MatchPos denotes the first encountered concept in the string or nil
// nothing could be matched.
func (m FuzzyDFAMatcher) Match(str string) MatchPos {
	for i := 0; i < len(str); {
		s := m.DFA.Initial(str[i:])
		var savepos int
		set := &matchset{m: make(map[*Concept]fuzzypos)}
		for !s.Empty() {
			s = m.DFA.Delta(s, func(k, pos int, c *Concept) {
				if pos <= 0 {
					return
				}
				if c == nil {
					panic("nil concept")
				}
				log.Printf("%q final state k=%d pos=(%d %d) url=%s", str[i:], k, i, i+pos, c.URL())
				pos--
				isws := str[i+pos] == ' '
				if savepos == 0 && isws {
					savepos = pos
				}
				set.insert(fuzzypos{c: c, l: k, s: i + 1, e: i + pos, isws: isws})
			})
			if len(set.m) > 0 {
				return set.makeMatch()
			}
		}
		i = next(i, savepos, str)
	}
	return MatchPos{}
}

type fuzzypos struct {
	l, s, e int
	isws    bool
	c       *Concept
}

type matchset struct {
	longest int
	m       map[*Concept]fuzzypos
}

func (m *matchset) makeMatch() MatchPos {
	var ps []fuzzypos
	left := math.MaxInt64
	for _, p := range m.m {
		if m.longest == p.e {
			ps = append(ps, p)
			if p.s < left {
				left = p.s
			}
		}
	}
	if len(ps) == 0 {
		return MatchPos{}
	}
	return MatchPos{Begin: left, End: m.longest, Concept: makeFuzzyConcept(ps)}
}

var fuzzyPredicate = NewConcept("http://bitbucket.org/fflo/semix/pkg/semix/fuzzy-predicate")

func makeFuzzyConcept(ps []fuzzypos) *Concept {
	switch len(ps) {
	case 0:
		return nil
	case 1:
		return ps[0].c
	default:
		c := NewConcept("http://bitbucket.org/fflo/semix/pkg/fuzzy-concept")
		for _, p := range ps {
			c.edges = append(c.edges, Edge{P: fuzzyPredicate, O: p.c, L: p.l})
		}
		return c
	}
}

func (m *matchset) insert(p fuzzypos) {
	if _, ok := m.m[p.c]; !ok {
		m.m[p.c] = p
	} else {
		m.m[p.c] = selectFuzzypos(m.m[p.c], p)
	}
	if m.longest < m.m[p.c].e {
		m.longest = m.m[p.c].e
	}
}

func selectFuzzypos(o, n fuzzypos) fuzzypos {
	if n.l < o.l {
		return n
	}
	if o.l < n.l {
		return o
	}
	if o.e == n.e { // does not matter
		return o
	}
	// We have two different end positions with the same error.
	// Choose the position that ends on white space.
	if o.isws && !n.isws {
		return o
	}
	if !o.isws && n.isws {
		return n
	}
	// We have two different end positions with the same error
	// that both end on whitespace.
	// Choose the right most position.
	if o.e < n.e {
		return n
	}
	return o
}
