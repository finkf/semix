package semix

import (
	"fmt"
	"testing"
)

func TestExpansion(t *testing.T) {
	tests := []struct {
		test, res string
		iserr     bool
	}{
		{"a,b,c", "[a,b,c]", false},
		{"a{b,c}", "[ab ac]", false},
		{"a{,b}", "[a ab]", false},
		{"a{b,}", "[ab a]", false},
		{"a{b,}c", "[abc ac]", false},
		{"a{,b}c", "[ac abc]", false},
		{"a{b{c,d}}", "[abc abd]", false},
		{"a,{b,c}", "[a,b a,c]", false},
		{`a{b\,\\,c}`, `[ab,\ ac]`, false},
		{`a\{b{c,d\}}`, `[a{bc a{bd}]`, false},
		{"a{b,c}{d,e}", "[abd abe acd ace]", false},
		{"a{b{c,d},e}{f,g}", "[abcf abcg abdf abdg aef aeg]", false},
		{`\{a\,b\}`, "[{a,b}]", false},
		{`a{b,c}\`, `[ab ac]`, false},
		{`a{b,c}\`, `[ab ac]`, false},
		{"Georg {von der ,}Marwitz", "[Georg von der Marwitz Georg Marwitz]", false},
		{"a{b,}}", "", true},
		{"a{{b,}", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			// fmt.Println("running " + tc.test)
			es, err := ExpandBraces(tc.test)
			if !tc.iserr && err != nil {
				t.Fatalf("got error: %v", err)
			}
			if tc.iserr && err == nil {
				t.Fatalf("expected error")
			}
			if !tc.iserr && fmt.Sprintf("%v", es) != tc.res {
				t.Fatalf("expected %v; got %v", tc.res, es)
			}
		})
	}
}
