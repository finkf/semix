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
		{"?(A, B({C, D}))", "?(A,B({C,D}))", false},
		{"?({C})", "?({C})", false},
		{"?(!A, B({C, D}))", "?(!A,B({C,D}))", false},
		{"?(*({C, D}))", "?(*({C,D}))", false},
		{"?(!*({C, D}))", "?(!*({C,D}))", false},
		{"?10(!*({C, D}))", "?10(!*({C,D}))", false},
		{"?*(!*({C, D}))", "?*(!*({C,D}))", false},
		{"?10*(*({C, D}))", "?*10(*({C,D}))", false},
		{"?*10(*({C, D}))", "?*10(*({C,D}))", false},
		{`?("A"({"B","C"}))`, `?(A({B,C}))`, false},
		{`?("A B"({"C D","E F"}))`, `?(A B({C D,E F}))`, false},
		{"", "", true},
		{"?(", "", true},
		{"?({}({}", "", true},
		{"?{}({})", "", true},
		{"?({'A, B}({C, D}))", "", true},
		{"?({'A, 10, B}({C, D}))", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			p := NewParser(tc.query)
			q, err := p.Parse()
			if tc.iserr {
				if err == nil {
					t.Fatalf("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("got an error: %v", err)
			}
			if str := fmt.Sprintf("%v", q); str != tc.want {
				t.Fatalf("expected %q; got %q", tc.want, str)
			}
		})
	}
}
