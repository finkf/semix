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
		ioutil.WriteFile(gold, got, 0644)
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
	}
	for _, tc := range tests {
		t.Run(tc.uri, func(t *testing.T) {
			d, err := NewFromURI(tc.uri, tc.ct)
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			defer func() { _ = d.Close() }()
			compareWithGold(t, tc.gold, d)
		})
	}
}
