package index

import (
	"context"

	"bitbucket.org/fflo/semix/pkg/semix"
)

// DirIndexOpt defines a functional argument setter.
type DirIndexOpt func(*index)

// WithBufferSize sets the optional buffer size of the directory index.
func WithBufferSize(n int) DirIndexOpt {
	return func(i *index) {
		i.n = n
	}
}

const (
	// DefaultIndexDirBufferSize is the default buffer size.
	DefaultIndexDirBufferSize = 1024
)

// New opens a directory index at the given directory path with
// and the given options.
func New(dir string, opts ...DirIndexOpt) (Index, error) {
	storage, err := OpenDirStorage(dir)
	if err != nil {
		return nil, err
	}
	i := &index{
		storage: storage,
		n:       DefaultIndexDirBufferSize,
		buffer:  make(map[string][]Entry),
		put:     make(chan putRequest),
		get:     make(chan getRequest),
	}
	for _, opt := range opts {
		opt(i)
	}
	go i.run()
	return i, nil
}

type putRequest struct {
	token semix.Token
	err   chan<- error
}

type getRequest struct {
	url string
	f   func(Entry)
	err chan<- error
}

type index struct {
	storage Storage
	buffer  map[string][]Entry
	cancel  context.CancelFunc
	put     chan putRequest
	get     chan getRequest
	dir     string
	n       int
}

// Put puts a token in the index.
func (i *index) Put(t semix.Token) error {
	err := make(chan error)
	i.put <- putRequest{token: t, err: err}
	return <-err
}

// Get queries the index for a concept and calls the callback function
// for each entry in the index.
func (i *index) Get(url string, f func(Entry)) error {
	err := make(chan error)
	i.get <- getRequest{url: url, f: f, err: err}
	return <-err
}

// Close closes the index and writes all buffered entries to disc.
func (i *index) Close() error {
	i.cancel()
	close(i.put)
	close(i.get)
	defer i.storage.Close()
	for url, es := range i.buffer {
		if len(es) == 0 {
			continue
		}
		if err := i.storage.Put(url, es); err != nil {
			return err
		}
	}
	return nil
}

func (i *index) run() {
	for {
		select {
		case r, ok := <-i.get:
			if !ok {
				return
			}
			r.err <- i.getEntries(r.url, r.f)
		case r, ok := <-i.put:
			if !ok {
				return
			}
			r.err <- i.putToken(r.token)
		}
	}
}

func (i *index) putToken(t semix.Token) error {
	return putAll(t, func(e Entry) error {
		url := e.ConceptURL
		i.buffer[url] = append(i.buffer[url], e)
		if len(i.buffer[url]) == i.n {
			if err := i.storage.Put(url, i.buffer[url]); err != nil {
				return err
			}
			i.buffer[url] = nil
		}
		return nil
	})
}

func (i *index) getEntries(url string, f func(Entry)) error {
	for _, e := range i.buffer[url] {
		f(e)
	}
	return i.storage.Get(url, f)
}

// NewMapIndex create a new in memory index, that uses
// a simple map of Entry slices for storage.
func NewMapIndex() Index {
	return mapIndex{index: make(map[string][]Entry)}
}

type mapIndex struct {
	index map[string][]Entry
}

func (i mapIndex) Put(t semix.Token) error {
	return putAll(t, func(e Entry) error {
		url := e.ConceptURL
		i.index[url] = append(i.index[url], e)
		return nil
	})
}

func (i mapIndex) Get(url string, f func(Entry)) error {
	for _, e := range i.index[url] {
		f(e)
	}
	return nil
}

func (i mapIndex) Close() error {
	return nil
}

func putAll(t semix.Token, f func(Entry) error) error {
	if t.Concept.Ambiguous() {
		return putAllAmbiguous(t, f)
	}
	return putAllWithError(t, 0, f)
}

// putAllWithError converts a semix.Token to an Entry and
// calls the callback function recursively for all
// connected concepts.
func putAllWithError(t semix.Token, k int, f func(Entry) error) error {
	url := t.Concept.URL()
	err := f(Entry{
		ConceptURL: url,
		Begin:      t.Begin,
		End:        t.End,
		Path:       t.Path,
		Token:      t.Token,
		K:          k,
	})
	if err != nil {
		return err
	}
	n := t.Concept.EdgesLen()
	for i := 0; i < n; i++ {
		edge := t.Concept.EdgeAt(i)
		objurl := edge.O.URL()
		err := f(Entry{
			ConceptURL:  objurl,
			Begin:       t.Begin,
			End:         t.End,
			Path:        t.Path,
			Token:       t.Token,
			RelationURL: edge.P.URL(),
			K:           k,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// putAllAmbiguous handles tokens with ambiguous concepts.
func putAllAmbiguous(t semix.Token, f func(Entry) error) error {
	c := t.Concept
	if !c.Ambiguous() {
		panic("putAllAmbiguous called with non ambiguous concept")
	}
	n := c.EdgesLen()
	for i := 0; i < n; i++ {
		e := c.EdgeAt(i)
		t.Concept = e.O
		if err := putAllWithError(t, e.L, f); err != nil {
			return err
		}
	}
	return nil
}
