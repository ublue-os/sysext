package store

import (
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/pkg/chattr"
	"os"
	"path"
	"path/filepath"
	"syscall"
)

var StoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Mount /usr/store to /nix/store",
	Long:  `Mount /usr/store to /nix/store so that your layered binaries may work`,
	RunE:  storeCmd,
}

var (
	fRefreshStore       *bool
	fStoreBindmountPath *string
)

func init() {
	fRefreshStore = StoreCmd.Flags().BoolP("refresh", "r", true, "Refresh the nix store")
	fStoreBindmountPath = StoreCmd.Flags().String("bindmount-path", "/tmp/nix-store-bindmount", "Path where an already existing nix store will be bind-mounted to")
}

func storeCmd(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat("/nix"); err != nil {
		root_dir, err := os.Open("/")
		if err != nil {
			return err
		}
		defer root_dir.Close()

		chattr.SetAttr(root_dir, chattr.FS_IMMUTABLE_FL)
		if err := os.Mkdir("/nix/store", 0755); err != nil {
			return err
		}
		chattr.UnsetAttr(root_dir, chattr.FS_IMMUTABLE_FL)
	}

	bindmount_path, err := filepath.Abs(path.Clean(*fStoreBindmountPath))
	if err != nil {
		return err
	}

	store_contents, err := os.ReadDir("/nix/store")
	if err != nil {
		return err
	}

	if _, err := os.Stat(bindmount_path); err != nil && len(store_contents) > 0 {
		if err := os.MkdirAll(bindmount_path, 0755); err != nil {
			return err
		}

		if err := syscall.Mount("/nix/store", bindmount_path, "bind", uintptr(syscall.MS_BIND), ""); err != nil {
			return err
		}

		if err := syscall.Mount(bindmount_path, "/nix/store", "bind", uintptr(syscall.MS_BIND), ""); err != nil {
			return err
		}
	}

	if _, err := os.Stat("/usr/store"); err != nil {
		return err
	}

	if _, err := os.Stat("/nix"); err == nil {
		if err := syscall.Mount("/usr/store", "/nix/store", "bind", uintptr(syscall.MS_BIND|syscall.MS_RDONLY), ""); err != nil {
			return err
		}
	}
	return nil
}
