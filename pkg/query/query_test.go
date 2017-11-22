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
	}{
		{"?(*({A}))", "[{A R} {A S}]", false, 0},
		{"?(R,S({A}))", "[{A R} {A S}]", false, 0},
		{"?(S({A}))", "[{A S}]", false, 0},
		{"?(!S({A}))", "[{A R}]", false, 0},
		{"?({A})", "[{A }]", false, 0},
		{"?(*({A,B}))", "[{A R} {A S} {B R} {B S}]", false, 0},
		{"?(R,S({A,B}))", "[{A R} {A S} {B R} {B S}]", false, 0},
		{"?0(R,S({A,B}))", "[{A R} {A S} {B R} {B S}]", false, 0},
		{"?1(R,S({A,B}))", "[]", false, 2},
		{"?2(R,S({A,B}))", "[{A R} {A S} {B R} {B S}]", false, 2},
		{"?3(R,S({A,B}))", "[{A R} {A S} {B R} {B S}]", false, 2},
		{"?(S({A,B}))", "[{A S} {B S}]", false, 0},
		{"?(!S({A,B}))", "[{A R} {B R}]", false, 0},
		{"?({A,B})", "[{A } {B }]", false, 0},
		{"?(}({A,B}))", "[]", true, 0},
		{"?({A,B}({C,D}))", "[]", true, 0},
		{"?(S({E,B}))", "", true, 0},
		{"?(E({A,B}))", "", true, 0},
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
			es, err := q.Execute(queryTestIndex{k: tc.k})
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
}

func (queryTestIndex) Put(semix.Token) error { return nil }
func (queryTestIndex) Close() error          { return nil }
func (i queryTestIndex) Get(url string, f func(e index.Entry)) error {
	f(index.Entry{ConceptURL: url, RelationURL: "", L: i.k})
	f(index.Entry{ConceptURL: url, RelationURL: "R", L: i.k})
	f(index.Entry{ConceptURL: url, RelationURL: "S", L: i.k})
	return i.err
}
