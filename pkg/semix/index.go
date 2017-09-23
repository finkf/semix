package semix

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// IndexEntry denotes a public available index entry
type IndexEntry struct {
	ConceptURL, Path, OriginURL, OriginRelationURL string
	Begin, End                                     int
	Token                                          string
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
	e := IndexEntry{
		Token:      t.Token,
		ConceptURL: t.Concept.URL(),
		Path:       t.Path,
		Begin:      t.Begin,
		End:        t.End,
	}
	i.doPut(id, e)
	t.Concept.Edges(func(edge Edge) {
		e.OriginURL = t.Concept.URL()
		e.ConceptURL = edge.O.URL()
		e.OriginRelationURL = edge.P.URL()
		i.doPut(int(edge.O.ID()), e)
	})
	return nil
}

func (i *dirIndex) doPut(id int, e IndexEntry) error {
	i.index[id] = append(i.index[id], e)
	if len(i.index[id]) >= i.n {
		return i.write(i.index[id])
	}
	return nil
}

func (i *dirIndex) write(es []IndexEntry) error {
	if len(es) == 0 {
		return nil
	}
	flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	path := i.indexFile(es[0].ConceptURL)
	of, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not write index: %s", path))
	}
	defer of.Close()
	e := gob.NewEncoder(of)
	logrus.Debugf("encoding entries to %s", path)
	if err := e.Encode(es); err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not write index %s: %v", path, err))
	}
	logrus.Infof("encoded entries to %s", path)
	// i.index
	return nil
}

func (i *dirIndex) Get(*Concept, func(IndexEntry)) error {
	return errors.New("Get: not implemented")
}

func (i *dirIndex) Flush() error {
	for _, arr := range i.index {
		if err := i.write(arr); err != nil {
			return err
		}
	}
	return nil
}

func (i *dirIndex) indexFile(url string) string {
	if i := strings.LastIndex(url, "/"); i != -1 {
		url = url[i+1:]
	}
	return filepath.Join(i.dir, url+".gob")
}
