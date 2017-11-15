package index

// Short var names for smaller gob entries.
// S is the string
// P is the document path
// B is the start position
// E is the end position
// R is the relation id
type dse struct {
	S, P string
	B, E uint32
	R    int32
	L    uint8
}

func newDSE(e Entry, register func(string) int) dse {
	return dse{
		S: e.Token,
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
		Token:       d.S,
		Path:        d.P,
		Begin:       int(d.B),
		End:         int(d.E),
		L:           l,
		Ambiguous:   a,
	}
}

// 0x3f -> 0011.1111
// 0x80 -> 1000.0000
// 0x40 -> 0100.0000
// max encodeable levenshtein distance -> 0x3f -> 63
func encodeL(l int, a, d bool) uint8 {
	x := uint8(l) & 0x3f
	if a {
		x |= 0x80
	}
	if d {
		x |= 0x40
	}
	return x
}

func decodeL(x uint8) (int, bool, bool) {
	return int(x & 0x3f), x&0x80 > 0, x&0x40 > 0
}
