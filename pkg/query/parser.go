package query

import (
	"errors"
	"fmt"
	"strconv"
)

type parserError string

// Parser represents a query parser.
type Parser struct {
	lexemes []Lexeme
	query   string
	pos     int
}

// NewParser create a new parser.
func NewParser(query string) *Parser {
	return &Parser{query: query}
}

// Parse parses a query.
func (p *Parser) Parse() (q Query, err error) {
	defer func() {
		if r, ok := recover().(parserError); ok {
			q = Query{}
			err = errors.New(string(r))
		}
	}()
	lexer := NewLexer(p.query)
	ls, err := lexer.Lex()
	if err != nil {
		return Query{}, fmt.Errorf("invalid query %q: %v", p.query, err)
	}
	p.lexemes = ls
	return p.parseQueryExp(), nil
}

func (p *Parser) parseQueryExp() Query {
	p.eat(LexemeQuest)
	var k int
	if p.peek().Typ == LexemeNumber { // parse optional number after `?`
		l := p.eat(LexemeNumber)
		tmp, err := strconv.ParseInt(l.Str, 10, 32)
		if err != nil {
			panic("not a number: " + l.Str)
		}
		k = int(tmp)
	}
	p.eat(LexemeOBrace)
	c, s := p.parseConstraint()
	p.eat(LexemeCBrace)
	return Query{set: s, constraint: c, l: k}
}

func (p *Parser) parseConstraint() (constraint, set) {
	var c constraint
	l := p.peek()
	// {<-...}
	if l.Typ == LexemeOBracet {
		return c, p.parseSet()
	}
	// !<-...
	if l.Typ == LexemeBang {
		p.eat(LexemeBang)
		c.not = true
		l = p.peek()
	}
	// *<-(...)
	if l.Typ == LexemeStar {
		p.eat(LexemeStar)
		c.all = true
		p.eat(LexemeOBrace)
		s := p.parseSet()
		p.eat(LexemeCBrace)
		return c, s
	}
	// A<-,...
	if l.Typ == LexemeIdent {
		c.set = p.parseList()
		p.eat(LexemeOBrace)
		s := p.parseSet()
		p.eat(LexemeCBrace)
		return c, s
	}
	p.die(l.Typ, LexemeIdent, LexemeStar, LexemeBang, LexemeOBracet)
	panic("not reached")
}

func (p *Parser) parseSet() set {
	p.eat(LexemeOBracet)
	set := p.parseList()
	p.eat(LexemeCBracet)
	return set
}

func (p *Parser) parseList() set {
	set := make(map[string]bool)
	// check for empty
	if p.peek().Typ != LexemeIdent {
		return set
	}
	set[p.eat(LexemeIdent).Str] = true
	for l := p.peek(); l.Typ == LexemeComma; l = p.peek() {
		p.eat(LexemeComma)
		set[p.eat(LexemeIdent).Str] = true
	}
	return set
}

func (p *Parser) peek() Lexeme {
	if p.pos >= len(p.lexemes) {
		return Lexeme{}
	}
	return p.lexemes[p.pos]
}

func (p *Parser) eat(typ int) Lexeme {
	if p.pos >= len(p.lexemes) {
		panic(parserError("premature EOF"))
	}
	l := p.lexemes[p.pos]
	p.pos++
	if l.Typ != typ {
		p.die(l.Typ, typ)
	}
	return l
}

func (p *Parser) die(not int, exp ...int) {
	str := fmt.Sprintf("at pos %d: expected", p.pos)
	c := ' '
	for _, i := range exp {
		str += fmt.Sprintf("%c%q", c, rune(i))
		c = ','
	}
	str += fmt.Sprintf("; got %q", rune(not))
	panic(parserError(str))
}
