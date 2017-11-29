package resolve

import (
	"bitbucket.org/fflo/semix/pkg/memory"
	"bitbucket.org/fflo/semix/pkg/rule"
	"bitbucket.org/fflo/semix/pkg/semix"
)

// Rules is a map that maps concept URLs to compiled disambiguation rules.
type Rules map[string]rule.Rule

// NewRules compiles all rules in a RulesDictionary.
func NewRules(rs semix.RulesDictionary, l func(string) int) (Rules, error) {
	rules := make(Rules, len(rs))
	for str, rstr := range rs {
		rule, err := rule.Compile(rstr, l)
		if err != nil {
			return nil, err
		}
		rules[str] = rule
	}
	return rules, nil
}

// Ruled is a resolver that uses the compiled rules to disambiguate concepts.
type Ruled struct {
	Rules Rules
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
