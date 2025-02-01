package cmd

import (
	"device-chronicle-server/config"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server",
	Long:  `Start the server to serve the web app.`,
	Run: func(cmd *cobra.Command, args []string) {
		//models.ConnectDb()
		config.Init()
	},
}
