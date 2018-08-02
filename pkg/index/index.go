package index

import (
	"context"

	"github.com/finkf/semix/pkg/say"
	"github.com/finkf/semix/pkg/semix"
)

// Entry denotes a public available index entry
type Entry struct {
	ConceptURL, Path, RelationURL, Token string
	Begin, End, L                        int
	Ambiguous                            bool
}

// Direct returns true iff the entry represents a direct index entry.
func (e Entry) Direct() bool {
	return e.RelationURL == ""
}

// Putter represents a simple interface to put tokens into an index.
type Putter interface {
	Put(semix.Token) error
}

// Interface represents the basic interface to put and get tokens from an index.
type Interface interface {
	Putter
	Get(string, func(Entry) bool) error
	Close() error
	Flush() error
}

// Put reads all tokens from a given stream into an index.
// Put does only insert Tokens into the index that have an associated concept.
func Put(ctx context.Context, putter Putter, s semix.Stream) semix.Stream {
	istream := make(chan semix.StreamToken)
	go func() {
		defer close(istream)
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-s:
				if !ok {
					return
				}
				// If there is an error,
				// let it be handled upstream.
				if t.Err != nil {
					istream <- t
					continue
				}
				// If the token is not associated, discard it.
				if t.Token.Concept == nil {
					continue
				}
				// Only put associated tokens.
				say.Debug("putting %q into index", t.Token.Token)
				err := putter.Put(t.Token)
				if err != nil {
					istream <- semix.StreamToken{Err: err}
				} else {
					istream <- t
				}
			}

		}
	}()
	return istream
}
