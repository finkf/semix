package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var putLocal bool
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "put [paths...]",
	Long: `
The put command puts files into the semantic index.
If a path is directory all files and directories are put recursively.
If a path looks like an URL (starts with either http:// or https://)
the URL is given to the semix daemon, which downloads the file and puts
its contents into the semantic index.`,
	Args: cobra.MinimumNArgs(1),
	RunE: put,
}

func init() {
	putCmd.Flags().BoolVarP(&putLocal, "local", "l", false, "do not upload files")
}

func put(cmd *cobra.Command, args []string) error {
	client := rest.NewClient(daemonHost)
	for _, arg := range args {
		if err := putPath(client, arg); err != nil {
			return err
		}
	}
	return nil
}

func putPath(client rest.Client, path string) error {
	if isURL(path) {
		return putFileOrURL(client, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return errors.Wrapf(err, "cannot put %s", path)
	}
	if info.IsDir() {
		return putDir(client, path)
	}
	return putFileOrURL(client, path)
}

func putDir(client rest.Client, path string) error {
	return filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "cannot put %s", p)
		}
		if i.IsDir() {
			return putDir(client, p)
		}
		return putFileOrURL(client, p)
	})
}

func putFileOrURL(client rest.Client, path string) error {
	var es []index.Entry
	var err error
	if isURL(path) {
		es, err = client.PutURL(path, nil, nil)
	} else if putLocal {
		es, err = client.PutLocalFile(path, nil, nil)
	} else {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()
		es, err = client.PutContent(file, path, "text/plain", nil, nil)
	}
	if err != nil {
		return errors.Wrapf(err, "cannot put %s", path)
	}
	fmt.Printf("%v\n", es)
	return nil
}

func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://")
}
