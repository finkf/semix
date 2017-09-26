package semix

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// IndexStorage puts IndexEntries into files.
type IndexStorage interface {
	Put(string, []IndexEntry) error
	Get(string, func(IndexEntry)) error
	Close() error
}

type dirIndexStorage struct {
	dir      string
	register *URLRegister
}

// OpenDirIndexStorage opens a new IndexStorage.
func OpenDirIndexStorage(dir string) (IndexStorage, error) {
	s := dirIndexStorage{dir: dir, register: NewURLRegister()}
	path := s.urlRegisterPath()
	is, err := os.Open(path)
	if err != nil {
		// ignore io errors
		return s, nil
	}
	defer is.Close()
	d := gob.NewDecoder(is)
	if err := d.Decode(s.register); err != nil {
		return dirIndexStorage{}, errors.Wrapf(err, "could not decode %q", path)
	}
	return s, nil
}

func (s dirIndexStorage) Put(url string, es []IndexEntry) error {
	if len(es) == 0 {
		return nil
	}
	ds := make([]dsentry, len(es))
	for i := range es {
		ds[i] = dsentry{
			S: es[i].Token,
			P: es[i].Path,
			B: es[i].Begin,
			E: es[i].End,
		}
		if es[i].OriginURL != "" {
			ds[i].O = s.register.Register(es[i].OriginURL)
		}
		if es[i].RelationURL != "" {
			ds[i].R = s.register.Register(es[i].RelationURL)
		}
	}
	return s.write(url, ds)
}

func (s dirIndexStorage) write(url string, ds []dsentry) error {
	path := s.path(url)
	flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	os, err := os.OpenFile(path, flags, 0666)
	if err != nil {
		return errors.Wrapf(err, "could not open %q", path)
	}
	defer os.Close()
	if err := writeBlock(os, ds); err != nil {
		return errors.Wrapf(err, "could not encode to %q", path)
	}
	return nil
}

func (s dirIndexStorage) Get(url string, f func(IndexEntry)) error {
	path := s.path(url)
	is, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "could not open %q", path)
	}
	defer is.Close()
	for {
		ds, err := readBlock(is)
		logrus.Infof("BLOCK: %v (%v)", ds, err)
		if err != nil {
			return errors.Wrapf(err, "could not decode %q", path)
		}
		if len(ds) == 0 {
			return nil
		}
		for _, d := range ds {
			f(IndexEntry{
				ConceptURL:  url,
				RelationURL: s.lookup(d.R),
				OriginURL:   s.lookup(d.O),
				Token:       d.S,
				Path:        d.P,
				Begin:       d.B,
				End:         d.E,
			})
		}
	}
}

func (s dirIndexStorage) Close() error {
	path := s.urlRegisterPath()
	os, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "could not write %q", path)
	}
	defer os.Close()
	e := gob.NewEncoder(os)
	return e.Encode(s.register)
}

func (s dirIndexStorage) path(url string) string {
	return filepath.Join(s.dir, escapeURL(url)+".gob")
}

func (s dirIndexStorage) urlRegisterPath() string {
	return filepath.Join(s.dir, escapeURL("http://bitbucket.org/fflo/semix/url-register")+".gob")
}

func (s dirIndexStorage) lookup(id int) string {
	if url, ok := s.register.LookupID(id); ok {
		return url
	}
	return ""
}

func writeBlock(w io.Writer, ds []dsentry) error {
	buffer := new(bytes.Buffer)
	e := gob.NewEncoder(buffer)
	if err := e.Encode(ds); err != nil {
		return err
	}
	header := int64(len(buffer.Bytes()))
	if err := binary.Write(w, binary.BigEndian, header); err != nil {
		return err
	}
	if _, err := w.Write(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}

func readBlock(r io.Reader) ([]dsentry, error) {
	var header int64
	if err := binary.Read(r, binary.BigEndian, &header); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}
	buffer := make([]byte, header)
	if _, err := r.Read(buffer); err != nil {
		return nil, err
	}
	dec := bytes.NewBuffer(buffer)
	d := gob.NewDecoder(dec)
	var ds []dsentry
	err := d.Decode(&ds)
	return ds, err
}

// Short var names for smaller gob indices.
// S is the string
// P is the document path
// B is the start position
// E is the end position
// R is the relation id
// O is the origin id
type dsentry struct {
	S, P       string
	B, E, R, O int
}

func escapeURL(u string) string {
	u = strings.Replace(u, "http://", "", 1)
	u = strings.Replace(u, "https://", "", 1)
	u = strings.Map(func(r rune) rune {
		if r == '/' {
			return '.'
		}
		return r
	}, u)
	return url.PathEscape(u)
}
