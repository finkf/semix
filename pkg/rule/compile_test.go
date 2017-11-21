package rule

import (
	"errors"
	"strings"
	"testing"
)

func checkSyntax(ast ast) (t astType, err error) {
	defer func() {
		if e, ok := recover().(astError); ok {
			err = errors.New(e.msg)
		}
	}()
	t = ast.check()
	return t, nil
}

func TestSyntaxCheck(t *testing.T) {
	tests := []struct {
		test  string
		want  astType
		iserr bool
	}{
		{"2", astNum, false},
		{"2+2", astNum, false},
		{"true+true", astBoolean, false},
		{"false+false", astBoolean, false},
		{"false*true", astBoolean, false},
		{"-false", astBoolean, false},
		{"-1", astNum, false},
		{"1>2", astBoolean, false},
		{"1<2", astBoolean, false},
		{"1=2", astBoolean, false},
		{"true=false", astBoolean, false},
		{`"ab"="cd"`, astBoolean, false},
		{`"ab"<"cd"`, astBoolean, false},
		{`"ab">"cd"`, astBoolean, false},
		{`{"a","b"}+es()`, astSet, false},
		{`min(len({"a","b"}),cs("foo"))`, astNum, false},
		{`len(es()-{"topnode"})`, astNum, false},
		{`len("abc")`, astNum, false},
		{"min(true,false,false,true)", astBoolean, false},
		{"2-true", 0, true},
		{"false+2", 0, true},
		{"false/true", 0, true},
		{"false=0", 0, true},
		{"false<true", 0, true},
		{"false>true", 0, true},
		{`min(len({"a","b"}),"foo")`, 0, true},
		{"min(1.0,false,false,true)", 0, true},
		{"min(true,false,1.0,true)", 0, true},
		{"es(true,false)", 0, true},
		{"max()", 0, true},
		{`max(1.0,"foo")`, 0, true},
		{`max("foo", "bar")`, 0, true},
		{"e(true)", 0, true},
		{"c()", 0, true},
		{`c({"abc"},1.0,true)`, 0, true},
		{`len("ab","foo")`, 0, true},
		{`len(1.0)`, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			ast, err := newParser(strings.NewReader(tc.test)).parse()
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			got, err := checkSyntax(ast)
			if tc.iserr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if got != tc.want {
				t.Fatalf("expected %d; got %d", tc.want, got)
			}
		})
	}
}

func TestCompile(t *testing.T) {
	tests := []struct {
		test  string
		want  float64
		iserr bool
	}{
		{"-2", -2, false},
		{"--2", 2, false},
		{"-true", 0, false},
		{"-false", 1, false},
		{"--true", 1, false},
		{"--false", 0, false},
		{"false=true", 0, false},
		{"true=true", 1, false},
		{"true=false", 0, false},
		{"false=false", 1, false},
		{`"abc"="abc"`, 1, false},
		{`"bc"="abc"`, 0, false},
		{`"abc"<"def"`, 1, false},
		{`"abc">"def"`, 0, false},
		{`"def">"abc"`, 1, false},
		{`"def"<"abc"`, 0, false},
		{"-{}", 0, true},
		{"-es()", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			rule, err := Compile(tc.test)
			if tc.iserr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			// t.Logf("rule = %s", rule)
			if got := rule.Execute(); got != tc.want {
				t.Fatalf("expected %f; got %f", tc.want, got)
			}
		})
	}
}
