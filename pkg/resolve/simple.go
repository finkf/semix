package resolve

import (
	"gitlab.com/finkf/semix/pkg/memory"
	"gitlab.com/finkf/semix/pkg/semix"
)

// Simple chooses the most occuring concept in the memory
// that occurres at least once in the memory.
type Simple struct{}

// Resolve chooses the most occuring concept in the memory
// that occurres at least once in the memory.
func (Simple) Resolve(c *semix.Concept, mem *memory.Memory) *semix.Concept {
	return resolve(c, func(c *semix.Concept) float64 {
		return float64(mem.CountIfS(func(cc *semix.Concept) bool {
			return cc.URL() == c.URL()
		}))
	})
}
