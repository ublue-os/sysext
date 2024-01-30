package layer

import (
	"github.com/spf13/cobra"
)

// listCmd represents the list command
func NewListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: listCmd,
	}
}

func listCmd(cmd *cobra.Command, args []string) error {
	return nil
}
