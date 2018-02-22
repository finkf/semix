package rest

import (
	"context"
	"fmt"
	"strings"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/resolve"
	"bitbucket.org/fflo/semix/pkg/rule"
	"bitbucket.org/fflo/semix/pkg/semix"
)

// DumpFileContent defines the content and the content type of a dump file.
type DumpFileContent struct {
	ContentType, Content, Path string
}

// PutData defines the data that is send to the server's put method
type PutData struct {
	URL         string
	Local       bool
	Errors      []int
	Resolvers   []Resolver
	ContentType string
	Content     string
}

func (p PutData) stream(
	ctx context.Context,
	dfa semix.DFA,
	rules rule.Map,
	idx index.Interface,
	dir string,
) (semix.Stream, error) {
	doc, err := p.document(dir)
	if err != nil {
		return nil, err
	}
	s := p.matchStream(ctx, dfa, semix.Normalize(ctx, semix.Read(ctx, doc)))
	s, err = p.resolveStream(ctx, rules, s)
	if err != nil {
		return nil, err
	}
	return index.Put(ctx, idx, semix.Filter(ctx, s)), nil
}

func (p PutData) matchStream(
	ctx context.Context,
	dfa semix.DFA,
	s semix.Stream,
) semix.Stream {
	for i := len(p.Errors); i > 0; i-- {
		l := p.Errors[i-1]
		if l <= 0 {
			continue
		}
		s = semix.Match(
			ctx, semix.FuzzyDFAMatcher{DFA: semix.NewFuzzyDFA(l, dfa)}, s)
	}
	return semix.Match(ctx, semix.DFAMatcher{DFA: dfa}, s)
}

func (p PutData) resolveStream(
	ctx context.Context,
	rules rule.Map,
	s semix.Stream,
) (semix.Stream, error) {
	for i := len(p.Resolvers); i > 0; i-- {
		resolver, err := p.Resolvers[i-1].resolver(rules)
		if err != nil {
			return nil, err
		}
		s = resolve.Resolve(ctx, p.Resolvers[i-1].MemorySize, resolver, s)
	}
	return s, nil
}

func (p PutData) document(dir string) (semix.Document, error) {
	if p.Content == "" {
		if p.Local {
			return semix.NewFileDocument(p.URL), nil
		}
		return semix.NewHTTPDocument(p.URL), nil
	}
	doc, err := newDumpFile(strings.NewReader(p.Content), dir, p.URL, p.ContentType)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// Resolver defines one of the three resolvers simple, automatic or ruled
// as defined in bitbucket.org/fflo/semix/pkg/resolve
type Resolver struct {
	Name       string
	Threshold  float64
	MemorySize int
}

// NewThematicResolver creates a new thematic resolver.
func NewThematicResolver(m int, t float64) Resolver {
	return Resolver{Name: "automatic", Threshold: t, MemorySize: m}
}

// NewRuledResolver creates a new ruled resolver.
func NewRuledResolver(m int) Resolver {
	return Resolver{Name: "ruled", MemorySize: m}
}

// NewSimpleResolver create a new simple resolver.
func NewSimpleResolver(m int) Resolver {
	return Resolver{Name: "simple", MemorySize: m}
}

func (r Resolver) resolver(rules rule.Map) (resolve.Interface, error) {
	switch strings.ToLower(r.Name) {
	case "automatic":
		return resolve.Automatic{Threshold: r.Threshold}, nil
	case "simple":
		return resolve.Simple{}, nil
	case "ruled":
		return resolve.Ruled{Rules: rules}, nil
	}
	return nil, fmt.Errorf("invalid resolver name: %s", r.Name)
}

// ConceptInfo holds information about a concept.
type ConceptInfo struct {
	Concept *semix.Concept
	Entries []string
}

// Predicates returns a map of the targets ordered by the predicates.
func (info ConceptInfo) Predicates() map[*semix.Concept][]*semix.Concept {
	m := make(map[*semix.Concept][]*semix.Concept)
	if info.Concept == nil {
		return m
	}
	info.Concept.EachEdge(func(edge semix.Edge) {
		m[edge.P] = append(m[edge.P], edge.O)
	})
	return m
}

// Context specifies the context of a match
type Context struct {
	Before, Match, After, URL string
	Begin, End, Len           int
}
