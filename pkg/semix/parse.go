package semix

import (
	"fmt"

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
	ambigs     map[string][]string
	rules      RulesDictionary
	traits     traits.Interface
}

func newParser(traits traits.Interface) *parser {
	return &parser{
		predicates: make(map[string]map[spo]bool),
		names:      make(map[string]string),
		labels:     make(map[string]label),
		ambigs:     make(map[string][]string),
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
	for e, urls := range parser.ambigs {
		var c *Concept
		if parser.traits.SplitAmbiguousURLs() {
			c = HandleAmbigsWithSplit(g, e, urls...)
		} else {
			c = HandleAmbigsWithMerge(g, e, urls...)
		}
		if c == nil {
			continue
		}
		if c.Name == "" {
			c.Name = e
		}
		d[e] = c.ID()
	}
	return d
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
		if _, ok := parser.ambigs[normalized]; ok {
			parser.ambigs[normalized] = append(parser.ambigs[normalized], url)
			return nil
		}
		if l, ok := parser.labels[normalized]; ok && l.url != url {
			delete(parser.labels, normalized)
			parser.ambigs[normalized] = append(parser.ambigs[normalized], url)
			parser.ambigs[normalized] = append(parser.ambigs[normalized], l.url)
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
