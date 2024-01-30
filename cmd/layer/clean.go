/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package layer

import (
	"github.com/spf13/cobra"
)

// cleanCmd represents the clean command
func NewCleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: cleanCmd,
	}
}

func cleanCmd(cmd *cobra.Command, args []string) error {
	return nil
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cleanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cleanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
