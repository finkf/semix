package reader

import (
	"io/ioutil"
	"testing"
)

func TestPlainTextDocument(t *testing.T) {
	tests := []struct{ gold, want string }{
		{"testdata/plain_text.txt", "this is a plain text file\n"},
	}
	for _, tc := range tests {
		t.Run(tc.gold, func(t *testing.T) {
			d, err := NewFromURI(tc.gold, PlainText)
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			defer func() { _ = d.Close() }()
			content, err := ioutil.ReadAll(d)
			if err != nil {
				t.Fatalf("got error: %s", err)
			}
			if got := string(content); tc.want != got {
				t.Fatalf("expected %q; got %q", tc.want, got)
			}
		})
	}
}
