package getProperty

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/logging"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var validOptions []string = []string{"NAME", "PACKAGES", "ARCH", "OS", "BINARIES", "ISMOUNTED"}

var GetPropertyCmd = &cobra.Command{
	Use:   "get-property",
	Short: "Get properties from a selected layer or configuration file",
	Long: fmt.Sprintf(`Get properties from a selected layer or configuration file

Supported properties:
    %s 
`, strings.Join(validOptions, "\n\t")),
	RunE: getPropertyCmd,
}

var (
	fFromFile  *string
	fSeparator *string
	fLogOnly   *bool
)

func init() {
	fFromFile = GetPropertyCmd.Flags().StringP("from-file", "f", "", "Read data from a file instead of layer")
	fLogOnly = GetPropertyCmd.Flags().Bool("log", false, "Do not make a table, just log everything")
	fSeparator = GetPropertyCmd.Flags().StringP("separator", "s", "\n", "Separator for listing things like arrays")
}

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

func getPropertyCmd(cmd *cobra.Command, args []string) error {
	var (
		target_layer        string = ""
		config_file_path    string
		raw_configuration   []byte
		unmarshalled_config = &internal.LayerConfiguration{}
	)

	if !*fLogOnly {
		slog.SetDefault(logging.NewMuteLogger())
	}

	if *fFromFile == "" {
		if len(args) < 1 {
			return internal.NewPositionalError("LAYER")
		}
		target_layer = args[0]
		args = remove(args, 0)
		config_file_path = path.Join(internal.Config.ExtensionsMount, target_layer, internal.MetadataFileName)
	} else {
		config_file_path = path.Clean(*fFromFile)
	}

	raw_configuration, err := os.ReadFile(config_file_path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw_configuration, unmarshalled_config); err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.SetTitle(unmarshalled_config.Name)
	t.Style().Options.SeparateRows = true
	t.SetColumnConfigs([]table.ColumnConfig{
		{Name: unmarshalled_config.Name, Align: text.AlignCenter, VAlign: text.VAlignMiddle},
	})

	if len(args) == 0 {
		args = validOptions
	}

	for _, property := range args {
		upper_prop_name := strings.ToUpper(property)
		property_name := cases.Title(language.English).String(property)

		var layer_mounted bool = true
		if _, err := os.Stat(path.Join(internal.Config.ExtensionsMount, unmarshalled_config.Name)); err != nil {
			layer_mounted = false
		}

		switch upper_prop_name {
		case "PACKAGES":
			{
				packages := strings.Join(unmarshalled_config.Packages, *fSeparator)
				slog.Info("packages", slog.String("value", packages))
				t.AppendRow(table.Row{property_name, packages})
				continue
			}
		case "BINARIES":
			{
				if !layer_mounted {
					slog.Info("binaries", slog.String("value", "Layer not mounted"))
					t.AppendRow(table.Row{"Binaries", "Layer not mounted"})
					continue
				}
				list_dir, err := os.ReadDir(path.Join(internal.Config.ExtensionsMount, unmarshalled_config.Name, "bin"))
				if err != nil {
					return err
				}
				var dir_contents []string
				for _, file := range list_dir {
					dir_contents = append(dir_contents, file.Name())
				}

				binaries := strings.Join(dir_contents, *fSeparator)
				slog.Info("binaries", slog.String("value", binaries))
				t.AppendRow(table.Row{"Binaries", binaries})
				continue
			}
		case "ISMOUNTED":
			{
				slog.Info("mounted", slog.Bool("value", layer_mounted))
				t.AppendRow(table.Row{"IsMounted", layer_mounted})
				continue
			}
		}

		value_get := string(internal.GetFieldFromStruct(unmarshalled_config, property_name).String())
		if value_get[0] == '<' {
			return internal.NewInvalidOptionError(property_name)
		}
		slog.Info(property_name, slog.String("value", value_get))
		t.AppendRow(table.Row{property_name, value_get})
	}

	if !*fLogOnly {
		fmt.Printf("%s\n", t.Render())
	}

	return nil
}
