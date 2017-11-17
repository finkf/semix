// +build isize1 isize2 !isize1,!isize2,!isize3,!isize4

package index

import (
	"fmt"
	"testing"
)

func TestRelationID(t *testing.T) {
	tests := []struct {
		L, ID int
		A, D  bool
	}{
		{1, 42, false, false},
		{2, 42, false, false},
		{3, 42, false, false},
		{4, 42, false, false},
		{5, 42, false, false},
		{1, 42, false, true},
		{2, 42, false, true},
		{3, 42, false, true},
		{4, 42, false, true},
		{5, 42, false, true},
		{1, 42, true, false},
		{2, 42, true, false},
		{3, 42, true, false},
		{4, 42, true, false},
		{5, 42, true, false},
		{1, 42, true, true},
		{2, 42, true, true},
		{3, 42, true, true},
		{4, 42, true, true},
		{5, 42, true, true},
		{12, 1, true, false},
		{12, 1, false, false},
		{12, 1, true, true},
		{12, 1, false, true},
		{27, 1, true, false},
		{27, 1, false, false},
		{27, 1, true, true},
		{27, 1, false, true},
		{42, 0xffffff, true, false},
		{42, 0xffffff, false, false},
		{42, 0xffffff, true, true},
		{42, 0xffffff, false, true},
		{0x3f, 1981, true, false},
		{0x3f, 1981, false, false},
		{0x3f, 1981, true, true},
		{0x3f, 1981, false, true},
		{0x3f, 0, true, false},
		{0x3f, 0, false, false},
		{0x3f, 0, true, true},
		{0x3f, 0, false, true},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			id := newRelationID(tc.ID, tc.L, tc.A, tc.D)
			if got := id.ID(); got != tc.ID {
				t.Errorf("expected %d; got %d", tc.ID, got)
			}
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
