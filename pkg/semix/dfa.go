package semix

import (
	"bytes"
	"log"
	"sort"

	"bitbucket.org/fflo/sparsetable"
)

// DFA is a simple wrapper around a sparsetable.DFA.
// It maps the ids to Concepts.
type DFA struct {
	dfa   *sparsetable.DFA
	graph *Graph
}

// NewDFA constructs a new DFA.
func NewDFA(d map[string]*Concept, graph *Graph) DFA {
	return DFA{graph: graph, dfa: newSparseTableDFA(d)}
}

// Initial returns the initial state of the DFA.
func (d DFA) Initial() uint32 {
	return d.dfa.Initial()
}

// Delta executes one transition in the DFA.
func (d DFA) Delta(s uint32, c byte) uint32 {
	return d.dfa.Delta(s, c)
}

// Final return the found Concept and true iff s denotes a final state.
// Otherwise it returns nil and false.
func (d DFA) Final(s uint32) (*Concept, bool) {
	data, final := d.dfa.Final(s)
	if !final {
		return nil, false
	}
	if c, ok := d.graph.FindByID(data); ok {
		return c, true
	}
	return nil, false
}

func newSparseTableDFA(d map[string]*Concept) *sparsetable.DFA {
	type pair struct {
		id  int32
		str string
	}
	var pairs []pair
	for str, c := range d {
		if c.ID() == 0 {
			log.Fatalf("concept %s: invalid id=%d for: %q", c.URL(), c.ID(), str)
		}
		pairs = append(pairs, pair{id: c.ID(), str: str})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return bytes.Compare([]byte(pairs[i].str), []byte(pairs[j].str)) < 0
	})
	b := sparsetable.NewBuilder()
	for _, p := range pairs {
		if err := b.Add(p.str, p.id); err != nil {
			log.Fatalf("could not add %v: %v", p, err)
		}
	}
	return b.Build()
}
