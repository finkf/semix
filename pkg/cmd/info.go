package cmd

import (
	"fmt"
	"strconv"

	"bitbucket.org/fflo/semix/pkg/rest"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info [URL|ID ...]",
	Short: "Print information about a concept",
	Long: `The info command prints out info about a concept.
The concept can be specified either with an ID or and URL.`,
	RunE:         info,
	SilenceUsage: true,
}

func info(cmd *cobra.Command, args []string) error {
	client := newClient()
	for _, concept := range args {
		if err := doInfo(client, concept); err != nil {
			return errors.Wrapf(err, "[info] could not get info for %s", concept)
		}
	}
	return nil
}

func doInfo(client *rest.Client, concept string) error {
	var info rest.ConceptInfo
	var err error
	id, err := strconv.Atoi(concept)
	if err == nil && id > 0 {
		info, err = client.InfoID(id)
	}
	info, err = client.InfoURL(concept)
	if err != nil {
		return err
	}
	printInfo(concept, info)
	return nil
}

func printInfo(concept string, info rest.ConceptInfo) {
	fmt.Printf("%s: %v\n", concept, info)
}
