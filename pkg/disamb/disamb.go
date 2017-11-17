package disamb

import "bitbucket.org/fflo/semix/pkg/semix"

// Decider defines the interface for the disambiguation.
type Decider interface {
	// Decide returns the disambiguated concept or nil if the
	// concept could not be disambiguated. It is an error
	// to call Decide with a non-ambigiuous concept.
	Decide(*semix.Concept) *semix.Concept
}

// Disambiguator disambiguates ambigous Concepts and handles a local memory.
type Disambiguator struct {
	Decider Decider
	Memory  *Memory
}

// Disambiguate tries to disambiguate an ambigous concept.
// if the given concept is not ambigous, the same concept is returned.
// Otherwise the disambiguated concept or nil is returned.
func (d Disambiguator) Disambiguate(c *semix.Concept) *semix.Concept {
	if c == nil {
		return nil
	}
	if !c.Ambiguous() {
		d.Memory.Push(c)
		return c
	}
	decided := d.Decider.Decide(c)
	if decided == nil {
		return nil
	}
	d.Memory.Push(decided)
	return decided
}
