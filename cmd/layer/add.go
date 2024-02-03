package layer

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/fileio"
)

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a built layer onto the cache and activate it",
	Long:  `Copy TARGET over to cache-dir as a blob with the TARGET's sha256 as the filename`,
	RunE:  addExec,
}

var (
	FNoSymlink  *bool
	FNoChecksum *bool
	FOverride   *bool
	FLayerName  *string
)

func init() {
	FNoSymlink = AddCmd.Flags().Bool("no-symlink", false, "Do not activate layer once added to cache")
	FNoChecksum = AddCmd.Flags().Bool("no-checksum", false, "Do not check if layer was properly added to cache")
	FOverride = AddCmd.Flags().Bool("override", false, "Override blob if they are already written to cache")
	FLayerName = AddCmd.Flags().String("layer-name", "", "Name of the layer that will be added onto")
}

func addExec(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		fmt.Println("Required argument TARGET")
		os.Exit(1)
	}
	target_layer := &internal.TargetLayerInfo{}
	target_layer.Path = path.Clean(args[0])

	var err error
	target_layer.FileInfo, err = os.Stat(target_layer.Path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(internal.Config.CacheDir, 0755); err != nil {
		return err
	}

	layer_sha := sha256.New()
	layer_sha.Write(target_layer.Data)
	target_layer.UUID = layer_sha.Sum(nil)
	if err != nil {
		return err
	}

	if *FLayerName != "" {
		target_layer.LayerName = *FLayerName
	} else {
		target_layer.LayerName = strings.Split(path.Base(target_layer.Path), ".")[0]
	}
	var blob_filepath string
	blob_filepath, err = filepath.Abs(path.Join(internal.Config.CacheDir, target_layer.LayerName, hex.EncodeToString(target_layer.UUID)))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path.Dir(blob_filepath), 0755); err != nil {
		return err
	}

	if fileio.FileExist(blob_filepath) && !*FOverride {
		fmt.Fprintln(os.Stderr, "Blob is already in cache")
		os.Exit(1)
	}

	if err := fileio.FileCopy(target_layer.Path, blob_filepath); err != nil {
		return err
	}

	if !*FNoChecksum {
		var written_file *os.File
		written_file, err = os.Open(blob_filepath)
		if err != nil {
			return err
		}
		defer written_file.Close()

		var tlayer_fileobj *os.File
		tlayer_fileobj, err = os.Open(target_layer.Path)
		if err != nil {
			return err
		}
		defer tlayer_fileobj.Close()

		_, err = internal.CheckFilesAreEqual(md5.New(), tlayer_fileobj, written_file)
		if err != nil {
			return err
		}
	}

	var current_blob_path string
	current_blob_path, err = filepath.Abs(path.Join(path.Dir(blob_filepath), internal.CurrentBlobName))
	if err != nil {
		return err
	}

	if _, err := os.Lstat(current_blob_path); err == nil {
		err = os.Remove(current_blob_path)
		if err != nil {
			return err
		}
	} else if errors.Is(err, os.ErrNotExist) {

	} else {
		return err
	}

	if *FNoSymlink == false {
		err = os.Symlink(blob_filepath, current_blob_path)
		if err != nil {
			return err
		}
	}

	return nil
}
