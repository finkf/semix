package rdfxml

import (
	"testing"
)

func TestTransitiveClosure(t *testing.T) {
	for _, tc := range []struct {
		test, want map[triple]bool
	}{
		{
			map[triple]bool{
				triple{"a", "p", "b"}: true,
				triple{"b", "p", "c"}: true,
				triple{"c", "p", "d"}: true,
				triple{"e", "p", "f"}: true,
			},
			map[triple]bool{
				triple{"a", "p", "a"}: false,
				triple{"a", "p", "b"}: true,
				triple{"a", "p", "c"}: true,
				triple{"a", "p", "d"}: true,
				triple{"a", "p", "e"}: false,
				triple{"a", "p", "f"}: false,

				triple{"b", "p", "a"}: false,
				triple{"b", "p", "b"}: false,
				triple{"b", "p", "c"}: true,
				triple{"b", "p", "d"}: true,
				triple{"b", "p", "e"}: false,
				triple{"b", "p", "f"}: false,

				triple{"c", "p", "a"}: false,
				triple{"c", "p", "b"}: false,
				triple{"c", "p", "c"}: false,
				triple{"c", "p", "d"}: true,
				triple{"c", "p", "e"}: false,
				triple{"e", "p", "f"}: false,

				triple{"d", "p", "a"}: false,
				triple{"d", "p", "b"}: false,
				triple{"d", "p", "c"}: false,
				triple{"d", "p", "d"}: false,
				triple{"d", "p", "e"}: false,
				triple{"d", "p", "f"}: false,

				triple{"e", "p", "a"}: false,
				triple{"e", "p", "b"}: false,
				triple{"e", "p", "c"}: false,
				triple{"e", "p", "d"}: false,
				triple{"e", "p", "e"}: false,
				triple{"e", "p", "f"}: true,

				triple{"f", "p", "a"}: false,
				triple{"f", "p", "b"}: false,
				triple{"f", "p", "c"}: false,
				triple{"f", "p", "d"}: false,
				triple{"f", "p", "e"}: false,
				triple{"f", "p", "f"}: false,
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
		test, want map[triple]bool
	}{
		{
			map[triple]bool{
				triple{"a", "p", "b"}: true,
				triple{"b", "p", "c"}: true,
				triple{"c", "p", "d"}: true,
				triple{"e", "p", "e"}: true,
			},
			map[triple]bool{
				triple{"a", "p", "a"}: false,
				triple{"a", "p", "b"}: true,
				triple{"a", "p", "c"}: false,
				triple{"a", "p", "d"}: false,
				triple{"a", "p", "e"}: false,

				triple{"b", "p", "a"}: true,
				triple{"b", "p", "b"}: false,
				triple{"b", "p", "c"}: true,
				triple{"b", "p", "d"}: false,
				triple{"b", "p", "e"}: false,

				triple{"c", "p", "a"}: false,
				triple{"c", "p", "b"}: true,
				triple{"c", "p", "c"}: false,
				triple{"c", "p", "d"}: true,
				triple{"c", "p", "e"}: false,

				triple{"d", "p", "a"}: false,
				triple{"d", "p", "b"}: false,
				triple{"d", "p", "c"}: true,
				triple{"d", "p", "d"}: false,
				triple{"d", "p", "e"}: false,

				triple{"e", "p", "a"}: false,
				triple{"e", "p", "b"}: false,
				triple{"e", "p", "c"}: false,
				triple{"e", "p", "d"}: false,
				triple{"e", "p", "e"}: true,
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
