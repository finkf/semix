package rule

import (
	"errors"
	"strings"
)

// Rule represents a compiled rule.
type Rule []instruction

// Execute executes a rule and returns its result.
func (r Rule) Execute() float64 {
	stack := new(stack)
	for _, o := range r {
		o.call(stack)
	}
	return stack.pop1()
}

func (r Rule) String() string {
	var strs []string
	for _, rule := range r {
		strs = append(strs, rule.String())
	}
	return strings.Join(strs, ";")
}

// Compile compiles a rule from an expression.
func Compile(expr string) (r Rule, err error) {
	defer func() {
		if e, ok := recover().(astError); ok {
			r = nil
			err = errors.New(e.msg)
		}
	}()
	ast, err := newParser(strings.NewReader(expr)).parse()
	if err != nil {
		return nil, err
	}
	ast.check()
	r = ast.compile(func(str string) int {
		return -1
	})
	return r, nil
}
