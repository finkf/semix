package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"bitbucket.org/fflo/semix/pkg/client"
	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	putLocal  bool
	resolvers []string
	levs      []int
	memsize   int
	threshold float64
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
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
	}
)

func init() {
	putCmd.Flags().BoolVarP(&putLocal, "local", "l", false,
		"do not upload files; use local files")
	putCmd.Flags().StringSliceVarP(&resolvers, "resolver", "r", []string{},
		"use resolvers in given order; allowed values are thematic,ruled,simple")
	putCmd.Flags().IntSliceVarP(&levs, "ks", "k", []int{},
		"add approximate searches with the given error limits")
	putCmd.Flags().IntVarP(&memsize, "memory-size", "m", 10,
		"set the memory size used by the resolvers")
	putCmd.Flags().Float64VarP(&threshold, "threshold", "t", 0.5,
		"set the threshold for the thematic resolver")
}

func put(cmd *cobra.Command, args []string) error {
	setupSay()
	rs, err := rest.MakeResolvers(threshold, memsize, resolvers)
	if err != nil {
		return errors.Wrapf(err, "put")
	}
	sort.Ints(levs)
	client := newClient(
		client.WithErrorLimits(levs...),
		client.WithResolvers(rs...),
	)
	for _, arg := range args {
		if err := putPath(client, arg); err != nil {
			return errors.Wrapf(err, "put")
		}
	}
	return nil
}

func putPath(client *client.Client, path string) error {
	if isURL(path) {
		return putFileOrURL(client, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return errors.Wrapf(err, "cannot index %s", path)
	}
	if info.IsDir() {
		return putDir(client, path)
	}
	return putFileOrURL(client, path)
}

func putDir(client *client.Client, path string) error {
	return filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "cannot index %s", p)
		}
		if i.IsDir() {
			return putDir(client, p)
		}
		return putFileOrURL(client, p)
	})
}

func putFileOrURL(client *client.Client, path string) error {
	es, err := doPutFileOrURL(client, path)
	if err != nil {
		return errors.Wrapf(err, "cannot index %s", path)
	}
	if jsonOutput {
		_ = json.NewEncoder(os.Stdout).Encode(es)
	} else {
		prettyPrintEntries(path, es)
	}
	return nil
}

func doPutFileOrURL(client *client.Client, path string) ([]index.Entry, error) {
	if isURL(path) {
		return client.PutURL(path)
	}
	if putLocal {
		return client.PutLocalFile(path)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	return client.PutContent(file, path, "text/plain")
}

func isURL(path string) bool {
	return strings.HasPrefix(path, "http://") ||
		strings.HasPrefix(path, "https://")
}
