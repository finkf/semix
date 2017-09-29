package semix

import (
	"context"
	"io/ioutil"
	"sync"
)

// StreamToken Wraps either a token or an error
type StreamToken struct {
	Token Token
	Err   error
}

// Stream repsents a stream to read tokens.
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
				// logrus.Debugf("FILTER: %v %t", t, ok)
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

// Take takes no more than n tokens from a stream.
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
				// logrus.Debugf("MATCH %v %t", t, ok)
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
			// logrus.Debugf("DO_MATCH: %v", match)
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
			// logrus.Debugf("DO_MATCH: %v", match)
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
			// logrus.Debugf("DO_MATCH: %v", match)
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
func Read(ctx context.Context, ds ...Document) Stream {
	rstream := make(chan StreamToken, len(ds))
	go func() {
		defer close(rstream)
		var wg sync.WaitGroup
		wg.Add(len(ds))
		for _, d := range ds {
			go func(d Document) {
				defer wg.Done()
				token := readToken(d)
				// logrus.Debugf("READ %v", token)
				select {
				case <-ctx.Done():
					return
				case rstream <- token:
					return
				}
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
			Token: " " + string(bs) + " ",
			Path:  d.Path(),
			Begin: 0,
			End:   len(bs) + 2,
		},
	}
}
