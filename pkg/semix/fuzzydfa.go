package semix

import (
	"fmt"

	"gitlab.com/finkf/sparsetable"
)

// FuzzyDFA is a simple wrapper around a sparsetable.FuzzyDFA.
// It maps the ids of the underlying sparsetable.DFA to the according Concepts.
type FuzzyDFA struct {
	dfa   *sparsetable.FuzzyDFA
	graph *Graph
}

// NewFuzzyDFA constructs a new FuzzyDFA with the given maximum error bound k.
func NewFuzzyDFA(k int, dfa DFA) FuzzyDFA {
	return FuzzyDFA{
		dfa:   sparsetable.NewFuzzyDFA(k, dfa.dfa),
		graph: dfa.graph,
	}
}

// MaxError returns the maximum allowed error for the this FuzzyDFA.
func (d FuzzyDFA) MaxError() int {
	return d.dfa.MaxError()
}

// Initial returns the initial state of this FuzzyDFA.
func (d FuzzyDFA) Initial(str string) *sparsetable.FuzzyStack {
	return d.dfa.Initial(str)
}

// Delta executes one fuzzy transition in this FuzzyDFA.
func (d FuzzyDFA) Delta(s *sparsetable.FuzzyStack, f func(int, int, *Concept)) bool {
	return d.dfa.Delta(s, func(k, pos int, id int32) {
		c, ok := d.graph.FindByID(id)
		if !ok {
			panic(fmt.Sprintf("invalid id: %d", id))
		}
		f(k, pos, c)
	})
}
