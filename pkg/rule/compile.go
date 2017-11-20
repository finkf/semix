package rule

import (
	"strings"
)

// Rule represents a compiled rule.
type Rule []optcode

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
func Compile(expr string) (Rule, error) {
	ast, err := newParser(strings.NewReader(expr)).parse()
	if err != nil {
		return nil, err
	}
	return compileAST(ast)
}

func compileAST(ast ast) (Rule, error) {
	switch ast.Type() {
	case astInfix:
		return compileInfixAST(ast.(infix))
	case astNum:
		return []optcode{optcode{code: optPushNum, arg: float64(ast.(num))}}, nil
	default:
		panic("invalid ast type")
	}
}

func compileInfixAST(i infix) (Rule, error) {
	// if i.left.Type() != i.right.Type() {
	// 	log.Printf("%d <-> %d", i.left.Type(), i.right.Type())
	// 	return nil, fmt.Errorf("type error: %s", i)
	// }
	left, err := compileAST(i.left)
	if err != nil {
		return nil, err
	}
	right, err := compileAST(i.right)
	if err != nil {
		return nil, err
	}
	var rule Rule
	rule = append(rule, left...)
	rule = append(rule, right...)
	rule = append(rule, optcode{code: optCode(i.op)})
	return rule, nil
}

func optCode(op operator) int {
	switch op {
	case '+':
		return optAdd
	case '-':
		return optSub
	case '*':
		return optMul
	case '/':
		return optDiv
	case '>':
		return optGT
	case '<':
		return optLT
	case '=':
		return optEQ
	case '!':
		return optNot

	default:
		panic("invalid operator")
	}
}
