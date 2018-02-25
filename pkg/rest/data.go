package rest

import (
	"context"
	"fmt"
	"strings"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/resolve"
	"bitbucket.org/fflo/semix/pkg/rule"
	"bitbucket.org/fflo/semix/pkg/semix"
	"upspin.io/errors"
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
	idx index.Putter,
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
	return index.Put(ctx, idx, s), nil
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

// Names for the different resolver types.
const (
	ThematicResolver = "thematic"
	RuledResolver    = "ruled"
	SimpleResolver   = "simple"
)

// MakeResolvers is a simple helper function to build resolvers from a list of strings.
func MakeResolvers(t float64, m int, rs []string) ([]Resolver, error) {
	res := make([]Resolver, len(rs))
	for i, r := range rs {
		switch strings.ToLower(r) {
		case ThematicResolver:
			res[i] = Resolver{
				Name:       ThematicResolver,
				MemorySize: m,
				Threshold:  t,
			}
		case RuledResolver:
			res[i] = Resolver{
				Name:       RuledResolver,
				MemorySize: m,
			}
		case SimpleResolver:
			res[i] = Resolver{
				Name:       SimpleResolver,
				MemorySize: m,
			}
		default:
			return nil, errors.Errorf("invalid resolver name: %s", r)
		}
	}
	return res, nil
}

func (r Resolver) resolver(rules rule.Map) (resolve.Interface, error) {
	switch strings.ToLower(r.Name) {
	case ThematicResolver:
		return resolve.Automatic{Threshold: r.Threshold}, nil
	case SimpleResolver:
		return resolve.Simple{}, nil
	case RuledResolver:
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
