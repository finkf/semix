package semix

import (
	"fmt"
	"testing"
)

func TestExpansion(t *testing.T) {
	tests := []struct {
		test, res string
	}{
		{"a,b,c", "[a b c]"},
		{"a{b,c}", "[ab ac]"},
		{"a{,b}", "[a ab]"},
		{"a{b{c,d}}", "[abc abd]"},
		{"a,{b,c}", "[a b c]"},
		{`a{b\,\\,c}`, `[ab,\ ac]`},
		{`a\{b{c,d\}}`, `[a{bc a{bd}]`},
		{"a{b,c}{d,e}", "[abd abe acd ace]"},
		{"a{b{c,d},e}{f,g}", "[abcf abcg abdf abdg aef aeg]"},
		{`\{a\,b\}`, "[{a,b}]"},
		{`a{b,c}\`, `[ab ac]`},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			// fmt.Println("running " + tc.test)
			es, err := ExpandBraces(tc.test)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprintf("%v", es) != tc.res {
				t.Fatalf("expected %v; got %v", tc.res, es)
			}
		})
	}
}
