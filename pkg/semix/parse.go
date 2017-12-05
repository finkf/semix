package semix

import (
	"fmt"
	"sort"

	"bitbucket.org/fflo/semix/pkg/traits"
)

// Parser defines a parser that parses (Subject, Predicate, Object) triples.
type Parser interface {
	Parse(func(string, string, string) error) error
}

// Parse creates a resource from a parser.
func Parse(p Parser, t traits.Interface) (*Resource, error) {
	parser := newParser(t)
	if err := p.Parse(parser.add); err != nil {
		return nil, err
	}
	return parser.parse()
}

type label struct {
	url       string
	ambiguous bool
}

type parser struct {
	predicates map[string]map[spo]bool
	names      map[string]string
	labels     map[string]label
	splits     map[string][]string
	rules      RulesDictionary
	traits     traits.Interface
}

func newParser(traits traits.Interface) *parser {
	return &parser{
		predicates: make(map[string]map[spo]bool),
		names:      make(map[string]string),
		labels:     make(map[string]label),
		splits:     make(map[string][]string),
		rules:      make(RulesDictionary),
		traits:     traits,
	}
}

func (parser *parser) parse() (*Resource, error) {
	g := parser.buildGraph()
	d := parser.buildDictionary(g)
	return NewResource(g, d, parser.rules), nil
}

func (parser *parser) buildGraph() *Graph {
	g := NewGraph()
	for p, spos := range parser.predicates {
		if parser.traits.IsInverted(p) {
			parser.predicates[p] = invert(spos)
		}
		if parser.traits.IsSymmetric(p) {
			parser.predicates[p] = calculateSymmetricClosure(spos)
		}
		if parser.traits.IsTransitive(p) {
			parser.predicates[p] = calculateTransitiveClosure(spos)
		}
		// insert triples into graph and set names of concepts
		for spo := range parser.predicates[p] {
			s, _, o := g.Add(spo.s, spo.p, spo.o)
			if name, ok := parser.names[s.url]; ok {
				s.Name = name
			}
			if name, ok := parser.names[o.url]; ok {
				o.Name = name
			}
		}
	}
	return g
}

func (parser *parser) buildDictionary(g *Graph) Dictionary {
	d := make(Dictionary)
	// simple entries
	for entry, label := range parser.labels {
		if c, ok := g.FindByURL(label.url); ok {
			id := c.ID()
			if label.ambiguous {
				id = -id
			}
			d[entry] = id
		}
	}
	// ambiguities
	var newURLs map[string]string
	if parser.traits.SplitAmbiguousURLs() {
		newURLs = parser.handleAmbiguitiesWithSplit(g)
	} else {
		newURLs = parser.handleAmbiguitiesWithMerge(g)
	}
	for entry, url := range newURLs {
		c, ok := g.FindByURL(url)
		if ok {
			// nicer names for concepts
			if c.Name == "" {
				c.Name = entry
			}
			d[entry] = c.ID()
		}
	}
	return d
}

func (parser *parser) handleAmbiguitiesWithSplit(g *Graph) map[string]string {
	newURLs := make(map[string]string)
	for entry := range parser.splits {
		urls := sortUnique(parser.splits[entry])
		newURL := CombineURLs(urls...)
		newURLs[entry] = newURL
		for _, url := range urls {
			g.Add(newURL, SplitURL, url)
		}
	}
	return newURLs
}

func (parser *parser) handleAmbiguitiesWithMerge(g *Graph) map[string]string {
	newURLs := make(map[string]string)
	for entry := range parser.splits {
		urls := sortUnique(parser.splits[entry])
		newURL := CombineURLs(urls...)
		newURLs[entry] = newURL
		edges := intersectEdges(g, urls...)
		g.Register(newURL)
		for p, os := range edges {
			for o := range os {
				g.Add(newURL, p, o)
			}
		}
	}
	return newURLs
}

func intersectEdges(g *Graph, urls ...string) map[string]map[string]struct{} {
	if len(urls) == 0 {
		return nil
	}
	a := makeEdgesMap(g, urls[0])
	for _, url := range urls[1:] {
		a = intersect(a, makeEdgesMap(g, url))
	}
	return a
}

func intersect(a, b map[string]map[string]struct{}) map[string]map[string]struct{} {
	c := make(map[string]map[string]struct{})
	for p, os := range a {
		for o := range os {
			if _, ok := b[p]; ok {
				if _, ok := b[p][o]; ok {
					if _, ok := c[p]; !ok {
						c[p] = make(map[string]struct{})
					}
					c[p][o] = struct{}{}
				}
			}
		}
	}
	return c
}

func makeEdgesMap(g *Graph, url string) map[string]map[string]struct{} {
	c, ok := g.FindByURL(url)
	if !ok {
		return nil
	}
	edges := make(map[string]map[string]struct{})
	for _, e := range c.edges {
		if edges[e.P.URL()] == nil {
			edges[e.P.URL()] = make(map[string]struct{})
		}
		edges[e.P.URL()][e.O.URL()] = struct{}{}
	}
	return edges
}

func sortUnique(urls []string) []string {
	urlset := make(map[string]bool)
	for _, url := range urls {
		urlset[url] = true
	}
	splits := make([]string, 0, len(urlset))
	for url := range urlset {
		splits = append(splits, url)
	}
	sort.Strings(splits)
	return splits
}

func (parser *parser) add(s, p, o string) error {
	if parser.traits.Ignore(p) {
		return nil
	}
	if parser.traits.IsRule(p) {
		parser.rules[s] = o
		return nil
	}
	if parser.traits.IsName(p) {
		return parser.addLabels(o, s, false, true)
	}
	if parser.traits.IsAmbiguous(p) {
		return parser.addLabels(o, s, true, false)
	}
	if parser.traits.IsDistinct(p) {
		return parser.addLabels(o, s, false, false)
	}
	return parser.addTriple(s, p, o)
}

func (parser *parser) addTriple(s, p, o string) error {
	triple := spo{s, p, o}
	if _, ok := parser.predicates[p]; !ok {
		parser.predicates[p] = make(map[spo]bool)
	}
	parser.predicates[p][triple] = true
	return nil
}

func (parser *parser) addLabels(entry, url string, ambig, name bool) error {
	labels, err := ExpandBraces(entry)
	if err != nil {
		return fmt.Errorf("could not expand: %v", err)
	}
	for _, expanded := range labels {
		normalized := NormalizeString(expanded, false)
		if _, ok := parser.splits[normalized]; ok {
			parser.splits[normalized] = append(parser.splits[normalized], url)
			return nil
		}
		if l, ok := parser.labels[normalized]; ok && l.url != url {
			delete(parser.labels, normalized)
			parser.splits[normalized] = append(parser.splits[normalized], url)
			parser.splits[normalized] = append(parser.splits[normalized], l.url)
			return nil
		}
		// name can/should never be part of a split
		if name && !ambig {
			if _, ok := parser.names[url]; !ok {
				// the name should not be normalized, so it looks nicer.
				// the name is still put normalized into the dictionary.
				parser.names[url] = expanded
			}
		}
		parser.labels[normalized] = label{url, ambig}
	}
	return nil
}
