package query

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

type parserError struct {
	msg string
}

// Parser represents a query parser.
type Parser struct {
	scanner *scanner.Scanner
	query   string
	p       rune
}

// NewParser create a new parser.
func NewParser(query string) *Parser {
	var s scanner.Scanner
	s.Init(strings.NewReader(query))
	s.Filename = "query"
	s.Error = parserFatal
	return &Parser{
		query:   query,
		scanner: &s,
	}
}

// Parse parses a query.
func (p *Parser) Parse() (q *Query, err error) {
	defer func() {
		if r, ok := recover().(parserError); ok {
			q = nil
			err = errors.New(r.msg)
		}
	}()
	return p.parseQueryExp(), nil
}

func (p *Parser) parseQueryExp() *Query {
	p.eat('?')
	var k int
	var a bool
loop:
	for {
		switch tok := p.peek(); tok {
		case '*':
			p.eat('*')
			a = true
		case scanner.Int:
			_, str := p.eat(scanner.Int)
			// ignore errors from strconv!
			k, _ = strconv.Atoi(str)
		default:
			break loop
		}
	}
	p.eat('(')
	c, s := p.parseConstraint()
	p.eat(')')
	return &Query{set: s, constraint: c, l: k, a: a}
}

func (p *Parser) parseConstraint() (constraint, set) {
	var c constraint
	l := p.peek()
	// {<-...}
	if l == '{' {
		return c, p.parseSet()
	}
	// !<-...
	if l == '!' {
		p.eat('!')
		c.not = true
		l = p.peek()
	}
	// *<-(...)
	if l == '*' {
		p.eat('*')
		c.all = true
		p.eat('(')
		s := p.parseSet()
		p.eat(')')
		return c, s
	}
	if l != scanner.Ident && l != scanner.String {
		p.fatalf("exepected %s (%d) or %s (%d); got %s (%d)",
			scanner.TokenString(scanner.Ident), int(scanner.Ident),
			scanner.TokenString(scanner.String), int(scanner.String),
			scanner.TokenString(l), int(l))
	}
	// A<-,...
	c.set = p.parseList()
	p.eat('(')
	s := p.parseSet()
	p.eat(')')
	return c, s
}

func (p *Parser) parseSet() set {
	p.eat('{')
	set := p.parseList()
	p.eat('}')
	return set
}

func (p *Parser) parseList() set {
	set := make(map[string]bool)
	// check for empty
	l := p.peek()
	if l != scanner.Ident && l != scanner.String {
		return set
	}
	str := p.parseString()
	set[str] = true
	for l := p.peek(); l == ','; l = p.peek() {
		p.eat(',')
		str := p.parseString()
		set[str] = true
	}
	return set
}

func (p *Parser) parseString() string {
	tok, str := p.eat(scanner.Ident, scanner.String)
	if tok == scanner.String {
		s, err := strconv.Unquote(str)
		if err != nil {
			p.fatalf("could not parse string: %s", err)
		}
		str = s
	}
	return str
}

func (p *Parser) peek() rune {
	if p.p == 0 {
		p.p = p.scanner.Scan()
	}
	return p.p
}

func (p *Parser) eat(toks ...rune) (rune, string) {
	peek := p.peek()
	for _, tok := range toks {
		if tok == peek {
			str := p.scanner.TokenText()
			p.p = p.scanner.Scan()
			return tok, str
		}
	}
	var strs []string
	for _, tok := range toks {
		strs = append(strs, fmt.Sprintf("%s", scanner.TokenString(tok)))
	}
	p.fatalf("expected %s; got %s",
		strings.Join(strs, " or "),
		scanner.TokenString(peek))
	panic("unreacheable")
}

func parserFatal(s *scanner.Scanner, msg string) {
	panic(parserError{fmt.Sprintf("%s: %s", s.Position, msg)})
}

func (p *Parser) fatalf(f string, args ...interface{}) {
	parserFatal(p.scanner, fmt.Sprintf(f, args...))
}
