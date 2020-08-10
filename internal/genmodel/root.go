package genmodel

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "genmodel",
	Short: "read / collect content from up to date migrations to generate models in go code.",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("some command invoked")
	},
}

func init() {
	rootCmd.AddCommand(genModelCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
