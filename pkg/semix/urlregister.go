package urlregister

import (
	"bytes"
	"encoding/gob"
)

// URLRegister is used to map urls to unique ids and vice versa.
type URLRegister struct {
	urls map[string]int
	ids  []string
}

// NewURLRegister creates a new URLRegister.
func NewURLRegister() *URLRegister {
	return &URLRegister{
		urls: make(map[string]int),
	}
}

// Register registers a new url and returs its associated id.
// If a given url does not yet exist, it is inserted and given a new id.
func (r *URLRegister) Register(url string) int {
	if id, ok := r.urls[url]; ok {
		return id
	}
	r.ids = append(r.ids, url)
	r.urls[url] = len(r.ids)
	return len(r.ids)
}

// LookupID searches for the given id and returs its associated url and true
// if it can be found or "" and false otherwise.
func (r *URLRegister) LookupID(id int) (string, bool) {
	if id <= 0 || id > len(r.ids) {
		return "", false
	}
	return r.ids[id-1], true
}

// LookupURL searches for the given url and returns its
// associated id and true if it can be found or 0 and false oterhwise.
func (r *URLRegister) LookupURL(url string) (int, bool) {
	if id, ok := r.urls[url]; ok {
		return id, true
	}
	return 0, false
}

// GobDecode implements gob.Decoder
func (r *URLRegister) GobDecode(bs []byte) error {
	buffer := bytes.NewBuffer(bs)
	d := gob.NewDecoder(buffer)
	if err := d.Decode(&r.urls); err != nil {
		return err
	}
	if err := d.Decode(&r.ids); err != nil {
		return err
	}
	return nil
}

// GobEncode implements gob.Encoder
func (r *URLRegister) GobEncode() ([]byte, error) {
	buffer := &bytes.Buffer{}
	e := gob.NewEncoder(buffer)
	if err := e.Encode(r.urls); err != nil {
		return nil, err
	}
	if err := e.Encode(r.ids); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
