package resolve

import (
	"github.com/finkf/semix/pkg/memory"
	"github.com/finkf/semix/pkg/rule"
	"github.com/finkf/semix/pkg/semix"
)

// Ruled is a resolver that uses the compiled rules to disambiguate concepts.
type Ruled struct {
	Rules rule.Map
}

// Resolve is used to resolve ambiguities.
func (r Ruled) Resolve(c *semix.Concept, mem *memory.Memory) *semix.Concept {
	return resolve(c, func(c *semix.Concept) float64 {
		if _, ok := r.Rules[c.URL()]; !ok {
			return 0
		}
		if r.Rules[c.URL()].Execute(mem) < 1 {
			return 0
		}
		return 1
	})
}
