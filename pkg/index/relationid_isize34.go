// +build isize3 isize4

package index

type relationID uint8

// 0x3f -> 0011.1111
// 0x80 -> 1000.0000
// 0x40 -> 0100.0000
// max encodeable levenshtein distance -> 0x3f -> 63
const (
	levflag relationID = 0x3f
	aflag   relationID = 0x80
	dflag   relationID = 0x40
)

func newRelationID(l int, a, d bool) relationID {
	x := relationID(l) & levflag
	if a {
		x |= aflag
	}
	if d {
		x |= dflag
	}
	return x
}

func (x relationID) Distance() int {
	return int(x & levflag)
}

func (x relationID) Ambiguous() bool {
	return x&aflag > 0
}

func (x relationID) Direct() bool {
	return x&dflag > 0
}
