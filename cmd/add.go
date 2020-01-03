package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cliAddCmd)
}

var cliAddCmd = &cobra.Command{
	Use:   "add",
	Short: "only adds new kv pairs",
	Long:  `adds blah blah blah`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add command")
	},
}
