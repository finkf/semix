package rule

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/scanner"
)

const (
	ruleTokens = scanner.ScanIdents |
		scanner.ScanFloats |
		scanner.SkipComments |
		scanner.ScanStrings |
		scanner.ScanRawStrings
)

type infixParseFunc func(ast) ast

type prefixParseFunc func() ast

type parser struct {
	scanner          *scanner.Scanner
	p                rune
	prefixParseFuncs map[rune]prefixParseFunc
	infixParseFuncs  map[rune]infixParseFunc
}

func newParser(r io.Reader) *parser {
	s := &scanner.Scanner{
		Error: die,
		Mode:  ruleTokens,
	}
	s.Init(r)
	s.Filename = "rule"
	p := &parser{
		scanner: s,
	}
	p.prefixParseFuncs = map[rune]prefixParseFunc{
		'{':            p.parseSet,
		'-':            p.parsePrefix,
		'(':            p.parseGroup,
		scanner.Ident:  p.parseBool,
		scanner.Int:    p.parseNum,
		scanner.Float:  p.parseNum,
		scanner.String: p.parseStr,
	}
	p.infixParseFuncs = map[rune]infixParseFunc{
		'-': p.parseInfix,
		'+': p.parseInfix,
		'*': p.parseInfix,
		'/': p.parseInfix,
		'=': p.parseInfix,
		'>': p.parseInfix,
		'<': p.parseInfix,
	}
	return p
}

func (p *parser) parse() (a ast, err error) {
	defer func() {
		if r, ok := recover().(parseError); ok {
			a = nil
			err = errors.New(r.msg)
		}
	}()
	a = p.parseExpression(lowest)
	p.eat(scanner.EOF)
	return a, nil
}

func (p *parser) parseExpression(prec int) ast {
	f, ok := p.prefixParseFuncs[p.peek()]
	if !ok {
		parserFatalf(p.scanner, "invalid expression: %s %s", scanner.TokenString(p.peek()), p.scanner.TokenText())
	}
	left := f()
	for p.peek() != scanner.EOF && prec < precedence(p.peek()) {
		f, ok := p.infixParseFuncs[p.peek()]
		if !ok {
			return left
		}
		left = f(left)
	}
	return left
}

func (p *parser) parseGroup() ast {
	p.eat('(')
	ast := p.parseExpression(lowest)
	p.eat(')')
	return ast
}

func (p *parser) parseSet() ast {
	set := make(set)
	for _, str := range p.parseStrList() {
		set[str] = true
	}
	return set
}

func (p *parser) parseStrList() []str {
	p.eat('{')
	var strs []str
	for p.peek() != '}' {
		strs = append(strs, p.parseStr().(str))
		tok, _ := p.eat(',', '}')
		if tok == '}' {
			return strs
		}
	}
	p.eat('}')
	return strs
}

func (p *parser) parsePrefix() ast {
	op := p.peek()
	p.eat('!', '-')
	return prefix{op: operator(op), expr: p.parseExpression(neg)}
}

func (p *parser) parseInfix(left ast) ast {
	op := p.peek()
	p.eat('-', '+', '*', '/', '=', '<', '>')
	prec := precedence(op)
	return infix{left: left, op: operator(op), right: p.parseExpression(prec)}
}

func (p *parser) parseStr() ast {
	_, s := p.eat(scanner.String)
	s, err := strconv.Unquote(s)
	if err != nil {
		parserFatalf(p.scanner, "invalid string: %s", err)
	}
	return str(s)
}

func (p *parser) parseBool() ast {
	switch _, str := p.eat(scanner.Ident); str {
	case "true":
		return boolean(true)
	case "false":
		return boolean(false)
	default:
		if p.peek() != '(' {
			parserFatalf(p.scanner, "invalid identifier: %s", str)
		}
		return function{name: str, args: p.parseArgs()}
	}
}

func (p *parser) parseArgs() []ast {
	p.eat('(')
	var args []ast
	for p.peek() != ')' {
		args = append(args, p.parseExpression(lowest))
		tok, _ := p.eat(',', ')')
		if tok == ')' {
			return args
		}
	}
	p.eat(')')
	return args
}

func (p *parser) parseNum() ast {
	_, str := p.eat(scanner.Float, scanner.Int)
	n, err := strconv.ParseFloat(str, 64)
	if err != nil {
		parserFatalf(p.scanner, "could not parse number %q: %s", str, err)
	}
	return num(n)
}

func (p *parser) eat(toks ...rune) (rune, string) {
	for _, tok := range toks {
		if p.p == tok {
			str := p.scanner.TokenText()
			p.p = p.scanner.Scan()
			return tok, str
		}
	}
	// error
	var xs []string
	for _, tok := range toks {
		xs = append(xs, scanner.TokenString(tok))
	}
	parserFatalf(p.scanner, "expected %s; got %s",
		strings.Join(xs, " or "), scanner.TokenString(p.p))
	panic("ureacheable")
}

func (p *parser) peek() rune {
	if p.p == 0 {
		p.p = p.scanner.Scan()
	}
	return p.p
}

type parseError struct {
	msg string
}

func die(s *scanner.Scanner, msg string) {
	panic(parseError{fmt.Sprintf("%s: %s", s.Position, msg)})
}

func parserFatalf(s *scanner.Scanner, f string, args ...interface{}) {
	msg := fmt.Sprintf(f, args...)
	die(s, msg)
}
