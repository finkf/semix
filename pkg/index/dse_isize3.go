// +build isize3,!isize1,!isize2,!isize4

package index

// Short var names for smaller gob entries.
// P is the document path
// B is the start position
// E is the end position
// L stores if entries are indirect, their levenshtein distance and their ambiguity.
type dse struct {
	P    string
	B, E uint32
	L    uint8
}

func newDSE(e Entry, register func(string) int) dse {
	return dse{
		P: e.Path,
		B: uint32(e.Begin),
		E: uint32(e.End),
		L: encodeL(e.L, e.Ambiguous, e.RelationURL != ""),
	}
}

func (d dse) entry(conceptURL string, lookup func(int) string) Entry {
	l, a, dir := decodeL(d.L)
	return Entry{
		ConceptURL:  conceptURL,
		RelationURL: getRelationURL(dir),
		Path:        d.P,
		Begin:       int(d.B),
		End:         int(d.E),
		L:           l,
		Ambiguous:   a,
	}
}

func getRelationURL(d bool) string {
	if d {
		return ""
	}
	return "http://bitbucket.org/fflo/semix/pkg/index/indirect"
}
