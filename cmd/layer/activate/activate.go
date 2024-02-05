package activate

import (
	"log"
	"os"
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

var fQuiet *bool

func init() {
	ActivateCmd.Flags().BoolP("quiet", "q", false, "Do not print anything on success")
}

func activateCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return internal.NewPositionalError("TARGET")
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

	if !*fQuiet {
		log.Println("Successfully activated layer")
	}
	return nil
}
