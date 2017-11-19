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
	scanner *scanner.Scanner
	p       rune
}

func newParser(r io.Reader) *parser {
	scanner := &scanner.Scanner{
		Error: die,
		Mode:  ruleTokens,
	}
	scanner.Init(r)
	scanner.Filename = "rule"
	return &parser{scanner: scanner}
}

func (p *parser) parse() (a ast, err error) {
	defer func() {
		if r, ok := recover().(parseError); ok {
			a = nil
			err = errors.New(r.msg)
		}
	}()
	switch p.peek() {
	case '{':
		a = p.parseSet()
	case scanner.Float:
		a = p.parseNum()
	case scanner.Int:
		a = p.parseNum()
	default:
		dief(p.scanner, "invalid: %s", scanner.TokenString(p.peek()))
	}
	p.eat(scanner.EOF)
	return a, nil
}

func (p *parser) parseSet() set {
	p.eat('{')
	set := make(set)
	for _, str := range p.parseStrList('}') {
		set[str] = true
	}
	p.eat('}')
	return set
}

func (p *parser) parseStrList(end rune) []string {
	var strs []string
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

func (p *parser) parseStr() string {
	_, str := p.eat(scanner.String)
	str, err := strconv.Unquote(str)
	if err != nil {
		dief(p.scanner, "invalid string: %s", err)
	}
	return str
}

func (p *parser) parseNum() num {
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
	panic("not reached")
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
