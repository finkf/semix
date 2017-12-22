package cmd

import (
	"fmt"
	"sort"

	"bitbucket.org/fflo/semix/pkg/rest"
	x "bitbucket.org/fflo/semix/pkg/semix"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "search [pattern...]",
	Long: `
The search command searches for concepts or predicates that match
a given pattern. A pattern matches either a normalized dictionary
entry of a concept or any part of a concept's URL or name.`,
	RunE: search,
}

var searchPredicates bool

func init() {
	searchCmd.Flags().BoolVarP(
		&searchPredicates,
		"predicates",
		"p",
		false,
		"search for matching predicates",
	)
}

func search(cmd *cobra.Command, args []string) error {
	client := client()
	doSearch := doConcepts
	if searchPredicates {
		doSearch = doPredicates
	}
	for _, pattern := range args {
		if err := doSearch(client, pattern); err != nil {
			return errors.Wrapf(err, "[search] error searching for %s", pattern)
		}
	}
	return nil
}

func doConcepts(client *rest.Client, pattern string) error {
	cs, err := client.Search(pattern)
	if err != nil {
		return err
	}
	printConcepts(pattern, cs)
	return nil
}

func doPredicates(client *rest.Client, pattern string) error {
	cs, err := client.Predicates(pattern)
	if err != nil {
		return err
	}
	printConcepts(pattern, cs)
	return nil
}

func printConcepts(pattern string, cs []*x.Concept) {
	sort.Slice(cs, func(i, j int) bool {
		return cs[i].Name < cs[j].Name
	})
	for i, c := range cs {
		fmt.Printf("%s:%d:%d: %s\n", pattern, i+1, len(cs), c)
	}
}
