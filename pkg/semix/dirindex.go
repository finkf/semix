package semix

import (
	"context"
	"encoding/gob"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type DirIndexOpt func(*dirIndex)

func WithBufferSize(n int) DirIndexOpt {
	return func(i *dirIndex) {
		i.n = n
	}
}

const (
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
	dir      string
	n        int
}

type dirIndexEntry struct {
	Token                 string
	Begin, End, Rel, Orig int
}

type dirIndexData struct {
	buffer   map[int][]dirIndexEntry
	register *URLRegister
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
		case t := <-i.put:
			i.putError(i.putToken(data, t))
		}
	}
}

func (i *dirIndex) putError(err error) {
	if err == nil {
		return
	}
	select {
	case i.err <- err:
		return
	default:
		// drop it
	}
}

func (i *dirIndex) getError() error {
	select {
	case err := <-i.err:
		return err
	default:
		return nil
	}
}

func (i *dirIndex) putToken(data dirIndexData, t Token) error {
	e := dirIndexEntry{
		Token: t.Token,
		Begin: t.Begin,
		End:   t.End,
	}
	id := data.register.Register(t.Concept.URL())
	if err := i.putEntry(data, id, e); err != nil {
		return err
	}
	for _, edge := range t.Concept.edges {
		e.Orig = id
		e.Rel = data.register.Register(edge.P.URL())
		oid := data.register.Register(edge.O.URL())
		if err := i.putEntry(data, oid, e); err != nil {
			return err
		}
	}
	return nil
}

func (i *dirIndex) putEntry(data dirIndexData, id int, e dirIndexEntry) error {
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
	return filepath.Join(dir, url.PathEscape(u))
}

// Put puts a token in the index.
func (i *dirIndex) Put(t Token) error {
	if err := i.getError(); err != nil {
		return err
	}
	i.put <- t
	return i.getError()
}

func (i *dirIndex) Close() error {
	return errors.New("not implemented")
}

func (i *dirIndex) Get(*Concept, func(Token)) error {
	return errors.New("not implemented")
}
