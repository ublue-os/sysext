package remove

import (
	"os"
	"path"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/fileio"
	"github.com/ublue-os/sysext/pkg/percentmanager"
)

var RemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a layer from your managed layers",
	Long:  `Remove either an entire layer or a specific hash in cache for that layer`,
	RunE:  removeCmd,
}

var (
	fHash *string
)

func init() {
	fHash = RemoveCmd.Flags().String("hash", "", "Remove specific hash from storage")
}

func removeCmd(cmd *cobra.Command, args []string) error {
	pw := percent.NewProgressWriter()
	go pw.Render()

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
	}

	delete_tracker.IncrementSection()
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
	err = os.Remove(deactivated_layer)
	if err != nil {
		return err
	}

	delete_tracker.IncrementSection()
	delete_tracker.Tracker.MarkAsDone()
	return nil
}
