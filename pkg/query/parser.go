package query

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
)

type parserError string

// Query represents a query.
type Query struct {
	constraint constraint
	set        set
}

// String returns a string representing the query.
func (q Query) String() string {
	return "?(" + q.constraint.String() + "(" + q.set.String() + "))"
}

type set map[string]bool

func (s set) String() string {
	if len(s) == 0 {
		return "{}"
	}
	sep := "{"
	str := ""
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		str += sep + k
		sep = ","
	}
	return str + "}"
}

type constraint struct {
	set      set
	not, all bool
}

func (c constraint) String() string {
	str := ""
	if c.not {
		str = "!"
	}
	if c.all {
		return str + "*"
	}
	return str + c.set.String()
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
	c, s := p.parseConstraint()
	return Query{set: s, constraint: c}
}

func (p *Parser) parseConstraint() (constraint, set) {
	var c constraint
	l := p.peek()
	if l.Typ == LexemeBang {
		p.eat(LexemeBang)
		c.not = true
		l = p.peek()
	}
	if l.Typ == LexemeStar {
		p.eat(LexemeStar)
		c.all = true
		l = p.peek()
	}
	if l.Typ == LexemeOBracet {
		c.set = p.parseSet()
	}
	p.eat(LexemeOBrace)
	s := p.parseSet()
	p.eat(LexemeCBrace)
	return c, s
}

func (p *Parser) parseSet() set {
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
