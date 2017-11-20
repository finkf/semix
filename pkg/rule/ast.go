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
)

type ast interface {
	fmt.Stringer
	Type() astType
}

type prefix struct {
	op   operator
	expr ast
}

func (prefix) Type() astType {
	return astPrefix
}

func (p prefix) String() string {
	return fmt.Sprintf("(%c%s)", p.op, p.expr)
}

type infix struct {
	op          operator
	left, right ast
}

func (infix) Type() astType {
	return astInfix
}

func (i infix) String() string {
	return fmt.Sprintf("(%s%c%s)", i.left, i.op, i.right)
}

type set map[str]bool

func (set) Type() astType {
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

type function struct {
	name string
	args []ast
}

func (function) Type() astType {
	return astFunction
}

func (f function) String() string {
	var strs []string
	for _, arg := range f.args {
		strs = append(strs, arg.String())
	}
	return fmt.Sprintf("%s(%s)", f.name, strings.Join(strs, ","))
}

type str string

func (str) Type() astType {
	return astStr
}

func (s str) String() string {
	return fmt.Sprintf("%q", string(s))
}

type num float64

func (num) Type() astType {
	return astNum
}

func (n num) String() string {
	return fmt.Sprintf("%.2f", n)
}

type boolean bool

func (boolean) Type() astType {
	return astBoolean
}

func (b boolean) String() string {
	return fmt.Sprintf("%t", b)
}
