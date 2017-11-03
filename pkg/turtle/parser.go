package turtle

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Parser parses a turtle knowledge base.
type Parser struct {
	l lexer
	p map[string]string
	b string
	n *token
}

// NewParser create a new parser instance.
func NewParser(r io.Reader) *Parser {
	return &Parser{
		l: lexer{r: bufio.NewReader(r)},
		p: make(map[string]string),
	}
}

// Parse reads the triples in the turtle file and calls the callback function
// for each encountered triple.
func (parser *Parser) Parse(f func(string, string, string) error) (err error) {
	defer func() {
		if le, ok := recover().(lexerError); ok {
			err = le.err
		}
	}()
	for done, err := parser.parse(f); !done; done, err = parser.parse(f) {
		if err != nil {
			return err
		}
	}
	return nil
}

func (parser *Parser) parse(f func(string, string, string) error) (bool, error) {
	switch t := parser.peek(); t.typ {
	case prefix:
		parser.parsePrefix()
		return false, nil
	case base:
		parser.parseBase()
		return false, nil
	case eof:
		return true, nil
	case word:
		return false, parser.parseTriples(f)
	default:
		return false, fmt.Errorf("invalid token: %s", t.typ)
	}
}

func (parser *Parser) parseTriples(f func(string, string, string) error) error {
	s := parser.nextWord()
	p := parser.nextWord()
	o := parser.nextWord()
	// TODO: error
	if err := f(s, p, o); err != nil {
		return err
	}
loop:
	for {
		switch t := parser.peek(); t.typ {
		case comma:
			parser.eat(comma)
			o = parser.nextWord()
			if err := f(s, p, o); err != nil {
				return err
			}
		case semicolon:
			parser.eat(semicolon)
			p = parser.nextWord()
			o = parser.nextWord()
			if err := f(s, p, o); err != nil {
				return err
			}
		case dot:
			break loop
		default:
			panic(lexerError{fmt.Errorf("invalid token: %s", t.typ)})
		}
	}
	parser.eat(dot)
	return nil
}

func (parser *Parser) parseBase() {
	t := parser.eat(base)
	base := parser.eat(word)
	if !t.specialSyntax {
		parser.eat(dot)
	}
	parser.b = base.str
}

func (parser *Parser) parsePrefix() {
	t := parser.eat(prefix)
	prefix := parser.eat(word)
	url := parser.nextWord()
	if !t.specialSyntax {
		parser.eat(dot)
	}
	parser.p[prefix.str] = url
}

func (parser *Parser) nextWord() string {
	t := parser.eat(word)
	if t.specialSyntax {
		return t.str
	}
	slashpos := strings.Index(t.str, "//")
	if slashpos > 0 {
		return t.str
	}
	i := strings.Index(t.str, ":")
	if i < 0 {
		return parser.b + t.str
	}
	if _, ok := parser.p[t.str[:i+1]]; !ok {
		panic(lexerError{fmt.Errorf("invalid prefix: %q", t.str[:i])})
	}
	return parser.p[t.str[:i+1]] + t.str[i+1:]
}

func (parser *Parser) peek() *token {
	if parser.n == nil {
		parser.n = parser.l.next()
	}
	return parser.n
}

func (parser *Parser) next() *token {
	t := parser.peek()
	parser.n = parser.l.next()
	return t
}

func (parser *Parser) eat(typ tokenType) *token {
	next := parser.next()
	if next.typ != typ {
		panic(lexerError{fmt.Errorf("expected %s; got %s", typ, next.typ)})
	}
	return next
}

type tokenType int

const (
	eof tokenType = iota
	dot
	comma
	semicolon
	word
	base
	prefix
)

func (t tokenType) String() string {
	switch t {
	case eof:
		return "EOF"
	case dot:
		return "."
	case comma:
		return ","
	case semicolon:
		return ";"
	case word:
		return "STRING"
	case base:
		return "@base"
	case prefix:
		return "@prefix"
	default:
		panic("invalid token type")
	}
}

type lexerError struct {
	err error
}

type token struct {
	str           string
	typ           tokenType
	specialSyntax bool
}

func (t token) String() string {
	return t.str
}

type lexer struct {
	r *bufio.Reader
}

func (l lexer) next() *token {
	l.skipWhiteSpace()
	switch l.peekChar() {
	case '@':
		return l.parseAnnotation()
	case '#':
		l.skipComment()
		return l.next()
	case '<':
		return l.parseURL()
	case '"':
		return l.parseQuotedString()
	case '.':
		l.eat('.')
		return &token{".", dot, false}
	case ',':
		l.eat(',')
		return &token{",", comma, false}
	case ';':
		l.eat(';')
		return &token{";", semicolon, false}
	case 0:
		return &token{"EOF", eof, false}
	default:
		return l.parseString()
	}
}

func (l lexer) parseAnnotation() *token {
	l.eat('@')
	str := l.nextWord(' ')
	switch str {
	case "prefix":
		return &token{str, prefix, false}
	case "base":
		return &token{str, base, false}
	default:
		panic(lexerError{fmt.Errorf("invalid annotation: @%s", str)})
	}
}

func (l lexer) skipComment() {
	l.eat('#')
	l.nextWord('\n')
}

func (l lexer) parseURL() *token {
	l.eat('<')
	str := l.nextWord('>')
	return &token{str, word, false}
}

func (l lexer) parseQuotedString() *token {
	l.eat('"')
	str := l.nextWord('"')
	// TODO: unescape xml
	return &token{str, word, true}
}

func (l lexer) parseString() *token {
	var bs []byte
	for {
		switch l.peekChar() {
		case '@', ',', '.', ';', 0, ' ', '\r', '\n', '\t', '\v', '"', '<':
			switch str := string(bs); str {
			case "PREFIX":
				return &token{str, prefix, true}
			case "BASE":
				return &token{str, base, true}
			default:
				return &token{str, word, false}
			}
		default:
			bs = append(bs, l.nextChar())
		}
	}
}

func (l lexer) nextWord(delim byte) string {
	bs, err := l.r.ReadBytes(delim)
	if err == io.EOF {
		panic(lexerError{errors.New("unexpected EOF")})
	}
	if err != nil {
		panic(lexerError{err})
	}
	return string(bs[:len(bs)-1])
}

func (l lexer) eat(b byte) {
	n := l.nextChar()
	if n != b {
		panic(lexerError{fmt.Errorf("expected %q; got %q", b, n)})
	}
}

func (l lexer) peekChar() byte {
	b := l.nextChar()
	if b == 0 {
		return 0
	}
	if err := l.r.UnreadByte(); err != nil {
		panic(lexerError{err})
	}
	return b
}

func (l lexer) nextChar() byte {
	b, err := l.r.ReadByte()
	if err == io.EOF {
		return 0
	}
	if err != nil {
		panic(lexerError{err})
	}
	return b
}

func (l lexer) skipWhiteSpace() {
	for b := l.peekChar(); unicode.IsSpace(rune(b)); b = l.peekChar() {
		l.nextChar()
	}
}
