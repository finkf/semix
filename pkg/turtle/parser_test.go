package turtle

import (
	"errors"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	tests := []struct {
		test, want string
		iserr      bool
	}{
		{"A B C.", "(A B C)", false},
		{"<A> B C.", "(A B C)", false},
		{"A B C. #comment\n", "(A B C)", false},
		{"#comment\nA B C.", "(A B C)", false},
		{"#A B C.\n", "", false},
		{"#comment", "", true},
		{"A B .", "", true},
		{"A B C. D E F.", "(A B C)(D E F)", false},
		{"A B C; D E.", "(A B C)(A D E)", false},
		{"A B C; D .", "", true},
		{"A B C, D, E.", "(A B C)(A B D)(A B E)", false},
		{"A B C; D, .", "", true},
		{",A B C.", "", true},
		{"@prefix x: abc.x:A x:B x:C.", "(abcA abcB abcC)", false},
		{"@prefix x: abc x:A x:B x:C.", "", true},
		{"PREFIX x: abc x:A x:B x:C.", "(abcA abcB abcC)", false},
		{"PREFIX x: abc. x:A x:B x:C.", "", true},
		{"@base <xyz>. A B C.", "(xyzA xyzB xyzC)", false},
		{"@base <xyz> A B C.", "", true},
		{"BASE <xyz> A B C.", "(xyzA xyzB xyzC)", false},
		{"BASE <xyz> . A B C.", "", true},
		{`"#p:x" http://B <#C>.`, "(#p:x http://B #C)", false},
		{"p:A B C.", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			var got string
			err := NewParser(strings.NewReader(tc.test)).Parse(func(s, p, o string) error {
				got += "(" + s + " " + p + " " + o + ")"
				return nil
			})
			if !tc.iserr && err != nil {
				t.Fatalf("got error: %v", err)
			}
			if tc.iserr && err == nil {
				t.Fatalf("expected an error")
			}
			if !tc.iserr && tc.want != got {
				t.Fatalf("expected %s; got %s", tc.want, got)
			}
			err = NewParser(strings.NewReader(tc.test)).Parse(func(s, p, o string) error {
				return errors.New("error")
			})
			if err == nil && got != "" { // if f is not called, no error will be returned.
				t.Fatalf("expected an error")
			}
		})
	}
}
