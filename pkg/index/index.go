package index

import (
	"context"

	"bitbucket.org/fflo/semix/pkg/semix"
)

// Entry denotes a public available index entry
type Entry struct {
	ConceptURL, Path, RelationURL, Token string
	Begin, End, L                        int
}

// Putter represents a simple interface to put tokens into an index.
type Putter interface {
	Put(semix.Token) error
}

// Index represents the basic interface to put and get tokens from an index.
type Index interface {
	Putter
	Get(string, func(Entry)) error
	Close() error
}

// Put reads all tokens from a given stream into an index.
func Put(ctx context.Context, index Putter, s semix.Stream) semix.Stream {
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
				err := index.Put(t.Token)
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
