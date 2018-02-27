package rdfxml

import (
	"encoding/xml"
	"fmt"
	"io"
)

// Concepts denotes a list of Concepts.
type Concepts struct {
	Concepts []Concept `xml:"Concept"`
}

// Concept denotes a RDF-XML Concept.
type Concept struct {
	About     string  `xml:"about,attr"`
	PrefLabel Label   `xml:"prefLabel"`
	AltLabels []Label `xml:"altLabel"`
	Links     []Link  `xml:",any"`
}

// Label denotes label declarations like:
// <skos:prefLabel>lable</skos:prefLabel>
type Label struct {
	XMLName xml.Name
	Label   string `xml:",chardata"`
}

// Link denotes a link to another concept like:
// <skos:narrower rdf:ressource="http://example.org"/>
type Link struct {
	XMLName xml.Name
	Object  string `xml:"resource,attr"`
}

// Parser parses a RDF-XML file.
type Parser struct {
	r io.Reader
}

// NewParser constructs a new parser instastance with a given reader.
func NewParser(r io.Reader) *Parser {
	return &Parser{r: r}
}

// Parse parses a RDF-XML file and calls the callback function for
// each triple in the knowledge base.
func (p *Parser) Parse(f func(string, string, string) error) error {
	d := xml.NewDecoder(p.r)
	var concepts Concepts
	if err := d.Decode(&concepts); err != nil {
		return fmt.Errorf("cannot decode xml file: %v", err)
	}
	for _, c := range concepts.Concepts {
		if err := f(c.PrefLabel.triple(c.About)); err != nil {
			return err
		}
		for _, label := range c.AltLabels {
			if err := f(label.triple(c.About)); err != nil {
				return err
			}
		}
		for _, link := range c.Links {
			if err := f(link.triple(c.About)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l Label) triple(about string) (string, string, string) {
	return about, xmlNameToString(l.XMLName), l.Label
}

func (l Link) triple(about string) (string, string, string) {
	return about, xmlNameToString(l.XMLName), l.Object
}

func xmlNameToString(name xml.Name) string {
	return name.Space + name.Local
}
