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
	return p.parseQuery(), nil
}

func (p *Parser) parseQuery() *Query {
	p.eat('?')
	k, a := p.parseQueryOpt()
	p.eat('(')
	c, s := p.parseQueryExp()
	p.eat(')')
	return &Query{set: s, constraint: c, l: k, a: a}
}

func (p *Parser) parseQueryOpt() (int, bool) {
	var k int64
	var a bool
loop:
	for {
		switch l := p.peek(); l {
		case '*':
			p.eat('*')
			a = true
		case scanner.Int:
			_, str := p.eat(scanner.Int)
			k, _ = strconv.ParseInt(str, 10, 32)
		default:
			break loop
		}
	}
	return int(k), a
}

func (p *Parser) parseQueryExp() (constraint, set) {
	var c constraint
	switch l := p.peek(); l {
	case '!':
		p.eat('!')
		c.not = true
		c.set = p.parseList()
		return c, p.parseSet()
	case '*':
		p.eat('*')
		c.all = true
		return c, p.parseSet()
	case scanner.String, scanner.Ident:
		set := p.parseList()
		if p.peek() == '(' {
			c.set = set
			return c, p.parseSet()
		}
		return c, set
	default:
		p.fatalf("exepected %s, %s, %s or %s; got %s",
			scanner.TokenString(scanner.Ident),
			scanner.TokenString(scanner.String),
			scanner.TokenString('*'),
			scanner.TokenString('!'),
			scanner.TokenString(l))
	}
	panic("unreacheable")
}

func (p *Parser) parseSet() set {
	p.eat('(')
	set := p.parseList()
	p.eat(')')
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
