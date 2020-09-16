package utilcmd

import (
	"fmt"
	"os"

	"github.com/huangc28/go-darkpanda-backend/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "util",
	Short: "utility commands that smoothen the development of darkpanda app",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	config.InitConfig()
}

var SecretF string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&SecretF, "secret", "s", "", "verbose output")
	rootCmd.AddCommand(GenJwtTokenByUuid)
	rootCmd.AddCommand(GenJwtTokenByName)
}
