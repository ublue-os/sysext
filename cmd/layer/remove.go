package layer

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"os"
	"path"
	"path/filepath"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a layer from your managed layers",
	Long:  `Remove either an entire layer or a specific hash in cache for that layer`,
	RunE:  removeCmd,
}

var (
	FHash *string
)

func init() {
	FHash = RemoveCmd.Flags().String("hash", "", "Remove specific hash from storage")
}

func removeCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Required positional argument TARGET")
		os.Exit(1)
	}

	target_layer := args[0]

	cache_dir, err := filepath.Abs(path.Clean(internal.Config.CacheDir))
	if err != nil {
		return err
	}

	if *FHash != "" {
		err := os.Remove(path.Join(cache_dir, target_layer, *FHash))
		if err != nil {
			return err
		}
	}

	err = os.RemoveAll(path.Join(cache_dir, target_layer))
	if err != nil {
		return err
	}

	deactivated_layer := path.Join(internal.Config.ExtensionsDir, target_layer) + internal.ValidSysextExtension
	_, err = os.Stat(deactivated_layer)
	if err != nil {
		return nil
	}
	err = os.Remove(deactivated_layer)
	if err != nil {
		return err
	}

	return nil
}
