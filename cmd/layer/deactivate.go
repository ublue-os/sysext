/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package layer

import (
	"github.com/spf13/cobra"
)

func NewDeactivateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deactivate",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: deactivateCmd,
	}
}

func deactivateCmd(cmd *cobra.Command, args []string) error {
	return nil
}
