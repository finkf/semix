// +build isize2,!isize1,!isize3,!isize4

package index

import "testing"

// Short var names for smaller gob entries.
// P is the document id
// R stores the relation id, if entries are direct, their levenshtein distance
// and their ambiguity
type dse struct {
	P uint32
	R relationID
}

func newDSE(e Entry, lookup lookupURLsFunc) dse {
	relID, docID := lookup(e.RelationURL, e.Path)
	return dse{
		P: uint32(docID),
		R: newRelationID(relID, e.L, e.Ambiguous, e.RelationURL != ""),
	}
}

func (d dse) entry(conceptURL string, lookup lookupIDsFunc) Entry {
	relURL, docURL := lookup(d.R.ID(), int(d.P))
	return Entry{
		ConceptURL:  conceptURL,
		RelationURL: relURL,
		Path:        docURL,
		L:           d.R.Distance(),
		Ambiguous:   d.R.Ambiguous(),
	}
}

func testEntries(t *testing.T, a, b Entry) {
	t.Helper()
	a.Token = b.Token
	a.Begin = b.Begin
	a.End = b.End
	if a != b {
		t.Fatalf("expected %v; got %v", b, a)
	}
}
