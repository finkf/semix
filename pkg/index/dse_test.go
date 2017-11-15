// +build !isize1,!isize2,!isize3,!isize4

package index

import (
	"fmt"
	"testing"
)

func testDSARegisterFunc(url string) int {
	switch url {
	case "A":
		return 1
	case "B":
		return 2
	case "C":
		return 42
	case "":
		return 0
	default:
		panic("invalid url: " + url)
	}
}

func testDSALookupFunc(id int) string {
	switch id {
	case 1:
		return "A"
	case 2:
		return "B"
	case 42:
		return "C"
	case 0:
		return ""
	default:
		panic(fmt.Sprintf("invalid id: %d", id))
	}
}

// type Entry struct {
// 	ConceptURL, Path, RelationURL, Token string
// 	Begin, End, L                        int
// 	Ambiguous                            bool
// }
func TestDSE(t *testing.T) {
	tests := []Entry{
		{"T", "", "", "", 0, 0, 0, false},
		{"T", "test-path", "A", "test-token", 1, 7, 1, false},
		{"T", "test-path", "B", "test-token", 2, 8, 2, true},
		{"T", "test-path", "C", "test-token", 3, 9, 3, false},
		{"T", "test-path", "", "test-token", 4, 10, 4, true},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			dse := newDSE(tc, testDSARegisterFunc)
			e := dse.entry("T", testDSALookupFunc)
			if tc != e {
				t.Errorf("expected %v; got %v", tc, e)
			}
		})
	}
}
