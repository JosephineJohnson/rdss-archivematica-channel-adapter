package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Archivematica RDSS Channel Adapter",
	RunE: func(cmd *cobra.Command, args []string) error {
		printVersion()
		return nil
	},
}

func printVersion() {
	fmt.Println("v0.1.1")
}
