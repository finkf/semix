// +build isize3 isize4

package index

import (
	"fmt"
	"testing"
)

func TestRelationID(t *testing.T) {
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
		{0x3f, true, false},
		{0x3f, false, false},
		{0x3f, true, true},
		{0x3f, false, true},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			id := newRelationID(tc.L, tc.A, tc.D)
			if got := id.Distance(); got != tc.L {
				t.Errorf("expected %d; got %d", tc.L, got)
			}
			if got := id.Direct(); got != tc.D {
				t.Errorf("expected %t; got %t", tc.D, got)
			}
			if got := id.Ambiguous(); got != tc.A {
				t.Errorf("expected %t; got %t", tc.A, got)
			}
		})
	}
}
