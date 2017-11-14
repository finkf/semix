package index

import (
	"fmt"
	"testing"
)

func TestEncodeL(t *testing.T) {
	tests := []struct {
		L int
		A bool
	}{
		{1, false},
		{2, false},
		{3, false},
		{4, false},
		{5, false},
		{1, true},
		{2, true},
		{3, true},
		{4, true},
		{5, true},
		{12, true},
		{12, false},
		{27, true},
		{27, false},
		{100, true},
		{100, false},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			x := encodeL(tc.L, tc.A)
			l, a := decodeL(x)
			if l != tc.L {
				t.Errorf("expected %d; got %d", tc.L, l)
			}
			if a != tc.A {
				t.Errorf("expected %t; got %t", tc.A, a)
			}
		})
	}
}
