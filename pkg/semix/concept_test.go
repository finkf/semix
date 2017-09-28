package semix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
)

func TestJSONMarshalling(t *testing.T) {
	tests := []struct {
		url, name string
		es        []string
		id        int32
	}{
		{url: "A", name: "A-name", id: 42, es: nil},
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
				t.Fatalf("could not encode concept: %v", err)
			}
			d := new(Concept)
			if err := json.NewDecoder(buffer).Decode(d); err != nil {
				t.Fatalf("could not decode concept: %v", err)
			}
			cstr := fmt.Sprintf("{%s %s %d %v}", c.url, c.Name, c.id, c.edges)
			dstr := fmt.Sprintf("{%s %s %d %v}", d.url, d.Name, d.id, d.edges)
			if dstr != cstr {
				t.Fatalf("expceted %s; got %s", cstr, dstr)
			}
		})
	}
}
