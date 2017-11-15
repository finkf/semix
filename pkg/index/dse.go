// +build !isize1,!isize2,!isize3,!isize4

package index

// Short var names for smaller gob entries.
// S is the string
// P is the document path
// B is the start position
// E is the end position
// R stores the relation id, if entries are direct, their levenshtein distance
// and their ambiguity
type dse struct {
	S, P string
	B, E uint32
	R    relationID
}

func newDSE(e Entry, register func(string) int) dse {
	return dse{
		S: e.Token,
		P: e.Path,
		B: uint32(e.Begin),
		E: uint32(e.End),
		R: newRelationID(register(e.RelationURL), e.L, e.Ambiguous, e.RelationURL != ""),
	}
}

func (d dse) entry(conceptURL string, lookup func(int) string) Entry {
	return Entry{
		ConceptURL:  conceptURL,
		RelationURL: lookup(int(d.R.ID())),
		Token:       d.S,
		Path:        d.P,
		Begin:       int(d.B),
		End:         int(d.E),
		L:           d.R.Distance(),
		Ambiguous:   d.R.Ambiguous(),
	}
}
