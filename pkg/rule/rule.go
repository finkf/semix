package rule

import (
	"errors"
	"strings"

	"github.com/finkf/semix/pkg/memory"
)

// Map maps concept URLs to compiled rules.
type Map map[string]Rule

// NewMap compiles a new Map from a map of rules.
func NewMap(rs map[string]string, lookup func(string) int) (Map, error) {
	m := make(Map, len(rs))
	for url, str := range rs {
		r, err := Compile(str, lookup)
		if err != nil {
			return nil, err
		}
		m[url] = r
	}
	return m, nil
}

// Rule represents a compiled rule.
type Rule []instruction

// Execute executes a rule and returns its result.
func (r Rule) Execute(memory *memory.Memory) float64 {
	stack := new(stack)
	for _, o := range r {
		o.call(memory, stack)
	}
	return stack.pop1()
}

func (r Rule) String() string {
	var strs []string
	for _, rule := range r {
		strs = append(strs, rule.String())
	}
	return strings.Join(strs, ";") + ";"
}

// Compile compiles a rule from an expression.
// The lookup function is used to map strings to concept ids.
// If lookup returns a number <= 0, the concept cannot be found and an
// error will be returned from Compile.
func Compile(expr string, lookup func(string) int) (r Rule, err error) {
	defer func() {
		if e := recover(); e != nil {
			switch t := e.(type) {
			case astError:
				err = errors.New(t.msg)
			case error:
				err = t
			case string:
				err = errors.New(t)
			}
		}
	}()
	ast, err := newParser(strings.NewReader(expr)).parse()
	if err != nil {
		return nil, err
	}
	ast.check()
	r = ast.compile(lookup)
	return r, nil
}
