// +build isize1 isize2 !isize1,!isize2,!isize3,!isize4,!isize5

package index

type relationID uint32

// 0x3f -> 0011.1111
// 0x80 -> 1000.0000
// 0x40 -> 0100.0000
// max encodeable levenshtein distance -> 0x3f -> 63
// max encodeable relation ID -> 0xffffff -> 16777216
const (
	idflag   relationID = 0x00ffffff
	levflag  relationID = 0x3f
	levshift relationID = 3 * 8
	aflag    relationID = 0x80000000
	dflag    relationID = 0x40000000
)

func newRelationID(id, l int, a, d bool) relationID {
	x := relationID(id) & idflag
	x |= (relationID(l) & levflag) << levshift
	if a {
		x |= 0x80000000
	}
	if d {
		x |= 0x40000000
	}
	return x
}

func (x relationID) ID() int {
	return int(x & idflag)
}

func (x relationID) Distance() int {
	return int((x >> levshift) & levflag)
}

func (x relationID) Ambiguous() bool {
	return x&aflag > 0
}

func (x relationID) Direct() bool {
	return x&dflag > 0
}
