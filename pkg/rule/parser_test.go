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
		{`{"abc",}`, `{"abc"}`, false},
		{`{"abc" "def"}`, ``, true},
		{`{"abc\" "def"}`, ``, true},
		{`{1,2}`, ``, true},
		{`??`, ``, true},
		{"4.2", "4.20", false},
		{"4a.2", "", true},
		{"10", "10.00", false},
		{"100a", "", true},
		{"-10", "(-10.00)", false},
		{"-10", "(-10.00)", false},
		{"true", "true", false},
		{"false", "false", false},
		{"-false", "(-false)", false},
		{"-true", "(-true)", false},
		{"tue", "", true},
		{"flse", "", true},
		{"-10 = 10.0", "((-10.00)=10.00)", false},
		{"-10 < 10.0", "((-10.00)<10.00)", false},
		{"-10 > 10.0", "((-10.00)>10.00)", false},
		{"-10 + - 10.0", "((-10.00)+(-10.00))", false},
		{"-10 - - 10.0", "((-10.00)-(-10.00))", false},
		{"-10 * - 10.0", "((-10.00)*(-10.00))", false},
		{"-10 / - 10.0", "((-10.00)/(-10.00))", false},
		{"2-3/3+1", "((2.00-(3.00/3.00))+1.00)", false},
		{"1+2*3", "(1.00+(2.00*3.00))", false},
		{"1/2-3", "((1.00/2.00)-3.00)", false},
		{`{"a","b"}-{"b"}`, `({"a","b"}-{"b"})`, false},
		{"(1+2)*3", "((1.00+2.00)*3.00)", false},
		{"1/(2-3)", "(1.00/(2.00-3.00))", false},
		{"(1+2)*-3", "((1.00+2.00)*(-3.00))", false},
		{"max(1,2,3)", "max(1.00,2.00,3.00)", false},
		{"max(1,2,)", "max(1.00,2.00)", false},
		{"min()", "min()", false},
		{"max(1,2,3", "", true},
		{`max({"abc","def"},1/2-3)`, `max({"abc","def"},((1.00/2.00)-3.00))`, false},
		{`min(len({"a","b"}),cs("foo"))`, `min(len({"a","b"}),cs("foo"))`, false},
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
