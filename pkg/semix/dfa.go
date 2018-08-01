package semix

import (
	"log"
	"sort"

	"gitlab.com/finkf/sparsetable"
)

// DFA is a simple wrapper around a sparsetable.DFA.
// It maps the ids to Concepts.
type DFA struct {
	dfa   *sparsetable.DFA
	graph *Graph
}

// NewDFA constructs a new DFA.
func NewDFA(d Dictionary, graph *Graph) DFA {
	return DFA{graph: graph, dfa: newSparseTableDFA(d)}
}

// Initial returns the initial state of the DFA.
func (d DFA) Initial() sparsetable.State {
	return d.dfa.Initial()
}

// Delta executes one transition in the DFA.
func (d DFA) Delta(s sparsetable.State, c byte) sparsetable.State {
	return d.dfa.Delta(s, c)
}

// Final return the found Concept and true iff s denotes a final state.
// Otherwise it returns nil and false.
func (d DFA) Final(s sparsetable.State) (*Concept, bool) {
	data, final := d.dfa.Final(s)
	if !final {
		return nil, false
	}
	if c, ok := d.graph.FindByID(data); ok {
		return c, true
	}
	return nil, false
}

func newSparseTableDFA(d Dictionary) *sparsetable.DFA {
	type pair struct {
		id  int32
		str string
	}
	var pairs []pair
	for str, id := range d {
		if id == 0 {
			// TODO: errors!
			log.Fatalf("concept %s: invalid id=%d", str, id)
		}
		pairs = append(pairs, pair{id: id, str: " " + str + " "})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].str < pairs[j].str
	})
	b := sparsetable.NewBuilder()
	for _, p := range pairs {
		if err := b.Add(p.str, p.id); err != nil {
			log.Fatalf("cannot add %v: %v", p, err)
		}
	}
	return b.Build()
}
