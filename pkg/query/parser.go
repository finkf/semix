package query

import (
	"fmt"

	"github.com/pkg/errors"
)

type parserError string

// Query represents a query.
type Query struct {
	str string
	qs  []Query
}

// String returns a string representing the query.
func (q Query) String() string {
	if len(q.qs) > 0 {
		return fmt.Sprintf("%s %v", q.str, q.qs)
	}
	return fmt.Sprintf("%s", q.str)
}

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
		return Query{}, errors.Wrapf(err, "invalid query %q", p.query)
	}
	p.lexemes = ls
	return p.parseQueryExp(), nil
}

func (p *Parser) parseQueryExp() Query {
	p.eat(LexemeQuest)
	p.eat(LexemeOBrace)
	q := Query{str: "?"}
	q.qs = append(q.qs, p.parseImg())
	p.eat(LexemeCBrace)
	return q
}

func (p *Parser) parseImg() Query {
	l := p.peek()
	if l.Typ != LexemeRight && l.Typ != LexemeLeft {
		die(l.Typ, LexemeLeft, LexemeRight)
	}
	p.eat(l.Typ)
	q := Query{str: l.Str}
	q.qs = append(q.qs, p.parseSet())
	return q
}

func (p *Parser) parseSet() Query {
	l := p.peek()
	switch {
	case l.Typ == LexemeStar:
		p.eat(l.Typ)
		return Query{str: l.Str}
	case l.Typ == LexemeOBracet:
		p.eat(l.Typ)
		q := Query{str: "SET"}
		for {
			l := p.peek()
			switch {
			case l.Typ == LexemeCBracet:
				p.eat(l.Typ)
				return q
			case l.Typ == LexemeIdent:
				p.eat(l.Typ)
				q.qs = append(q.qs, Query{str: l.Str})
				l = p.peek()
				if l.Typ == LexemeCBracet {
					p.eat(l.Typ)
					return q
				}
				p.eat(LexemeComma)
			default:
				die(l.Typ, LexemeCBracet, LexemeComma)
			}
		}
		// panic(parserError("missing `}`"))
	default:
		die(l.Typ, LexemeStar, LexemeOBracet)
	}
	panic("not reached")
}

func (p *Parser) parseIdentSet() Query {
	p.eat(LexemeOBracet)
	set := make(map[string]bool)
	for l := p.peek(); l.Typ != LexemeCBracet; l = p.peek() {
		ident := p.eat(LexemeIdent)
		set[ident.Str] = true
		if p.peek().Typ == LexemeComma {
			p.eat(LexemeComma)
		}
	}
	p.eat(LexemeCBracet)
	return Query{}
}

func (p *Parser) peek() Lexeme {
	if p.pos >= len(p.lexemes) {
		return Lexeme{}
	}
	return p.lexemes[p.pos]
}

func (p *Parser) eat(typ int) Lexeme {
	if p.pos >= len(p.lexemes) {
		return Lexeme{}
	}
	l := p.lexemes[p.pos]
	p.pos++
	if l.Typ != typ {
		die(l.Typ, typ)
	}
	return l
}

func die(not int, exp ...int) {
	str := "expected"
	c := ' '
	for _, i := range exp {
		str += fmt.Sprintf("%c%q", c, rune(i))
		c = ','
	}
	str += fmt.Sprintf("; got %q", rune(not))
	panic(parserError(str))
}
