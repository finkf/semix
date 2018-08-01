package resolve

import (
	"context"

	"gitlab.com/finkf/semix/pkg/memory"
	"gitlab.com/finkf/semix/pkg/semix"
)

// Interface defines the interface for the disambiguation.
type Interface interface {
	// Resolve returns the disambiguated concept or nil if the
	// concept cannot be disambiguated. It is an error
	// to call Resolve with a non-ambigiuous concept.
	Resolve(*semix.Concept, *memory.Memory) *semix.Concept
}

// Resolve resolves ambiguities using the given Interface.
func Resolve(ctx context.Context, n int, r Interface, s semix.Stream) semix.Stream {
	rstream := make(chan semix.StreamToken)
	go func() {
		defer close(rstream)
		mem := make(map[string]*memory.Memory)
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-s:
				if !ok {
					return
				}
				if t.Err == nil && mem[t.Token.Path] == nil {
					mem[t.Token.Path] = memory.New(n)
				}
				if t.Err == nil && t.Token.Concept != nil && t.Token.Concept.Ambig() {
					t.Token = doResolve(t.Token, r, mem[t.Token.Path])
				}
				if t.Err == nil && t.Token.Concept != nil && !t.Token.Concept.Ambig() {
					mem[t.Token.Path].Push(t.Token.Concept)
				}
				rstream <- t
			}
		}
	}()
	return rstream
}

func doResolve(t semix.Token, r Interface, mem *memory.Memory) semix.Token {
	if c := r.Resolve(t.Concept, mem); c != nil {
		t.Concept = c
	}
	return t
}
