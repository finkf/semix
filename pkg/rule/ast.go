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
	compile(func(string) int) Rule
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

func (p prefix) compile(f func(string) int) Rule {
	rule := p.expr.compile(f)
	if p.expr.check() == astBoolean {
		return append(rule, instruction{opcode: opNOT})
	}
	return append(rule, instruction{opcode: opNEG})
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

func (s set) compile(func(string) int) Rule {
	astFatalf("cannot compile %s", s)
	panic("unreacheable")
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

func (s str) compile(func(string) int) Rule {
	astFatalf("cannot compile %s", s)
	panic("unreacheable")
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

func (n num) compile(func(string) int) Rule {
	return Rule{instruction{opcode: opPushNUM, arg: float64(n)}}
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

func (b boolean) compile(func(string) int) Rule {
	if b {
		return Rule{instruction{opcode: opPushTRUE}}
	}
	return Rule{instruction{opcode: opPushFALSE}}
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
