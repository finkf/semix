package query

import (
	"errors"
	"unicode"
)

// Lexer identifies the lexems in a query.
type Lexer struct {
	str string
	pos int
}

type lexerError string

// LexemTypes
const (
	LexemeIdent   = 0
	LexemeLeft    = '<'
	LexemeRight   = '>'
	LexemeQuest   = '?'
	LexemeOBrace  = '('
	LexemeCBrace  = ')'
	LexemeOBracet = '{'
	LexemeCBracet = '}'
	LexemeComma   = ','
	LexemeStar    = '*'
	LexemeBang    = '!'
)

// Lexeme represents a lexem.
type Lexeme struct {
	Str string
	Typ int
}

// String returns the string representation of the lexem.
func (l Lexeme) String() string {
	return l.Str
}

// NewLexer create a new Lexer.
func NewLexer(str string) *Lexer {
	return &Lexer{str: str}
}

// Lex returns a slice of lexemes.
func (l *Lexer) Lex() (ls []Lexeme, err error) {
	defer func() {
		if r, ok := recover().(lexerError); ok {
			ls = nil
			err = errors.New(string(r))
		}
	}()
	for {
		lexeme, done := l.nextLexeme()
		if done {
			return ls, nil
		}
		ls = append(ls, lexeme)
	}
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.str) {
		return 0
	}
	return l.str[l.pos]
}

func (l *Lexer) next() byte {
	if l.pos >= len(l.str) {
		return 0
	}
	b := l.str[l.pos]
	l.pos++
	return b
}

func (l *Lexer) skip() {
	// just recognices ascii whitespace
	for isws(l.peek()) {
		l.next()
	}
}

func (l *Lexer) nextLexeme() (Lexeme, bool) {
	l.skip()
	c := l.peek()
	switch {
	case isop(c):
		l.next()
		return Lexeme{Typ: int(c), Str: l.str[l.pos-1 : l.pos]}, false
	case c == '\'' || c == '"':
		l.next()
		return l.parseQuotedIdent(c), false
	case c == 0:
		return Lexeme{}, true
	default:
		return l.parseIdent(), false
	}
}

func (l *Lexer) parseQuotedIdent(q byte) Lexeme {
	spos := l.pos
	for {
		c := l.next()
		switch {
		case c == q:
			return Lexeme{Typ: LexemeIdent, Str: l.str[spos : l.pos-1]}
		case c == 0:
			panic(lexerError("missing quotation"))
		}
	}
	// panic(lexerError("not reached"))
}

func (l *Lexer) parseIdent() Lexeme {
	spos := l.pos
	for {
		c := l.peek()
		switch {
		case isop(c) || c == '\'' || c == '"' || isws(c):
			return Lexeme{Typ: LexemeIdent, Str: l.str[spos:l.pos]}
		case c == 0:
			return Lexeme{Typ: LexemeIdent, Str: l.str[spos:]}
		default:
			l.next()
		}
	}
	// panic(lexerError("not reached"))
}

func isws(c byte) bool {
	return unicode.IsSpace(rune(c))
}

func isop(c byte) bool {
	switch {
	case c == '?' || c == '<' || c == '>' || c == '(' || c == ')' ||
		c == '{' || c == '}' || c == ',' || c == '*' || c == '!':
		return true
	default:
		return false
	}
}
