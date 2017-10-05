package semix

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestReadStream(t *testing.T) {
	tests := []struct {
		test, want string
		iserr      bool
	}{
		{"A,B,C", "A,B,C", false},
		{"A,B,C,D,E,F,G", "A,B,C,D,E,F,G", false},
		{"error,A,B", "", true},
		{"A,error,B", "", true},
		{"A,B,error", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			ds := makeTestDocuments(tc.test)
			ctx, cancel := context.WithCancel(context.Background())
			checkStream(t, tc, cancel, Read(ctx, ds...))
		})
	}
}

func TestNormalizerStream(t *testing.T) {
	tests := []struct {
		test, want string
		iserr      bool
	}{
		{"A+B,C+D", " A B , C D ", false},
		{"  A   B   ", " A B ", false},
		{" A?+B?", " A B ", false},
		{"A?+CB?? ,B", " A CB , B ", false},
		{"A?,error,B", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			ds := makeTestDocuments(tc.test)
			ctx, cancel := context.WithCancel(context.Background())
			checkStream(t, tc, cancel, Normalize(ctx, Read(ctx, ds...)))
		})
	}
}

func TestMatcherStream(t *testing.T) {
	tests := []struct {
		test, want string
		iserr      bool
	}{
		{"AmatchB", "A,match,B", false},
		{"A match B", "A ,match, B", false},
		{"match A", "match, A", false},
		{"match", "match", false},
		{"A B", "A B", false},
		{"error,A match B", "", true},
		{"A match B,error", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			ds := makeTestDocuments(tc.test)
			ctx, cancel := context.WithCancel(context.Background())
			checkStream(t, tc, cancel, Match(ctx, testm{}, Read(ctx, ds...)))
		})
	}
}

func TestFilterStream(t *testing.T) {
	tests := []struct {
		test, want string
		iserr      bool
	}{
		{"AmatchB", "match", false},
		{"Amatch,error", "match", true},
		{"error,A match", "match", true},
	}
	for _, tc := range tests {
		t.Run(tc.test, func(t *testing.T) {
			ds := makeTestDocuments(tc.test)
			ctx, cancel := context.WithCancel(context.Background())
			checkStream(t, tc, cancel, Filter(ctx, Match(ctx, testm{}, Read(ctx, ds...))))
		})
	}
}

func TestStreamCancel(t *testing.T) {
	ds := makeTestDocuments("A,B,C,D")
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Wait()
		s := Filter(ctx, Match(ctx, testm{}, Normalize(ctx, Read(ctx, ds...))))
		for token := range s {
			t.Fatalf("sould not read token %v", token)
		}
	}()
	cancel()
	wg.Done()
}

func checkStream(
	t *testing.T,
	tc struct {
		test, want string
		iserr      bool
	},
	cancel context.CancelFunc,
	s Stream,
) {
	t.Helper()
	res, err := stream2string(cancel, s)
	if tc.iserr && err == nil {
		t.Fatalf("expceted error; got %q %v", res, err)
	}
	if !tc.iserr && err != nil {
		t.Fatalf("got error: %v", err)
	}
	if !tc.iserr && !equals(res, tc.want) {
		t.Fatalf("expceted %q; got %q", tc.want, res)
	}
}

func equals(t, w string) bool {
	want := strings.Split(w, ",")
	test := strings.Split(t, ",")
	if len(test) != len(want) {
		return false
	}
	contains := func(strs []string, str string) bool {
		for _, test := range strs {
			if test == str {
				return true
			}
		}
		return false
	}
	for _, str := range test {
		if !contains(want, str) {
			return false
		}
	}
	return true
}

func stream2string(cancel context.CancelFunc, s Stream) (string, error) {
	defer cancel()
	var strs []string
	for t := range s {
		if t.Err != nil {
			return "", t.Err
		}
		strs = append(strs, string(t.Token.Token))
	}
	return strings.Join(strs, ","), nil
}

func makeTestDocuments(str string) []Document {
	var ds []Document
	for _, s := range strings.Split(str, ",") {
		if s == "error" {
			ds = append(ds, errDoc{})
		} else {
			ds = append(ds, NewStringDocument(s, s))
		}
	}
	return ds
}

type errDoc struct{}

func (errDoc) Close() error             { return nil }
func (errDoc) Path() string             { return "errdoc" }
func (errDoc) Read([]byte) (int, error) { return 0, errors.New("errdoc") }

type testm struct{}

func (testm) Match(str string) MatchPos {
	i := strings.Index(str, string("match"))
	if i < 0 {
		return MatchPos{End: len("match")}
	}
	return MatchPos{Begin: i, End: i + len("match"), Concept: &Concept{url: "match"}}
}
