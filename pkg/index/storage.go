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
	Get(string, func(Entry) bool) error
	Close() error
}

type dirStorage struct {
	dir                      string
	relationReg, documentReg *semix.URLRegister
}

// OpenDirStorage opens a new IndexStorage.
func OpenDirStorage(dir string) (Storage, error) {
	rel, err := semix.ReadURLRegister(relationRegisterPath(dir))
	if err != nil {
		return dirStorage{}, err
	}
	doc, err := semix.ReadURLRegister(documentRegisterPath(dir))
	if err != nil {
		return dirStorage{}, err
	}
	return dirStorage{dir, rel, doc}, nil
}

func (s dirStorage) Put(url string, es []Entry) error {
	if len(es) == 0 {
		return nil
	}
	ds := make([]dse, len(es))
	for i := range es {
		ds[i] = newDSE(es[i], s.lookupURLs)
	}
	return s.write(url, ds)
}

func (s dirStorage) write(url string, ds []dse) error {
	path := preparePath(s.dir, url)
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

func (s dirStorage) Get(url string, f func(Entry) bool) error {
	path := preparePath(s.dir, url)
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
			if !f(d.entry(url, s.lookupIDs)) {
				return nil
			}
		}
	}
}

func (s dirStorage) Close() error {
	if err := s.relationReg.Write(relationRegisterPath(s.dir)); err != nil {
		return err
	}
	return s.documentReg.Write(documentRegisterPath(s.dir))
}

type lookupIDsFunc func(int, int) (string, string)

func (s dirStorage) lookupIDs(relID, docID int) (string, string) {
	var relURL, docURL string
	if url, ok := s.relationReg.LookupID(relID); ok {
		relURL = url
	}
	if url, ok := s.documentReg.LookupID(docID); ok {
		docURL = url
	}
	return relURL, docURL
}

type lookupURLsFunc func(string, string) (int, int)

func (s dirStorage) lookupURLs(relURL, docURL string) (int, int) {
	var relID, docID int
	if relURL != "" {
		relID = s.relationReg.Register(relURL)
	}
	if docURL != "" {
		docID = s.documentReg.Register(docURL)
	}
	return relID, docID
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

func relationRegisterPath(dir string) string {
	return preparePath(dir, "http://bitbucket.org/fflo/semix/relation-register.gob")
}

func documentRegisterPath(dir string) string {
	return preparePath(dir, "http://bitbucket.org/fflo/semix/document-register.gob")
}

func preparePath(dir, u string) string {
	u = strings.Replace(u, "https://", "", 1)
	u = strings.Replace(u, "http://", "", 1)
	u = filepath.Join(dir, u)
	p := filepath.Dir(u)
	if err := os.MkdirAll(p, os.ModePerm); err != nil {
		// log.Printf("could no prepare: %s: %s", p, err)
	}
	return u
}
