package rule

import (
	"fmt"
	"sort"
	"strings"
)

type astType int

const (
	astSet = iota
	astStr
	astNum
	astFunction
	astBoolean
	astPrefix
	astInfix
	astNumArray
)

type ast interface {
	fmt.Stringer
	typ() astType
	check() astType
}

type prefix struct {
	op   operator
	expr ast
}

func (prefix) typ() astType {
	return astPrefix
}

func (p prefix) check() astType {
	if p.op != '-' {
		astFatalf("invalid prefix operator in: %s", p)
	}
	t := p.expr.check()
	checkTypIn(p, t, astBoolean, astNum)
	return t
}

func (p prefix) String() string {
	return fmt.Sprintf("(%c%s)", p.op, p.expr)
}

type infix struct {
	op          operator
	left, right ast
}

func (infix) typ() astType {
	return astInfix
}

func (i infix) check() astType {
	left := i.left.check()
	right := i.right.check()
	if left != right {
		astFatalf("invalid expression: %s", i)
	}
	switch i.op {
	case '=':
		return astBoolean
	case '>':
		checkTypIn(i, left, astNum, astStr)
		return astBoolean
	case '<':
		checkTypIn(i, left, astNum, astStr)
		return astBoolean
	case '+':
		checkTypIn(i, left, astBoolean, astNum, astSet, astStr)
		return left
	case '-':
		checkTypIn(i, left, astNum, astSet)
		return left
	case '/':
		checkTypIn(i, left, astNum)
		return left
	case '*':
		checkTypIn(i, left, astBoolean, astNum, astSet)
		return left
	default:
		astFatalf("invalid expression: %s", i)
	}
	panic("unreacheable")
}

func (i infix) String() string {
	return fmt.Sprintf("(%s%c%s)", i.left, i.op, i.right)
}

type set map[str]bool

func (set) typ() astType {
	return astSet
}

func (set) check() astType {
	return astSet
}

func (s set) String() string {
	var strs []string
	for str := range s {
		strs = append(strs, str.String())
	}
	sort.Strings(strs)
	return fmt.Sprintf("{%s}", strings.Join(strs, ","))
}

type str string

func (str) typ() astType {
	return astStr
}

func (str) check() astType {
	return astStr
}

func (s str) String() string {
	return fmt.Sprintf("%q", string(s))
}

type num float64

func (num) typ() astType {
	return astNum
}

func (num) check() astType {
	return astNum
}

func (n num) String() string {
	return fmt.Sprintf("%.2f", n)
}

type boolean bool

func (boolean) typ() astType {
	return astBoolean
}

func (boolean) check() astType {
	return astBoolean
}

func (b boolean) String() string {
	return fmt.Sprintf("%t", b)
}

type astError struct {
	msg string
}

func checkTypIn(ast ast, t astType, set ...astType) {
	for _, s := range set {
		if t == s {
			return
		}
	}
	astFatalf("invalid expression: %s", ast)
	panic("unreacheable")
}

func astFatalf(f string, args ...interface{}) {
	panic(astError{msg: fmt.Sprintf(f, args)})
}
