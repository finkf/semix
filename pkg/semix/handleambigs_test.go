package semix

import (
	"testing"
)

func TestHandleAmbigsWithSplit(t *testing.T) {
	tests := []struct {
		test string
		urls []string
	}{
		{"A-B-C", []string{"A", "B", "C"}},
		{"A-B-C", []string{"A", "B", "C", "A"}},
		{"A-B-C", []string{"C", "B", "A", "A"}},
		{"A-B", []string{"B", "A", "A"}},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			g := NewGraph()
			c, err := HandleAmbigsWithSplit(g, tc.urls...)
			if err != nil {
				t.Fatalf("got %s", err)
			}
			if c.URL() != tc.test {
				t.Fatalf("expected %s; got %s", tc.test, c.URL())
			}
			if !c.Ambig() {
				t.Fatalf("concept is not ambiguous")
			}
			for _, url := range tc.urls {
				if _, ok := c.FindEdge(SplitURL, url); !ok {
					t.Fatalf("cannot find Edge{%s,%s}", SplitURL, url)
				}
			}
		})
	}
}

func TestHandleAmbigsWithMerge(t *testing.T) {
	tests := []struct {
		test        string
		urls, edges []string
	}{
		{"A1-A2", []string{"A1", "A2"}, []string{"PA", "a"}},
		{"A1-B1", []string{"A1", "B1"}, []string{"PA", "a"}},
		{"A1-C1", []string{"A1", "C1"}, []string{"PA", "a"}},
		{"B1-B2", []string{"B1", "B2"}, []string{"PB", "b", "PA", "a"}},
		{"A1-B1-C1", []string{"A1", "B1", "C1"}, []string{"PA", "a"}},
		{"A1-C2", []string{"A1", "C2"}, []string{}},
	}
	ng := func() *Graph {
		g := NewGraph()
		g.Add("A1", "PA", "a")
		g.Add("A2", "PA", "a")
		g.Add("B1", "PA", "a")
		g.Add("B1", "PB", "b")
		g.Add("B2", "PA", "a")
		g.Add("B2", "PB", "b")
		g.Add("C1", "PA", "a")
		g.Add("C2", "PB", "b")
		return g
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			g := ng()
			c, err := HandleAmbigsWithMerge(g, tc.urls...)
			if err != nil {
				t.Fatalf("got %s", err)
			}
			if c.URL() != tc.test {
				t.Fatalf("expected %s; got %s", tc.test, c.URL())
			}
			if c.Ambig() {
				t.Fatalf("concept is ambiguous")
			}
			if len(c.edges) != len(tc.edges)/2 {
				t.Fatalf("expected %d edges; got %d",
					len(tc.edges)/2, len(c.edges))
			}
			for i := 1; i < len(tc.edges); i += 2 {
				if _, ok := c.FindEdge(tc.edges[i-1], tc.edges[i]); !ok {
					t.Fatalf("cannot find Edge{%s,%s}",
						tc.edges[i-1], tc.edges[i])
				}
			}
		})
	}
}
