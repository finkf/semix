package semix

import (
	"io"
	"net/http"
	"strings"
)

// Document defines an interface for readeable documents.
type Document interface {
	io.ReadCloser
	Path() string
}

// ReaderDocument wraps an io.Reader.
type ReaderDocument struct {
	path string
	r    io.Reader
}

// NewReaderDocument create a new ReaderDocument.
func NewReaderDocument(path string, r io.Reader) Document {
	return ReaderDocument{path: path, r: r}
}

// NewStringDocument returns a document that reads from a string.
func NewStringDocument(path, str string) Document {
	return NewReaderDocument(path, strings.NewReader(str))
}

// Path returns the path of this StringDocument.
func (d ReaderDocument) Path() string {
	return d.path
}

// Close returns nil.
func (ReaderDocument) Close() error {
	return nil
}

// Read implements the io.Reader interface.
func (d ReaderDocument) Read(b []byte) (int, error) {
	return d.r.Read(b)
}

// HTTPDocument is a document that reads from HTTP.
type HTTPDocument struct {
	r   io.ReadCloser
	url string
}

// NewHTTPDocument create a new HTTPDocument with the given url.
func NewHTTPDocument(url string) Document {
	return &HTTPDocument{url: url, r: nil}
}

// Path returns the url of the HTTPDocument.
func (d *HTTPDocument) Path() string {
	return d.url
}

// Close closes the underlying body of the http GET
// resoponse of the HTTPDocument.
func (d *HTTPDocument) Close() error {
	if d.r != nil {
		return d.r.Close()
	}
	return nil
}

// Read implements the io.Reader interface.
func (d *HTTPDocument) Read(b []byte) (int, error) {
	if d.r == nil {
		resp, err := http.Get(d.url)
		if err != nil {
			return 0, err
		}
		d.r = resp.Body
	}
	return d.r.Read(b)
}
