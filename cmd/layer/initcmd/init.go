package initcmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"

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
)

func init() {
	fOutPath = InitCmd.Flags().StringP("output-path", "o", "config.json", "Output path for new configuration")
	fTemplate = InitCmd.Flags().StringP("template", "t", "", "URL for template configuration")
	fOverride = InitCmd.Flags().Bool("override", false, "Override configuration if it already exists in output-path")
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
		slog.Debug("Using default template")
		var err error
		json_default, err := json.MarshalIndent(defaultConfiguration, "", structures.INDENTATION)
		if err != nil {
			return err
		}
		json_config = &json_default
	} else {
		slog.Debug("Using remote template", slog.String("remote_url", *fTemplate))
		var fetchedConfig = &internal.LayerConfiguration{}
		json_conf, err := structures.FetchJsonConfig(*fTemplate, fetchedConfig)
		if err != nil {
			return err
		}
		json_config = &json_conf
	}

	if fileio.FileExist(*fOutPath) && !*fOverride {
		slog.Warn("Failed writing, file already exists")
		return nil
	}

	os.Remove(*fOutPath)

	bytes_written, err := fileio.FileAppend(*fOutPath, *json_config)
	if err != nil {
		slog.Warn(fmt.Sprintf("Failed writing configuration file, bytes written: %d\n", bytes_written))
		return err
	}

	slog.Debug("configuration", slog.String("value", string(*json_config)))
	slog.Info("Successfully written configuration file "+path.Base(*fOutPath), slog.String("out_path", *fOutPath))
	return nil
}
