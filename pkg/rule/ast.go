package rule

import (
	"fmt"
	"sort"
	"strings"
)

type visitor interface {
	visitSet(set)
	visitStr(str)
	visitNum(num)
}

type ast interface {
	visit(visitor)
}

type set map[string]bool

func (s set) visit(v visitor) {
	v.visitSet(s)
}

func (s set) String() string {
	var strs []string
	for str := range s {
		strs = append(strs, fmt.Sprintf("%q", str))
	}
	sort.Strings(strs)
	return fmt.Sprintf("{%s}", strings.Join(strs, ","))
}

type str string

func (s str) visit(v visitor) {
	v.visitStr(s)
}

func (s str) String() string {
	return fmt.Sprintf("%q", string(s))
}

type num float64

func (n num) visit(v visitor) {
	v.visitNum(n)
}

func (n num) String() string {
	return fmt.Sprintf("%f", n)
}
