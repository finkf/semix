package rest

import (
	"net/url"
	"reflect"
	"testing"
)

func TestQueryInvalid(t *testing.T) {
	url, _ := url.Parse("http://example.org?a=b&c=d")
	var i int
	if err := DecodeQuery(url.Query(), &i); err == nil {
		t.Fatalf("expected an error")
	}
	var x struct{ A, C string }
	if err := DecodeQuery(url.Query(), x); err == nil {
		t.Fatalf("expected an error")
	}
}

func TestQuery(t *testing.T) {
	type data struct {
		B   bool
		I   int
		F32 float32
		F64 float64
		S   string
	}
	tests := []struct {
		url   string
		want  data
		iserr bool
	}{
		{"http://example.org.org", data{}, false},
		{"http://example.org.org?b=true", data{B: true}, false},
		{"http://example.org.org?b=not-a-bool", data{}, true},
		{"http://example.org.org?i=true", data{}, true},
		{"http://example.org.org?i=27", data{I: 27}, false},
		{"http://example.org.org?f32=2.7", data{F32: 2.7}, false},
		{"http://example.org.org?f32=not-a-float", data{}, true},
		{"http://example.org.org?f64=2.7", data{F64: 2.7}, false},
		{"http://example.org.org?f64=not-a-float", data{}, true},
		{"http://example.org.org?s=some%20string", data{S: "some string"}, false},
	}
	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			url, err := url.Parse(tc.url)
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			var got data
			err = DecodeQuery(url.Query(), &got)
			if tc.iserr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			if got != tc.want {
				t.Fatalf("expected %v; got %v", tc.want, got)
			}
		})
	}
}

func TestQueryArray(t *testing.T) {
	type data struct {
		B   []bool
		I   []int
		F32 []float32
		F64 []float64
		S   []string
	}
	// eq := func(a, b data) bool {
	// 	return false
	// }
	tests := []struct {
		url   string
		want  data
		iserr bool
	}{
		{"http://example.org", data{}, false},
		{"http://example.org?b=true&b=false", data{B: []bool{true, false}}, false},
		{"http://example.org?b=true&b=not-a-bool", data{}, true},
		{"http://example.org?i=1&i=2", data{I: []int{1, 2}}, false},
		{"http://example.org?i=1&i=not-an-int", data{}, true},
		{"http://example.org?f32=1.1&f32=2.1", data{F32: []float32{1.1, 2.1}}, false},
		{"http://example.org?f32=1.1&f32=not-a-float", data{}, true},
		{"http://example.org?f64=1.1&f64=2.1", data{F64: []float64{1.1, 2.1}}, false},
		{"http://example.org?f64=1.1&f64=not-a-float", data{}, true},
		{"http://example.org?s=first&s=second", data{S: []string{"first", "second"}}, false},
		{"http://example.org?s=first&i=1&s=second&i=2", data{S: []string{"first", "second"}, I: []int{1, 2}}, false},
	}
	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			url, err := url.Parse(tc.url)
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			var got data
			err = DecodeQuery(url.Query(), &got)
			if tc.iserr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("expected %v; got %v", tc.want, got)
			}

		})
	}
}
