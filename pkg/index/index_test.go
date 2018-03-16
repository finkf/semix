package index

import (
	"context"
	"strings"
	"testing"

	"bitbucket.org/fflo/semix/pkg/semix"
)

func TestIndex(t *testing.T) {
	tests := []struct {
		test string
		urls map[string]int
	}{
		{"a", map[string]int{"A": 1, "B": 1, "C": 1, "D": 0}},
		{"b", map[string]int{"A": 0, "B": 1, "C": 1, "D": 0}},
		{"c", map[string]int{"A": 0, "B": 0, "C": 1, "D": 0}},
		{"a, b oder c", map[string]int{"A": 1, "B": 2, "C": 3, "D": 0}},
		{"a, a oder a", map[string]int{"A": 3, "B": 3, "C": 3, "D": 0}},
		{"b, b oder b", map[string]int{"A": 0, "B": 3, "C": 3, "D": 0}},
		{"c, c oder c", map[string]int{"A": 0, "B": 0, "C": 3, "D": 0}},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			m := matcher()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			i := NewMemory(1)
			d := semix.NewStringDocument(tc.test, tc.test)
			s := Put(ctx, i,
				semix.Match(ctx, m,
					semix.Normalize(ctx,
						semix.Read(ctx, d))))
			for token := range s {
				if token.Err != nil {
					t.Fatalf("got error: %v", token.Err)
				}
			}
			for url, c := range tc.urls {
				if tmp := count(i, url); c != tmp {
					t.Fatalf("expected count(%s)=%d; got %d", url, c, tmp)
				}
			}
		})
	}
}

func count(i Interface, url string) int {
	var count int
	i.Get(url, func(e Entry) bool {
		count++
		return true
	})
	return count
}

func matcher() semix.Matcher {
	g := semix.NewGraph()
	d := make(semix.Dictionary)
	for _, ts := range strings.Split("A,P,B.B,P,C.A,P,C", ".") {
		t := strings.Split(ts, ",")
		if len(t) != 3 {
			panic("invalid triple: " + ts)
		}
		s, p, o := g.Add(t[0], t[1], t[2])
		d[strings.ToLower(t[0])] = s.ID()
		d[strings.ToLower(t[1])] = p.ID()
		d[strings.ToLower(t[2])] = o.ID()
	}
	return semix.DFAMatcher{DFA: semix.NewDFA(d, g)}
}
