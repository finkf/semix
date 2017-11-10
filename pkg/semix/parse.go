package semix

import (
	"fmt"
	"sort"
)

// Parser defines a parser that parses (Subject, Predicate, Object) triples.
type Parser interface {
	Parse(func(string, string, string) error) error
}

// Traits define the different traits of predicates.
type Traits interface {
	Ignore(string) bool
	IsSymmetric(string) bool
	IsTransitive(string) bool
	IsName(string) bool
	IsDistinct(string) bool
	IsAmbiguous(string) bool
	IsInverted(string) bool
}

// Dictionary is a dictionary that maps the labels of the concepts
// to their apporpriate IDs. Negative IDs mark ambigous dictionary entries.
// The map to the according positve ID.
type Dictionary map[string]int32

// Parse creates a graph and a dictionary from a parser.
func Parse(p Parser, t Traits) (*Graph, Dictionary, error) {
	parser := newParser(t)
	if err := p.Parse(parser.add); err != nil {
		return nil, nil, err
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
	traits     Traits
}

func newParser(traits Traits) *parser {
	return &parser{
		predicates: make(map[string]map[spo]bool),
		names:      make(map[string]string),
		labels:     make(map[string]label),
		splits:     make(map[string][]string),
		traits:     traits,
	}
}

func (parser *parser) parse() (*Graph, Dictionary, error) {
	g := parser.buildGraph()
	d := parser.buildDictionary(g)
	return g, d, nil
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
	// splits
	splitURLs := make(map[string]string)
	splitPreds := make(map[spo]bool)
	for entry := range parser.splits {
		urls := sortUnique(parser.splits[entry])
		splitURL := CombineURLs(urls...)
		splitURLs[entry] = splitURL
		for _, url := range urls {
			splitPreds[spo{splitURL, SplitURL, url}] = true
		}
	}
	for spo := range splitPreds {
		g.Add(spo.s, spo.p, spo.o)
	}
	for entry, url := range splitURLs {
		c, ok := g.FindByURL(url)
		if ok {
			d[entry] = c.ID()
		}
	}
	return d
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
				parser.names[url] = normalized
			}
		}
		parser.labels[normalized] = label{url, ambig}
	}
	return nil
}
