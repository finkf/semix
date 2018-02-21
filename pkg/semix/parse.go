package semix

import "github.com/pkg/errors"

// Parser defines a parser that parses (Subject, Predicate, Object) triples.
type Parser interface {
	Parse(func(string, string, string) error) error
}

// Parse creates a resource from a parser.
func Parse(p Parser, t Traits) (*Resource, error) {
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
	traits     Traits
}

func newParser(traits Traits) *parser {
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
	d, err := parser.buildDictionary(g)
	if err != nil {
		return nil, errors.Wrapf(err, "could not build dictionary")
	}
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

func (parser *parser) buildDictionary(g *Graph) (Dictionary, error) {
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
		h := parser.traits.HandleAmbigs()
		c, err := h(g, urls...)
		if err != nil {
			return nil, errors.Wrapf(err, "could not handle internal ambiguity %s", e)
		}
		if c == nil {
			continue
		}
		if c.Name == "" {
			c.Name = e
		}
		d[e] = c.ID()
	}
	return d, nil
}

func (parser *parser) add(s, p, o string) error {
	if parser.traits.Ignore(p) {
		return nil
	}
	if parser.traits.IsRule(p) {
		parser.rules[s] = o
		return nil
	}
	// Names will never be expanded or normalized and will not be
	// used as an entry in the lexikon.
	// It is possible to set the name predicate to be
	// a distinct or ambiguous lexicon entry.
	if parser.traits.IsName(p) {
		parser.names[s] = o
	}
	if parser.traits.IsAmbig(p) {
		return parser.addLabels(o, s, true)
	}
	if parser.traits.IsDistinct(p) {
		return parser.addLabels(o, s, false)
	}
	// If p is a name we are done.
	// We checked for lexicon entries with IsAmbig and IsDistinct.
	// but it will never be a normal predicate.
	if parser.traits.IsName(p) {
		return nil
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

func (parser *parser) addLabels(entry, url string, ambig bool) error {
	labels, err := ExpandBraces(entry)
	if err != nil {
		return errors.Wrapf(err, "could not expand %s", entry)
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
		parser.labels[normalized] = label{url, ambig}
	}
	return nil
}
