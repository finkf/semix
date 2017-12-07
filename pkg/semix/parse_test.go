package semix

import (
	"testing"
)

func TestParse(t *testing.T) {
	parser := makeNewTestParser()
	r, err := Parse(parser, testTraits{})
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	// dictionray names are normalized
	for _, name := range []string{"name", "distinct", "ambiguous", "abd", "acd",
		"split name", "second split name"} {
		if _, ok := r.Dictionary[name]; !ok {
			t.Fatalf("could not find %q in dictionary", name)
		}
	}
	// test `normal` concepts
	for _, url := range []string{"A", "B", "C", "AS", "BS", "AT", "BT", "CT",
		"http://example.org/A", "http://example.org/B", "http://example.org/C",
	} {
		c, ok := r.Graph.FindByURL(url)
		if !ok {
			t.Fatalf("could not find concept %s", url)
		}
		if tmp := c.URL(); tmp != url {
			t.Fatalf("expected url=%s; got %s", url, tmp)
		}
		if tmp := c.Ambig(); tmp != false {
			t.Fatalf("expected ambiguous = false; got %t", tmp)
		}
	}
	// test `ambiguous` concepts
	for _, url := range []string{"A-B", "http://example.org/A-B-C"} {
		c, ok := r.Graph.FindByURL(url)
		if !ok {
			t.Fatalf("could not find concept %s", url)
		}
		if tmp := c.URL(); tmp != url {
			t.Fatalf("expected url=%s; got %s", url, tmp)
		}
		if tmp := c.Ambig(); tmp != true {
			t.Fatalf("expected ambiguous = true; got %t", tmp)
		}
	}
	for _, url := range []string{"X", "Y", "Z"} {
		if _, ok := r.Graph.FindByURL(url); ok {
			t.Fatalf("found something for url=%s", url)
		}
	}
	if c, _ := r.Graph.FindByURL("A"); c.Name != "name" {
		t.Fatalf("expected name=%s; got %s", "name", c.Name)
	}
	if got, ok := r.Rules["R"]; !ok || got != "rule" {
		t.Fatalf("expected rule=%s; got %s", "rule", got)
	}
	a, _ := r.Graph.FindByURL("A")
	edgesExist(t, a, "p", "B")
	as, _ := r.Graph.FindByURL("AS")
	edgesExist(t, as, "s", "BS")
	bs, _ := r.Graph.FindByURL("BS")
	edgesExist(t, bs, "s", "AS")
	at, _ := r.Graph.FindByURL("AT")
	edgesExist(t, at, "t", "BT", "t", "CT")
	bv, _ := r.Graph.FindByURL("BV")
	edgesExist(t, bv, "v", "AV")
}

func edgesExist(t *testing.T, c *Concept, urls ...string) {
	t.Helper()
	for i := 0; i < len(urls); i += 2 {
		var found bool
		for _, edge := range c.edges {
			if edge.P.URL() == urls[i] && edge.O.URL() == urls[i+1] {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("could not find edge {%s %s}", urls[i], urls[i+1])
		}
	}
}

func newTestParser(ts ...string) Parser {
	return testParser{ts: ts}
}

type testParser struct {
	ts []string
}

func (p testParser) Parse(f func(string, string, string) error) error {
	for i := 0; i < len(p.ts); i += 3 {
		if err := f(p.ts[i], p.ts[i+1], p.ts[i+2]); err != nil {
			return err
		}
	}
	return nil
}

func makeNewTestParser() Parser {
	return newTestParser(
		"A", "p", "B", // normal
		"B", "p", "C", // normal
		"X", "i", "X", // ignore
		"A", "n", "name", // name
		"A", "d", "distinct", // distinct label
		"A", "d", "a{b,c}d", // distinct label
		"A", "a", "ambiguous", // ambiguous label
		"AS", "s", "BS", // symmetric
		"AT", "t", "BT", // transitive
		"BT", "t", "CT", // transitive
		"AV", "v", "BV", // inverted
		"A", "d", "split-name", // split
		"B", "d", "split-name", // split
		"R", "r", "rule", // rule
		"http://example.org/A", "d", "second-split-name", // split
		"http://example.org/B", "d", "second-split-name", // split
		"http://example.org/C", "d", "second-split-name", // split
	)
}

type testTraits struct{}

func (testTraits) Ignore(p string) bool       { return p == "i" }
func (testTraits) IsName(p string) bool       { return p == "n" }
func (testTraits) IsDistinct(p string) bool   { return p == "d" }
func (testTraits) IsAmbig(p string) bool      { return p == "a" }
func (testTraits) IsSymmetric(p string) bool  { return p == "s" }
func (testTraits) IsTransitive(p string) bool { return p == "t" }
func (testTraits) IsInverted(p string) bool   { return p == "v" }
func (testTraits) IsRule(p string) bool       { return p == "r" }
func (testTraits) SplitAmbiguousURLs() bool   { return true }
