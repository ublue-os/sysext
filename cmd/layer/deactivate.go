package layer

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

var DeactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate a layer and refresh sysext",
	Long:  `Deativate a selected layer (unsymlink it from /var/lib/extensions) and refresh the system extensions store.`,
	RunE:  deactivateCmd,
}

var (
	fDeactivateNoRefresh *bool
	fDeactivateForce     *bool
)

func init() {
	fDeactivateNoRefresh = DeactivateCmd.Flags().BoolP("no-refresh", "r", false, "Do not refresh systemd-sysext on run")
	fDeactivateForce = DeactivateCmd.Flags().BoolP("force", "f", false, "Pass the --force flag to systemd-sysext")
}

func deactivateCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Required positional argument TARGET")
		os.Exit(1)
	}

	target_layer := args[0]

	extensions_dir, err := filepath.Abs(path.Clean(internal.Config.ExtensionsDir))
	if err != nil {
		return err
	}

	target_layer_path := path.Join(extensions_dir, target_layer+internal.ValidSysextExtension)

	if _, err := os.Stat(target_layer_path); err != nil {
		return err
	}

	if err := os.Remove(target_layer_path); err != nil {
		return err
	}

	if *fDeactivateNoRefresh {
		return nil
	}

	var forceflag string = ""
	if *fDeactivateForce {
		forceflag = "--force"
	}

	if err := exec.Command("systemd-sysext", "refresh", forceflag).Run(); err != nil {
		return err
	}

	return nil
}
