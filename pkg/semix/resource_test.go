package semix

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"
)

func TestResourceGOB(t *testing.T) {
	r, err := Parse(makeNewTestParser(), testTraits{})
	if err != nil {
		t.Fatalf("got error: %s", err)
	}
	b := new(bytes.Buffer)
	if err := gob.NewEncoder(b).Encode(r); err != nil {
		t.Fatalf("got error: %s", err)
	}
	x := new(Resource)
	if err := gob.NewDecoder(b).Decode(x); err != nil {
		t.Fatalf("got error: %s", err)
	}
	if !reflect.DeepEqual(*x, *r) {
		t.Fatalf("decoded resources do not equal encoded resources")
	}

}
