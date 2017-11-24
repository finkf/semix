package disamb

import (
	"math"

	"bitbucket.org/fflo/semix/pkg/memory"
	"bitbucket.org/fflo/semix/pkg/semix"
)

// SimpleDecider chooses the most occuring concept in the memory.
type SimpleDecider struct{}

// Decide chooses the most occuring concept in the memory.
func (SimpleDecider) Decide(mem *memory.Memory, c *semix.Concept) *semix.Concept {
	return decide(c, func(c *semix.Concept) float64 {
		return float64(mem.CountIfS(func(cc *semix.Concept) bool { return cc.URL() == c.URL() }))
	})
}

func maxConcept(cs []*semix.Concept, scores []float64) *semix.Concept {
	var idx int
	var maxcount int
	max := -math.MaxFloat64
	for i := range cs {
		if scores[i] > max {
			maxcount = 1
			idx = i
			max = scores[i]
		}
		if scores[i] == max {
			maxcount++
		}
	}
	if maxcount != 1 {
		return nil
	}
	return cs[idx]
}

func referencedConcepts(c *semix.Concept) []*semix.Concept {
	var cs []*semix.Concept
	c.EachEdge(func(e semix.Edge) {
		if e.O.Ambiguous() {
			e.O.EachEdge(func(e semix.Edge) {
				cs = append(cs, e.O)
			})
		} else {
			cs = append(cs, e.O)
		}
	})
	return cs
}

func decide(c *semix.Concept, f func(*semix.Concept) float64) *semix.Concept {
	cs := referencedConcepts(c)
	scores := make([]float64, len(cs))
	for i := range cs {
		scores[i] = f(cs[i])
	}
	return maxConcept(cs, scores)
}
