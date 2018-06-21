package semix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestJSONMarshalling(t *testing.T) {
	tests := []struct {
		url, name string
		es        []string
		id        int32
		a         bool
	}{
		{url: "A", name: "A-name", id: 42, es: nil, a: false},
		{url: SplitURL, name: "A-name", id: 42, es: nil, a: true},
		{url: "A", name: "A-name", id: 38, es: []string{"B", "C"}},
		{url: "A", name: "A-name", id: 38, es: []string{"B", "C", "B", "D"}},
	}
	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			c := &Concept{url: tc.url, Name: tc.name, id: tc.id}
			// add edges
			for i := 0; i < len(tc.es); i += 2 {
				p := &Concept{url: tc.es[i], Name: tc.es[i] + "-name", id: rand.Int31()}
				o := &Concept{url: tc.es[i+1], Name: tc.es[i+1] + "-name", id: rand.Int31()}
				c.edges = append(c.edges, Edge{P: p, O: o})
			}
			buffer := new(bytes.Buffer)
			if err := json.NewEncoder(buffer).Encode(c); err != nil {
				t.Fatalf("cannot encode concept: %v", err)
			}
			d := new(Concept)
			if err := json.NewDecoder(buffer).Decode(d); err != nil {
				t.Fatalf("cannot decode concept: %v", err)
			}
			cstr := fmt.Sprintf("{%s %s %d %v %t}",
				c.URL(), c.Name, c.ID(), c.edges, c.Ambig())
			dstr := fmt.Sprintf("{%s %s %d %v %t}",
				d.URL(), d.Name, d.ID(), d.edges, c.Ambig())
			if dstr != cstr {
				t.Fatalf("expceted %s; got %s", cstr, dstr)
			}
			if d.String() != c.String() {
				t.Fatalf("expceted %s; got %s", c, d)
			}
		})
	}
}

func TestCombineURLs(t *testing.T) {
	tests := []struct {
		urls []string
		want string
	}{
		{[]string{}, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b", "c"}, "a-b-c"},
		{[]string{"http://a/x", "http://a/y", "http://a/z"}, "http://a/x-y-z"},
		{[]string{"http://a/b/x", "http://a/c/y", "http://a/d/z"}, "http://a/b/x-y-z"},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc.urls), func(t *testing.T) {
			got := CombineURLs(tc.urls...)
			if got != tc.want {
				t.Errorf("expected %s; got %s", tc.want, got)
			}
		})
	}
}

func TestReduceTransitive(t *testing.T) {
	a := &Concept{url: "a", id: 1}
	b := &Concept{url: "b", id: 2}
	c := &Concept{url: "c", id: 3}
	p := &Concept{url: "p", id: 4}
	q := &Concept{url: "q", id: 5}
	d := &Concept{url: "d", id: 6}
	e := &Concept{url: "e", id: 7}

	a.edges = append(a.edges, Edge{P: p, O: b})
	a.edges = append(a.edges, Edge{P: p, O: c})
	a.edges = append(a.edges, Edge{P: p, O: d})
	a.edges = append(a.edges, Edge{P: q, O: e})
	b.edges = append(b.edges, Edge{P: p, O: c})
	b.edges = append(b.edges, Edge{P: p, O: d})
	b.edges = append(b.edges, Edge{P: q, O: e})
	c.edges = append(c.edges, Edge{P: p, O: d})

	a.ReduceTransitive()
	if a.HasLinkP(p, c) {
		t.Fatalf("a should not have p link to c")
	}
	if !a.HasLinkP(q, e) {
		t.Fatalf("a should have q link to e")
	}
	if a.HasLinkP(p, d) {
		t.Fatalf("a should not have p link to d")
	}
	if !a.HasLinkP(p, b) {
		t.Fatalf("a should have a p link to b")
	}
	if b.HasLinkP(p, d) {
		t.Fatalf("b should not have p link to d")
	}
	if !b.HasLinkP(p, c) {
		t.Fatalf("b should have a p link to c")
	}
	if !b.HasLinkP(q, e) {
		t.Fatalf("b should have q link to e")
	}
	if !c.HasLinkP(p, d) {
		t.Fatalf("c should have a p link to d")
	}
}

func TestVisitAll(t *testing.T) {
	a := &Concept{url: "a", id: 1}
	b := &Concept{url: "b", id: 2}
	c := &Concept{url: "c", id: 3}
	p := &Concept{url: "p", id: 4}
	q := &Concept{url: "q", id: 5}
	d := &Concept{url: "d", id: 6}

	a.edges = append(a.edges, Edge{P: p, O: b})
	a.edges = append(a.edges, Edge{P: p, O: c})
	a.edges = append(a.edges, Edge{P: p, O: d})
	a.edges = append(a.edges, Edge{P: q, O: a})
	b.edges = append(b.edges, Edge{P: p, O: c})
	b.edges = append(b.edges, Edge{P: p, O: d})
	c.edges = append(c.edges, Edge{P: p, O: d})

	got := make(map[*Concept]bool)
	a.VisitAll(func(c *Concept) {
		if got[c] {
			t.Fatalf("concept %s was visited already", c)
		}
		got[c] = true
	})
	want := map[*Concept]bool{a: true, b: true, c: true, d: true}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("expected %v; got %v", want, got)
	}
}
