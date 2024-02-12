package store

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/chattr"
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
	bindmount_path, err := filepath.Abs(path.Clean(*fStoreBindmountPath))
	if err != nil {
		return err
	}

	if *internal.Config.UnmountFlag {
		slog.Debug("Unmounting store", slog.String("target", "/nix/store"))
		if err := syscall.Unmount("/nix/store", 0); err != nil {
			return err
		}

		slog.Debug("Unmounting bindmount", slog.String("target", bindmount_path))
		if err := syscall.Unmount(bindmount_path, 0); err != nil {
			return err
		}
		slog.Info("Successfully unmounted store and bindmount", slog.String("store_path", "/nix/store"), slog.String("bindmount_path", bindmount_path))
		return nil
	}

	if _, err := os.Stat("/nix"); err != nil {
		root_dir, err := os.Open("/")
		if err != nil {
			return err
		}
		defer root_dir.Close()

		slog.Debug("Creating nix store", slog.String("target", "/nix/store"))
		chattr.SetAttr(root_dir, chattr.FS_IMMUTABLE_FL)
		if err := os.Mkdir("/nix/store", 0755); err != nil {
			return err
		}
		chattr.UnsetAttr(root_dir, chattr.FS_IMMUTABLE_FL)
	}

	store_contents, err := os.ReadDir("/nix/store")
	if err != nil {
		return err
	}

	if _, err := os.Stat(bindmount_path); err != nil && len(store_contents) > 0 {
		if err := os.MkdirAll(bindmount_path, 0755); err != nil {
			return err
		}

		syscall.Unmount("/nix/store", 0)
		syscall.Unmount(bindmount_path, 0)

		slog.Debug("Mounting store to itself", slog.String("source", "/nix/store"), slog.String("target", bindmount_path))
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
		slog.Info("Mounting /usr/store to /nix/store", slog.String("source", "/usr/store"), slog.String("target", "/nix/store"))
		if err := syscall.Mount("/usr/store", "/nix/store", "bind", uintptr(syscall.MS_BIND|syscall.MS_RDONLY), ""); err != nil {
			return err
		}
	}
	return nil
}
