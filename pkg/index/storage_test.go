package index

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

type tmpdir struct {
	dir string
}

func openTmpdir() tmpdir {
	dir, err := ioutil.TempDir("", "semix-pkg-index")
	if err != nil {
		log.Fatal(err)
	}
	return tmpdir{dir: dir}
}

func (t tmpdir) Close() error {
	if err := os.RemoveAll(t.dir); err != nil {
		log.Fatal(err)
	}
	return nil
}

func TestStorage(t *testing.T) {
	tests := []struct {
		url string
		es  []Entry
	}{
		{"empty", []Entry{}},
		{"url1", []Entry{
			{"url1", "path1", "", "token1", 8, 10, 5, false},
			{"url1", "path2", "rel1", "token4", 8, 10, 5, true},
			{"url1", "path3", "rel2", "token3", 8, 12, 0, true},
		}},
		{"url2", []Entry{
			{"url2", "path1", "", "token1", 8, 10, 5, false},
			{"url2", "path2", "rel3", "token4", 8, 10, 0, true},
		}},
	}
	dir := openTmpdir()
	defer dir.Close()
	storage, err := OpenDirStorage(dir.dir)
	if err != nil {
		t.Fatalf("could not open storage: %v", err)
	}
	for _, tc := range tests {
		if err := storage.Put(tc.url, tc.es); err != nil {
			t.Fatalf("could not put %v into storage: %v", tc.es, err)
		}
	}
	if err := storage.Close(); err != nil {
		t.Fatalf("could not close storage: %v", err)
	}
	storage, err = OpenDirStorage(dir.dir)
	defer storage.Close()
	if err != nil {
		t.Fatalf("could not open storage: %v", err)
	}
	for _, tc := range tests {
		var es []Entry
		storage.Get(tc.url, func(e Entry) {
			es = append(es, e)
		})
		if len(es) != len(tc.es) {
			t.Fatalf("expected %d entries; got %d entries", len(tc.es), len(es))
		}
		for i := range es {
			testEntries(t, es[i], tc.es[i])
		}
	}
}
