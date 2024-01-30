package layer

import (
	"os/exec"

	"github.com/spf13/cobra"
)

func NewAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: addCmd,
	}
}

func addCmd(cmd *cobra.Command, args []string) error {
	othercmd := exec.Command("yourcommand", "some", "args")
	if err := othercmd.Run(); err != nil {
		return err
	}
	return nil
}
