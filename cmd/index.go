package main

import (
	"ci-recipe-finder-bot/index"
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(indexCmd)
}

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Rebuild Cooks Illustrated Index",
	Long:  "Rebuild Cooks Illustrated Index",
	Run: func(cmd *cobra.Command, args []string) {
		err := index.RefreshIndex()
		if err != nil {
			fmt.Println(err.Error())
		}
	},
}
