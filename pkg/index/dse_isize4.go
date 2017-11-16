// +build isize4,!isize1,!isize2,!isize3

package index

// Short var names for smaller gob entries.
// P is the document id
// L stores if entries are indirect, their levenshtein distance and their ambiguity.
type dse struct {
	P uint32
	L uint8
}

func newDSE(e Entry, lookup lookupURLsFunc) dse {
	_, docID := lookup("", e.Path)
	return dse{
		P: uint32(docID),
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
		L:           l,
		Ambiguous:   a,
	}
}
