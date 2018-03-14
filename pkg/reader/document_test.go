package reader

import (
	"flag"
	"io"
	"io/ioutil"
	"testing"
)

var update = flag.Bool("update", false, "update gold files")

func compareWithGold(t *testing.T, gold string, r io.Reader) {
	t.Helper()
	got, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("got error: %s", err)
	}
	if *update {
		if err := ioutil.WriteFile(gold, got, 0644); err != nil {
			panic(err)
		}
	}
	want, err := ioutil.ReadFile(gold)
	if err != nil {
		panic(err)
	}
	if string(got) != string(want) {
		t.Fatalf("expected %q; got %q", want, got)
	}
}

func TestFileDocuments(t *testing.T) {
	tests := []struct{ uri, gold, ct string }{
		{"testdata/plain_text.txt", "testdata/plain_text.txt", PlainText},
		{"testdata/example.org.html", "testdata/example.org.html.gold", HTML},
		{"http://example.org", "testdata/example.org.html.gold", HTTP},
	}
	for _, tc := range tests {
		t.Run(tc.uri, func(t *testing.T) {
			d, err := NewFromURI(tc.uri, tc.ct)
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			defer func() { _ = d.Close() }()
			if got := d.Path(); got != tc.uri {
				t.Fatalf("expected %q; got %q", tc.uri, got)
			}
			compareWithGold(t, tc.gold, d)
		})
	}
}
