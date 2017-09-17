package semix

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// IndexEntry denotes a public available index entry
type IndexEntry struct {
	URL        string
	Begin, End int
	Token      string
}

// Index represents the basic interface to put and get tokens from an index.
type Index interface {
	Put(Token) error
	Get(*Concept, func(IndexEntry)) error
	Flush() error
}

// NewDirIndex builds a simple directory index.
func NewDirIndex(dir string, n int) Index {
	return &dirIndex{dir: dir, n: n}
}

type dirIndex struct {
	index map[int][]IndexEntry
	dir   string
	n     int
}

func (i *dirIndex) Put(t Token) error {
	if i.index == nil {
		i.index = make(map[int][]IndexEntry)
	}
	if t.Concept == nil || t.Concept.ID() == 0 {
		return nil
	}
	id := int(t.Concept.ID())
	if id < 0 {
		id = -id
	}
	i.index[id] = append(i.index[id], IndexEntry{
		Token: t.Token,
		URL:   t.Concept.URL(),
		Begin: t.Begin,
		End:   t.End,
	})
	if len(i.index[id]) < i.n {
		return i.write(id)
	}
	return nil
}

func (i *dirIndex) write(id int) error {
	if id <= 0 {
		panic(fmt.Sprintf("dirIndex: invalid id: %d", id))
	}
	if len(i.index[id]) == 0 {
		return nil
	}
	flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	path := i.indexFile(id)
	of, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not write index: %s", path))
	}
	defer of.Close()
	e := gob.NewEncoder(of)
	logrus.Debugf("encoding id %d to %s", id, path)
	if err := e.Encode(i.index[id]); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not write index %s: %v", path, err))
	}
	logrus.Infof("encoded id %d to %s", id, path)
	return nil
}

func (i *dirIndex) Get(*Concept, func(IndexEntry)) error {
	return errors.New("Get: not implemented")
}

func (i *dirIndex) Flush() error {
	for id := range i.index {
		if err := i.write(id); err != nil {
			return err
		}
	}
	return nil
}

func (i *dirIndex) indexFile(id int) string {
	return filepath.Join(i.dir, fmt.Sprintf("%016x.gob", id))
}
