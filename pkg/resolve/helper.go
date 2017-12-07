package resolve

import (
	"math"

	"bitbucket.org/fflo/semix/pkg/semix"
)

// maxConcept chooses the concept with the biggest score.
// if more than one concept with the biggest score can be found,
// nil is returned.  Both slices must have the same length.
func maxConcept(cs []*semix.Concept, scores []float64) *semix.Concept {
	var idx int
	var maxcount int
	max := -math.MaxFloat64
	for i := range cs {
		if scores[i] > max {
			maxcount = 1
			idx = i
			max = scores[i]
		} else if scores[i] == max {
			maxcount++
		}
	}
	if maxcount != 1 {
		return nil
	}
	return cs[idx]
}

// refernecedConcepts returns an array that contains
// all concepts referenced by the given ambigiuous concept.
// The given concept must be ambigiuous.
func referencedConcepts(c *semix.Concept) []*semix.Concept {
	var cs []*semix.Concept
	c.EachEdge(func(e semix.Edge) {
		if e.O.Ambig() {
			e.O.EachEdge(func(e semix.Edge) {
				cs = append(cs, e.O)
			})
		} else {
			cs = append(cs, e.O)
		}
	})
	return cs
}

// resolve calculates a score for each referenced concept and returns the
// concept with the maximal score.  The given concept must be ambigiuous.
func resolve(c *semix.Concept, f func(*semix.Concept) float64) *semix.Concept {
	cs := referencedConcepts(c)
	scores := make([]float64, len(cs))
	for i := range cs {
		scores[i] = f(cs[i])
	}
	return maxConcept(cs, scores)
}
