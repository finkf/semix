package rdfxml

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"bitbucket.org/fflo/semix/pkg/semix"

	"github.com/pkg/errors"
)

// Concepts denotes a list of Concepts.
type Concepts struct {
	Concepts []Concept `xml:"Concept"`
}

// Concept denotes a RDF-XML Concept.
type Concept struct {
	About     string   `xml:"about,attr"`
	PrefLabel string   `xml:"prefLabel"`
	AltLabel  []string `xml:"altLabel"`
	Links     []Link   `xml:",any"`
}

// Link denotes a link to another Concept.
type Link struct {
	XMLName xml.Name
	Object  string `xml:"resource,attr"`
}

// ParserOpt function to set Parser options.
type ParserOpt func(*Parser)

// WithSplitRelationURL sets the URL of the split relation.
// Set this to the empty string to disable splitting of multiple defined strings.
func WithSplitRelationURL(url string) ParserOpt {
	return func(p *Parser) {
		p.traits.splitRelationURL = url
	}
}

// WithTransitiveURLs sets the URLs of the transitive relations.
func WithTransitiveURLs(urls ...string) ParserOpt {
	return func(p *Parser) {
		for _, url := range urls {
			p.traits.transitive[url] = true
		}
	}
}

// WithSymmetricURLs sets the URLs of the symmetric relations.
func WithSymmetricURLs(urls ...string) ParserOpt {
	return func(p *Parser) {
		for _, url := range urls {
			p.traits.symmetric[url] = true
		}
	}
}

// WithIgnoreURLs specifies URLs of relations that should be ignored.
func WithIgnoreURLs(urls ...string) ParserOpt {
	return func(p *Parser) {
		for _, url := range urls {
			p.traits.ignore[url] = true
		}
	}
}

// Parser is parses a RDF-XML file.
type Parser struct {
	dictionary map[string]string
	relations  map[string]map[triple]bool
	names      map[string]string
	traits     traits
}

// NewParser constructs a new Parser.
func NewParser(args ...ParserOpt) *Parser {
	p := &Parser{
		dictionary: make(map[string]string),
		relations:  make(map[string]map[triple]bool),
		names:      make(map[string]string),
		traits:     newTraits(),
	}
	for _, arg := range args {
		arg(p)
	}
	return p
}

// ParseFile parses a RDF-XML file.
func (p *Parser) ParseFile(path string) error {
	is, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not open file %q", path))
	}
	defer is.Close()
	if err := p.Parse(is); err != nil {
		return errors.Wrap(err, fmt.Sprintf("error reading file: %s", path))
	}
	return nil
}

// Parse parses an io.Reader.
func (p *Parser) Parse(r io.Reader) error {
	d := xml.NewDecoder(r)
	var concepts Concepts
	if err := d.Decode(&concepts); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not decode xml file"))
	}
	for _, c := range concepts.Concepts {
		if err := p.add(c); err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not add concept %s", c.About))
		}
	}
	return nil
}

// Get build the graph and the dictionary from the parsed data.
// Get should be called after Parse().
func (p *Parser) Get() (*semix.Graph, map[string]*semix.Concept) {
	g := semix.NewGraph()
	for r, ts := range p.relations {
		if p.traits.ignoreURL(r) {
			continue
		}
		if p.traits.isSymmetricURL(r) {
			p.relations[r] = calculateSymmetricClosure(ts)
		}
		if p.traits.isTransitiveURL(r) {
			p.relations[r] = calculateTransitiveClosure(ts)
		}
		// insert triples into graph and set names of URL's.
		for t := range p.relations[r] {
			s, _, o := g.Add(t.s, t.p, t.o)
			if s.Name == "" {
				if name, ok := p.names[t.s]; ok {
					s.Name = name
				}
			}
			if o.Name == "" {
				if name, ok := p.names[t.o]; ok {
					o.Name = name
				}
			}
		}
	}
	d := make(map[string]*semix.Concept, len(p.dictionary))
	for str, url := range p.dictionary {
		if c, ok := g.FindByURL(url); ok {
			d[str] = c
		}
	}
	return g, d
}

func (p *Parser) add(c Concept) error {
	s := c.About
	for _, l := range c.Links {
		pred := l.XMLName.Space + l.XMLName.Local
		o := l.Object
		t := triple{s: s, p: pred, o: o}
		if err := p.addTriple(t); err != nil {
			return nil
		}
	}
	if err := p.addLabel(c.PrefLabel, s); err != nil {
		return err
	}
	for _, l := range c.AltLabel {
		if err := p.addLabel(l, s); err != nil {
			return err
		}
	}
	// add the prefLabel as name for this subject.
	p.names[s] = c.PrefLabel
	return nil
}

func (p *Parser) addLabel(l, url string) error {
	if l == "" {
		return fmt.Errorf("invalid empty string for url: %s", url)
	}
	l = " " + l + " "
	if uurl, ok := p.dictionary[l]; ok && url != uurl {
		if p.traits.splitRelationURL == "" {
			return fmt.Errorf("multiple strings for: %q", l)
		}
		spliturl := p.split(url, uurl)
		// log.Printf("split: {%q %q}", spliturl, p.traits.splitURL)
		p.dictionary[l] = spliturl
		// log.Printf("adding triple: %v", triple{s: spliturl, p: p.traits.splitURL, o: url})
		if err := p.addTriple(triple{s: spliturl, p: p.traits.splitRelationURL, o: url}); err != nil {
			return err
		}
		// log.Printf("adding triple: %v", triple{s: spliturl, p: p.traits.splitURL, o: uurl})
		return p.addTriple(triple{s: spliturl, p: p.traits.splitRelationURL, o: uurl})
	}
	p.dictionary[l] = url
	return nil
}

func (p *Parser) addTriple(t triple) error {
	if t.s == "" || t.p == "" || t.o == "" {
		return fmt.Errorf("invalid triple: %v", t)
	}
	if _, ok := p.relations[t.p]; !ok {
		p.relations[t.p] = make(map[triple]bool)
	}
	p.relations[t.p][t] = true
	return nil
}

func (p *Parser) split(a, b string) string {
	ai := strings.LastIndex(a, "/")
	bi := strings.LastIndex(b, "/")
	if ai == -1 || bi == -1 || ai != bi || a[:ai] != b[:bi] {
		return a + "+" + b
	}
	return a[:ai+1] + a[ai+1:] + "+" + b[bi+1:]
}

type triple struct {
	s, p, o string
}
