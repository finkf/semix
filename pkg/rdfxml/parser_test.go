package rdfxml

import (
	"strings"
	"testing"
)

const content = `
<rdf:RDF xmlns:rdf="rdf"
	xmlns:rdfs="r:"
	xmlns:skos="s:">
	<rdfs:Class rdf:about="A">
		<rdfs:comment>A rdfs Class used for topics.</rdfs:comment>
		<rdfs:prefLabel>a</rdfs:prefLabel>
	</rdfs:Class>
	<skos:Concept rdf:about="T">
		<skos:prefLabel>T</skos:prefLabel>
		<skos:altLabel>t</skos:altLabel>
		<skos:narrower rdf:resource="ID1"/>
		<skos:narrower rdf:resource="ID2"/>
	</skos:Concept>
	<skos:Concept rdf:about="ID1">
		<skos:prefLabel>id1</skos:prefLabel>
		<skos:broader rdf:resource="T"/>
	</skos:Concept>
	<skos:Concept rdf:about="ID2">
		<skos:prefLabel>id2</skos:prefLabel>
		<skos:broader rdf:resource="T"/>
	</skos:Concept>
</rdf:RDF>
`

var want = []string{
	"(T s:prefLabel T)",
	"(T s:altLabel t)",
	"(T s:narrower ID1)",
	"(T s:narrower ID2)",
	"(ID1 s:prefLabel id1)",
	"(ID1 s:broader T)",
	"(ID2 s:prefLabel id2)",
	"(ID2 s:broader T)",
}

func TestRdfXMLParser(t *testing.T) {
	parser := NewParser(strings.NewReader(content))
	got := make(map[string]bool)
	err := parser.Parse(func(s, p, o string) error {
		got["("+s+" "+p+" "+o+")"] = true
		return nil
	})
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d entries; got %d", len(want), len(got))
	}
	for _, w := range want {
		if !got[w] {
			t.Fatalf("could not find %s", w)
		}
	}
	parser = NewParser(strings.NewReader("<invalid>xml</"))
	err = parser.Parse(func(s, p, o string) error {
		return nil
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}
