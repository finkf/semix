package rule

import (
	"errors"
	"math"
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

func lookupID(str string) int {
	switch str {
	case "a":
		return 1
	case "b":
		return 2
	case "c":
		return 3
	}
	return -1
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
		{`min(len({"a","b"}),len("foo"))`, astNum, false},
		{`c("a","b")+cs("c","d")`, astSet, false},
		{"max()", astNum, false},
		{"min(1.0,false,false,true)", astNum, false},
		{"min(true,false,1.0,true)", astNum, false},
		{`len(es()-{"topnode"})`, astNum, false},
		{`len("abc")`, astNum, false},
		{"min(true,false,false,true)", astNum, false},
		{`cs({"a","b"})`, astSet, false},
		{"2-true", 0, true},
		{"false+2", 0, true},
		{"false/true", 0, true},
		{"false=0", 0, true},
		{"false<true", 0, true},
		{"false>true", 0, true},
		{`min(len({"a","b"}),"foo")`, 0, true},
		{"es(true,false)", 0, true},
		{`max(1.0,"foo")`, 0, true},
		{`max("foo", "bar")`, 0, true},
		{"e(true)", 0, true},
		{"c()", 0, true},
		{`c({"abc"},1.0,true)`, 0, true},
		{`len("ab","foo")`, 0, true},
		{`len(1.0)`, 0, true},
		{`LEN(1.0)`, 0, true},
		{`c(1.0)`, 0, true},
		{"log()", 0, true},
		{"exp(1.0,2.0,3.0)", 0, true},
		{"pow(1.0,{})", 0, true},
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
			if err != nil {
				t.Fatalf("go error: %s", err)
			}
			if got != tc.want {
				t.Fatalf("expected %d; got %d", tc.want, got)
			}
		})
	}
}

func TestCompileRule(t *testing.T) {
	tests := []struct {
		test, want string
	}{
		{"-2", "PUSH 2.00;NEG;"},
		{"min(1)", "PUSH 1.00;PUSH 1;MIN;"},
		{"max(1,2)", "PUSH 1.00;PUSH 2.00;PUSH 2;MAX;"},
		{"max({})", "PUSH 0;MAX;"},
		{"min(1+2,3-4)", "PUSH 1.00;PUSH 2.00;ADD;PUSH 3.00;PUSH 4.00;SUB;PUSH 2;MIN;"},
		{"pow(1,2)", "PUSH 1.00;PUSH 2.00;POW;"},
		{`min({"a","b","c"})`, "PUSH 1;PUSH 2;PUSH 3;PUSH 3;MIN;"},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			rule, err := Compile(tc.test, lookupID)
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			if got := rule.String(); got != tc.want {
				t.Fatalf("expected %q; got %q.", tc.want, got)
			}
		})
	}
}

func TestExecuteRule(t *testing.T) {
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
		{"(false=true)+(false=false)", 1, false},
		{"(false=true)*(false=false)", 0, false},
		{"-false*true", 1, false},
		{"-3*5<8+3", 1, false},
		{"-3*5>8-3", 0, false},
		{"4/2=2", 1, false},
		{"2+3*2=8", 1, false},
		{"2+3*2=10", 0, false},
		{`{"a","b","c"}`, 3, false},
		{`{"a","b","c"}={"b","c","a"}`, 1, false},
		{`{}={}`, 1, false},
		{`{"a","b"}={"c","a"}`, 0, false},
		{`{}={"c","a"}`, 0, false},
		{`{"a","b","c"}={"a"}`, 0, false},
		{`{"a","b","c"}+{"a"}={"a","b","c"}`, 1, false},
		{`{"a","b"}+{"a"}={"a","b"}`, 1, false},
		{`{"a","b"}+{"c"}={"a","b","c"}`, 1, false},
		{`{"b"}+{"a"}={"a","b"}`, 1, false},
		{`{}+{"a"}={"a"}`, 1, false},
		{`{"b"}+{}={"b"}`, 1, false},
		{`{"a"}*{"b"}={}`, 1, false},
		{`{"a"}*{"a","b"}={"a"}`, 1, false},
		{`{"a","c"}*{"a","b"}={"a"}`, 1, false},
		{`{"a","c"}*{"a","c"}={"a","c"}`, 1, false},
		{`{"a","c"}-{"a","c"}={}`, 1, false},
		{`{"a","c"}-{"a","b"}={"c"}`, 1, false},
		{`{"a","b"}-{"c"}={"a","b"}`, 1, false},
		{`"abc"="abc"`, 1, false},
		{`"bc"="abc"`, 0, false},
		{`"abc"<"def"`, 1, false},
		{`"abc">"def"`, 0, false},
		{`"def">"abc"`, 1, false},
		{`"def"<"abc"`, 0, false},
		{`len("")=0`, 1, false},
		{`len("my-string")=9`, 1, false},
		{`len({})=0`, 1, false},
		{`len({"a","b","c"})=3`, 1, false},
		{`len({"a","b","c"})+1=4`, 1, false},
		{`1-len({"a","b","c"})=-2`, 1, false},
		{`log(exp(1+1))=2*1`, 1, false},
		{`pow(1*2,2+1)=8`, 1, false},
		{"min()", -math.MaxFloat64, false},
		{"max()", math.MaxFloat64, false},
		{`min(true,true,false)`, 0, false},
		{`max(1,2,3+5)`, 8, false},
		{`min(-1,2,-3*5)`, -15, false},
		{`min(-1,2,-3*5)`, -15, false},
		{`min({"a","b","c"})`, 1, false}, // a little bit silly: this checks for the *minimal ID*.
		{"-{}", 0, true},
		{"-es()", 0, true},
		{`{"a","not","b"}`, 0, true},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			rule, err := Compile(tc.test, lookupID)
			if tc.iserr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			if got := rule.Execute(); got != tc.want {
				t.Fatalf("expected %f; got %f", tc.want, got)
			}
		})
	}
}
