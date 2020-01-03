package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cliDeleteCmd)
}

var cliDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "deletes kv pairs not found in the files",
	Long:  `deletes blah blah blah`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("the delete command")
	},
}
