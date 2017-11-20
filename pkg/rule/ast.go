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
	visitBoolean(boolean)
	visitInfix(infix)
	visitPrefix(prefix)
	visitFunction(function)
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

func (p *astprinter) visitBoolean(b boolean) {
	p.buffer.WriteString(b.String())
}

func (p *astprinter) visitInfix(i infix) {
	p.buffer.WriteString(i.String())
}

func (p *astprinter) visitPrefix(x prefix) {
	p.buffer.WriteString(x.String())
}

func (p *astprinter) visitFunction(f function) {
	p.buffer.WriteString(f.String())
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
	return fmt.Sprintf("(%c%s)", p.op, printer)
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
	return fmt.Sprintf("(%s%c%s)", l, i.op, r)
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

type function struct {
	name string
	args []ast
}

func (f function) visit(v visitor) {
	v.visitFunction(f)
}

func (f function) String() string {
	var strs []string
	for _, arg := range f.args {
		p := newAstPrinter()
		arg.visit(p)
		strs = append(strs, fmt.Sprintf("%s", p))
	}
	return fmt.Sprintf("%s(%s)", f.name, strings.Join(strs, ","))
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
	return fmt.Sprintf("%.2f", n)
}

type boolean bool

func (b boolean) visit(v visitor) {
	v.visitBoolean(b)
}

func (b boolean) String() string {
	return fmt.Sprintf("%t", b)
}
