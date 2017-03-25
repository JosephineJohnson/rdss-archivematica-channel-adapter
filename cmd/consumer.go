package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// consumerCmd represents the consumer command
var consumerCmd = &cobra.Command{
	Use:   "consumer",
	Short: "Inbound server (RDSS » Archivematica)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("consumer called")
	},
}

func init() {
	RootCmd.AddCommand(consumerCmd)
}
