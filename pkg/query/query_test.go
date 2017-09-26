package query

import (
	"errors"
	"fmt"
	"sort"
	"testing"

	"bitbucket.org/fflo/semix/pkg/semix"
)

func TestQuery(t *testing.T) {
	tests := []struct {
		query, want string
		err         bool
	}{
		{"?(*({A}))", "[{A R} {A S}]", false},
		{"?(R,S({A}))", "[{A R} {A S}]", false},
		{"?(S({A}))", "[{A S}]", false},
		{"?(!S({A}))", "[{A R}]", false},
		{"?({A})", "[{A }]", false},
		{"?(*({A,B}))", "[{A R} {A S} {B R} {B S}]", false},
		{"?(R,S({A,B}))", "[{A R} {A S} {B R} {B S}]", false},
		{"?(S({A,B}))", "[{A S} {B S}]", false},
		{"?(!S({A,B}))", "[{A R} {B R}]", false},
		{"?({A,B})", "[{A } {B }]", false},
		{"?(}({A,B}))", "[]", true},
		{"?({A,B}({C,D}))", "[]", true},
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
				return es[i].ConceptURL < es[j].ConceptURL
			})
			if str := tostring(es); str != tc.want {
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

func tostring(es []semix.IndexEntry) string {
	type pair struct {
		first, second string
	}
	var pairs []pair
	for _, e := range es {
		pairs = append(pairs, pair{
			first:  e.ConceptURL,
			second: e.RelationURL,
		})
	}
	return fmt.Sprintf("%v", pairs)
}

type queryTestIndex struct {
	err error
}

func (queryTestIndex) Put(semix.Token) error { return nil }
func (queryTestIndex) Close() error          { return nil }
func (i queryTestIndex) Get(url string, f func(e semix.IndexEntry)) error {
	f(semix.IndexEntry{ConceptURL: url, RelationURL: ""})
	f(semix.IndexEntry{ConceptURL: url, RelationURL: "R"})
	f(semix.IndexEntry{ConceptURL: url, RelationURL: "S"})
	return i.err
}
