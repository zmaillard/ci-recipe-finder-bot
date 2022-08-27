package main

import (
	"ci-recipe-finder-bot/db"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

func init() {
	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import [pathToFile] [issueNumber]",
	Short: "Import New Cooks Illustrated Issue",
	Long:  "Import New Cooks Illustrated Issue",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("2 Arguments Are Required")
		}
		if !fileExists(args[0]) {
			return fmt.Errorf("%s Does Not Exist", args[0])
		}

		issueNumber, err := strconv.Atoi(args[1])
		if err != nil || (issueNumber <= 92) {
			return fmt.Errorf("%s Needs To Be A Number Greater Than 92", args[0])
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		issueNumber, _ := strconv.Atoi(args[1])

		err := db.ImportFile(args[0], issueNumber)

		if err != nil {
			fmt.Println(err.Error())
		}

	},
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
