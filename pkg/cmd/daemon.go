package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "daemon [options...]",
	Long:  `The daemon command starts the semix daemon.`,
	RunE:  daemon,
}

var (
	daemonDir      string
	daemonResource string
)

func init() {
	daemonCmd.Flags().StringVarP(&daemonDir, "dir", "d",
		filepath.Join(os.Getenv("HOME"), "semix"), "set semix index directory")
	daemonCmd.Flags().StringVarP(&daemonResource, "resource", "r",
		"semix.toml", "set path of resource file")
}

func daemon(cmd *cobra.Command, args []string) error {
	return nil
}
