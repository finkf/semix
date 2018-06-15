package cmd

import (
	"fmt"

	"bitbucket.org/fflo/semix/pkg/client"
	"bitbucket.org/fflo/semix/pkg/semix"
	"github.com/spf13/cobra"
)

var dotCmd = &cobra.Command{
	Use:   "dot concept",
	Short: "get the dot graph for the given concept",
	Long: `
The dot command queries the daemon for a concept
and returns the dot-graph of all its connections.
`,
	RunE:         dot,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true,
}

var (
	dotDir string
)

func init() {
	getCmd.Flags().StringVarP(&dotDir, "orientation", "o", "BT", "set the orientation of the graph")
}

func dot(cmd *cobra.Command, args []string) error {
	setupSay()
	defer func() { fmt.Println("} // DOTCODE") }()
	fmt.Println("digraph semix { // DOTCODE")
	fmt.Printf("rankdir=%s // DOTCODE\n", dotDir)

	c := client.New(DaemonHost())
	cs, err := c.Download(args[0])
	if err != nil {
		return err
	}
	for _, c := range cs {
		fmt.Printf("%d [label=%s] // DOTCODE\n", c.ID(), name(c))
		//		say.Debug("concept: %q", c.ShortName())
	}
	for _, c := range cs {
	edges:
		for _, e := range c.Edges() {
			t := e.O
			p := e.P
			for _, ee := range c.Edges() {
				if ee.O.ID() != t.ID() && ee.O.HasLinkP(p, t) {
					continue edges
				}
			}
			fmt.Printf("%d -> %d [label=%q] // DOTCODE\n", c.ID(), t.ID(), p.ShortName())
		}
	}
	return nil
}

func canReach(a, b *semix.Concept) bool {
	if a.ID() == b.ID() {
		return false // cannot reach itself
	}
	for _, e := range a.Edges() {
		if e.O.ID() == b.ID() {
			return true
		}
	}
	return false
}

func register(ids map[int32]bool, c *semix.Concept) {
	if !ids[c.ID()] {
		fmt.Printf("%d [label=%s] // DOTCODE\n", c.ID(), name(c))
		ids[c.ID()] = true
	}
}

func name(c *semix.Concept) string {
	name := fmt.Sprintf("%q", c.ShortName())
	len := len(name)
	for i := len/2 + 1; i < len; i++ {
		if name[i] == ' ' {
			return name[0:i] + "\\n" + name[i+1:]
		}
	}
	return name
}
