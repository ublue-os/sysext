package extensions

import (
	"fmt"
	"github.com/spf13/cobra"
	"os/exec"
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
	var (
		force_flag     string = ""
		sysext_command        = "refresh"
	)
	if *fForce {
		force_flag = "--force"
	}

	if !*fRefresh {
		sysext_command = "merge"
	}

	out, err := exec.Command("systemd-sysext", sysext_command, force_flag).Output()
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", out)
	return nil
}
