package semix

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
func OpenDirIndex(dir string, opts ...DirIndexOpt) Index {
	ctx, cancel := context.WithCancel(context.Background())
	i := &dirIndex{
		register: NewURLRegister(),
		dir:      dir,
		n:        DefaultIndexDirBufferSize,
		put:      make(chan Token),
		get:      make(chan dirIndexQuery),
		err:      make(chan error),
		cancel:   cancel,
	}
	for _, opt := range opts {
		opt(i)
	}
	go i.start(ctx)
	return i
}

type dirIndex struct {
	register *URLRegister
	ctx      context.Context
	cancel   context.CancelFunc
	err      chan error
	put      chan Token
	get      chan dirIndexQuery
	dir      string
	n        int
}

// Short var names for smaller gob indices.
// S is the string
// P is the document path
// B is the start position
// E is the end position
// R is the relation id
// O is the origin id
type dirIndexEntry struct {
	S, P       string
	B, E, R, O int
}

type dirIndexQuery struct {
	f   func(IndexEntry)
	url string
}

type dirIndexData struct {
	buffer   map[int][]dirIndexEntry
	register *URLRegister
}

// Put puts a token in the index.
func (i *dirIndex) Put(t Token) error {
	if err := i.getError(); err != nil {
		return err
	}
	i.put <- t
	return i.getError()
}

// Get queries the index for a concept and calls the callback function
// for each entry in the index.
func (i *dirIndex) Get(c *Concept, f func(IndexEntry)) error {
	if c == nil {
		return nil
	}
	if err := i.getError(); err != nil {
		return err
	}
	i.get <- dirIndexQuery{url: c.URL(), f: f}
	return i.getError()
}

// Close closes the index and writes all buffered entries to disc.
func (i *dirIndex) Close() error {
	return errors.New("not implemented")
}

func (i *dirIndex) start(ctx context.Context) {
	data := dirIndexData{
		buffer:   make(map[int][]dirIndexEntry),
		register: NewURLRegister(),
	}
	for {
		select {
		case <-ctx.Done():
			return
		case q := <-i.get:
			if err := i.getEntries(data, q); err != nil {
				logrus.Infof("i.getEntries error: %v", err)
				i.putError(err)
			}
		case t := <-i.put:
			if err := i.putToken(data, t); err != nil {
				logrus.Infof("i.putToken error: %v", err)
				i.putError(err)
			}
		}
	}
}

func (i *dirIndex) putError(err error) {
	if err == nil {
		// logrus.Infof("not putting a nil error")
		return
	}
	select {
	case i.err <- err:
		// logrus.Infof("put error: %v", err)
		return
	default:
		// logrus.Infof("put nothing")
		// drop it
	}
}

func (i *dirIndex) getError() error {
	select {
	case err := <-i.err:
		// logrus.Infof("got error: %v", err)
		return err
	default:
		// logrus.Infof("got nothing")
		return nil
	}
}

func (i *dirIndex) putToken(data dirIndexData, t Token) error {
	e := dirIndexEntry{
		S: t.Token,
		P: t.Path,
		B: t.Begin,
		E: t.End,
	}
	id := data.register.Register(t.Concept.URL())
	if err := i.putEntry(data, id, e); err != nil {
		return err
	}
	e.O = id
	for _, edge := range t.Concept.edges {
		e.R = data.register.Register(edge.P.URL())
		oid := data.register.Register(edge.O.URL())
		if err := i.putEntry(data, oid, e); err != nil {
			return err
		}
	}
	return nil
}

func (i *dirIndex) putEntry(data dirIndexData, id int, e dirIndexEntry) error {
	// logrus.Infof("putting entry %v (id: %d)", e, id)
	data.buffer[id] = append(data.buffer[id], e)
	if len(data.buffer[id]) == i.n {
		if err := i.write(data, id); err != nil {
			return err
		}
	}
	return nil
}

func (i *dirIndex) write(data dirIndexData, id int) error {
	if len(data.buffer[id]) == 0 {
		return nil
	}
	url, ok := data.register.LookupID(id)
	if !ok {
		return fmt.Errorf("invalid internal id: %d", id)
	}
	path := getFilenameFromURL(i.dir, url)
	flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	os, err := os.OpenFile(path, flags, 0666)
	if err != nil {
		return errors.Wrapf(err, "could not open %q", path)
	}
	defer os.Close()
	e := gob.NewEncoder(os)
	if err := e.Encode(data.buffer[id]); err != nil {
		return errors.Wrapf(err, "could not encode %q", path)
	}
	// clear the buffer
	data.buffer[id] = data.buffer[id][:0]
	return nil
}

func getFilenameFromURL(dir, u string) string {
	return filepath.Join(dir, url.PathEscape(u)+".gob")
}

func (i *dirIndex) getEntries(data dirIndexData, q dirIndexQuery) error {
	// handle entries that are still in the buffer.
	id, _ := data.register.LookupURL(q.url)
	if es, ok := data.buffer[id]; ok {
		for _, e := range es {
			ie, err := makeIndexEntry(data.register, q.url, e)
			if err != nil {
				return err
			}
			q.f(ie)
		}
	}
	// open file
	path := getFilenameFromURL(i.dir, q.url)
	is, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "could not open %q", path)
	}
	defer is.Close()
	if err := i.getEntriesReader(data, is, q); err != nil {
		return errors.Wrapf(err, "could not read %q", path)
	}
	return nil
}

func (i *dirIndex) getEntriesReader(data dirIndexData, r io.Reader, q dirIndexQuery) error {
	var es []dirIndexEntry
	for {
		logrus.Infof("start of for")
		d := gob.NewDecoder(r)
		logrus.Infof("decoding")
		if err := d.Decode(&es); err != nil {
			logrus.Infof("returning error %v", err)
			return errors.Wrap(err, "could not decode")
		}
		logrus.Infof("decoded: %v", es)
		logrus.Infof("decoded: %d", len(es))
		if len(es) == 0 {
			break
		}
		logrus.Infof("here")
		for _, e := range es {
			ie, err := makeIndexEntry(data.register, q.url, e)
			if err != nil {
				logrus.Infof("returning error")
				return err
			}
			q.f(ie)
		}
		logrus.Infof("end of for")
	}
	return nil
}

func makeIndexEntry(register *URLRegister, url string, e dirIndexEntry) (IndexEntry, error) {
	// logrus.Infof("entry: %v", e)
	ie := IndexEntry{
		ConceptURL: url,
		Token:      e.S,
		Path:       e.P,
		Begin:      e.B,
		End:        e.E,
	}
	// direct hit
	if e.O == 0 && e.R == 0 {
		// logrus.Infof("ientry: %v", ie)
		return ie, nil
	}

	O, ok := register.LookupID(e.O)
	// logrus.Infof("%s %t", O, ok)
	if !ok {
		return ie, fmt.Errorf("invalid internal id: %d", e.O)
	}
	R, ok := register.LookupID(e.R)
	// logrus.Infof("%s %t", R, ok)
	if !ok {
		return ie, fmt.Errorf("invalid internal id: %d", e.R)
	}
	ie.OriginRelationURL = R
	ie.OriginURL = O
	// logrus.Infof("ientry: %v", ie)
	return ie, nil
}
