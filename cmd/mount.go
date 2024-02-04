package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/chattr"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

var MountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount, refresh, and manage system extensions",
	Long:  `Manage and mount the nix store, your layers, and path variables.`,
	RunE:  mountCmd,
}

var (
	fRefreshStore       *bool
	fRefreshPath        *bool
	fRefreshExtensions  *bool
	fForce              *bool
	fPathPath           *string
	fStoreBindmountPath *string
)

func init() {
	fRefreshStore = MountCmd.Flags().BoolP("refresh-store", "s", true, "Refresh the nix store")
	fRefreshPath = MountCmd.Flags().BoolP("refresh-path", "p", true, "Refresh the shared layer path")
	fRefreshExtensions = MountCmd.Flags().BoolP("refresh-extensions", "e", true, "Refresh the activated systemd-sysext(s)")
	fForce = MountCmd.Flags().BoolP("force", "f", false, "Pass the --force flag to systemd-sysext")
	fPathPath = MountCmd.Flags().String("path-path", "/tmp/extensions.d/bin", "Path where all shared binaries will be mounted to")
	fStoreBindmountPath = MountCmd.Flags().String("bindmount-path", "/tmp/nix-store-bindmount", "Path where an already existing nix store will be bind-mounted to")
}

func mountCmd(cmd *cobra.Command, args []string) error {
	var force_flag string = ""
	if *fForce {
		force_flag = "--force"
	}

	if *fRefreshExtensions {
		out, err := exec.Command("systemd-sysext", "refresh", force_flag).Output()
		if err != nil {
			return err
		}
		fmt.Printf("%s", out)
	}

	if *fRefreshStore {
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

			if err := syscall.Mount("/nix/store", bindmount_path, "bind", uintptr(syscall.MS_BIND|syscall.MS_RDONLY), ""); err != nil {
				return err
			}

			if err := syscall.Mount(bindmount_path, "/nix/store", "bind", uintptr(syscall.MS_BIND|syscall.MS_RDONLY), ""); err != nil {
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
	}

	if *fRefreshPath {
		extensions_mount, err := filepath.Abs(path.Clean(internal.Config.ExtensionsMount))
		if err != nil {
			return nil
		}
		path_path, err := filepath.Abs(path.Clean(*fPathPath))

		layers, err := os.ReadDir(extensions_mount)
		if err != nil {
			return err
		}

		var valid_layers []string
		for _, layer := range layers {
			if _, err := os.Stat(path.Join(extensions_mount, layer.Name(), "bin")); err != nil {
				continue
			}
			valid_layers = append(valid_layers, layer.Name())
		}

		if len(layers) == 0 {
			fmt.Fprintln(os.Stderr, "Error: No layers are mounted")
			os.Exit(1)
		} else if len(layers) == 1 {
			mount_path := path.Join(extensions_mount, layers[0].Name(), "bin")

			if _, err := os.Stat(mount_path); err == nil {
				err := syscall.Unmount(mount_path, int(syscall.MNT_FORCE))
				if err != nil {
					return err
				}
			}

			if err := syscall.Mount(mount_path, path_path, "bind", uintptr(syscall.MS_BIND|syscall.MS_RDONLY), ""); err != nil {
				return err
			}
		} else {
			if _, err := os.Stat(path_path); err == nil {
				err := syscall.Unmount(path_path, int(syscall.MNT_FORCE))
				if err != nil {
					return err
				}
			}

			err = syscall.Mount("none", path_path, "overlayfs", uintptr(syscall.MS_RDONLY|syscall.MS_NODEV|syscall.MS_NOATIME), "lowerdir="+strings.Join(valid_layers, ":"))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
