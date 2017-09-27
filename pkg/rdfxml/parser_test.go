package rdfxml

import (
	"bytes"
	"testing"

	"bitbucket.org/fflo/semix/pkg/semix"
)

var content = `
<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
    xmlns:skos="http://example.org/"
	xmlns:rdfs="http://www.w3.org/2000/01/rdf-schema#">
	<!-- Lizensed TopicZoom Ontologie Version indexer20120229 -->
	<rdfs:Class rdf:about="http://rdf.internal.topiczoom.de/#Topic">
		<rdfs:label>Topic</rdfs:label>
		<rdfs:comment>A rdfs Class used for topics.</rdfs:comment>
	</rdfs:Class>
	<skos:Concept rdf:about="http://example.org/#id1">
		<skos:prefLabel>TOPNODE</skos:prefLabel>
		<skos:altLabel>Topnode</skos:altLabel>
		<skos:narrower rdf:resource="http://example.org/#id2"/>
		<skos:narrower rdf:resource="http://example.org/#id3"/>
	</skos:Concept>
	<skos:Concept rdf:about="http://example.org/#id2">
		<skos:prefLabel>string for id 2</skos:prefLabel>
		<skos:altLabel>string for id 2</skos:altLabel>
		<skos:altLabel>ambigous for id 2 and 4</skos:altLabel>
		<skos:broader rdf:resource="http://example.org/#id1"/>
		<skos:narrower rdf:resource="http://example.org/#id4"/>
	</skos:Concept>
	<skos:Concept rdf:about="http://example.org/#id3">
		<skos:prefLabel>string for id 3</skos:prefLabel>
		<skos:altLabel>other string for id 3</skos:altLabel>
		<skos:broader rdf:resource="http://example.org/#id1"/>
	</skos:Concept>
	<skos:Concept rdf:about="http://example.org/#id4">
		<skos:prefLabel>string for id 4</skos:prefLabel>
		<skos:altLabel>ambigous for id 2 and 4</skos:altLabel>
		<skos:broader rdf:resource="http://example.org/#id4"/>
	</skos:Concept>
</rdf:RDF>
`

const (
	tid1     = "http://example.org/#id1"
	tid2     = "http://example.org/#id2"
	tid3     = "http://example.org/#id3"
	tid4     = "http://example.org/#id4"
	tid2and4 = "http://example.org/#id4+#id2"
	broader  = "http://example.org/broader"
	narrower = "http://example.org/narrower"
	split    = "http://example.org/split"
)

func makeTestParser(t *testing.T) *Parser {
	t.Helper()
	p := NewParser(
		WithTransitiveURLs(
			"http://example.org/broader",
			"http://example.org/narrower"),
		WithSplitRelationURL("http://example.org/split"))
	if err := p.Parse(bytes.NewBufferString(content)); err != nil {
		t.Fatalf("parsing error: %v", err)
	}
	return p
}

func TestDictionary(t *testing.T) {
	_, d := makeTestParser(t).Get()
	for _, tc := range []struct {
		test, want string
	}{
		{" TOPNODE ", tid1},
		{" Topnode ", tid1},
		{" string for id 2 ", tid2},
		{" string for id 3 ", tid3},
		{" other string for id 3 ", tid3},
		{" string for id 4 ", tid4},
		{" ambigous for id 2 and 4 ", tid2and4},
	} {
		t.Run(tc.test, func(t *testing.T) {
			if d[tc.test] == nil {
				t.Fatalf("Could not find %q", tc.test)
			}
			if url := d[tc.test].URL(); url != tc.want {
				t.Fatalf("expected url %s; got %s", tc.want, url)
			}
		})
	}
}

func TestGraphConcepts(t *testing.T) {
	g, _ := makeTestParser(t).Get()
	for _, tc := range []string{tid1, tid2, tid3, tid4, tid2and4, broader, narrower, split} {
		t.Run(tc, func(t *testing.T) {
			if _, ok := g.FindByURL(tc); !ok {
				t.Fatalf("could not find concept %s", tc)
			}
			c, _ := g.FindByURL(tc)
			if url := c.URL(); url != tc {
				t.Fatalf("expected URL = %s; got %s", tc, url)
			}

			if tmp, _ := g.FindById(c.ID()); tmp.URL() != tc {
				t.Fatalf("expected URL = %s; got %s", tc, tmp.URL())
			}
		})
	}
}

func containsLink(c *semix.Concept, p, o string) bool {
	var found bool
	c.Edges(func(edge semix.Edge) {
		if edge.P.URL() == p && edge.O.URL() == o {
			found = true
		}
	})
	return found
}

func TestGraphLinks(t *testing.T) {
	g, _ := makeTestParser(t).Get()
	for _, tc := range []struct {
		test  string
		links []string
	}{} {
		t.Run(tc.test, func(t *testing.T) {
			c, ok := g.FindByURL(tc.test)
			if !ok || c == nil {
				t.Fatalf("could not find concept %s", tc.test)
			}
			for i := 1; i < len(tc.links); i++ {
				if !containsLink(c, tc.links[0], tc.links[i]) {
					t.Fatalf("%s does not contain {%s %s}",
						c.URL(), tc.links[0], tc.links[i])
				}
			}
		})
	}
}
