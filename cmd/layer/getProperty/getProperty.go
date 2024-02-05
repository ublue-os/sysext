package getProperty

import (
	"encoding/json"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"path"
	"strings"
)

var GetPropertyCmd = &cobra.Command{
	Use:   "get-property",
	Short: "Get properties from a selected layer or configuration file",
	Long: `Get properties from a selected layer or configuration file

Supported properties:
    NAME
    PACKAGES
    ARCH
    OS 
    BINARIES
    ISMOUNTED
`,
	// not supported yet: UNITS SIGNINGKEY
	RunE: getPropertyCmd,
}

var (
	fFromFile    *string
	fSeparator   *string
	validOptions []string = []string{"NAME", "PACKAGES", "ARCH", "OS", "BINARIES", "ISMOUNTED"}
)

func init() {
	fFromFile = GetPropertyCmd.Flags().StringP("from-file", "f", "", "Read data from a file instead of layer")
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
				t.AppendRow(table.Row{property_name, strings.Join(unmarshalled_config.Packages, *fSeparator)})
				continue
			}
		case "BINARIES":
			{
				if !layer_mounted {
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

				t.AppendRow(table.Row{"Binaries", strings.Join(dir_contents, *fSeparator)})
				continue
			}
		case "ISMOUNTED":
			{
				t.AppendRow(table.Row{"IsMounted", layer_mounted})
				continue
			}
		}

		value_get := string(internal.GetFieldFromStruct(unmarshalled_config, property_name).String())
		if value_get[0] == '<' {
			return internal.NewInvalidOptionError(property_name)
		}
		t.AppendRow(table.Row{property_name, value_get})
	}
	fmt.Printf("%s\n", t.Render())

	return nil
}
