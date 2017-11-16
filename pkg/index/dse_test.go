// +build !isize1,!isize2,!isize3,!isize4

package index

import (
	"fmt"
	"testing"
)

func testDSELookupIDFunc(url string) int {
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

func testDSELookupIDsFunc(relURL, docURL string) (int, int) {
	return testDSELookupIDFunc(relURL), testDSELookupIDFunc(docURL)
}

func testDSELookupURLFunc(id int) string {
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

func testDSELookupURLsFunc(relID, docID int) (string, string) {
	return testDSELookupURLFunc(relID), testDSELookupURLFunc(docID)
}

// type Entry struct {
// 	ConceptURL, Path, RelationURL, Token string
// 	Begin, End, L                        int
// 	Ambiguous                            bool
// }
func TestDSE(t *testing.T) {
	tests := []Entry{
		{"T", "", "", "", 0, 0, 0, false},
		{"T", "B", "A", "test-token", 1, 7, 1, false},
		{"T", "C", "B", "test-token", 2, 8, 2, true},
		{"T", "A", "C", "test-token", 3, 9, 3, false},
		{"T", "A", "", "test-token", 4, 10, 4, true},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("%v", tc), func(t *testing.T) {
			dse := newDSE(tc, testDSELookupIDsFunc)
			e := dse.entry("T", testDSELookupURLsFunc)
			if tc != e {
				t.Errorf("expected %v; got %v", tc, e)
			}
		})
	}
}
