// +build isize4

package index

import (
	"testing"
)

// Short var names for smaller gob entries.
// P is the document id
// B is the start position
// E is the end position
// L stores if entries are indirect, their levenshtein distance and their ambiguity.
type dse struct {
	P uint32
	L relationID
}

func newDSE(e Entry, lookup lookupURLsFunc) dse {
	_, docID := lookup("", e.Path)
	return dse{
		P: uint32(docID),
		L: newRelationID(e.L, e.Ambiguous, e.RelationURL == ""),
	}
}

func (d dse) entry(conceptURL string, lookup lookupIDsFunc) Entry {
	_, docURL := lookup(0, int(d.P))
	return Entry{
		ConceptURL:  conceptURL,
		RelationURL: dseRelationURL(d.L.Direct()),
		Path:        docURL,
		L:           d.L.Distance(),
		Ambiguous:   d.L.Ambiguous(),
	}
}

func testEntries(t *testing.T, a, b Entry) {
	t.Helper()
	a.Token = b.Token
	a.Begin = b.Begin
	a.End = b.End
	if a.RelationURL != "" {
		a.RelationURL = b.RelationURL
	}
	if a != b {
		t.Fatalf("expected %v; got %v", b, a)
	}
}
