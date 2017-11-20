package rule

import (
	"fmt"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	tests := []struct {
		test, want string
		iserr      bool
	}{
		{`{"abc", "def"}`, `{"abc","def"}`, false},
		{`{}`, `{}`, false},
		{`{"abc"}`, `{"abc"}`, false},
		{`{"abc" "def"}`, ``, true},
		{`{"abc\" "def"}`, ``, true},
		{`{1,2}`, ``, true},
		{`??`, ``, true},
		{"4.2", "4.200000", false},
		{"4a.2", "", true},
		{"10", "10.000000", false},
		{"100a", "", true},
		{"- 10", "-10.000000", false},
		{"! 10", "!10.000000", false},
		{"true", "true", false},
		{"false", "false", false},
		{"! false", "!false", false},
		{"! true", "!true", false},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			p := newParser(strings.NewReader(tc.test))
			ast, err := p.parse()
			if tc.iserr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			if got := fmt.Sprintf("%s", ast); got != tc.want {
				t.Fatalf("expected %s; got %s", tc.want, got)
			}
		})
	}
}
