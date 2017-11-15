package index

import (
	"fmt"
	"testing"
)

func TestEncodeL(t *testing.T) {
	tests := []struct {
		L    int
		A, D bool
	}{
		{1, false, false},
		{2, false, false},
		{3, false, false},
		{4, false, false},
		{5, false, false},
		{1, false, true},
		{2, false, true},
		{3, false, true},
		{4, false, true},
		{5, false, true},
		{1, true, false},
		{2, true, false},
		{3, true, false},
		{4, true, false},
		{5, true, false},
		{1, true, true},
		{2, true, true},
		{3, true, true},
		{4, true, true},
		{5, true, true},
		{12, true, false},
		{12, false, false},
		{12, true, true},
		{12, false, true},
		{27, true, false},
		{27, false, false},
		{27, true, true},
		{27, false, true},
		{42, true, false},
		{42, false, false},
		{42, true, true},
		{42, false, true},
		{0x3f, true, false},
		{0x3f, false, false},
		{0x3f, true, true},
		{0x3f, false, true},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			x := encodeL(tc.L, tc.A, tc.D)
			l, a, d := decodeL(x)
			if l != tc.L {
				t.Errorf("expected %d; got %d", tc.L, l)
			}
			if a != tc.A {
				t.Errorf("expected %t; got %t", tc.A, a)
			}
			if d != tc.D {
				t.Errorf("expected %t; got %t", tc.D, d)
			}
		})
	}
}
