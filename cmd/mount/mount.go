package mount

import (
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/cmd/mount/extensions"
	"github.com/ublue-os/sysext/cmd/mount/path"
	"github.com/ublue-os/sysext/cmd/mount/store"
)

var MountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount, refresh, and manage system extensions",
	Long:  `Manage and mount the nix store, your layers, and path variables.`,
}

func init() {
	MountCmd.AddCommand(extensions.ExtensionsCmd)
	MountCmd.AddCommand(store.StoreCmd)
	MountCmd.AddCommand(path.PathCmd)
}
