package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"bitbucket.org/fflo/semix/pkg/index"
	"bitbucket.org/fflo/semix/pkg/rest"
	"bitbucket.org/fflo/semix/pkg/say"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:          "get [query...]",
	Short:        "Query the semantic index",
	Long:         `The get command sends queries to the daemon.`,
	RunE:         get,
	SilenceUsage: true,
}

var (
	getMax  int
	getSkip int
)

func init() {
	getCmd.Flags().IntVarP(&getMax, "max", "m", 0, "set max number of entries")
	getCmd.Flags().IntVarP(&getSkip, "skip", "s", 0, "set number of entries to skip")
}

func get(cmd *cobra.Command, args []string) error {
	say.SetDebug(debug)
	client := newClient(rest.WithSkip(getSkip), rest.WithMax(getMax))
	for _, query := range args {
		if err := doGet(client, query); err != nil {
			return err
		}
	}
	return nil
}

func doGet(client *rest.Client, query string) error {
	ts, err := client.Get(query)
	if err != nil {
		return errors.Wrapf(err, "[get] could not execute query %s", query)
	}
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Path < ts[j].Path
	})
	if jsonOutput {
		_ = json.NewEncoder(os.Stdout).Encode(ts)
	} else {
		prettyPrintEntries(query, ts)
	}
	return nil
}

func prettyPrintEntries(query string, es []index.Entry) {
	for i, e := range es {
		fmt.Printf("%s:%d:%d: %q %q %q %q\n",
			query, getSkip+i+1, getSkip+len(es),
			e.Token, e.RelationURL, e.ConceptURL, e.Path)
	}
}
