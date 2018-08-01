package memory

import (
	"fmt"
	"testing"

	"gitlab.com/finkf/semix/pkg/semix"
)

func TestInc(t *testing.T) {
	tests := []struct {
		n  int
		is []uint
	}{
		{1, []uint{0, 0, 0, 0, 0}},
		{5, []uint{0, 1, 2, 3, 4, 0, 1, 2, 3, 4}},
		{10, []uint{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1}},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", tc.n), func(t *testing.T) {
			i := uint(0)
			for x := range tc.is {
				if i != tc.is[x] {
					t.Fatalf("(%d) expected %d; got %d", tc.n, tc.is[x], i)
				}
				i = inc(i, tc.n)
			}
		})
	}
}

func equalsURL(url string) func(*semix.Concept) bool {
	return func(c *semix.Concept) bool {
		return c.URL() == url
	}
}

func pushURLs(mem *Memory, urls []string) {
	d := semix.NewConcept("D")
	p := semix.NewConcept("P")
	for _, url := range urls {
		mem.Push(semix.NewConcept(url, semix.WithEdges(p, d)))
	}
}

func TestMemory(t *testing.T) {
	tests := []struct {
		c, e, es, n int
		urls        []string
	}{
		{0, 0, 0, 0, []string{}},
		{1, 3, 4, 3, []string{"A", "B", "C"}},
		{2, 2, 3, 3, []string{"A", "A", "C"}},
		{3, 1, 2, 3, []string{"A", "A", "A"}},
		{3, 3, 4, 5, []string{"A", "B", "A", "C", "A"}},
		{3, 3, 4, 5, []string{"A", "B", "A", "C", "A", "A"}},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", tc.n), func(t *testing.T) {
			m := New(5)
			if got := m.N(); got != 5 {
				t.Fatalf("expected %d; got %d", 5, got)
			}
			pushURLs(m, tc.urls)
			if got := m.Len(); got != tc.n {
				t.Fatalf("expected %d; got %d", tc.n, got)
			}
			if got := m.CountIf(equalsURL("A")); got != tc.c {
				t.Fatalf("expected %d; got %d", tc.c, got)
			}
			if got := m.CountIfS(equalsURL("D")); got != tc.n {
				t.Fatalf("expected %d; got %d", tc.n, got)
			}
			if got := m.CountIfS(equalsURL("A")); got != tc.c {
				t.Fatalf("expected %d; got %d", tc.c, got)
			}
			if got := len(m.Elements()); got != tc.e {
				t.Fatalf("expected %d; got %d", tc.e, got)
			}
			if got := len(m.ElementsS()); got != tc.es {
				t.Fatalf("expected %d; got %d", tc.es, got)
			}
		})
	}
}
