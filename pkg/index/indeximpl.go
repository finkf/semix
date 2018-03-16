package index

import (
	"sync"

	"bitbucket.org/fflo/semix/pkg/semix"
	"github.com/pkg/errors"
)

const (
	// DefaultBufferSize is the default buffer size.
	DefaultBufferSize = 1024
)

// NewMemoryMap create a new in memory index, that uses a simple map
// of Entry slices for storage. It is a shortcut for
// NewIndexStorage(OpenMemStorage()).
func NewMemoryMap(n int) Interface {
	return NewIndexStorage(OpenMemStorage(), n)
}

// NewIndexStorage returns a new Interface with a given buffer size
// and storage.
func NewIndexStorage(s Storage, n int) Interface {
	return &index{
		storage: s,
		buffer:  make(map[string][]Entry),
		mutex:   new(sync.RWMutex),
		n:       n,
		pool: &sync.Pool{New: func() interface{} {
			return make([]Entry, 0, n)
		}},
	}
}

// New opens a directory index at the given directory path with
// and the given options.
func New(dir string, size int) (Interface, error) {
	storage, err := OpenDirStorage(dir)
	if err != nil {
		return nil, err
	}
	return NewIndexStorage(storage, size), nil
}

type index struct {
	storage Storage
	buffer  map[string][]Entry
	pool    *sync.Pool
	mutex   *sync.RWMutex
	n       int
}

func (i *index) newBuffer() interface{} {
	return make([]Entry, 0, i.n)
}

func (i *index) putBuffer(url string) {
	buf := i.buffer[url]
	if buf == nil {
		return
	}
	i.buffer[url] = nil
	buf = buf[:0]
	i.pool.Put(buf)
}

func (i *index) getBuffer(url string) {
	if i.buffer[url] != nil {
		return
	}
	i.buffer[url] = i.pool.Get().([]Entry)
}

// Put puts a token in the index.
func (i *index) Put(t semix.Token) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	return putAll(t, func(e Entry) error {
		url := e.ConceptURL
		i.getBuffer(url)
		i.buffer[url] = append(i.buffer[url], e)
		if len(i.buffer[url]) == i.n {
			if err := i.storage.Put(url, i.buffer[url]); err != nil {
				return errors.Wrapf(err, "cannot put entries")
			}
			i.putBuffer(url)
		}
		return nil
	})
}

// Get queries the index for a concept and calls the callback function
// for each entry in the index.
func (i *index) Get(url string, f func(Entry) bool) error {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	for _, e := range i.buffer[url] {
		if !f(e) {
			return nil
		}
	}
	return i.storage.Get(url, f)
}

// Flush flushes the index.
// All non empty buffers are written to the storage.
func (i *index) Flush() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	return i.putAll()
}

// putAll puts all non empty buffers into the index.
// Must be called with a locked mutex.
func (i *index) putAll() error {
	for url, es := range i.buffer {
		if len(es) == 0 {
			continue
		}
		if err := i.storage.Put(url, es); err != nil {
			return errors.Wrapf(err, "cannot write index buffer")
		}
		i.buffer[url] = i.buffer[url][:0]
	}
	return nil
}

// Close closes the index and writes all buffered entries to disc.
func (i *index) Close() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if err := i.putAll(); err != nil {
		return errors.Wrapf(err, "cannot close index")
	}
	if err := i.storage.Close(); err != nil {
		return errors.Wrapf(err, "cannot close index")
	}
	return nil
}

func putAll(t semix.Token, f func(Entry) error) error {
	if t.Concept.Ambig() {
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
		L:          k,
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
			L:           k,
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
	if !c.Ambig() {
		panic("putAllAmbiguous called with non ambiguous concept")
	}
	n := c.EdgesLen()
	for i := 0; i < n; i++ {
		e := c.EdgeAt(i)
		t.Concept = e.O
		l := e.L
		if l == 0 {
			l = -1
		}
		if err := putAllWithError(t, l, f); err != nil {
			return err
		}
	}
	return nil
}
