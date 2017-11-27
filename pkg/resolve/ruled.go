package resolve

import (
	"bitbucket.org/fflo/semix/pkg/memory"
	"bitbucket.org/fflo/semix/pkg/rule"
	"bitbucket.org/fflo/semix/pkg/semix"
)

type ruled struct {
	rules map[string]rule.Rule
}

// NewRuled creates an new rule-based resolver.
func NewRuled(rs semix.RulesDictionary, l func(string) int) (Interface, error) {
	ruled := &ruled{make(map[string]rule.Rule, len(rs))}
	for url, r := range rs {
		compiled, err := rule.Compile(r, l)
		if err != nil {
			return nil, err
		}
		ruled.rules[url] = compiled
	}
	return ruled, nil
}

func (r *ruled) Resolve(c *semix.Concept, mem *memory.Memory) *semix.Concept {
	return resolve(c, func(c *semix.Concept) float64 {
		if _, ok := r.rules[c.URL()]; !ok {
			return 0
		}
		if r.rules[c.URL()].Execute(mem) < 1 {
			return 0
		}
		return 1
	})
}
