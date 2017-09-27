package semix

import "io"

// Document defines an interface for readeable documents.
type Document interface {
	io.ReadCloser
	Path() string
}
