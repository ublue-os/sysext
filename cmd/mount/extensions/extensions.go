package extensions

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
)

var ExtensionsCmd = &cobra.Command{
	Use:   "extensions",
	Short: "Mount systemd-sysextensions stored in /var/lib/extensions and /run/extensions",
	Long:  `Mount systemd-sysextensions stored in /var/lib/extensions and /run/extensions`,
	RunE:  extensionsCmd,
}

var (
	fRefresh *bool
	fForce   *bool
)

func init() {
	fRefresh = ExtensionsCmd.Flags().BoolP("refresh", "r", true, "Refresh instead of erroring on already mounted directories")
	fForce = ExtensionsCmd.Flags().BoolP("force", "f", false, "Pass the --force flag to systemd-sysext")
}

func extensionsCmd(cmd *cobra.Command, args []string) error {

	var command []string

	if *fForce {
		command = append(command, "--force")
	}

	if *fRefresh {
		command = append(command, "refresh")
	} else if *internal.Config.UnmountFlag {
		command = append(command, "unmerge")
	} else {
		command = append(command, "merge")
	}

	out, err := exec.Command("systemd-sysext", command...).Output()
	if err != nil {
		fmt.Printf("%s\n", out)
		return err
	}
	fmt.Printf("%s\n", out)
	return nil
}
