package remove

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/fileio"
	"github.com/ublue-os/sysext/pkg/logging"
	"github.com/ublue-os/sysext/pkg/percentmanager"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a layer from your managed layers",
	Long:  `Remove either an entire layer or a specific hash in cache for that layer`,
	RunE:  removeCmd,
}

var (
	fHash   *string
	fDryRun *bool
)

func init() {
	fHash = RemoveCmd.Flags().StringP("hash", "h", "", "Remove specific hash from storage")
	fDryRun = RemoveCmd.Flags().Bool("dry-run", false, "Do not remove anything")
}

func removeCmd(cmd *cobra.Command, args []string) error {
	pw := percent.NewProgressWriter()
	if !*internal.Config.NoProgress {
		go pw.Render()
		slog.SetDefault(logging.NewMuteLogger())
	}

	if len(args) < 1 {
		return internal.NewPositionalError("TARGET")
	}

	target_layer := args[0]

	cache_dir, err := filepath.Abs(path.Clean(internal.Config.CacheDir))
	if err != nil {
		return err
	}

	var message string = "Deleting layer " + target_layer
	var expectedSections = 4

	if *fHash != "" {
		message = "Deleting hash " + target_layer
		expectedSections = expectedSections + 1
	}
	delete_tracker := percent.NewIncrementTracker(&progress.Tracker{Message: message, Total: int64(100), Units: progress.UnitsDefault}, expectedSections)
	pw.AppendTracker(delete_tracker.Tracker)

	if *fHash != "" {
		delete_tracker.IncrementSection()
		err := os.Remove(path.Join(cache_dir, target_layer, *fHash))
		if err != nil {
			return err
		}
		slog.Info("Successfuly deleted " + *fHash)
		return nil
	}

	delete_tracker.IncrementSection()
	slog.Debug("Deleting layer", slog.String("target", target_layer))
	err = os.RemoveAll(path.Join(cache_dir, target_layer))
	if err != nil {
		return err
	}

	delete_tracker.IncrementSection()
	deactivated_layer := path.Join(internal.Config.ExtensionsDir, target_layer) + internal.ValidSysextExtension
	if !fileio.FileExist(deactivated_layer) {
		return nil
	}

	delete_tracker.IncrementSection()
	slog.Debug("Deactivating layer", slog.String("target", deactivated_layer))
	err = os.Remove(deactivated_layer)
	if err != nil {
		return err
	}

	slog.Info("Successfuly deleted "+deactivated_layer, slog.String("target", deactivated_layer))

	delete_tracker.IncrementSection()
	delete_tracker.Tracker.MarkAsDone()
	return nil
}
