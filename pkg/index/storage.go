package index

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/fflo/semix/pkg/semix"
)

// Storage puts IndexEntries into files.
type Storage interface {
	Put(string, []Entry) error
	Get(string, func(Entry)) error
	Close() error
}

type dirStorage struct {
	dir      string
	register *semix.URLRegister
}

// OpenDirStorage opens a new IndexStorage.
func OpenDirStorage(dir string) (Storage, error) {
	s := dirStorage{dir: dir, register: semix.NewURLRegister()}
	path := s.urlRegisterPath()
	is, err := os.Open(path)
	if err != nil {
		// ignore io errors
		// log.Printf("ignoring error: %v", err)
		return s, nil
	}
	defer is.Close()
	d := gob.NewDecoder(is)
	if err := d.Decode(s.register); err != nil {
		return dirStorage{}, fmt.Errorf("could not decod %q: %v", path, err)
	}
	return s, nil
}

func (s dirStorage) Put(url string, es []Entry) error {
	if len(es) == 0 {
		return nil
	}
	ds := make([]dse, len(es))
	for i := range es {
		ds[i] = dse{
			S: es[i].Token,
			P: es[i].Path,
			B: uint32(es[i].Begin),
			E: uint32(es[i].End),
			L: encodeL(es[i].L, es[i].Ambiguous),
		}
		if es[i].RelationURL != "" {
			ds[i].R = int32(s.register.Register(es[i].RelationURL))
		}
	}
	return s.write(url, ds)
}

func (s dirStorage) write(url string, ds []dse) error {
	path := s.path(url)
	flags := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	// log.Printf("wrting %d entries to %s", len(ds), path)
	os, err := os.OpenFile(path, flags, 0666)
	if err != nil {
		return fmt.Errorf("could not open %q: %v", path, err)
	}
	defer os.Close()
	if err := writeBlock(os, ds); err != nil {
		return fmt.Errorf("could not encode %q: %v", path, err)
	}
	return nil
}

func (s dirStorage) Get(url string, f func(Entry)) error {
	path := s.path(url)
	is, err := os.Open(path)
	if os.IsNotExist(err) { // nothing in the index
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not open %q: %v", path, err)
	}
	defer is.Close()
	// log.Printf("reading path %s", path)
	for {
		ds, err := readBlock(is)
		if err != nil {
			return fmt.Errorf("could not decode %q: %v", path, err)
		}
		if len(ds) == 0 {
			return nil
		}
		for _, d := range ds {
			l, a := decodeL(d.L)
			f(Entry{
				ConceptURL:  url,
				RelationURL: s.lookup(int(d.R)),
				Token:       d.S,
				Path:        d.P,
				Begin:       int(d.B),
				End:         int(d.E),
				L:           l,
				Ambiguous:   a,
			})
		}
	}
}

func (s dirStorage) Close() error {
	path := s.urlRegisterPath()
	// log.Printf("wrting register to %s", path)
	os, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot write %q: %v", path, err)
	}
	defer os.Close()
	e := gob.NewEncoder(os)
	return e.Encode(s.register)
}

func (s dirStorage) path(url string) string {
	return s.preparePath(url + ".gob")
}

func (s dirStorage) urlRegisterPath() string {
	return s.path("http://bitbucket.org/fflo/semix/url-register")
}

func (s dirStorage) lookup(id int) string {
	if url, ok := s.register.LookupID(id); ok {
		return url
	}
	return ""
}

func writeBlock(w io.Writer, ds []dse) error {
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
	// log.Printf("wrote %d entries", len(ds))
	return nil
}

func readBlock(r io.Reader) ([]dse, error) {
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
	var ds []dse
	err := d.Decode(&ds)
	// log.Printf("read %d entries", len(ds))
	return ds, err
}

// Short var names for smaller gob entries.
// S is the string
// P is the document path
// B is the start position
// E is the end position
// R is the relation id
type dse struct {
	S, P string
	B, E uint32
	R    int32
	L    uint8
}

func encodeL(l int, a bool) uint8 {
	x := uint8(l) & 0x7f
	if a {
		x |= 0x80
	}
	return x
}

func decodeL(x uint8) (int, bool) {
	return int(x & 0x7f), x&0x80 > 0
}

func (s dirStorage) preparePath(u string) string {
	u = strings.Replace(u, "https://", "", 1)
	u = strings.Replace(u, "http://", "", 1)
	u = filepath.Join(s.dir, u)
	p := filepath.Dir(u)
	// log.Printf("preparing: %s", p)
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		// log.Printf("could no prepare: %s: %s", p, err)
	}
	return u
}
