// +build isize2,!isize1,!isize3,!isize4

package index

// Short var names for smaller gob entries.
// P is the document path
// B is the start position
// E is the end position
// R stores the relation id, if entries are direct, their levenshtein distance
// and their ambiguity
type dse struct {
	P string
	R relationID
}

func newDSE(e Entry, register func(string) int) dse {
	return dse{
		P: e.Path,
		R: newRelationID(register(e.RelationURL), e.L, e.Ambiguous, e.RelationURL != ""),
	}
}

func (d dse) entry(conceptURL string, lookup func(int) string) Entry {
	return Entry{
		ConceptURL:  conceptURL,
		RelationURL: lookup(int(d.R.ID())),
		Path:        d.P,
		L:           d.R.Distance(),
		Ambiguous:   d.R.Ambiguous(),
	}
}
