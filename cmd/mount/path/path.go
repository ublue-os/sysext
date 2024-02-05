package path

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

var PathCmd = &cobra.Command{
	Use:   "path",
	Short: "Mount all /bin paths from each layer to target destination",
	Long:  `Mount all /bin paths from each layer to target destination`,
	RunE:  pathCmd,
}

var (
	fPathPath *string
)

func init() {
	fPathPath = PathCmd.Flags().StringP("path", "p", "/tmp/extensions.d/bin", "Path where all shared binaries will be mounted to")
}

func pathCmd(cmd *cobra.Command, args []string) error {
	extensions_mount, err := filepath.Abs(path.Clean(internal.Config.ExtensionsMount))
	if err != nil {
		return nil
	}
	path_path, err := filepath.Abs(path.Clean(*fPathPath))
	if err != nil {
		return err
	}

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

	return nil
}
