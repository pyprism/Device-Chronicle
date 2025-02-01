package cmd

import (
	"device-chronicle-server/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "device-chronicle",
	Short: "device-chronicle is a tool to store and display device metrics",
	Long:  `device-chronicle is a tool to store and display device metrics`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		config.Logger.Error("Execution error:", zap.Error(err))
		os.Exit(1)
	}
}
