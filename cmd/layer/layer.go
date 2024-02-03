package layer

import (
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/cmd/layer/activate"
	"github.com/ublue-os/sysext/cmd/layer/add"
	//"github.com/ublue-os/sysext/cmd/layer/build"
	"github.com/ublue-os/sysext/cmd/layer/clean"
	"github.com/ublue-os/sysext/cmd/layer/deactivate"
	//"github.com/ublue-os/sysext/cmd/layer/getProperty"
	"github.com/ublue-os/sysext/cmd/layer/initcmd"
	//"github.com/ublue-os/sysext/cmd/layer/list"
	"github.com/ublue-os/sysext/cmd/layer/remove"
	"github.com/ublue-os/sysext/internal"
)

var LayerCmd = &cobra.Command{
	Use:   "layer",
	Short: "Execute layer-based operations",
}

func init() {
	LayerCmd.PersistentFlags().StringVar(&internal.Config.CacheDir, "cache-root", "/var/cache/extensions", "root directory for the layer cache")
	LayerCmd.PersistentFlags().StringVar(&internal.Config.ExtensionsDir, "extensions-root", "/var/lib/extensions", "root directory for the systemd-sysext layers")
	LayerCmd.AddCommand(activate.ActivateCmd)
	LayerCmd.AddCommand(add.AddCmd)
	LayerCmd.AddCommand(clean.CleanCmd)
	LayerCmd.AddCommand(deactivate.DeactivateCmd)
	//LayerCmd.AddCommand(GetPropertyCmd)
	LayerCmd.AddCommand(initcmd.InitCmd)
	//LayerCmd.AddCommand(ListCmd)
	//LayerCmd.AddCommand(BuildCmd)
	LayerCmd.AddCommand(remove.RemoveCmd)
}
