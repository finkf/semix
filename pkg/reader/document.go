package reader

import (
	"bytes"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"

	"bitbucket.org/fflo/semix/pkg/semix"
	"github.com/pkg/errors"
)

type plainTextDocument struct {
	r   io.ReadCloser
	uri string
}

func (r plainTextDocument) Path() string {
	return r.uri
}

func (r plainTextDocument) Close() error {
	return r.r.Close()
}

func (r plainTextDocument) Read(b []byte) (int, error) {
	return r.r.Read(b)
}

var _ semix.Document = plainTextDocument{}

type xmlDocument struct {
	plainTextDocument
	parsed bool
}

func (r *xmlDocument) Read(b []byte) (int, error) {
	if !r.parsed {
		if err := r.parse(); err != nil {
			return 0, err
		}
	}
	return r.r.Read(b)
}

func (r *xmlDocument) parse() error {
	// make sure that the original reader is closed
	reader := r.r
	defer func() { _ = reader.Close() }()
	// parse markup
	d := xml.NewDecoder(reader)
	newreader, err := parseMarkup(d, nil)
	if err != nil {
		return errors.Wrapf(err, "cannot parse XML: %s", r.Path())
	}
	r.r = newreader
	return nil
}

var _ semix.Document = &xmlDocument{}

type htmlDocument struct {
	plainTextDocument
	parsed bool
}

func (r *htmlDocument) Read(b []byte) (int, error) {
	if !r.parsed {
		if err := r.parse(); err != nil {
			return 0, err
		}
	}
	return r.Read(b)
}

var ignoreHTMLTags = map[string]struct{}{
	"body": {},
}

func (r *htmlDocument) parse() error {
	// make sure to close underlying reader
	reader := r.r
	defer func() { _ = reader.Close() }()
	// parse html
	d := xml.NewDecoder(reader)
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity
	newreader, err := parseMarkup(d, ignoreHTMLTags)
	if err != nil {
		return errors.Wrapf(err, "cannot parse HTML: %s", r.Path())
	}
	r.r = newreader
	return nil
}

type httpDocument struct {
	plainTextDocument
}

func (r *httpDocument) Read(b []byte) (int, error) {
	if r.r == nil {
		if err := r.get(); err != nil {
			return 0, err
		}
	}
	return r.r.Read(b)
}

func (r *httpDocument) get() error {
	res, err := http.Get(r.uri)
	if err != nil {
		return errors.Wrapf(err, "cannot get: %s", r.uri)
	}
	ct := res.Header.Get("Content-Type")
	reader, err := New(res.Body, r.uri, ct)
	if err != nil {
		return errors.Wrapf(err, "cannot read content from: %s", r.uri)
	}
	r.r = reader
	return nil
}

var _ semix.Document = &httpDocument{}

func parseMarkup(d *xml.Decoder, ignore map[string]struct{}) (io.ReadCloser, error) {
	buf := &bytes.Buffer{}
	var err error
	var t xml.Token
	for t, err = d.Token(); err != nil; {
		switch token := t.(type) {
		case xml.CharData:
			if _, e2 := buf.Write(token); e2 != nil {
				return nil, errors.Wrapf(e2, "cannot write char data")
			}
			// append ' '
			if e2 := buf.WriteByte(' '); e2 != nil {
				return nil, errors.Wrapf(e2, "cannot write char data")
			}
		case xml.StartElement:
			if _, ok := ignore[token.Name.Local]; ok {
				if e2 := d.Skip(); e2 != nil {
					return nil, e2
				}
			}
		}
	}
	if err != nil && err != io.EOF {
		return nil, errors.Wrapf(err, "cannot parse markup")
	}
	return ioutil.NopCloser(buf), nil
}
