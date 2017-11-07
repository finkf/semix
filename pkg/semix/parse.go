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
	traits     Traits
}

func newParser(traits Traits) *parser {
	return &parser{
		predicates: make(map[string]map[spo]bool),
		names:      make(map[string]string),
		labels:     make(map[string]label),
		traits:     traits,
	}
}

func (parser *parser) parse() (*Graph, Dictionary, error) {
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
	d := make(Dictionary)
	for entry, label := range parser.labels {
		if c, ok := g.FindByURL(label.url); ok {
			id := c.ID()
			if label.ambiguous {
				id = -id
			}
			d[entry] = id
		}
	}
	return g, d, nil
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
		if name {
			if _, ok := parser.names[url]; !ok {
				parser.names[url] = expanded
			}
		}
		if l, ok := parser.labels[expanded]; ok && l.url != url {
			return parser.addSplit(expanded, l.url, url)
		}
		parser.labels[expanded] = label{url, ambig}
	}
	return nil
}

func (parser *parser) addSplit(entry, aurl, burl string) error {
	splits := []string{burl}
	var founds []spo
	for t := range parser.predicates[SplitURL] {
		if t.s == aurl {
			founds = append(founds, t)
			splits = append(splits, t.o)
		}
	}
	if len(founds) == 0 {
		splits = append(splits, aurl)
	}
	for _, t := range founds {
		delete(parser.predicates[SplitURL], t)
	}
	// sort for stability
	sort.Strings(splits)
	splitURL := CombineURLs(splits...)
	parser.labels[entry] = label{splitURL, false}
	for _, url := range splits {
		if err := parser.add(splitURL, SplitURL, url); err != nil {
			return err
		}
	}
	return nil
}
