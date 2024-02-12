package add

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/filecomp"
	"github.com/ublue-os/sysext/pkg/fileio"
	"github.com/ublue-os/sysext/pkg/logging"
	"github.com/ublue-os/sysext/pkg/percentmanager"
)

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a built layer onto the cache and activate it",
	Long:  `Copy TARGET over to cache-dir as a blob with the TARGET's sha256 as the filename`,
	RunE:  addCmd,
}

var (
	fNoSymlink  *bool
	fNoChecksum *bool
	fOverride   *bool
	fLayerName  *string
)

func init() {
	fNoSymlink = AddCmd.Flags().Bool("no-symlink", false, "Do not activate layer once added to cache")
	fNoChecksum = AddCmd.Flags().Bool("no-checksum", false, "Do not check if layer was properly added to cache")
	fOverride = AddCmd.Flags().Bool("override", false, "Override blob if they are already written to cache")
	fLayerName = AddCmd.Flags().String("layer-name", "", "Name of the layer that will be added onto")
}

func addCmd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return internal.NewPositionalError("TARGET")
	}
	target_layer := &internal.TargetLayerInfo{}
	target_layer.Path = path.Clean(args[0])

	var err error
	target_layer.FileInfo, err = os.Stat(target_layer.Path)
	if err != nil {
		return err
	}
	pw := percent.NewProgressWriter()
	if !*internal.Config.NoProgress {
		go pw.Render()
		slog.SetDefault(logging.NewMuteLogger())
	}
	var expectedSections int = 4

	if !*fNoSymlink {
		expectedSections++
	}
	if !*fNoChecksum {
		expectedSections += 2
	}

	add_tracker := percent.NewIncrementTracker(&progress.Tracker{Message: "Adding layer " + path.Base(target_layer.Path) + " to cache", Total: int64(100), Units: progress.UnitsDefault}, expectedSections)
	pw.AppendTracker(add_tracker.Tracker)

	add_tracker.IncrementSection()
	if err := os.MkdirAll(internal.Config.CacheDir, 0755); err != nil {
		return err
	}

	add_tracker.IncrementSection()
	layer_sha := sha256.New()
	layer_sha.Write(target_layer.Data)
	target_layer.UUID = layer_sha.Sum(nil)
	if err != nil {
		return err
	}

	if *fLayerName != "" {
		slog.Warn("The path inside /usr/lib/sysext/extensions-* must be the same as the layer's name in order for it to function, please check if this is actually the case")
		target_layer.LayerName = *fLayerName
	} else {
		target_layer.LayerName = strings.Split(path.Base(target_layer.Path), ".")[0]
	}
	var blob_filepath string
	blob_filepath, err = filepath.Abs(path.Join(internal.Config.CacheDir, target_layer.LayerName, hex.EncodeToString(target_layer.UUID)))
	if err != nil {
		add_tracker.Tracker.MarkAsErrored()
		return err
	}

	add_tracker.IncrementSection()
	if err := os.MkdirAll(path.Dir(blob_filepath), 0755); err != nil {
		add_tracker.Tracker.MarkAsErrored()
		return err
	}

	if fileio.FileExist(blob_filepath) && !*fOverride {
		slog.Warn("Blob is already in cache")
		add_tracker.Tracker.MarkAsErrored()
		os.Exit(1)
	}

	add_tracker.IncrementSection()
	if err := fileio.FileCopy(target_layer.Path, blob_filepath); err != nil {
		return err
	}

	if !*fNoChecksum {
		add_tracker.Tracker.Message = "Checking blob"

		add_tracker.IncrementSection()
		var written_file *os.File
		written_file, err = os.Open(blob_filepath)
		if err != nil {
			add_tracker.Tracker.MarkAsErrored()
			return err
		}
		defer written_file.Close()

		var tlayer_fileobj *os.File
		tlayer_fileobj, err = os.Open(target_layer.Path)
		if err != nil {
			return err
		}
		defer tlayer_fileobj.Close()

		add_tracker.IncrementSection()
		_, err = filecomp.CheckFilesAreEqual(md5.New(), tlayer_fileobj, written_file)
		if err != nil {
			slog.Warn("Copied blobs did not match")
			return err
		}
	}

	if *fNoSymlink {
		slog.Info("Successfully added blob to cache", slog.String("blob_path", blob_filepath))
		add_tracker.Tracker.MarkAsDone()
		return nil
	}

	var current_blob_path string
	current_blob_path, err = filepath.Abs(path.Join(path.Dir(blob_filepath), internal.CurrentBlobName))
	if err != nil {
		return err
	}
	add_tracker.Tracker.Message = "Refreshing symlink"
	slog.Debug("Refreshing symlink", slog.String("path", current_blob_path))
	add_tracker.IncrementSection()
	if _, err := os.Lstat(current_blob_path); err == nil {
		err = os.Remove(current_blob_path)
		if err != nil {
			add_tracker.Tracker.MarkAsErrored()
			return err
		}
	} else if errors.Is(err, os.ErrNotExist) {

	} else {
		add_tracker.Tracker.MarkAsErrored()
		return err
	}

	err = os.Symlink(blob_filepath, current_blob_path)
	if err != nil {
		add_tracker.Tracker.MarkAsErrored()
		return err
	}
	add_tracker.Tracker.MarkAsDone()

	slog.Info("Successfully added blob to cache", slog.String("blob_path", blob_filepath))
	return nil
}
