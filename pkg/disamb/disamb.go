package disamb

import "bitbucket.org/fflo/semix/pkg/semix"

// Decider is an interface
type Decider interface {
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

func maxConcept(cs []*semix.Concept, scores []float64) *semix.Concept {
	return nil
}

func referencedConcepts(c *semix.Concept) []*semix.Concept {
	return nil
}
