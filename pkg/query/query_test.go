package query

import (
	"errors"
	"fmt"
	"sort"
	"testing"
)

func TestQuery(t *testing.T) {
	tests := []struct {
		query, want string
		err         bool
	}{
		{"?(*({A}))", "[{A R} {A S}]", false},
		{"?({R,S}({A}))", "[{A R} {A S}]", false},
		{"?({S}({A}))", "[{A S}]", false},
		{"?(!{S}({A}))", "[{A R}]", false},
		{"?({}({A}))", "[{A }]", false},
		{"?(*({A,B}))", "[{A R} {A S} {B R} {B S}]", false},
		{"?({R,S}({A,B}))", "[{A R} {A S} {B R} {B S}]", false},
		{"?({S}({A,B}))", "[{A S} {B S}]", false},
		{"?(!{S}({A,B}))", "[{A R} {B R}]", false},
		{"?({}({A,B}))", "[{A } {B }]", false},
		{"?(}({A,B}))", "[]", true},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			es, err := Execute(tc.query, queryTestIndex{})
			if tc.err && err == nil {
				t.Fatalf("expected error")
			}
			if !tc.err && err != nil {
				t.Fatalf("got error: %v", err)
			}
			sort.Slice(es, func(i, j int) bool {
				return es[i].URL < es[j].URL
			})
			if str := fmt.Sprintf("%v", es); str != tc.want {
				t.Fatalf("expected %q; got %q", tc.want, str)
			}
			_, err = Execute(tc.query, queryTestIndex{errors.New("test")})
			if !tc.err {
				if err.Error() != "test" {
					t.Fatalf("expceted error")
				}
			}
		})
	}
}

type queryTestIndex struct {
	err error
}

func (i queryTestIndex) Get(url string, f func(e IndexEntry)) error {
	f(IndexEntry{URL: url, Rel: ""})
	f(IndexEntry{URL: url, Rel: "R"})
	f(IndexEntry{URL: url, Rel: "S"})
	return i.err
}
