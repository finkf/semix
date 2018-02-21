package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rest"
	"bitbucket.org/fflo/semix/pkg/say"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	putLocal  bool
	resolvers []string
	levs      []int
	putCmd    = &cobra.Command{
		Use:   "put [paths...]",
		Short: "Put a file into the semantic index",
		Long: `The put command puts files into the semantic index.
If a given path denotes a directory all files and
directories are indexed recrusively.

Files can either be uploaded to the daemon or used locally. In the first
case, put uploads the file's contents to the daemon and the daemon indexes
the file's content and stores it in a local file.
In the latter case the daemon reads the file itself from the given path
and no local copy of the file is made.

If a path looks like an URL (starts with either http:// or https://)
the URL is given to the semix daemon, which downloads the file and puts
its contents into the semantic index. The contents of the URL
are not stored locally`,
		RunE:         put,
		SilenceUsage: true,
	}
)

func init() {
	putCmd.Flags().BoolVarP(&putLocal, "local", "l", false,
		"do not upload files; use local files")
	putCmd.Flags().StringSliceVarP(&resolvers, "resolver", "r", []string{},
		"use resolvers in given order. Allowed values are thematic,ruled,simple")
	putCmd.Flags().IntSliceVarP(&levs, "lev", "d", []int{},
		"add approximate searches with the given error limits (order does not matter)")
}

func put(cmd *cobra.Command, args []string) error {
	say.SetDebug(debug)
	sort.Ints(levs)
	say.Debug("Resolvers: %v", resolvers)
	say.Debug("Levs:      %v", levs)
	client := newClient()
	for _, arg := range args {
		if err := putPath(client, arg); err != nil {
			return err
		}
	}
	return nil
}

func putPath(client *rest.Client, path string) error {
	if isURL(path) {
		return putFileOrURL(client, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return errors.Wrapf(err, "[put] cannot index %s", path)
	}
	if info.IsDir() {
		return putDir(client, path)
	}
	return putFileOrURL(client, path)
}

func putDir(client *rest.Client, path string) error {
	return filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "[put] cannot index %s", p)
		}
		if i.IsDir() {
			return putDir(client, p)
		}
		return putFileOrURL(client, p)
	})
}

func putFileOrURL(client *rest.Client, path string) error {
	es, err := doPutFileOrURL(client, path)
	if err != nil {
		return errors.Wrapf(err, "[put] cannot index %s", path)
	}
	if jsonOutput {
		_ = json.NewEncoder(os.Stdout).Encode(es)
	} else {
		prettyPrintEntries(path, es)
	}
	return nil
}

func doPutFileOrURL(client *rest.Client, path string) ([]index.Entry, error) {
	if isURL(path) {
		return client.PutURL(path, nil, nil)
	}
	if putLocal {
		return client.PutLocalFile(path, nil, nil)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	return client.PutContent(file, path, "text/plain", nil, nil)
}

func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://")
}
