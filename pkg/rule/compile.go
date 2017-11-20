package rule

import (
	"errors"
	"fmt"
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
func Compile(expr string) (r Rule, err error) {
	defer func() {
		if e, ok := recover().(compileError); ok {
			r = nil
			err = errors.New(e.msg)
		}
	}()
	ast, err := newParser(strings.NewReader(expr)).parse()
	if err != nil {
		return nil, err
	}
	typecheck(ast)
	return compileAST(ast), nil
}

func compileAST(ast ast) Rule {
	switch ast.Type() {
	case astInfix:
		return compileInfixAST(ast.(infix))
	case astNum:
		return []optcode{optcode{code: optPushNum, arg: float64(ast.(num))}}
	default:
		panic("invalid ast type")
	}
}

func compileInfixAST(i infix) Rule {
	// if i.left.Type() != i.right.Type() {
	// 	log.Printf("%d <-> %d", i.left.Type(), i.right.Type())
	// 	return nil, fmt.Errorf("type error: %s", i)
	// }
	left := compileAST(i.left)
	right := compileAST(i.right)
	var rule Rule
	rule = append(rule, left...)
	rule = append(rule, right...)
	rule = append(rule, optcode{code: optCode(i.op)})
	return rule
}

func typecheck(ast ast) astType {
	switch t := ast.(type) {
	case prefix:
		at := typecheck(t.expr)
		if t.op == '!' && at != astBoolean {
			abortCompilation("invalid type in prefix expression: %s", ast)
		}
		if t.op == '-' && at != astNum {
			abortCompilation("invalid type in prefix expression: %s", ast)
		}
		return at
	case infix:
		l := typecheck(t.left)
		r := typecheck(t.right)
		if l != r {
			abortCompilation("types in infix expression do not match: %s", ast)
		}
		return l
	default:
		return ast.Type()
	}
}

type compileError struct {
	msg string
}

func abortCompilation(f string, args ...interface{}) {
	panic(compileError{fmt.Sprintf(f, args...)})
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
