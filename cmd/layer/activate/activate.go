package activate

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
)

var ActivateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate a layer and refresh sysext",
	Long:  `Activate a selected layer (symlink it to /var/lib/extensions) and refresh the system extensions store.`,
	RunE:  activateCmd,
}

var (
	fNoRefresh *bool
	fForce     *bool
)

func init() {
	fNoRefresh = ActivateCmd.Flags().BoolP("no-refresh", "r", false, "Do not refresh systemd-sysext on run")
	fForce = ActivateCmd.Flags().BoolP("force", "f", false, "Pass the --force flag to systemd-sysext")
}

func activateCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Required positional argument TARGET")
		os.Exit(1)
	}

	target_layer := args[0]

	cache_dir, err := filepath.Abs(path.Clean(internal.Config.CacheDir))
	if err != nil {
		return err
	}

	current_blob_path := path.Join(cache_dir, target_layer, internal.CurrentBlobName)
	if _, err := os.Stat(current_blob_path); err != nil {
		return err
	}

	if err := os.MkdirAll(internal.Config.ExtensionsDir, 0755); err != nil {
		return err
	}

	extensions_dir, err := filepath.Abs(path.Clean(internal.Config.ExtensionsDir))
	if err != nil {
		return err
	}

	if err := os.Symlink(current_blob_path, path.Join(extensions_dir, path.Base(path.Dir(current_blob_path))+internal.ValidSysextExtension)); err != nil {
		return err
	}

	if *fNoRefresh {
		return nil
	}

	var forceflag string = ""
	if *fForce {
		forceflag = "--force"
	}

	if err := exec.Command("systemd-sysext", "refresh", forceflag).Run(); err != nil {
		return err
	}

	return nil
}
