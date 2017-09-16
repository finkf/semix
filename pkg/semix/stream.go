package semix

import (
	"context"
	"io"
	"io/ioutil"
	"sync"

	"github.com/sirupsen/logrus"
)

type StreamToken struct {
	Token Token
	Err   error
}

// TokenStream repsents a strem to read tokens.
type Stream <-chan StreamToken

// Filter filters all tokens for which f returns true.
func Filter(ctx context.Context, s Stream) Stream {
	filterstream := make(chan StreamToken)
	go func() {
		defer close(filterstream)
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-s:
				logrus.Debugf("FILTER: %v %t", t, ok)
				if !ok {
					return
				}
				if t.Token.Concept != nil {
					filterstream <- t
				}
			}
		}
	}()
	return filterstream
}

func Take(ctx context.Context, n int, stream Stream) Stream {
	takestream := make(chan StreamToken)
	go func() {
		defer close(takestream)
		for i := 0; i < n; i++ {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-stream:
				if !ok {
					return
				}
				takestream <- t
			}
		}
	}()
	return takestream
}

func Repeat(ctx context.Context, tokens ...string) Stream {
	rs := make(chan StreamToken)
	go doRepeat(ctx, rs, tokens...)
	return rs
}

func doRepeat(ctx context.Context, rs chan StreamToken, tokens ...string) {
	defer close(rs)
	for {
		for _, token := range tokens {
			t := StreamToken{
				Token: Token{
					Token:   token,
					Concept: nil,
					Begin:   0,
					End:     len(token),
				},
				Err: nil,
			}
			select {
			case <-ctx.Done():
				return
			case rs <- t:
			}
		}
	}
}

func Match(ctx context.Context, m Matcher, s Stream) Stream {
	ms := make(chan StreamToken, 2)
	go func() {
		defer close(ms)
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-s:
				logrus.Debugf("MATCH %v %t", t, ok)
				if !ok {
					return
				}
				if t.Err != nil {
					ms <- t
					continue
				}
				doMatch(ms, t.Token, m)
			}
		}
	}()
	return ms
}

// TODO: this does not use cancellation. It just insert tokens into the stream.
func doMatch(s chan StreamToken, t Token, m Matcher) {
	rest := t.Token
	ofs := 0
	for len(rest) > 0 {
		match := m.Match(rest)
		if match.Concept == nil {
			logrus.Debugf("DO_MATCH: %v", match)
			s <- StreamToken{
				Token: Token{
					Token:   rest,
					Begin:   ofs,
					End:     ofs + len(rest),
					Concept: nil,
				},
			}
			rest = ""
		} else if match.Begin == 0 {
			logrus.Debugf("DO_MATCH: %v", match)
			s <- StreamToken{
				Token: Token{
					Token:   rest[0:match.End],
					Begin:   ofs,
					End:     ofs + match.End,
					Concept: match.Concept,
				},
			}
			rest = rest[match.End:]
			ofs += match.End
		} else {
			logrus.Debugf("DO_MATCH: %v", match)
			s <- StreamToken{
				Token: Token{
					Token:   rest[0:match.Begin],
					Begin:   ofs,
					End:     ofs + match.Begin,
					Concept: nil,
				},
			}
			s <- StreamToken{
				Token: Token{
					Token:   rest[match.Begin:match.End],
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

func Read(ctx context.Context, rs ...io.Reader) Stream {
	rstream := make(chan StreamToken, len(rs))
	go func() {
		defer close(rstream)
		var wg sync.WaitGroup
		wg.Add(len(rs))
		for _, r := range rs {
			go func(r io.Reader) {
				defer wg.Done()
				token := readToken(r)
				logrus.Debugf("READ %v", token)
				select {
				case <-ctx.Done():
					return
				case rstream <- token:
					return
				}
			}(r)
		}
		wg.Wait()
	}()
	return rstream
}

func readToken(r io.Reader) StreamToken {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return StreamToken{Err: err}
	}
	return StreamToken{
		Token: Token{
			Token: " " + string(bs) + " ",
			Begin: 0,
			End:   len(bs) + 2,
		},
	}
}
