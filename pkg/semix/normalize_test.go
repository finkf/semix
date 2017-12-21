package semix

import "testing"

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		test, want string
		sourround  bool
	}{
		{"a,b,c", "a b c", false},
		{"a,b,c", " a b c ", true},
		{"a, b, c", "a b c", false},
		{"a, b, c", " a b c ", true},
		{" a,b,\u00a0c ", " a b c ", true},
		{"(abc)", "abc", false},
		{" abc ", "abc", false},
		{" (abc) ", "abc", false},
		{" (abc) ", " abc ", true},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			got := NormalizeString(tc.test, tc.sourround)
			if got != tc.want {
				t.Fatalf("expected %q; got %q", tc.want, got)
			}
		})
	}
}
