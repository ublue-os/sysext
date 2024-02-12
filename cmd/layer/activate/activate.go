package activate

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	fFromFile *string
)

func init() {
	fFromFile = ActivateCmd.Flags().StringP("file", "f", "", "Activate directly from file instead of cache")
}

func activateCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return internal.NewPositionalError("TARGET")
	}

	target_layer := args[0]

	extensions_dir, err := filepath.Abs(path.Clean(internal.Config.ExtensionsDir))
	if err != nil {
		return err
	}

	if *fFromFile != "" {
		if !strings.HasSuffix(target_layer, internal.ValidSysextExtension) {
			target_layer, err = filepath.Abs(path.Clean(target_layer + internal.ValidSysextExtension))
			if err != nil {
				return err
			}
		}

		target_path := path.Join(extensions_dir, target_layer)
		slog.Debug("acitavate",
			slog.String("fromfile", *fFromFile),
			slog.String("layer", target_layer),
			slog.String("path", target_path),
		)
		if err := os.Symlink(target_layer, path.Join(extensions_dir, target_layer)); err != nil {
			return err
		}
		slog.Info(fmt.Sprintf("Successfully activated layer %s\n", path.Base(target_layer)))
		return nil
	}

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

	target_path := path.Join(extensions_dir, path.Base(path.Dir(current_blob_path))+internal.ValidSysextExtension)
	slog.Debug("acitavate",
		slog.String("fromfile", *fFromFile),
		slog.String("layer", target_layer),
		slog.String("blob", current_blob_path),
	)
	if err := os.Symlink(current_blob_path, target_path); err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Successfully activated layer %s\n", path.Base(target_layer)))
	return nil
}
