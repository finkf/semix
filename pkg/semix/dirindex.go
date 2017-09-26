package semix

import (
	"context"
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
	storage, err := OpenDirIndexStorage(dir)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	i := &dirIndex{
		storage: storage,
		n:       DefaultIndexDirBufferSize,
		buffer:  make(map[string][]IndexEntry),
		put:     make(chan Token),
		get:     make(chan dirIndexQuery),
		err:     make(chan error),
		cancel:  cancel,
	}
	for _, opt := range opts {
		opt(i)
	}
	go i.run(ctx)
	return i, nil
}

type dirIndex struct {
	storage IndexStorage
	buffer  map[string][]IndexEntry
	cancel  context.CancelFunc
	err     chan error
	put     chan Token
	get     chan dirIndexQuery
	dir     string
	n       int
}

type dirIndexQuery struct {
	f   func(IndexEntry)
	url string
}

// Put puts a token in the index.
func (i *dirIndex) Put(t Token) error {
	i.put <- t
	return <-i.err
}

// Get queries the index for a concept and calls the callback function
// for each entry in the index.
func (i *dirIndex) Get(url string, f func(IndexEntry)) error {
	i.get <- dirIndexQuery{url: url, f: f}
	return <-i.err
}

// Close closes the index and writes all buffered entries to disc.
func (i *dirIndex) Close() error {
	i.cancel()
	close(i.put)
	close(i.get)
	close(i.err)
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

func (i *dirIndex) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case q, ok := <-i.get:
			if !ok {
				return
			}
			i.err <- i.getEntries(q.url, q.f)
		case t, ok := <-i.put:
			if !ok {
				return
			}
			i.err <- i.putToken(t)
		}
	}
}

func (i *dirIndex) putToken(t Token) error {
	url := t.Concept.URL()
	i.buffer[url] = append(i.buffer[url], IndexEntry{
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
	for _, edge := range t.Concept.edges {
		objurl := edge.O.URL()
		i.buffer[objurl] = append(i.buffer[objurl], IndexEntry{
			ConceptURL:  objurl,
			Begin:       t.Begin,
			End:         t.End,
			Path:        t.Path,
			Token:       t.Token,
			RelationURL: edge.P.URL(),
			OriginURL:   url,
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

func (i *dirIndex) getEntries(url string, f func(IndexEntry)) error {
	for _, e := range i.buffer[url] {
		f(e)
	}
	return i.storage.Get(url, f)
}
