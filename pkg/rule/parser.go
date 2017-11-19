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
	p.registerPrefixParseFunc('{', p.parseSet)
	p.registerPrefixParseFunc('!', p.parsePrefix)
	p.registerPrefixParseFunc('-', p.parsePrefix)
	p.registerPrefixParseFunc(scanner.Int, p.parseNum)
	p.registerPrefixParseFunc(scanner.Float, p.parseNum)
	p.registerInfixParseFunc('-', p.parseInfix)
	p.registerInfixParseFunc('+', p.parseInfix)
	p.registerInfixParseFunc('*', p.parseInfix)
	p.registerInfixParseFunc('/', p.parseInfix)
	p.registerInfixParseFunc('=', p.parseInfix)
	p.registerInfixParseFunc('>', p.parseInfix)
	p.registerInfixParseFunc('<', p.parseInfix)
	return p
}

type infixParseFunc func(ast) ast

func (p *parser) registerInfixParseFunc(tok rune, f infixParseFunc) {
	if p.infixParseFuncs == nil {
		p.infixParseFuncs = make(map[rune]infixParseFunc)
	}
	p.infixParseFuncs[tok] = f
}

type prefixParseFunc func() ast

func (p *parser) registerPrefixParseFunc(tok rune, f prefixParseFunc) {
	if p.prefixParseFuncs == nil {
		p.prefixParseFuncs = make(map[rune]prefixParseFunc)
	}
	p.prefixParseFuncs[tok] = f
}

func (p *parser) parse() (a ast, err error) {
	defer func() {
		if r, ok := recover().(parseError); ok {
			a = nil
			err = errors.New(r.msg)
		}
	}()
	a = p.parseExpression()
	p.eat(scanner.EOF)
	return a, nil
}

func (p *parser) parseExpression() ast {
	peek := p.peek()
	if f, ok := p.prefixParseFuncs[peek]; ok {
		return f()
	}
	if f, ok := p.infixParseFuncs[peek]; ok {
		return f()
	}
	dief(p.scanner, "invalid: %s", scanner.TokenString(peek))
	panic("ureacheable")
}

func (p *parser) parseSet() ast {
	p.eat('{')
	set := make(set)
	for _, str := range p.parseStrList('}') {
		set[str] = true
	}
	p.eat('}')
	return set
}

func (p *parser) parseStrList(end rune) []str {
	var strs []str
loop:
	for p.peek() != end {
		strs = append(strs, p.parseStr())
		switch next := p.peek(); next {
		case ',':
			p.eat(',')
		case end:
			break loop
		default:
			dief(p.scanner, `expected "," or "%c"; got %s`,
				end, scanner.TokenString(next))
		}
	}
	return strs
}

func (p *parser) parsePrefix() ast {
	op := p.peek()
	p.eat('!', '-')
	return prefix{op: operator(op), expr: p.parseExpression()}
}

func (p *parser) parseInfix(left ast) ast {
	op := p.peek()
	p.eat('-', '+', '*', '/', '=', '<', '>')
	return infix{left: left, op: operator(op), right: p.parseExpression()}
}

func (p *parser) parseStr() str {
	_, s := p.eat(scanner.String)
	s, err := strconv.Unquote(s)
	if err != nil {
		dief(p.scanner, "invalid string: %s", err)
	}
	return str(s)
}

func (p *parser) parseNum() ast {
	_, str := p.eat(scanner.Float, scanner.Int)
	n, err := strconv.ParseFloat(str, 64)
	if err != nil {
		dief(p.scanner, "could not parse number %q: %s", str, err)
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
	dief(p.scanner, "expected %s; got %s",
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

func dief(s *scanner.Scanner, f string, args ...interface{}) {
	msg := fmt.Sprintf(f, args...)
	die(s, msg)
}
