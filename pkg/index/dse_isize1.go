// +build isize1,!isize2,!isize3,!isize4

package index

import "testing"

// Short var names for smaller gob entries.
// P is the document id
// B is the start position
// E is the end position
// R stores the relation id, if entries are direct, their levenshtein distance
// and their ambiguity
type dse struct {
	P, B, E uint32
	R       relationID
}

func newDSE(e Entry, lookup lookupURLsFunc) dse {
	relID, docID := lookup(e.RelationURL, e.Path)
	return dse{
		P: uint32(docID),
		B: uint32(e.Begin),
		E: uint32(e.End),
		R: newRelationID(relID, e.L, e.Ambiguous, e.RelationURL != ""),
	}
}

func (d dse) entry(conceptURL string, lookup lookupIDsFunc) Entry {
	relURL, docURL := lookup(d.R.ID(), int(d.P))
	return Entry{
		ConceptURL:  conceptURL,
		RelationURL: relURL,
		Path:        docURL,
		Begin:       int(d.B),
		End:         int(d.E),
		L:           d.R.Distance(),
		Ambiguous:   d.R.Ambiguous(),
	}
}

func testEntries(t *testing.T, a, b Entry) {
	t.Helper()
	a.Token = b.Token
	if a != b {
		t.Fatalf("expected %v; got %v", b, a)
	}
}
