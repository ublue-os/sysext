package list

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/fileio"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List layers in cache and in activation",
	Long:  `List layers in the cache directory, their blobs and symlinks in their cache`,
	RunE:  listCmd,
}

var (
	fLayer     *string
	fQuiet     *bool
	fVerbose   *bool
	fActivated *bool
	fSeparator *string
	fLogOnly   *bool
)

func init() {
	fVerbose = ListCmd.Flags().BoolP("verbose", "v", false, "List all layer's hashes")
	fLayer = ListCmd.Flags().StringP("layer", "l", "", "List hashes inside the target layer")
	fQuiet = ListCmd.Flags().BoolP("quiet", "q", false, "Only check for layer or hash existence instead of listing")
	fActivated = ListCmd.Flags().Bool("activated", false, "List only activated layers")
	fSeparator = ListCmd.Flags().StringP("separator", "s", "\n", "Separator for listing things like arrays")
	fLogOnly = ListCmd.Flags().Bool("log", false, "Do not make a table, just log everything")
}

func Btoi(b bool) int {
	if b {
		return 0
	}
	return 1
}

func listCmd(cmd *cobra.Command, args []string) error {
	cache_dir, err := filepath.Abs(path.Clean(internal.Config.CacheDir))
	if err != nil {
		return err
	}

	if *fQuiet {
		if len(args) < 1 {
			return internal.NewPositionalError("LAYER")
		}
		layer := args[0]
		if len(args) > 1 {
			hash := args[1]
			os.Exit(Btoi(fileio.FileExist(path.Join(cache_dir, layer, hash))))
		}
		os.Exit(Btoi(fileio.FileExist(path.Join(cache_dir, layer))))
	}

	dirdata, err := os.ReadDir(cache_dir)
	if err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.SetTitle("Layers")
	t.Style().Options.SeparateRows = true
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: "Layers", Align: text.AlignCenter, VAlign: text.VAlignMiddle},
		{Name: "Binaries", Align: text.AlignCenter, VAlign: text.VAlignMiddle},
	})

	for _, dir := range dirdata {
		if !dir.IsDir() {
			continue
		}
		if *fLayer != "" && dir.Name() != *fLayer {
			continue
		}
		if _, err := os.Stat(path.Join(internal.Config.ExtensionsDir, dir.Name()+internal.ValidSysextExtension)); err != nil && *fActivated {
			continue
		}

		var blobs []string
		if *fVerbose {
			layerdir, err := os.ReadDir(path.Join(cache_dir, dir.Name()))
			if err != nil {
				return err
			}

			for _, blob := range layerdir {
				if blob.Name() == internal.CurrentBlobName {
					fstat, err := filepath.EvalSymlinks(path.Join(cache_dir, dir.Name(), blob.Name()))
					if err != nil {
						return err
					}
					blobs = append(blobs, fmt.Sprintf("%s -> %s", blob.Name(), path.Base(fstat)))
					continue
				}
				blobs = append(blobs, blob.Name())
			}
		}

		if len(blobs) == 0 {
			slog.Info(dir.Name(), slog.String("blobs", ""))
			t.AppendRow(table.Row{dir.Name()})
			continue
		}
		slog.Info(dir.Name(), slog.String("blobs", strings.Join(blobs, ":")))
		t.AppendRow(table.Row{dir.Name(), strings.Join(blobs, *fSeparator)})
	}

	if t.Length() == 0 {
		slog.Warn("No layers found")
		return nil
	}

	if !*fLogOnly {
		fmt.Printf("%s\n", t.Render())
	}
	return nil
}
