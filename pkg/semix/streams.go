package semix

import (
	"context"
	"sync"
)

// StreamToken Wraps either a token or an error
type StreamToken struct {
	Token Token
	Err   error
}

// Stream repsents a stream to read tokens.
type Stream <-chan StreamToken

// Normalize normalizes the token input.
// It prepends and appends one ' ' character to the token.
// All sequences of one or more unicode punctuation or unicode whitespaces
// are replaced by exactly one whitespace character ' '.
func Normalize(ctx context.Context, s Stream) Stream {
	nstream := make(chan StreamToken)
	go func() {
		defer close(nstream)
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-s:
				if !ok {
					return
				}
				if t.Err == nil {
					t.Token.Token = NormalizeString(t.Token.Token, true)
					t.Token.End = t.Token.Begin + len(t.Token.Token)
				}
				nstream <- t
			}
		}
	}()
	return nstream
}

// Match matches concepts in the stream and splits the tokens accordingly.
// So one token ' text <match> text ' is split into ' text ',
// '<match>' and ' text '.
func Match(ctx context.Context, m Matcher, s Stream) Stream {
	ms := make(chan StreamToken, 2) // matcher will put 2 token into the stream.
	go func() {
		defer close(ms)
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-s:
				if !ok {
					return
				}
				if t.Err != nil || t.Token.Concept != nil {
					ms <- t
				} else {
					doMatch(ctx, ms, t.Token, m)
				}
			}
		}
	}()
	return ms
}

func doMatch(ctx context.Context, s chan StreamToken, t Token, m Matcher) {
	if t.Concept != nil {
		panic("t.Token.Concept != nil")
	}
	rest := t.Token
	ofs := t.Begin
	// say.Info("### MATCHING TOKEN %v", t)
	for len(rest) > 0 {
		// // check for cancel
		// select {
		// case <-ctx.Done():
		// 	return
		// default:
		// }
		match := m.Match(rest)
		// say.Info("match: %v", match)
		if match.Concept == nil {
			putMatches(ctx, s, Token{
				Token:   rest,
				Path:    t.Path,
				Begin:   ofs,
				End:     ofs + len(rest),
				Concept: nil,
			})
			rest = ""
		} else if match.Begin == 0 {
			// say.Info("DO_MATCH: %v", match)
			putMatches(ctx, s, Token{
				Token:   rest[0:match.End],
				Path:    t.Path,
				Begin:   ofs,
				End:     ofs + match.End,
				Concept: match.Concept,
			})
			rest = rest[match.End:]
			ofs += match.End
		} else {
			// say.Info("DO_MATCH: %v", match)
			putMatches(ctx, s, Token{
				Token:   rest[0:match.Begin],
				Path:    t.Path,
				Begin:   ofs,
				End:     ofs + match.Begin,
				Concept: nil,
			},
				Token{
					Token:   rest[match.Begin:match.End],
					Path:    t.Path,
					Begin:   ofs + match.Begin,
					End:     ofs + match.End,
					Concept: match.Concept,
				})
			rest = rest[match.End:]
			ofs += match.End
		}
	}
}

func putMatches(ctx context.Context, out chan StreamToken, ts ...Token) {
	for _, t := range ts {
		select {
		case <-ctx.Done():
			return
		case out <- StreamToken{Token: t}:
			// say.Info("put token: %v", t)
		}
	}
}

// Read reads documents into tokens.
func Read(ctx context.Context, ds ...Document) Stream {
	rstream := make(chan StreamToken, len(ds))
	go func() {
		var wg sync.WaitGroup
		defer close(rstream)
		wg.Add(len(ds))
		for _, d := range ds {
			go func(d Document) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					return
				case rstream <- ReadStreamToken(d):
				}
			}(d)
		}
		wg.Wait()
	}()
	return rstream
}

// ReadStreamToken reads a StreamToken from a document.
// It simply wraps ReadToken and returns a StreamToken
func ReadStreamToken(d Document) StreamToken {
	token, err := ReadToken(d)
	if err != nil {
		return StreamToken{Err: err}
	}
	return StreamToken{Token: token}
}
