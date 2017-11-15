// +build isize1,!isize2,!isize3,!isize4

package index

// Short var names for smaller gob entries.
// P is the document path
// B is the start position
// E is the end position
// R is the relation id
type dse struct {
	P    string
	B, E uint32
	R    int32
	L    uint8
}

func newDSE(e Entry, register func(string) int) dse {
	return dse{
		P: e.Path,
		B: uint32(e.Begin),
		E: uint32(e.End),
		L: encodeL(e.L, e.Ambiguous, e.RelationURL != ""),
		R: int32(register(e.RelationURL)),
	}
}

func (d dse) entry(conceptURL string, lookup func(int) string) Entry {
	l, a, _ := decodeL(d.L)
	return Entry{
		ConceptURL:  conceptURL,
		RelationURL: lookup(int(d.R)),
		Path:        d.P,
		Begin:       int(d.B),
		End:         int(d.E),
		L:           l,
		Ambiguous:   a,
	}
}
