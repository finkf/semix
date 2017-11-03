package semix

import "testing"

func TestParse(t *testing.T) {
	parser := newTestParser(
		"A", "p", "B", // normal
		"B", "p", "C", // normal
		"X", "i", "X", // ignore
		"name", "n", "A", // name
		"distinct", "d", "A", // distinct label
		"ambiguous", "a", "A", // ambiguous label
		"AS", "s", "BS", // symmetric
		"AT", "t", "BT", // transitive
		"BT", "t", "CT", // transitive
		"split-name", "d", "A", // split
		"split-name", "d", "B", // split
		"second-split-name", "d", "http://example.org/A", // split
		"second-split-name", "d", "http://example.org/B", // split
	)
	g, d, err := Parse(parser, testTraits{})
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	for _, name := range []string{"name", "distinct", "ambiguous", "split-name", "second-split-name"} {
		if _, ok := d[name]; !ok {
			t.Fatalf("could not find %s in dictionary", name)
		}
	}
	for _, url := range []string{
		"A", "B", "C", "AS", "BS", "AT", "BT", "CT", "A+B",
		"http://example.org/A", "http://example.org/B", "http://example.org/A+B",
	} {
		c, ok := g.FindByURL(url)
		if !ok {
			t.Fatalf("could not find concept %s", url)
		}
		if tmp := c.URL(); tmp != url {
			t.Fatalf("expected url=%s; got %s", url, tmp)
		}
	}
	for _, url := range []string{"X", "Y", "Z"} {
		if _, ok := g.FindByURL(url); ok {
			t.Fatalf("found something for url=%s", url)
		}
	}
	if c, _ := g.FindByURL("A"); c.Name != "name" {
		t.Fatalf("expected name=%s; got %s", "name", c.Name)
	}
	a, _ := g.FindByURL("A")
	edgesExist(t, a, "p", "B")
	as, _ := g.FindByURL("AS")
	edgesExist(t, as, "s", "BS")
	bs, _ := g.FindByURL("BS")
	edgesExist(t, bs, "s", "AS")
	at, _ := g.FindByURL("AT")
	edgesExist(t, at, "t", "BT", "t", "CT")
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

type testTraits struct{}

func (testTraits) Ignore(p string) bool       { return p == "i" }
func (testTraits) IsName(p string) bool       { return p == "n" }
func (testTraits) IsDistinct(p string) bool   { return p == "d" }
func (testTraits) IsAmbiguous(p string) bool  { return p == "a" }
func (testTraits) IsSymmetric(p string) bool  { return p == "s" }
func (testTraits) IsTransitive(p string) bool { return p == "t" }
