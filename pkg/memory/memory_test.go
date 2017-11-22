package memory

import (
	"fmt"
	"testing"

	"bitbucket.org/fflo/semix/pkg/semix"
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

func TestMemory(t *testing.T) {
	tests := []struct {
		c, e, n int
		urls    []string
	}{
		{0, 0, 0, []string{}},
		{1, 3, 3, []string{"A", "B", "C"}},
		{2, 2, 3, []string{"A", "A", "C"}},
		{3, 1, 3, []string{"A", "A", "A"}},
		{3, 3, 5, []string{"A", "B", "A", "C", "A"}},
		{3, 3, 5, []string{"A", "B", "A", "C", "A", "A"}},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%d", tc.n), func(t *testing.T) {
			m := NewMemory(5)
			for _, url := range tc.urls {
				m.Push(semix.NewConcept(url))
			}
			if got := m.N(); got != tc.n {
				t.Fatalf("expected %d; got %d", tc.n, got)
			}
			if got := m.Count("A"); got != tc.c {
				t.Fatalf("expected %d; got %d", tc.c, got)
			}
			if got := len(m.Elements()); got != tc.e {
				t.Fatalf("expected %d; got %d", tc.e, got)
			}
		})
	}
}
