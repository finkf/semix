package semix

import (
	"io"
	"net/http"
	"os"
	"strings"
)

// Document defines an interface for readeable documents.
type Document interface {
	io.ReadCloser
	Path() string
}

// ReaderDocument wraps an io.Reader.
type ReaderDocument struct {
	io.Reader
	path string
}

// NewReaderDocument create a new ReaderDocument.
func NewReaderDocument(path string, r io.Reader) Document {
	return ReaderDocument{r, path}
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

// HTTPDocument is a document that reads from HTTP.
type HTTPDocument struct {
	r   io.ReadCloser
	url string
}

// NewHTTPDocument creates a new HTTPDocument with the given url.
// The first call to Read will trigger an http.Get request to be sent.
// Any errors from this request will be returned in the Read method.
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

// FileDocument wraps an os.File and a path.
type FileDocument struct {
	file *os.File
	path string
}

// NewFileDocument creates a new FileDocument with the given path.
// The first call to Read will trigger an os.Open.
// Any errors from os.Open will be returned in the Read method.
func NewFileDocument(path string) Document {
	return &FileDocument{path: path, file: nil}
}

// Path returns the url of the HTTPDocument.
func (d *FileDocument) Path() string {
	return d.path
}

// Close closes the underlying body of the http GET
// resoponse of the HTTPDocument.
func (d *FileDocument) Close() error {
	if d.file != nil {
		return d.file.Close()
	}
	return nil
}

// Read implements the io.Reader interface.
func (d *FileDocument) Read(b []byte) (int, error) {
	if d.file == nil {
		is, err := os.Open(d.path)
		if err != nil {
			return 0, err
		}
		d.file = is
	}
	return d.file.Read(b)
}
