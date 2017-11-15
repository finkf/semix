package index

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
