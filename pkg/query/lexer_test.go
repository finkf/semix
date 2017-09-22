package query

import (
	"fmt"
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		query, want string
		iserr       bool
	}{
		{"?(<R{A,B})", "[? ( < R { A , B } )]", false},
		{"?(>R{'Qu ote',B})", "[? ( > R { Qu ote , B } )]", false},
		{`?(>R{"()?{}<>",B})`, "[? ( > R { ()?{}<> , B } )]", false},
		{`?(> R { "quot ed" , A } )`, "[? ( > R { quot ed , A } )]", false},
		{`?(<{ident1 ident2`, "[? ( < { ident1 ident2]", false},
		{`?(>R{"Missing quote, A})`, "[]", true},
		{"", "[]", false},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			lexer := NewLexer(tc.query)
			ls, err := lexer.Lex()
			if tc.iserr && err == nil {
				t.Fatalf("expected error")
			}
			if !tc.iserr && err != nil {
				t.Fatalf("got error: %v", err)
			}
			if str := fmt.Sprintf("%v", ls); str != tc.want {
				t.Fatalf("expected %q; got %q", tc.want, str)
			}
		})
	}
}
