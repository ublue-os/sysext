package layer

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
)

var CleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean every unused cache blob",
	Long:  `Clean every unused blob from cache except the current blob and its symlink`,
	RunE:  cleanCmd,
}

var (
	FExclude *[]string
	FDryRun  *bool
)

func init() {
	FExclude = CleanCmd.Flags().StringSliceP("exclude", "e", make([]string, 0), "Exclude directories from cleaning")
	FDryRun = CleanCmd.Flags().Bool("dry-run", false, "Do not actually clean anything, just print what would be deleted")
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

	for _, entry := range target_cache {
		if !entry.IsDir() {
			continue
		}
		entry_dir_path := path.Join(cache_dir, entry.Name())
		entry_dir, err := os.ReadDir(entry_dir_path)
		if err != nil {
			return err
		}

		if len(entry_dir) < 1 {
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

		var do_not_clean []string
		for _, cleanthing := range clean {
			fstat, err := os.Lstat(cleanthing)
			if err != nil {
				return err
			}
			if fstat.Mode().Type() == os.ModeSymlink && fstat.Name() == internal.CurrentBlobName {
				eval_link, err := filepath.EvalSymlinks(cleanthing)
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				} else if errors.Is(err, os.ErrNotExist) {
					continue
				}
				do_not_clean = append(do_not_clean, eval_link)
				do_not_clean = append(do_not_clean, cleanthing)
				break
			}
		}

		for _, provided_path := range *FExclude {
			managed_path, err := filepath.Abs(path.Clean(provided_path))
			if err != nil {
				return err
			}
			do_not_clean = append(do_not_clean, managed_path)
		}

		for _, cleanthing := range clean {
			if slices.Contains(do_not_clean, cleanthing) {
				continue
			}
			if *FDryRun {
				fmt.Printf("d: %s\n", cleanthing)
				continue
			}
			if err := os.Remove(cleanthing); err != nil {
				return err
			}
		}
	}

	return nil
}
