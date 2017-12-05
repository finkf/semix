package resource

import "testing"

func TestConfig(t *testing.T) {
	c, err := Read("testdata/test.toml")
	if err != nil {
		t.Fatalf("got error: %s", err)
	}
	if got := c.File.Path; got != "testdata/test.toml" {
		t.Fatalf("invalid config file path: %s", got)
	}
	if got := c.File.Type; got != "TESTTYPE" {
		t.Fatalf("invalid config file type: %s", got)
	}
	if got := c.File.Cache; got != "/tmp/test.cache" {
		t.Fatalf("invalid config file cache: %s", got)
	}
	traits := c.Traits()
	if !traits.IsTransitive("http://example.org/transitive") {
		t.Fatalf("missing transitive predicate")
	}
	if !traits.IsSymmetric("http://example.org/symmetric") {
		t.Fatalf("missing symmetric predicate")
	}
	if !traits.Ignore("http://example.org/ignore") {
		t.Fatalf("missing ignore predicate")
	}
	if !traits.IsAmbiguous("http://example.org/ambiguous") {
		t.Fatalf("missing ambiguous predicate")
	}
	if !traits.IsName("http://example.org/name") {
		t.Fatalf("missing name predicate")
	}
	if !traits.IsDistinct("http://example.org/distinct") {
		t.Fatalf("missing distinct predicate")
	}
	if !traits.IsRule("http://example.org/rule") {
		t.Fatalf("missing rule predicate")
	}
	if traits.SplitAmbiguousURLs() {
		t.Fatalf("should merge ambiguous URLs")
	}
}
