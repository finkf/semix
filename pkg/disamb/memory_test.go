package disamb

import (
	"fmt"
	"testing"
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
