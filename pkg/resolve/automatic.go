package resolve

import (
	"bitbucket.org/fflo/semix/pkg/memory"
	"bitbucket.org/fflo/semix/pkg/semix"
)

// Automatic disambiguates concepts by calculating the thematic overlap
// of the concepts with the memory.
type Automatic struct {
	Threshold float64
}

// Resolve resolves ambiguities using the automatic method.
func (a Automatic) Resolve(c *semix.Concept, mem *memory.Memory) *semix.Concept {
	elems := mem.ElementsS()
	return resolve(c, func(c *semix.Concept) float64 {
		o := overlap(c, elems)
		if o > a.Threshold {
			return o
		}
		return 0
	})
}

func overlap(c *semix.Concept, elems map[string]*semix.Concept) float64 {
	celems := make(map[string]bool)
	c.EachEdge(func(e semix.Edge) {
		celems[e.O.URL()] = true
	})
	var n float64
	for url := range celems {
		if elems[url] != nil {
			n++
		}
	}
	return n / float64(len(elems))
}
