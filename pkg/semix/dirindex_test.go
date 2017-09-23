package semix

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func makeRandomDirIndexEntries() []dirIndexEntry {
	n := rand.Intn(20) + 1
	es := make([]dirIndexEntry, n)
	for i := range es {
		es[i] = dirIndexEntry{
			S: fmt.Sprintf("random string %d", rand.Int()),
			P: fmt.Sprintf("random path %d", rand.Int()),
			B: rand.Int(),
			E: rand.Int(),
			R: rand.Int(),
			O: rand.Int(),
		}
	}
	return es
}

func TestDirIndexReadWrite(t *testing.T) {
	b := new(bytes.Buffer)
	seed := time.Now().Unix()
	rand.Seed(seed)
	want := make([][]dirIndexEntry, 10)
	for i := 0; i < 10; i++ {
		want[i] = makeRandomDirIndexEntries()
		if err := writeDirIndexEntries(b, want[i]); err != nil {
			t.Fatalf("got error: %v (seed = %d)", err, seed)
		}
	}
	var test [][]dirIndexEntry
	for {
		es, err := readDirIndexEntries(b)
		if err != nil {
			t.Fatalf("got error: %v (seed = %d)", err, seed)
		}
		if len(es) == 0 {
			break
		}
		test = append(test, es)
	}
	wantstr := fmt.Sprintf("%v", want)
	teststr := fmt.Sprintf("%v", test)
	if wantstr != teststr {
		t.Fatalf("expected %q; got %q", wantstr, teststr)
	}
}
