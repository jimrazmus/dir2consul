package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cliSyncCmd)
}

var cliSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "synchronize all kv pairs",
	Long:  `synchronize blah blah blah`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sync cmd")
	},
}
