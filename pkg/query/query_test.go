package query

import (
	"errors"
	"fmt"
	"sort"
	"testing"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/semix"
)

func TestQueryExecute(t *testing.T) {
	tests := []struct {
		query, want string
		iserr       bool
		k           int
		a           bool
	}{
		{"?(*({A}))", "[{A } {A R} {A S}]", false, 0, false},
		{"?(R,S({A}))", "[{A } {A R} {A S}]", false, 0, false},
		{"?(S({A}))", "[{A } {A S}]", false, 0, false},
		{"?(!S({A}))", "[{A } {A R}]", false, 0, false},
		{"?({A})", "[{A }]", false, 0, false},
		{"?(*({A,B}))", "[{A } {A R} {A S} {B } {B R} {B S}]", false, 0, false},
		{"?(R,S({A,B}))", "[{A } {A R} {A S} {B } {B R} {B S}]", false, 0, false},
		{"?0(R,S({A,B}))", "[{A } {A R} {A S} {B } {B R} {B S}]", false, 0, false},
		{"?1(R,S({A,B}))", "[]", false, 2, false},
		{"?2(R,S({A,B}))", "[{A } {A R} {A S} {B } {B R} {B S}]", false, 2, false},
		{"?3(R,S({A,B}))", "[{A } {A R} {A S} {B } {B R} {B S}]", false, 2, false},
		{"?(S({A,B}))", "[{A } {A S} {B } {B S}]", false, 0, false},
		{"?(!S({A,B}))", "[{A } {A R} {B } {B R}]", false, 0, false},
		{"?({A,B})", "[{A } {B }]", false, 0, false},
		{"?(R({A}))", "[]", false, 0, true},
		{"?*(R({A}))", "[{A } {A R}]", false, 0, true},
		{"?(R({A}))", "[]", false, 0, true},
		{"?*(R({A}))", "[{A } {A R}]", false, 0, true},
		{"?1(R({A}))", "[]", false, 2, true},
		{"?2(R({A}))", "[]", false, 2, true},
		{"?*1(R({A}))", "[]", false, 2, true},
		{"?*2(R({A}))", "[{A } {A R}]", false, 0, true},
		{"?1*(R({A}))", "[]", false, 2, true},
		{"?2*(R({A}))", "[{A } {A R}]", false, 0, true},
		{"?(}({A,B}))", "[]", true, 0, false},
		{"?({A,B}({C,D}))", "[]", true, 0, false},
		{"?(S({E,B}))", "", true, 0, false},
		{"?(E({A,B}))", "", true, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.query, func(t *testing.T) {
			q, err := New(tc.query, func(str string) ([]string, error) {
				if str == "E" {
					return nil, errors.New("ERROR")
				}
				return []string{str}, nil
			})
			if tc.iserr && err != nil {
				return
			}
			es, err := q.Execute(queryTestIndex{k: tc.k, a: tc.a})
			if tc.iserr {
				if err == nil {
					t.Fatalf("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("got error: %v", err)
			}
			sort.Slice(es, func(i, j int) bool {
				return es[i].ConceptURL < es[j].ConceptURL
			})
			if str := tostring(es); str != tc.want {
				t.Fatalf("expected %q; got %q", tc.want, str)
			}
			_, err = q.Execute(queryTestIndex{err: errors.New("test")})
			if !tc.iserr {
				if err.Error() != "test" {
					t.Fatalf("expceted error")
				}
			}
		})
	}
}

func tostring(es []index.Entry) string {
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
	k   int
	a   bool
}

func (queryTestIndex) Put(semix.Token) error { return nil }
func (queryTestIndex) Close() error          { return nil }
func (i queryTestIndex) Get(url string, f func(e index.Entry)) error {
	f(index.Entry{ConceptURL: url, RelationURL: "", L: i.k, Ambiguous: i.a})
	f(index.Entry{ConceptURL: url, RelationURL: "R", L: i.k, Ambiguous: i.a})
	f(index.Entry{ConceptURL: url, RelationURL: "S", L: i.k, Ambiguous: i.a})
	return i.err
}
