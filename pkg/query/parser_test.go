package query

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	tests := []struct {
		query, want string
		iserr       bool
	}{
		{"?(<{A,B,C})", "? [< [SET [A B C]]]", false},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			p := NewParser(tc.query)
			q, err := p.Parse()
			if tc.iserr && err == nil {
				t.Fatalf("expected an error")
			} else if !tc.iserr && err != nil {
				t.Fatalf("got an error: %v", err)
			}
			if str := fmt.Sprintf("%v", q); str != tc.want {
				t.Fatalf("expected %q; got %q", tc.want, str)
			}
		})
	}
}
