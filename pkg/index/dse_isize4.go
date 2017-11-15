// +build isize4,!isize1,!isize2,!isize3

package index

// Short var names for smaller gob entries.
// P is the document path
// L stores if entries are indirect, their levenshtein distance and their ambiguity.
type dse struct {
	P string
	L uint8
}

func newDSE(e Entry, register func(string) int) dse {
	return dse{
		P: e.Path,
		L: dseEncodeL(e.L, e.Ambiguous, e.RelationURL != ""),
	}
}

func (d dse) entry(conceptURL string, lookup func(int) string) Entry {
	l, a, dir := dseDecodeL(d.L)
	return Entry{
		ConceptURL:  conceptURL,
		RelationURL: dseRelationURL(dir),
		Path:        d.P,
		L:           l,
		Ambiguous:   a,
	}
}
