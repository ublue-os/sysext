package clean

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/percentmanager"
)

var CleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean every unused cache blob",
	Long:  `Clean every unused blob from cache except the current blob and its symlink`,
	RunE:  cleanCmd,
}

var (
	fExclude *[]string
	fDryRun  *bool
)

func init() {
	fExclude = CleanCmd.Flags().StringSliceP("exclude", "e", make([]string, 0), "Exclude directories from cleaning")
	fDryRun = CleanCmd.Flags().Bool("dry-run", false, "Do not actually clean anything, just print what would be deleted")
}

func getWhatNotToClean(clean []string) ([]string, error) {
	var do_not_clean []string
	for _, cleanthing := range clean {
		fstat, err := os.Lstat(cleanthing)
		if err != nil {
			return nil, err
		}
		if fstat.Mode().Type() == os.ModeSymlink && fstat.Name() == internal.CurrentBlobName {
			eval_link, err := filepath.EvalSymlinks(cleanthing)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			} else if errors.Is(err, os.ErrNotExist) {
				continue
			}
			do_not_clean = append(do_not_clean, eval_link)
			do_not_clean = append(do_not_clean, cleanthing)
			break
		}
	}
	return do_not_clean, nil
}

func cleanCmd(cmd *cobra.Command, args []string) error {
	cache_dir, err := filepath.Abs(internal.Config.CacheDir)
	if err != nil {
		return err
	}
	target_cache, err := os.ReadDir(cache_dir)
	if err != nil {
		return err
	}

	base_message := "Cleaning "
	pw := percent.NewProgressWriter()
	go pw.Render()

	var expectedSections int
	for _, entry := range target_cache {
		if !entry.IsDir() {
			continue
		}

		entry_dir_path := path.Join(cache_dir, entry.Name())
		entry_dir, err := os.ReadDir(entry_dir_path)
		if err != nil {
			return err
		}

		var clean []string

		for _, cache_blob := range entry_dir {
			if cache_blob.IsDir() {
				continue
			}

			clean = append(clean, path.Join(entry_dir_path, cache_blob.Name()))
		}

		do_not_clean, err := getWhatNotToClean(clean)
		if err != nil {
			return err
		}

		if len(entry_dir) < 1 {
			expectedSections++
			continue
		}

		for _, cache_blob := range entry_dir {
			if cache_blob.IsDir() {
				continue
			}
			expectedSections++
		}
		expectedSections = expectedSections - len(do_not_clean)
	}

	delete_tracker := percent.NewIncrementTracker(&progress.Tracker{Message: base_message, Total: int64(100), Units: progress.UnitsDefault}, expectedSections)
	go pw.Render()
	pw.AppendTracker(delete_tracker.Tracker)

	for _, entry := range target_cache {
		if !entry.IsDir() {
			continue
		}
		delete_tracker.Tracker.Message = base_message + entry.Name()
		entry_dir_path := path.Join(cache_dir, entry.Name())
		entry_dir, err := os.ReadDir(entry_dir_path)
		if err != nil {
			delete_tracker.Tracker.MarkAsErrored()
			return err
		}

		if len(entry_dir) < 1 {
			delete_tracker.IncrementSection()
			os.Remove(entry_dir_path)
			continue
		}

		var clean []string

		for _, cache_blob := range entry_dir {
			if cache_blob.IsDir() {
				continue
			}

			clean = append(clean, path.Join(entry_dir_path, cache_blob.Name()))
		}

		do_not_clean, err := getWhatNotToClean(clean)
		if err != nil {
			delete_tracker.Tracker.MarkAsErrored()
			return err
		}

		for _, provided_path := range *fExclude {
			managed_path, err := filepath.Abs(path.Clean(provided_path))
			if err != nil {
				return err
			}
			do_not_clean = append(do_not_clean, managed_path)
		}

		for _, cleanthing := range clean {
			if slices.Contains(do_not_clean, cleanthing) || *fDryRun {
				continue
			}
			delete_tracker.IncrementSection()
			if err := os.Remove(cleanthing); err != nil {
				delete_tracker.Tracker.MarkAsErrored()
				return err
			}
		}
	}
	delete_tracker.Tracker.MarkAsDone()

	return nil
}
