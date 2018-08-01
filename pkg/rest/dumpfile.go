package rest

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gitlab.com/finkf/semix/pkg/semix"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func openDumpFile(dir, path string) semix.Document {
	return semix.NewFileDocument(filepath.Join(dir, "dump", path))
}

func newDumpFile(r io.Reader, dir, pre, ct string) (semix.Document, error) {
	if err := os.MkdirAll(filepath.Join(dir, "dump"), os.ModePerm); err != nil {
		return nil, err
	}
	path, err := makeFileName(pre, ct)
	if err != nil {
		return nil, err
	}
	os, err := os.Create(filepath.Join(dir, "dump", path))
	if err != nil {
		return nil, err
	}
	return dumpFile{r: r, w: os, p: path}, nil // dumpFile closes the file
}

func makeFileName(pre, ct string) (string, error) {
	switch strings.ToLower(ct) {
	case "text/plain":
		ct = "text-plain"
	default:
		return "", fmt.Errorf("invalid Content-Type: %s", ct)
	}
	pre = regexp.MustCompile("\\s+").ReplaceAllLiteralString(pre, "-")
	pre = regexp.MustCompile("/+").ReplaceAllLiteralString(pre, "-")
	pre = regexp.MustCompile("-+").ReplaceAllLiteralString(pre, "-")
	pre = strings.ToLower(pre)
	return fmt.Sprintf("semix-%s-%s-%d-%s",
		pre, time.Now().Format("2006-01-02-15-04-05"), rand.Int(), ct), nil
}

type dumpFile struct {
	r io.Reader
	w io.WriteCloser
	p string
}

func (d dumpFile) ContentType() string {
	if strings.HasSuffix(d.p, "text-plain") {
		return "text/plain"
	}
	return ""
}

func (d dumpFile) Close() error {
	return d.w.Close()
}

func (d dumpFile) Read(bs []byte) (int, error) {
	n, err := d.r.Read(bs)
	if err != nil {
		return 0, err
	}
	if _, err := d.w.Write(bs[:n]); err != nil {
		return 0, err
	}
	return n, nil
}

func (d dumpFile) Path() string {
	return d.p
}
