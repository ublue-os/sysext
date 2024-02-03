package layer

import (
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
)

var LayerCmd = &cobra.Command{
	Use:   "layer",
	Short: "Execute layer-based operations",
}

func init() {
	LayerCmd.PersistentFlags().StringVar(&internal.Config.CacheDir, "cache-root", "/var/cache/extensions", "root directory for the layer cache")
	LayerCmd.PersistentFlags().StringVar(&internal.Config.ExtensionsDir, "extensions-root", "/var/lib/extensions", "root directory for the systemd-sysext layers")

	LayerCmd.AddCommand(ActivateCmd)
	LayerCmd.AddCommand(AddCmd)
	LayerCmd.AddCommand(CleanCmd)
	LayerCmd.AddCommand(DeactivateCmd)
	LayerCmd.AddCommand(NewGetPropertyCmd())
	LayerCmd.AddCommand(NewInitCmd())
	LayerCmd.AddCommand(NewListCmd())
	LayerCmd.AddCommand(NewRefreshCmd())
	LayerCmd.AddCommand(RemoveCmd)
}
