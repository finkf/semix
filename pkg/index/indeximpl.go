package index

import (
	"context"

	"bitbucket.org/fflo/semix/pkg/semix"
)

// DirIndexOpt defines a functional argument setter.
type DirIndexOpt func(*dirIndex)

// WithBufferSize sets the optional buffer size of the directory index.
func WithBufferSize(n int) DirIndexOpt {
	return func(i *dirIndex) {
		i.n = n
	}
}

const (
	// DefaultIndexDirBufferSize is the default buffer size.
	DefaultIndexDirBufferSize = 1024
)

// OpenDirIndex opens a directory index at the given directory path with
// and the given options.
func OpenDirIndex(dir string, opts ...DirIndexOpt) (Index, error) {
	storage, err := OpenDirStorage(dir)
	if err != nil {
		return nil, err
	}
	i := &dirIndex{
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

type dirIndex struct {
	storage Storage
	buffer  map[string][]Entry
	cancel  context.CancelFunc
	put     chan putRequest
	get     chan getRequest
	dir     string
	n       int
}

// Put puts a token in the index.
func (i *dirIndex) Put(t semix.Token) error {
	err := make(chan error)
	i.put <- putRequest{token: t, err: err}
	return <-err
}

// Get queries the index for a concept and calls the callback function
// for each entry in the index.
func (i *dirIndex) Get(url string, f func(Entry)) error {
	err := make(chan error)
	i.get <- getRequest{url: url, f: f, err: err}
	return <-err
}

// Close closes the index and writes all buffered entries to disc.
func (i *dirIndex) Close() error {
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

func (i *dirIndex) run() {
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

func (i *dirIndex) putToken(t semix.Token) error {
	url := t.Concept.URL()
	i.buffer[url] = append(i.buffer[url], Entry{
		ConceptURL: url,
		Begin:      t.Begin,
		End:        t.End,
		Path:       t.Path,
		Token:      t.Token,
	})
	if len(i.buffer[url]) == i.n {
		if err := i.storage.Put(url, i.buffer[url]); err != nil {
			return err
		}
		i.buffer[url] = nil
	}
	for j := 0; j < t.Concept.EdgesLen(); j++ {
		edge := t.Concept.EdgeAt(j)
		objurl := edge.O.URL()
		i.buffer[objurl] = append(i.buffer[objurl], Entry{
			ConceptURL:  objurl,
			Begin:       t.Begin,
			End:         t.End,
			Path:        t.Path,
			Token:       t.Token,
			RelationURL: edge.P.URL(),
		})
		if len(i.buffer[objurl]) == i.n {
			if err := i.storage.Put(objurl, i.buffer[objurl]); err != nil {
				return err
			}
			i.buffer[objurl] = nil
		}
	}
	return nil
}

func (i *dirIndex) getEntries(url string, f func(Entry)) error {
	for _, e := range i.buffer[url] {
		f(e)
	}
	return i.storage.Get(url, f)
}
