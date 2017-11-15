// +build isize3 isize4

package index

func dseRelationURL(d bool) string {
	if d {
		return ""
	}
	return "http://bitbucket.org/fflo/semix/pkg/index/indirect"
}

func dseEncodeL(l int, a, d bool) uint8 {
	return uint8(newRelationID(0, l, a, d) >> levshift)
}

func dseDecodeL(x uint8) (int, bool, bool) {
	id := relationID(x << levshift)
	return id.Distance(), id.Ambiguous(), id.Direct()
}
