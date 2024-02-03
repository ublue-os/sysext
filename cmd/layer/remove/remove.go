package remove

import (
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/fileio"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a layer from your managed layers",
	Long:  `Remove either an entire layer or a specific hash in cache for that layer`,
	RunE:  removeCmd,
}

var (
	fHash *string
)

func init() {
	fHash = RemoveCmd.Flags().String("hash", "", "Remove specific hash from storage")
}

func removeCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return internal.NewPositionalError("TARGET")
	}

	target_layer := args[0]

	cache_dir, err := filepath.Abs(path.Clean(internal.Config.CacheDir))
	if err != nil {
		return err
	}

	if *fHash != "" {
		err := os.Remove(path.Join(cache_dir, target_layer, *fHash))
		if err != nil {
			return err
		}
	}

	err = os.RemoveAll(path.Join(cache_dir, target_layer))
	if err != nil {
		return err
	}

	deactivated_layer := path.Join(internal.Config.ExtensionsDir, target_layer) + internal.ValidSysextExtension
	if !fileio.FileExist(deactivated_layer) {
		return nil
	}
	err = os.Remove(deactivated_layer)
	if err != nil {
		return err
	}

	return nil
}
