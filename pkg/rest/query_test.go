package rest

import (
	"net/url"
	"testing"
)

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
