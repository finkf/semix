package rule

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type visitor interface {
	visitSet(set)
	visitStr(str)
	visitNum(num)
	visitInfix(infix)
	visitPrefix(prefix)
}

type astprinter struct {
	buffer *bytes.Buffer
}

func newAstPrinter() visitor {
	return &astprinter{buffer: new(bytes.Buffer)}
}

func (p *astprinter) visitSet(s set) {
	p.buffer.WriteString(s.String())
}

func (p *astprinter) visitStr(s str) {
	p.buffer.WriteString(s.String())
}

func (p *astprinter) visitNum(n num) {
	p.buffer.WriteString(n.String())
}

func (p *astprinter) visitInfix(i infix) {
	p.buffer.WriteString(i.String())
}

func (p *astprinter) visitPrefix(x prefix) {
	p.buffer.WriteString(x.String())
}

func (p astprinter) String() string {
	return p.buffer.String()
}

type ast interface {
	visit(visitor)
}

type prefix struct {
	op   operator
	expr ast
}

func (p prefix) visit(v visitor) {
	v.visitPrefix(p)
}

func (p prefix) String() string {
	printer := newAstPrinter()
	p.expr.visit(printer)
	return fmt.Sprintf("%c%s", p.op, printer)
}

type infix struct {
	op          operator
	left, right ast
}

func (i infix) visit(v visitor) {
	v.visitInfix(i)
}

func (i infix) String() string {
	l := newAstPrinter()
	r := newAstPrinter()
	i.left.visit(l)
	i.right.visit(r)
	return fmt.Sprintf("%s%c%s", l, i.op, r)
}

type set map[str]bool

func (s set) visit(v visitor) {
	v.visitSet(s)
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
