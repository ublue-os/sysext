package layer

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"os"
	"path"
	"path/filepath"
)

var DeactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate a layer and refresh sysext",
	Long:  `Deativate a selected layer (unsymlink it from /var/lib/extensions) and refresh the system extensions store.`,
	RunE:  deactivateCmd,
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

	return nil
}
