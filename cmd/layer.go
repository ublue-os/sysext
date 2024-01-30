/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/cmd/layer"
)

// layerCmd represents the layer command
var layerCmd = &cobra.Command{
	Use:   "layer",
	Short: "Execute layer-based operations",
}

func init() {
	layerCmd.AddCommand(layer.NewActivateCmd())
	layerCmd.AddCommand(layer.NewAddCmd())
	layerCmd.AddCommand(layer.NewCleanCmd())
	layerCmd.AddCommand(layer.NewDeactivateCmd())
	layerCmd.AddCommand(layer.NewGetPropertyCmd())
	layerCmd.AddCommand(layer.NewInitCmd())
	layerCmd.AddCommand(layer.NewListCmd())
	layerCmd.AddCommand(layer.NewRefreshCmd())
	layerCmd.AddCommand(layer.NewRemoveCmd())

	rootCmd.AddCommand(layerCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// layerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// layerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
