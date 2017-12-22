package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rest"
	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put",
	Short: "put [paths...]",
	Args:  cobra.MinimumNArgs(1),
	RunE:  put,
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
		return err
	}
	if info.IsDir() {
		return putDir(client, path)
	}
	return putFileOrURL(client, path)
}

func putDir(client rest.Client, path string) error {
	return filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
		if err != nil {
			return err
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
	var local bool
	if isURL(path) {
		es, err = client.PutURL(path, nil, nil)
	} else if local {
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
		return err
	}
	fmt.Printf("%v\n", es)
	return nil
}

func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://")
}
