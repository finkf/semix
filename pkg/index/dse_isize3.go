// +build isize3,!isize1,!isize2,!isize4

package index

// Short var names for smaller gob entries.
// P is the document id
// B is the start position
// E is the end position
// L stores if entries are indirect, their levenshtein distance and their ambiguity.
type dse struct {
	P, B, E uint32
	L       uint8
}

func newDSE(e Entry, lookup lookupURLsFunc) dse {
	_, docID := lookup("", e.Path)
	return dse{
		P: uint32(docID),
		B: uint32(e.Begin),
		E: uint32(e.End),
		L: dseEncodeL(e.L, e.Ambiguous, e.RelationURL != ""),
	}
}

func (d dse) entry(conceptURL string, lookup lookupIDsFunc) Entry {
	_, docURL := lookup(0, int(d.P))
	l, a, dir := dseDecodeL(d.L)
	return Entry{
		ConceptURL:  conceptURL,
		RelationURL: dseRelationURL(dir),
		Path:        docURL,
		Begin:       int(d.B),
		End:         int(d.E),
		L:           l,
		Ambiguous:   a,
	}
}
