package restd

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/fflo/semix/pkg/semix"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func openDumpFile(dir, path string) semix.Document {
	return semix.NewFileDocument(filepath.Join(dir, "dump", path))
}

func newDumpFile(r io.Reader, dir, content string) (semix.Document, error) {
	var path string
	switch strings.ToLower(content) {
	case "text/plain":
		path = fmt.Sprintf(
			"semix-%s-%d-%s",
			time.Now().Format("2006-01-02-15-04-05"),
			rand.Int(),
			"text-plain",
		)
	default:
		return nil, fmt.Errorf("invalid Content-Type: %s", content)
	}
	os, err := os.Create(filepath.Join(dir, "dump", path))
	if err != nil {
		return nil, err
	}
	return dumpFile{r: r, w: os, p: path}, nil // dumpFile closes the file
}

type dumpFile struct {
	r io.Reader
	w io.WriteCloser
	p string
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
