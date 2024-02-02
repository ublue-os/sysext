package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/cmd/layer"
)

var RootCmd = &cobra.Command{
	Use:   "bext",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	SilenceUsage: true,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(layer.LayerCmd)
	RootCmd.AddCommand(AddToPathCmd)
	RootCmd.AddCommand(ServiceCmd)
	RootCmd.AddCommand(StoreCmd)
}
