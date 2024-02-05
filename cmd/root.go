package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/cmd/layer"
	"github.com/ublue-os/sysext/cmd/mount"
)

var RootCmd = &cobra.Command{
	Use:   "bext",
	Short: "",
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
	RootCmd.AddCommand(mount.MountCmd)
	RootCmd.AddCommand(AddToPathCmd)
}
