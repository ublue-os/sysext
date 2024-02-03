package initcmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/fileio"
	"github.com/ublue-os/sysext/pkg/structures"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an example configuration for building a sample layer",
	Long:  `Initialize a configuration file for later building a layer`,
	RunE:  initCmd,
}

var (
	fOutPath  *string
	fTemplate *string
	fOverride *bool
	fYaml     *bool
)

func init() {
	fOutPath = InitCmd.Flags().StringP("output-path", "o", "config.json", "Output path for new configuration")
	fTemplate = InitCmd.Flags().StringP("template", "t", "", "URL for template configuration")
	fOverride = InitCmd.Flags().Bool("override", false, "Override configuration if it already exists in output-path")
	fYaml = InitCmd.Flags().Bool("yaml", false, "Write configuration as Yaml instead of Json")
}

var defaultConfiguration = &internal.LayerConfiguration{
	Name:     "example",
	Arch:     "x86-64",
	Os:       "_any",
	Packages: []string{"hello", "rsync", "rclone"},
}

func initCmd(cmd *cobra.Command, args []string) error {
	var json_config *[]byte = nil
	if *fTemplate == "" {
		var err error
		json_default, err := json.MarshalIndent(defaultConfiguration, "", structures.INDENTATION)
		if err != nil {
			return err
		}
		json_config = &json_default
	} else {
		var fetchedConfig = &internal.LayerConfiguration{}
		json_conf, err := structures.FetchJsonConfig(*fTemplate, fetchedConfig)
		if err != nil {
			return err
		}
		json_config = &json_conf
	}

	if fileio.FileExist(*fOutPath) && !*fOverride {
		fmt.Fprintf(os.Stderr, "Failed writing, file already exists\n")
		return nil
	}

	_ = os.Remove(*fOutPath)

	if *fYaml {
		var configFormat = &internal.LayerConfiguration{}
		yaml_config, err := structures.JsonToYaml(*json_config, configFormat)
		if err != nil {
			return nil
		}
		json_config = &yaml_config
	}

	bytes_written, err := fileio.FileAppend(*fOutPath, *json_config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed writing configuration file, bytes written: %d\n", bytes_written)
		return err
	}

	return nil
}
