package semix

import (
	"testing"
)

func TestTransitiveClosure(t *testing.T) {
	for _, tc := range []struct {
		test, want map[spo]bool
	}{
		{
			map[spo]bool{
				spo{"a", "p", "b"}: true,
				spo{"b", "p", "c"}: true,
				spo{"c", "p", "d"}: true,
				spo{"e", "p", "f"}: true,
			},
			map[spo]bool{
				spo{"a", "p", "a"}: false,
				spo{"a", "p", "b"}: true,
				spo{"a", "p", "c"}: true,
				spo{"a", "p", "d"}: true,
				spo{"a", "p", "e"}: false,
				spo{"a", "p", "f"}: false,

				spo{"b", "p", "a"}: false,
				spo{"b", "p", "b"}: false,
				spo{"b", "p", "c"}: true,
				spo{"b", "p", "d"}: true,
				spo{"b", "p", "e"}: false,
				spo{"b", "p", "f"}: false,

				spo{"c", "p", "a"}: false,
				spo{"c", "p", "b"}: false,
				spo{"c", "p", "c"}: false,
				spo{"c", "p", "d"}: true,
				spo{"c", "p", "e"}: false,
				spo{"e", "p", "f"}: false,

				spo{"d", "p", "a"}: false,
				spo{"d", "p", "b"}: false,
				spo{"d", "p", "c"}: false,
				spo{"d", "p", "d"}: false,
				spo{"d", "p", "e"}: false,
				spo{"d", "p", "f"}: false,

				spo{"e", "p", "a"}: false,
				spo{"e", "p", "b"}: false,
				spo{"e", "p", "c"}: false,
				spo{"e", "p", "d"}: false,
				spo{"e", "p", "e"}: false,
				spo{"e", "p", "f"}: true,

				spo{"f", "p", "a"}: false,
				spo{"f", "p", "b"}: false,
				spo{"f", "p", "c"}: false,
				spo{"f", "p", "d"}: false,
				spo{"f", "p", "e"}: false,
				spo{"f", "p", "f"}: false,
			},
		},
	} {
		closure := calculateTransitiveClosure(tc.test)
		for tr, b := range tc.want {
			if closure[tr] != b {
				t.Errorf("expected closure[%v] = %t; got %t", tr, b, closure[tr])
			}
		}
	}
}

func TestSymmetricClosure(t *testing.T) {
	for _, tc := range []struct {
		test, want map[spo]bool
	}{
		{
			map[spo]bool{
				spo{"a", "p", "b"}: true,
				spo{"b", "p", "c"}: true,
				spo{"c", "p", "d"}: true,
				spo{"e", "p", "e"}: true,
			},
			map[spo]bool{
				spo{"a", "p", "a"}: false,
				spo{"a", "p", "b"}: true,
				spo{"a", "p", "c"}: false,
				spo{"a", "p", "d"}: false,
				spo{"a", "p", "e"}: false,

				spo{"b", "p", "a"}: true,
				spo{"b", "p", "b"}: false,
				spo{"b", "p", "c"}: true,
				spo{"b", "p", "d"}: false,
				spo{"b", "p", "e"}: false,

				spo{"c", "p", "a"}: false,
				spo{"c", "p", "b"}: true,
				spo{"c", "p", "c"}: false,
				spo{"c", "p", "d"}: true,
				spo{"c", "p", "e"}: false,

				spo{"d", "p", "a"}: false,
				spo{"d", "p", "b"}: false,
				spo{"d", "p", "c"}: true,
				spo{"d", "p", "d"}: false,
				spo{"d", "p", "e"}: false,

				spo{"e", "p", "a"}: false,
				spo{"e", "p", "b"}: false,
				spo{"e", "p", "c"}: false,
				spo{"e", "p", "d"}: false,
				spo{"e", "p", "e"}: true,
			},
		},
	} {
		closure := calculateSymmetricClosure(tc.test)
		for tr, b := range tc.want {
			if closure[tr] != b {
				t.Errorf("expected closure[%v] = %t; got %t", tr, b, closure[tr])
			}
		}
	}
}
