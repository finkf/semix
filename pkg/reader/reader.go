package reader

import (
	"io"
	"os"

	"github.com/pkg/errors"
	"gitlab.com/finkf/semix/pkg/semix"
)

// Content-Types for the documents
const (
	PlainText = "text/plain"
	XML       = "application/xml"
	ALTO      = "application/xml+alto"
	HTML      = "text/html"
	HTTP      = "application/http"
)

// NewFromURI is a simple convenience function to open Documents from
// a given path. If the given Content-Type is HTTP, an according
// httpReader is returned.
func NewFromURI(path, ct string) (semix.Document, error) {
	if ct == HTTP {
		return NewFromURL(path)
	}
	is, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open file")
	}
	// close is handled by the documents
	return New(is, path, ct)
}

// NewFromURL returns a new reader that executes an HTTP request and
// reads the returned document.
func NewFromURL(url string) (semix.Document, error) {
	return &httpDocument{plainTextDocument{uri: url}}, nil
}

// New returns a new Document that reads the given Content-Type. If
// the Content-Type is empty or unknown, the Content-Type is
// automatically determined.
func New(r io.ReadCloser, uri, ct string) (semix.Document, error) {
	switch ct {
	case PlainText:
		return plainTextDocument{r: r, uri: uri}, nil
	case XML:
		return &xmlDocument{plainTextDocument{r: r, uri: uri}, false}, nil
	case HTML:
		return &htmlDocument{plainTextDocument{r: r, uri: uri}, false}, nil
	}
	return nil, errors.Errorf("cannot determine document type: %s (%s)", uri, ct)
}
