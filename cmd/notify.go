package main

import (
	"ci-recipe-finder-bot/notify"
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(notifyCmd)
}

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Notify Subscribers Of Changes",
	Long:  "Notify Subscribers Of Changes",
	Run: func(cmd *cobra.Command, args []string) {
		err := notify.Send()
		if err != nil {
			fmt.Println(err.Error())
		}
	},
}
