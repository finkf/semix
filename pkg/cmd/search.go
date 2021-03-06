package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"bitbucket.org/fflo/semix/pkg/client"
	x "bitbucket.org/fflo/semix/pkg/semix"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [pattern...]",
	Short: "Search concepts",
	Long: `
The search command searches for concepts or predicates that match
a given pattern. A pattern matches either a normalized dictionary
entry of a concept or any part of a concept's URL or name.`,
	RunE:         search,
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
}

var (
	searchPredicates bool
)

func init() {
	searchCmd.Flags().BoolVarP(&searchPredicates, "predicates", "p",
		false, "search for matching predicates")
}

func search(cmd *cobra.Command, args []string) error {
	setupSay()
	client := client.New(DaemonHost())
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

func doConcepts(client *client.Client, pattern string) error {
	cs, err := client.Search(pattern)
	if err != nil {
		return err
	}
	printConcepts(pattern, cs)
	return nil
}

func doPredicates(client *client.Client, pattern string) error {
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
	if jsonOutput {
		_ = json.NewEncoder(os.Stdout).Encode(cs)
	} else {
		for i, c := range cs {
			prettyPrintConcept(pattern, i, len(cs), c)
		}
	}
}

func prettyPrintConcept(pattern string, i, n int, c *x.Concept) {
	fmt.Printf("%s:%d:%d: %s (%d)\n", pattern, i+1, n, c.ShortName(), c.ID())
	c.EachEdge(func(edge x.Edge) {
		fmt.Printf("%s:%d:%d: + %d %s (%d) -> %s (%d)\n",
			pattern, i+1, n,
			edge.L, edge.P.ShortName(), edge.P.ID(),
			edge.O.ShortName(), edge.O.ID())
	})
}
