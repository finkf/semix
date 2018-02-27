package semix

import (
	"bytes"
	"encoding/gob"
	"testing"
)

func TestLookup(t *testing.T) {
	tests := []struct {
		url string
		id  int
		ok  bool
	}{
		{"first-url", 1, true},
		{"second-url", 2, true},
		{"third-url", 3, true},
		{"fourth-url", 4, true},
		{"fifth-url", 0, false},
		{"sixth-url", 0, false},
		{"id-0", 0, false},
		{"id-smaller-than-0", -8, false},
		{"id-larger-than-len", 8, false},
	}
	r := NewURLRegister()
	for _, tc := range tests {
		t.Run("register-"+tc.url, func(t *testing.T) {
			if tc.ok {
				if id := r.Register(tc.url); id != tc.id {
					t.Fatalf("expected id = %d; got %d", tc.id, id)
				}
				// try it a second time
				if id := r.Register(tc.url); id != tc.id {
					t.Fatalf("expected id = %d; got %d", tc.id, id)
				}

			}
		})
	}
	for _, tc := range tests {
		t.Run("lookup-id-"+tc.url, func(t *testing.T) {
			url, ok := r.LookupID(tc.id)
			if ok != tc.ok {
				t.Fatalf("expected ok = %t; got %t", tc.ok, ok)
			}
			if ok && url != tc.url { // check valid urls only
				t.Fatalf("expceted url = %q; got %q", tc.url, url)
			}
			id, ok := r.LookupURL(tc.url)
			if ok != tc.ok {
				t.Fatalf("expected ok = %t; got %t", tc.ok, ok)
			}
			if ok && id != tc.id { // check valid ids only
				t.Fatalf("expceted id = %d; got %d", tc.id, id)
			}

		})
	}

}

func TestGob(t *testing.T) {
	tests := []struct {
		url string
		id  int
		ok  bool
	}{
		{"first-url", 1, true},
		{"second-url", 2, true},
		{"third-url", 3, true},
		{"fourth-url", 4, true},
		{"fifth-url", 0, false},
		{"sixth-url", 0, false},
		{"id-0", 0, false},
		{"id-smaller-than-0", -8, false},
		{"id-larger-than-len", 8, false},
	}
	r := NewURLRegister()
	for _, tc := range tests {
		if !tc.ok {
			continue
		}
		r.Register(tc.url)
	}
	b := &bytes.Buffer{}
	e := gob.NewEncoder(b)
	if err := e.Encode(r); err != nil {
		t.Fatalf("cannot encode register: %v", err)
	}
	r2 := new(URLRegister)
	d := gob.NewDecoder(b)
	if err := d.Decode(r2); err != nil {
		t.Fatalf("cannot decode register: %v", err)
	}
	for _, tc := range tests {
		t.Run("lookup-id-"+tc.url, func(t *testing.T) {
			url, ok := r2.LookupID(tc.id)
			if ok != tc.ok {
				t.Fatalf("expected ok = %t; got %t", tc.ok, ok)
			}
			if ok && url != tc.url { // check valid urls only
				t.Fatalf("expceted url = %q; got %q", tc.url, url)
			}
			id, ok := r2.LookupURL(tc.url)
			if ok != tc.ok {
				t.Fatalf("expected ok = %t; got %t", tc.ok, ok)
			}
			if ok && id != tc.id { // check valid ids only
				t.Fatalf("expceted id = %d; got %d", tc.id, id)
			}

		})
	}
}

func TestGobErrors(t *testing.T) {
	tests := []struct {
		name string
		test interface{}
	}{
		{"int", int(15)},
		{"string", "some string"},
		{"url-map", map[string]int{"a": 1, "b": 2}},
		{"ids", []string{"a", "b", "c"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := &bytes.Buffer{}
			e := gob.NewEncoder(b)
			if err := e.Encode(tc.test); err != nil {
				t.Fatalf("cannot encode %v: %v", tc.test, err)
			}
			r := new(URLRegister)
			d := gob.NewDecoder(b)
			if err := d.Decode(r); err == nil {
				t.Fatalf("should not be possible to decode %v", tc.test)
			}

		})
	}
}
