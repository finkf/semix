package semix

import (
	"io/ioutil"
	"regexp"
	"sync"
)

// StreamToken Wraps either a token or an error
type StreamToken struct {
	Token Token
	Err   error
}

// Stream repsents a stream to read tokens.
type Stream <-chan StreamToken

// Filter discards all tokens that do not match a concept.
func Filter(s Stream) Stream {
	fstream := make(chan StreamToken)
	go func() {
		defer close(fstream)
		for t := range s {
			if t.Err == nil && t.Token.Concept == nil {
				continue
			}
			fstream <- t
		}
	}()
	return fstream
}

// Normalize normalizes the token input.
// It prepends and appends one ' ' character to the token.
// All sequences of one or more unicode punctuation or unicode whitespaces
// are replaced by exactly one whitespace character ' '.
func Normalize(s Stream) Stream {
	nstream := make(chan StreamToken)
	go func() {
		defer close(nstream)
		for t := range s {
			if t.Err == nil {
				t.Token.Token = regexp.MustCompile(`(\s|\pP|\pS)+`).
					ReplaceAllLiteralString(" "+t.Token.Token+" ", " ")
			}
			nstream <- t
		}
	}()
	return nstream
}

// Match matches concepts in the stream and splits the tokens accordingly.
// So one token ' text <match> text ' is split into ' text ',
// '<match>' and ' text '.
func Match(m Matcher, s Stream) Stream {
	ms := make(chan StreamToken, 2) // matcher will put 2 token into the stream.
	go func() {
		defer close(ms)
		for t := range s {
			if t.Err != nil {
				ms <- t
			} else {
				doMatch(ms, t.Token, m)
			}
		}
	}()
	return ms
}

func doMatch(s chan StreamToken, t Token, m Matcher) {
	rest := t.Token
	ofs := 0
	for len(rest) > 0 {
		match := m.Match(rest)
		if match.Concept == nil {
			// log.Printf("DO_MATCH: %v", match)
			s <- StreamToken{
				Token: Token{
					Token:   rest,
					Path:    t.Path,
					Begin:   ofs,
					End:     ofs + len(rest),
					Concept: nil,
				},
			}
			rest = ""
		} else if match.Begin == 0 {
			// log.Printf("DO_MATCH: %v", match)
			s <- StreamToken{
				Token: Token{
					Token:   rest[0:match.End],
					Path:    t.Path,
					Begin:   ofs,
					End:     ofs + match.End,
					Concept: match.Concept,
				},
			}
			rest = rest[match.End:]
			ofs += match.End
		} else {
			// log.Printf("DO_MATCH: %v", match)
			s <- StreamToken{
				Token: Token{
					Token:   rest[0:match.Begin],
					Path:    t.Path,
					Begin:   ofs,
					End:     ofs + match.Begin,
					Concept: nil,
				},
			}
			s <- StreamToken{
				Token: Token{
					Token:   rest[match.Begin:match.End],
					Path:    t.Path,
					Begin:   ofs + match.Begin,
					End:     ofs + match.End,
					Concept: match.Concept,
				},
			}
			rest = rest[match.End:]
			ofs += match.End
		}
	}
}

// Read reads documents into tokens.
func Read(ds ...Document) Stream {
	rstream := make(chan StreamToken, len(ds))
	go func() {
		var wg sync.WaitGroup
		defer close(rstream)
		wg.Add(len(ds))
		for _, d := range ds {
			go func(d Document) {
				defer wg.Done()
				rstream <- readToken(d)
			}(d)
		}
		wg.Wait()
	}()
	return rstream
}

func readToken(d Document) StreamToken {
	defer d.Close()
	bs, err := ioutil.ReadAll(d)
	if err != nil {
		return StreamToken{Err: err}
	}
	return StreamToken{
		Token: Token{
			Token: string(bs),
			Path:  d.Path(),
			Begin: 0,
			End:   len(bs) + 2,
		},
	}
}
