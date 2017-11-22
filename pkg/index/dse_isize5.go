// +build isize5

package index

import (
	"testing"
)

// Short var names for smaller gob entries.
// S is the string
// P is the document id
// B is the start position
// P is the document id
// B is the start position
// E is the end position
// L stores if entries are indirect, their levenshtein distance and their ambiguity.
type dse struct {
	S       string
	P, B, E uint32
	L       relationID
}

func newDSE(e Entry, lookup lookupURLsFunc) dse {
	_, docID := lookup("", e.Path)
	return dse{
		S: e.Token,
		P: uint32(docID),
		B: uint32(e.Begin),
		E: uint32(e.End),
		L: newRelationID(e.L, e.Ambiguous, e.RelationURL == ""),
	}
}

func (d dse) entry(conceptURL string, lookup lookupIDsFunc) Entry {
	_, docURL := lookup(0, int(d.P))
	return Entry{
		ConceptURL:  conceptURL,
		RelationURL: dseRelationURL(d.L.Direct()),
		Token:       d.S,
		Path:        docURL,
		Begin:       int(d.B),
		End:         int(d.E),
		L:           d.L.Distance(),
		Ambiguous:   d.L.Ambiguous(),
	}
}

func testEntries(t *testing.T, a, b Entry) {
	t.Helper()
	if a.RelationURL != "" {
		a.RelationURL = b.RelationURL
	}
	if a != b {
		t.Fatalf("expected %v; got %v", b, a)
	}
}
